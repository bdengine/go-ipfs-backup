package allocate

import "fmt"

var (
	errOrgNotEnough       = fmt.Errorf("组织数不满足条件")
	errPeerNotEnough      = fmt.Errorf("节点数不满足条件")
	errOwnerPeerNotEnough = fmt.Errorf("本组织节点数不满足条件")
	ErrBackupNotEnough    = fmt.Errorf("备份数不足")
)
