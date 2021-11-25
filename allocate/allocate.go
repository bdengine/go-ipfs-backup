package allocate

import (
	bsmsg "github.com/ipfs/go-bitswap/message"
	model "github.com/ipfs/go-ipfs-auth/standard/model"
	"math/rand"
	"time"
)

type Setting struct {
	Strategy  uint
	TargetNum int
}

// AllocateBlocks_LOOP 时间复杂度 len(backupLoadList)*3  --- len(backupLoadList)*len(serverList)
func AllocateBlocks_LOOP(backupLoadList []bsmsg.Load, serverList []model.CorePeer, targetNum int, ownOrg string, orgMap map[string][]string, orgFileMap map[string]map[string]int, peerFileMap map[string]map[string]int) error {
	// 对于blockList[0],在本组织随机选择三个节点保存文件
	// 需要维护的信息
	// 1.每个组织的不同文件数量
	// 2.每个组织的相同文件数量
	// 3.每个节点的文件
	orgLen := len(orgMap)
	if orgLen <= targetNum {
		return errOrgNotEnough
	}
	serverLen := len(serverList)
	if serverLen < orgLen*4 {
		return errPeerNotEnough
	}
	// 分配头文件
	ownPeer := orgMap[ownOrg]
	n := len(ownPeer)
	if n < targetNum {
		return errOwnerPeerNotEnough
	}
	ownPeerCopy := make([]string, n)
	copy(ownPeerCopy, ownPeer)
	rand.Shuffle(n, func(i, j int) {
		ownPeerCopy[i], ownPeerCopy[j] = ownPeerCopy[j], ownPeerCopy[i]
	})
	backupLoadList[0].TargetPeerList = ownPeerCopy[:3]
	headCid := backupLoadList[0].Block.Cid().String()
	// 记录头文件分配
	for _, s := range backupLoadList[0].TargetPeerList {
		fl := peerFileMap[s]
		if fl == nil {
			fl = map[string]int{headCid: 1}
		} else {
			fl[headCid] = fl[headCid] + 1
		}
		peerFileMap[s] = fl
	}
	fl := orgFileMap[ownOrg]
	if fl == nil {
		fl = map[string]int{headCid: 3}
	}
	orgFileMap[ownOrg] = fl

	// 分配其他文件
	blockLen := len(backupLoadList)
	rand.Seed(time.Now().Unix())
	for i := 1; i < blockLen; i++ {
		tempCid := backupLoadList[i].Block.Cid().String()
		//serverList打乱顺序
		rand.Shuffle(serverLen, func(i, j int) {
			serverList[i], serverList[j] = serverList[j], serverList[i]
		})
		// 最多循环节点个数的次数
		for j := 0; j < serverLen; j++ {
			tempOrg := serverList[j].Org
			tempPeer := serverList[j].PeerId
			// 组织条件
			// 1.没有所有文件 2.没有所有备份, 备份数小于节点数
			o1 := len(orgFileMap[tempOrg]) < blockLen-1
			o2 := orgFileMap[tempOrg][tempCid] < targetNum-1 && orgFileMap[tempOrg][tempCid] < len(orgMap[tempOrg])
			// 节点条件 p1 没有所有文件 p2 没有重复文件
			p1 := len(peerFileMap[tempPeer]) < blockLen-1
			_, p2 := peerFileMap[tempPeer][tempCid]

			// 满足所有条件
			if o1 && o2 && p1 && !p2 {
				// 维护组织信息
				temp := orgFileMap[tempOrg]
				if temp == nil {
					temp = map[string]int{tempCid: 1}
				} else {
					temp[tempCid] = temp[tempCid] + 1
				}
				orgFileMap[tempOrg] = temp

				// 维护节点信息
				temp = peerFileMap[tempPeer]
				if temp == nil {
					temp = map[string]int{tempCid: 1}
				} else {
					temp[tempCid] = temp[tempCid] + 1
				}
				peerFileMap[tempPeer] = temp

				// 维护Load信息
				tempPl := backupLoadList[i].TargetPeerList
				if tempPl == nil {
					tempPl = []string{tempPeer}
				} else {
					tempPl = append(tempPl, tempPeer)
				}
				backupLoadList[i].TargetPeerList = tempPl

				if len(tempPl) >= targetNum {
					break
				}
			}
		}
		if len(backupLoadList[i].TargetPeerList) != targetNum {
			return ErrBackupNotEnough
		}
	}
	return nil
}
