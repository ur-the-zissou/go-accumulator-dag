package dag

import (
	"context"
	. "github.com/ur-the-zissou/go-accumulator-dag/codec"
	"sync"

	bserv "github.com/ipfs/go-blockservice"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	mdag "github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-merkledag/traverse"
)

type dagService interface {
	Add(ctx context.Context, nd ipld.Node) error
	AddMany(ctx context.Context, nds []ipld.Node) error
	Get(ctx context.Context, c cid.Cid) (ipld.Node, error)
	GetLinks(ctx context.Context, c cid.Cid) ([]*ipld.Link, error)
	Remove(ctx context.Context, c cid.Cid) error
	RemoveMany(ctx context.Context, cids []cid.Cid) error
	GetMany(context.Context, []cid.Cid) <-chan *ipld.NodeOption
}

func Node(d []byte) *mdag.ProtoNode {
	return mdag.NodeWithData(d)
}

type DagService struct {
	Blocks bserv.BlockService
	Srv    dagService
}

func New(bs bserv.BlockService) *DagService {
	d := new(DagService)
	d.Blocks = bs
	d.Srv = mdag.NewDAGService(d.Blocks)
	return d
}

type AccDag struct {
	*DagService
	sync.Mutex
	Accumulator []*ProtoNode
	Size        int
	Depth       uint64
	Epoch       uint64
}

func NewAccDag(bs bserv.BlockService, size int) *AccDag {
	d := new(AccDag)
	d.DagService = New(bs)
	d.Size = size
	d.Accumulator = make([]*ProtoNode, size)
	return d
}

// Store node - bypassing a position in Accumulator DAG
func (ad *AccDag) Store(ctx context.Context, node *ProtoNode) error {
	return ad.DagService.Srv.Add(ctx, node)
}

// Store node - bypassing a position in Accumulator DAG
func (ad *AccDag) GetDepth() uint64 {
	defer func() { ad.Unlock() }()
	ad.Lock()
	return ad.Depth
}

func (ad *AccDag) GetEpoch() uint64 {
	defer func() { ad.Unlock() }()
	ad.Lock()
	return ad.Depth
}

func (ad *AccDag) SetEpoch(e uint64) {
	defer func() { ad.Unlock() }()
	ad.Lock()
	ad.Epoch = e
}

func (ad *AccDag) get(i int) *ProtoNode {
	return ad.Accumulator[i]
}

func (ad *AccDag) set(i int, n *ProtoNode) {
	ad.Accumulator[i] = n
}

func (ad *AccDag) weld(ctx context.Context, a *ProtoNode, b *ProtoNode) (n *ProtoNode, err error) {
	n = new(ProtoNode)
	n.AddNodeLink(WeldLeft, a)
	n.AddNodeLink(WeldRight, b)
	err = ad.DagService.Srv.Add(ctx, n)
	return n, err
}

func (ad *AccDag) Append(ctx context.Context, n *ProtoNode) (depth uint64, ok bool) {
	defer func() { ad.Unlock() }()
	ad.Lock()
	err := ad.DagService.Srv.Add(ctx, n)
	ok = false
	if err != nil {
		return ad.Depth, ok
	}
	for i := 0; i < ad.Size; i++ {
		target := ad.get(i)
		if target == nil {
			ad.set(i, n)
			ok = true
			break
		} else {
			n, _ = ad.weld(ctx, n, target)
			ad.set(i, nil)
		}
	}
	if ok {
		ad.Depth++
	}
	return ad.Depth, ok
}

// reduce the accumulator slice to produce a merkle root
func (ad *AccDag) TruncateRoot(ctx context.Context) (n *ProtoNode, ok bool) {
	defer func() { ad.Unlock() }()
	ad.Lock()
	for i := 0; i < ad.Size; i++ {
		target := ad.get(i)
		if n == nil {
			n = target
			continue
		}
		if target != nil {
			n, _ = ad.weld(ctx, n, target)
			ad.set(i, nil)
		}
	}
	ad.Depth = 0
	return n, n != nil
}

// walk the DAG w/ BFS
func (ad *AccDag) Traverse(n *ProtoNode, traverseFunc traverse.Func) error {
	return traverse.Traverse(n, traverse.Options{
		Order:          traverse.BFS,
		SkipDuplicates: true, // NOTE: this suppresses duplicate nodes!!
		Func:           traverseFunc,
		DAG:            ad.DagService.Srv,
	})
}
