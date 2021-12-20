package backup

import (
	"encoding/json"
	"fmt"
	"github.com/ipfs/go-bitswap/message"
	"github.com/ipfs/go-datastore"
)

const (
	fileBackupPrefix = "fileBackup/"
)

type FileInfo struct {
	Cid       string
	Size      uint64
	Uid       string
	BlockList []blockInfo
}

type blockInfo struct {
	Cid            string
	TargetPeerList []string
}

func transLoadToInfo(load *message.Load) (res blockInfo) {
	res.Cid = load.Block.Cid().String()
	res.TargetPeerList = load.TargetPeerList
	return res
}

func batchTransLoadToInfo(loads []message.Load) []blockInfo {
	l := len(loads)
	res := make([]blockInfo, l)
	for i, load := range loads {
		res[i] = transLoadToInfo(&load)
	}
	return res
}

func AddFileBackupInfo(ds datastore.Datastore, loadList []message.Load, uid string, size uint64) (*FileInfo, error) {
	cid := loadList[0].Block.Cid().String()
	key := datastore.NewKey(fileBackupPrefix + cid)
	_, err := ds.Get(key)
	if err == nil {
		return nil, fmt.Errorf("文件备份信息已存在")
	} else if err != datastore.ErrNotFound {
		return nil, err
	}
	file := FileInfo{
		Cid:       cid,
		Size:      size,
		Uid:       uid,
		BlockList: batchTransLoadToInfo(loadList),
	}
	marshal, err := json.Marshal(file)
	if err != nil {
		return nil, err
	}
	err = ds.Put(key, marshal)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func GetFileBackupInfo(ds datastore.Datastore, cid string) (*FileInfo, error) {
	key := datastore.NewKey(fileBackupPrefix + cid)
	marshal, err := ds.Get(key)
	if err != nil {
		return nil, err
	}
	var file FileInfo
	err = json.Unmarshal(marshal, &file)
	return &file, err
}
