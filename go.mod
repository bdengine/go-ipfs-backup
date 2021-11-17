module github.com/ipfs/go-ipfs-backup

go 1.15

require (
	github.com/Hyperledger-TWGC/tjfoc-gm v1.4.0
	github.com/golang/protobuf v1.5.0
	github.com/google/uuid v1.2.0
	github.com/ipfs/go-bitswap v0.3.4
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-datastore v0.4.5
	github.com/ipfs/go-ipfs-auth/standard v0.0.0
	google.golang.org/protobuf v1.27.1

)

replace (
	github.com/ipfs/go-bitswap => ../go-bitswap
	github.com/ipfs/go-cid => ../ipld/cid/go-cid
	github.com/ipfs/go-ipfs-auth/auth-source-fabric => ../go-ipfs-auth/auth-source-fabric
	github.com/ipfs/go-ipfs-auth/selector => ../go-ipfs-auth/selector
	github.com/ipfs/go-ipfs-auth/standard => ../go-ipfs-auth/standard
	github.com/ipfs/go-ipfs-backup => ../go-ipfs-backup
)
