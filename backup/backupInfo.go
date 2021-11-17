package backup

import (
	"encoding/base64"
	"fmt"
	"github.com/Hyperledger-TWGC/tjfoc-gm/sm3"
	"github.com/golang/protobuf/proto"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
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

type Info struct {
	IdHashPin      map[string]bool
	IdHashUnpin    map[string]string
	TargetPeerList []string
}

func (b *Info) GetPb() pb.BackupInfo {
	return pb.BackupInfo{
		IdHashPin:      b.IdHashPin,
		IdHashUnpin:    b.IdHashUnpin,
		TargetPeerList: b.TargetPeerList,
	}
}

func (b *Info) LoadPb(pbi *pb.BackupInfo) {
	b.IdHashPin = pbi.IdHashPin
	b.IdHashUnpin = pbi.IdHashUnpin
	b.TargetPeerList = pbi.TargetPeerList
}

// todo pb化的改动
func Marshal(b Info) ([]byte, error) {
	getPb := b.GetPb()
	return proto.Marshal(&getPb)
}

// todo pb化的改动
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
	key := datastore.NewKey(Prefix + cid)
	marshal, err := Marshal(b)
	if err != nil {
		return err
	}
	err = ds.Put(key, marshal)
	return err
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
		err = Put(ds, cid, b)
		if err != nil {
			return 0, false, err
		}
		return len(b.IdHashPin), true, nil
	}

	if _, f := b.IdHashUnpin[hash]; f {
		return len(b.IdHashPin), false, nil
	}

	b.IdHashUnpin[hash] = u
	err = Put(ds, cid, b)

	return len(b.IdHashPin), false, err
}
