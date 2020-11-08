package accumulator

import (
	"context"
	"fmt"
	"sync"

	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-merkledag/traverse"
	. "github.com/ur-the-zissou/go-accumulator-dag/codec"
	"github.com/ur-the-zissou/go-accumulator-dag/dag"
)

// Max Depth of Accumulator
var DagSize = 20

// access to DagService used to build IPFS DAGS

type Dag interface {
	GetDepth() uint64
	Store(ctx context.Context, n *ProtoNode) error
	Append(ctx context.Context, n *ProtoNode) (depth uint64, ok bool)
	TruncateRoot(ctx context.Context) (n *ProtoNode, ok bool)
	Traverse(n *ProtoNode, traverseFunc TraverseFunc) error
}

// function used to read a dag using BFS
type TraverseFunc = traverse.Func

// finite FiniteService
type Service struct {
	Dag
	*ProtoNode
	sync.Mutex
	Epoch uint64
}

func New(bs blockservice.BlockService, dagSize ...uint64) *Service {
	if len(dagSize) == 1 {
		DagSize = int(dagSize[0])
	}
	Service := new(Service)
	Service.Dag = dag.NewAccDag(bs, DagSize)
	return Service
}

// Truncating Dag terminate with Ref node
func (e *Service) Anchor(ctx context.Context) (a *Ref, ok bool) {
	e.Lock()
	defer func() { e.Unlock() }()
	a = new(Ref)
	a.Epoch = e.Epoch
	a.Depth = e.GetDepth()
	e.Append(ctx, NodeWithData([]byte(fmt.Sprintf(`%s:%v`, Epoch, e.Epoch)))) // write Ref message
	e.ProtoNode, ok = e.TruncateRoot(ctx)
	a.Cid = e.Cid().String()

	if ok {
		e.Append(ctx, a.ToNode())
		e.Append(ctx, e.ProtoNode)
		e.Epoch++
	} else {
		panic("truncate failed")
	}
	return a, ok
}

// Store bytes bypassing Dag Service
func (e *Service) StoreBytes(ctx context.Context, b []byte) (n *ProtoNode, ok bool) {
	n = NodeWithData(b)
	err := e.Store(ctx, n)
	return n, err == nil
}

// Store string data bypassing Dag Service
func (e *Service) StoreString(ctx context.Context, str string) (n *ProtoNode, ok bool) {
	return e.StoreBytes(ctx, []byte(str))
}

// convenience wrapper to write bytes to the Dag FiniteService
func (e *Service) AppendBytes(ctx context.Context, b []byte) (n *ProtoNode, depth uint64, ok bool) {
	n = NodeWithData(b)
	depth, ok = e.Append(ctx, n)
	return n, depth, ok
}

// append human-readable messages < 128 chars to the DAG
func (e *Service) AppendTxt(ctx context.Context, str string) (n *ProtoNode, depth uint64, ok bool) {
	if len(str) < 128 {
		return e.AppendBytes(ctx, []byte(fmt.Sprintf(`%s:%s`, Txt, str)))
	} else {
		return nil, e.GetDepth(), false
	}
}
