package mock

import (
	testinstance "github.com/ipfs/go-bitswap/testinstance"
	tn "github.com/ipfs/go-bitswap/testnet"
	bs "github.com/ipfs/go-blockservice"
	bsrv "github.com/ipfs/go-blockservice"
	ds "github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	delay "github.com/ipfs/go-ipfs-delay"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	mockrouting "github.com/ipfs/go-ipfs-routing/mock"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
)

// BlockServices returns |n| connected mock Blockservices
func BlockServices(n int) []bs.BlockService {
	net := tn.VirtualNetwork(mockrouting.NewServer(), delay.Fixed(0))
	sg := testinstance.NewTestInstanceGenerator(net, nil, nil)

	instances := sg.Instances(n)

	var servs []bs.BlockService
	for _, i := range instances {
		servs = append(servs, bs.New(i.Blockstore(), i.Exchange))
	}
	return servs
}

// DAGService returns a new thread-safe, mock DAGService.
func DAGService() ipld.DAGService {
	return dag.NewDAGService(BlockService())
}

// BlockService returns a new, thread-safe, mock BlockService.
func BlockService() bsrv.BlockService {
	bstore := blockstore.NewBlockstore(dssync.MutexWrap(ds.NewMapDatastore()))
	return bsrv.New(bstore, offline.Exchange(bstore))
}
