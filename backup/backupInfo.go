package backup

import (
	"encoding/base64"
	"fmt"
	"github.com/Hyperledger-TWGC/tjfoc-gm/sm3"
	"github.com/golang/protobuf/proto"
	"github.com/ipfs/go-bitswap/message"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	pb "github.com/ipfs/go-ipfs-backup/backup/pb"
)

const (
	Prefix = "backup/"
)

type idHash string
type uid []byte

func GetIdHash(c string, u string) (string, error) {
	decode, err := cid.Decode(c)
	if err != nil {
		return "", err
	}
	if decode.Version() > 2 {
		return "", fmt.Errorf("错误的版本号")
	}
	h := sm3.New()
	h.Write([]byte(c))
	decodeString, err := base64.StdEncoding.DecodeString(u)
	if err != nil {
		return "", nil
	}
	if len(decodeString) != 16 {
		return "", fmt.Errorf("错误的uid")
	}
	h.Write(decodeString)
	sum := h.Sum(nil)
	res := base64.StdEncoding.EncodeToString(sum)
	return res, nil
}

type StringSet map[string]struct{}

func getStringSet(in []string) StringSet {
	var res StringSet = map[string]struct{}{}
	for _, s := range in {
		res[s] = struct{}{}
	}
	return res
}

func (ss StringSet) GetArray() []string {
	var res []string
	for s, _ := range ss {
		res = append(res, s)
	}
	return res
}

func (s1 StringSet) append(ss StringSet) {
	for s, s2 := range ss {
		ss[s] = s2
	}
}

func (s1 StringSet) Add(peer string) {
	s1[peer] = struct{}{}
}

type Info struct {
	IdHashPin      map[string]bool
	IdHashUnpin    map[string]string
	TargetPeerList StringSet
}

func (b *Info) GetPb() pb.BackupInfo {
	return pb.BackupInfo{
		IdHashPin:      b.IdHashPin,
		IdHashUnpin:    b.IdHashUnpin,
		TargetPeerList: b.TargetPeerList.GetArray(),
	}
}

func (b *Info) LoadPb(pbi *pb.BackupInfo) {
	b.IdHashPin = pbi.IdHashPin
	b.IdHashUnpin = pbi.IdHashUnpin
	b.TargetPeerList = getStringSet(pbi.TargetPeerList)
}

func Marshal(b Info) ([]byte, error) {
	getPb := b.GetPb()
	return proto.Marshal(&getPb)
}

func Unmarshal(m []byte, b *Info) error {
	pbi := &pb.BackupInfo{}
	err := proto.Unmarshal(m, pbi)
	if err != nil {
		return err
	}
	b.LoadPb(pbi)
	return nil
}

func Put(ds datastore.Datastore, cid string, b Info) error {
	exist, err := Get(ds, cid)
	if err == datastore.ErrNotFound {
	} else if err != nil {
		return err
	} else {
		b = merge(exist, b, cid)
	}
	return put(ds, cid, b)
}

func put(ds datastore.Datastore, cid string, b Info) error {
	key := datastore.NewKey(Prefix + cid)
	marshal, err := Marshal(b)
	if err != nil {
		return err
	}
	err = ds.Put(key, marshal)
	return err
}

func Puts(ds datastore.Datastore, loadList []message.Load) error {
	for _, load := range loadList {
		c := load.Block.Cid().String()
		err := Put(ds, c, Info{
			IdHashPin:      map[string]bool{load.IdHash: true},
			IdHashUnpin:    nil,
			TargetPeerList: getStringSet(load.TargetPeerList),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// merge 将相同键的两个备份信息和二为一
func merge(i1 Info, i2 Info, c string) Info {
	// 先将idHashPin的map合为一体
	for s, b := range i2.IdHashPin {
		t := i1.IdHashPin
		if t == nil {
			t = map[string]bool{s: b}
		} else {
			t[s] = b
		}
		i1.IdHashPin = t
	}
	// 将idHashUnpin合并，并且将unPin从Pin中删除
	for s, s2 := range i2.IdHashUnpin {
		// 验证是否正确
		if t, err := GetIdHash(c, s2); err == nil && t == s {
			i1.IdHashUnpin[s] = s2
			delete(i1.IdHashPin, s)
		}
	}
	for s, s2 := range i1.IdHashUnpin {
		// 验证是否正确
		if t, err := GetIdHash(c, s2); err == nil && t == s {
			delete(i1.IdHashPin, s)
		}
	}

	i1.TargetPeerList.append(i2.TargetPeerList)
	return i1
}

func Get(ds datastore.Datastore, cid string) (Info, error) {
	var b Info
	key := datastore.NewKey(Prefix + cid)
	bytes, err := ds.Get(key)
	if err != nil {
		return b, err
	}
	err = Unmarshal(bytes, &b)
	return b, err
}

func Delete(ds datastore.Datastore, cid string, u string) (int, bool, error) {
	/* 描述: 根据uuid删除某个备份信息
	 * 入参:[datastore,cid,uuid]
	 * 前置状态:
	 * 出参：[len(b.idHashPin),是否进行了删除idHash动作，error]
	 * 后置状态:
	 */
	hash, err := GetIdHash(cid, u)
	if err != nil {
		return 0, false, err
	}
	b, err := Get(ds, cid)
	if err != nil {
		return 0, false, err
	}

	if _, f := b.IdHashPin[hash]; f {
		delete(b.IdHashPin, hash)
		b.IdHashUnpin[hash] = u

		// todo 删除后如果为空，考虑删除该条记录（或者不删）

		err = put(ds, cid, b)
		if err != nil {
			return 0, false, err
		}
		return len(b.IdHashPin), true, nil
	}

	if _, f := b.IdHashUnpin[hash]; f {
		return len(b.IdHashPin), false, nil
	}

	b.IdHashUnpin[hash] = u
	err = put(ds, cid, b)

	return len(b.IdHashPin), false, err
}

// 移除指定cid的备份信息
func Remove(ds datastore.Datastore, cids ...string) error {
	for _, s := range cids {
		err := ds.Delete(datastore.NewKey(Prefix + s))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetAll(ds datastore.Datastore) (map[string]interface{}, error) {
	q := query.Query{
		Prefix: Prefix,
	}
	results, err := ds.Query(q)
	if err != nil {
		return nil, err
	}
	rest, err := results.Rest()
	if err != nil {
		return nil, err
	}
	res := map[string]interface{}{}
	for _, entry := range rest {
		var b Info
		err := Unmarshal(entry.Value, &b)
		if err != nil {
			res[entry.Key] = err.Error()
		} else {
			res[entry.Key] = b
		}
	}

	return res, nil
}
