package accumulator_test

import (
	"bytes"
	"context"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag/traverse"
	"github.com/ur-the-zissou/go-accumulator-dag/accumulator"
	. "github.com/ur-the-zissou/go-accumulator-dag/codec"
	"github.com/ur-the-zissou/go-accumulator-dag/mock"
	"testing"
)

func TestEventDags(t *testing.T) {
	var svc = accumulator.New(mock.BlockService(), 10)

	e :=ProtoEvent{
		Schema:  "Counter",
		Oid:      NewUuid(),
		Event:     NewUuid(),
		Parent:   NewLinkID(),
		Multi:   1,
		Command: []byte("inc"),
		Payload: []byte("hello world"),
		State:   StateVector{1, 0},
	}

	ctx := context.Background()
	svc.Anchor(ctx) // EPOCH:0

	sub, _ := e.EventTree()
	svc.Store(ctx, sub.Payload)
	svc.Store(ctx, sub.Action)
	svc.Store(ctx, sub.Msg)
	svc.Append(ctx, sub.Node())

	svc.AppendTxt(ctx, "0.01")
	svc.Anchor(ctx) // EPOCH:1
	svc.AppendTxt(ctx, "1.01")
	svc.Anchor(ctx) // EPOCH:2
	svc.AppendTxt(ctx, "2.01")
	svc.AppendTxt(ctx, "2.02")
	svc.AppendTxt(ctx, "2.03")
	svc.Anchor(ctx) // EPOCH:3
	svc.Anchor(ctx) // EPOCH:4
	svc.Anchor(ctx) // EPOCH:5

	// format links
	show := func(ls []*ipld.Link) *bytes.Buffer {
		b := new(bytes.Buffer)
		for _, v := range ls {
			s := v.Cid.String()
			b.WriteString(" ")
			b.WriteString(s)
		}
		return b
	}

	_ = svc.Traverse(svc.ProtoNode, func(state traverse.State) error {
		n := state.Node.(*ProtoNode)
		if n.Data() != nil {
			t.Logf("%s, %s, %v", n.Cid(), n.Data(), show(n.Links()))
		} else {
			//t.Logf("weld: %s =>%v", n.Cid(), show(n.Links()))
		}
		return nil
	})
}
