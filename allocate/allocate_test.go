package allocate

import (
	"fmt"
	"github.com/ipfs/go-bitswap/message"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-ipfs-auth/standard/model"
	"testing"
)

func TestAllocateBlocks_LOOP(t *testing.T) {
	time := 6
	eCh := make(chan int, time)
	fCh := make(chan int, time)

	for j := 0; j < time; j++ {
		go func() {
			failTime := 0
			errTime := 0
		test:
			for i := 0; i < 100; i++ {
				targetNum := 3
				blockLen := 1 * 1024 * 1024 / 256
				ownOrg := "org1"
				//fmt.Printf("****第%v次测试****\n",i)
				serverList, orgMap, orgFileMap, peerFileMap := getFakeServerList()
				bList := getFakeBlockList(blockLen)
				err := AllocateBlocks_LOOP(bList, serverList, targetNum, ownOrg, orgMap, orgFileMap, peerFileMap)
				if err != nil {
					errTime++
					// try again
					if err == ErrBackupNotEnough {
						serverList, orgMap, orgFileMap, peerFileMap = getFakeServerList()
						bList = getFakeBlockList(blockLen)
						err = AllocateBlocks_LOOP(bList, serverList, targetNum, ownOrg, orgMap, orgFileMap, peerFileMap)
						if err != nil {
							failTime++
							continue test
						}
					}
					//t.Fatal(err)
				}

				bL1 := 0
				bL2 := 0
				for org := range orgFileMap {
					fileNum := len(orgFileMap[org])
					//s1 := fmt.Sprintf("%v持有不同的文件片比列为%v", org, float64(fileNum)/float64(blockLen))
					if fileNum > blockLen {
						failTime++
						continue test
						//t.Fatalf("%v持有不同的文件片比列为%v", org, float64(fileNum)/float64(blockLen))
					}
					//fmt.Println(s1)
					oFNum := 0
					for f := range orgFileMap[org] {
						sFileNum := orgFileMap[org][f]
						//s2 := fmt.Sprintf("%v持有的文件片%v数量为%v", org, f, sFileNum)
						if org == ownOrg && f == bList[0].Block.Cid().String() {
							if sFileNum != targetNum {
								failTime++
								continue test
							}
						} else if sFileNum >= targetNum || sFileNum > len(orgMap[org]) {
							failTime++
							continue test
						}
						//fmt.Println(s2)
						bL1 += sFileNum
						oFNum += sFileNum
					}
					//fmt.Printf("%v持有不同的文件片比列为%v,持有文件片比例为%v\n", org, float64(fileNum)/float64(blockLen), float64(oFNum)/(3*float64(blockLen)))
					peerList := orgMap[org]
					for _, peer := range peerList {
						fileMap := peerFileMap[peer]
						for _, i := range fileMap {
							if i != 1 {
								failTime++
								continue test
								//t.Fatalf("%v在%v中数量为%v",s,peer,i)
							}
						}
						bL2 += len(fileMap)
						//fmt.Printf("%v:%v,",peer,len(fileMap))
					}
					//fmt.Println("")
				}
				if bL1 != blockLen*targetNum {
					failTime++
					continue test
					//t.Fatal("组织记录备份数量错误")
				}
				if bL2 != blockLen*targetNum {
					failTime++
					continue test
					//t.Fatal("节点记录备份数量错误")
				}
				for _, load := range bList {
					if len(load.TargetPeerList) != 3 {
						failTime++
						continue test
						//t.Fatalf("%v落点为%v",load.Block.Cid().String(),load.TargetPeerList)
					}
				}
			}
			eCh <- errTime
			fCh <- failTime
		}()
	}
	errTime := 0
	failTime := 0
	for i := 0; i < time; i++ {
		errTime += <-eCh
		failTime += <-fCh
	}
	fmt.Printf("错误次数%v,不满足条件次数%v", errTime, failTime)

}

func getFakeServerList() ([]model.CorePeer, map[string][]string, map[string]map[string]int, map[string]map[string]int) {
	var serverList []model.CorePeer
	orgMap := map[string][]string{}
	orgFileMap := map[string]map[string]int{}
	peerFileMap := map[string]map[string]int{}

	for i := 0; i < 4; i++ {
		tempOrg := fmt.Sprintf("org%v", i)
		for j := 0; j < 4; j++ {
			tempPeer := fmt.Sprintf("org%vpeer%v", i, j)
			serverList = append(serverList, model.CorePeer{
				Org: tempOrg,
				Peer: model.Peer{
					UserCode: "",
					PeerId:   tempPeer,
					Role:     0,
				},
				Addresses: nil,
			})
			orgMap[tempOrg] = append(orgMap[tempOrg], tempPeer)
		}
	}

	return serverList, orgMap, orgFileMap, peerFileMap
}

func getFakeBlockList(n int) []message.Load {
	var res []message.Load
	for i := 0; i < n; i++ {
		tempHash := fmt.Sprintf("idHash%v", i)
		tempBlockData := fmt.Sprintf("block%v", i)
		res = append(res, message.Load{
			TargetPeerList: nil,
			IdHash:         tempHash,
			Block:          blocks.NewBlock([]byte(tempBlockData)),
		})
	}
	return res
}
