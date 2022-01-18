package allocate

import (
	model "github.com/bdengine/go-ipfs-blockchain-standard/model"
	bsmsg "github.com/ipfs/go-bitswap/message"
	"github.com/ipfs/go-ipfs-backup/backup"
	"math/rand"
	"time"
)

type Setting struct {
	Strategy  uint
	TargetNum int
}

// AllocateBlocks_LOOP 时间复杂度 len(backupLoadList)*3  --- len(backupLoadList)*len(serverList)
func AllocateBlocks_LOOP(backupLoadList []bsmsg.Load, serverList []model.CorePeer, targetNum int, filePeerMap map[string]backup.StringSet) error {
	// 3.每个节点的文件
	serverLen := len(serverList)
	if serverLen < targetNum {
		return errPeerNotEnough
	}
	// 分配文件
	blockLen := len(backupLoadList)
	rand.Seed(time.Now().Unix())
	for i := 0; i < blockLen; i++ {
		tempCid := backupLoadList[i].Block.Cid().String()
		if len(backupLoadList[i].TargetPeerList) < targetNum {
			//serverList打乱顺序
			rand.Shuffle(serverLen, func(i, j int) {
				serverList[i], serverList[j] = serverList[j], serverList[i]
			})
			// 最多循环节点个数的次数
			for j := 0; j < serverLen; j++ {
				peerSet := filePeerMap[tempCid]
				tempPeer := serverList[j].PeerId
				if peerSet == nil {
					peerSet = backup.StringSet{tempPeer: {}}
				} else if _, f := peerSet[tempPeer]; f {
					continue
				} else {
					peerSet.Add(tempPeer)
				}
				filePeerMap[tempCid] = peerSet
				if len(peerSet) >= targetNum {
					break
				}
			}
			backupLoadList[i].TargetPeerList = filePeerMap[tempCid].GetArray()
			if len(backupLoadList[i].TargetPeerList) != targetNum {
				return ErrBackupNotEnough
			}
		}
	}
	return nil
}
