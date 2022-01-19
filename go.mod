module github.com/ipfs/go-ipfs-backup

go 1.15

require (
	github.com/Hyperledger-TWGC/tjfoc-gm v1.4.0
	github.com/golang/protobuf v1.5.0
	github.com/google/uuid v1.2.0
	github.com/ipfs/go-bitswap v0.3.4
	github.com/ipfs/go-block-format v0.0.3
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-datastore v0.4.5
	github.com/bdengine/go-ipfs-blockchain-standard v0.0.1
	google.golang.org/protobuf v1.27.1
)

replace (
	github.com/ipfs/go-bitswap => ../go-bitswap
	github.com/ipfs/go-cid => ../go-cid
	github.com/bdengine/go-ipfs-blockchain-eth => ../go-ipfs-blockchain-eth
	github.com/bdengine/go-ipfs-blockchain-selector => ../go-ipfs-blockchain-selector
	github.com/bdengine/go-ipfs-blockchain-standard => ../go-ipfs-blockchain-standard
	github.com/ipfs/go-ipfs-backup => ../go-ipfs-backup
)
