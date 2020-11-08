package dag_test

import (
	"bytes"
	"context"
	accumulator "github.com/ur-the-zissou/go-accumulator-dag/accumulator"
	"github.com/ur-the-zissou/go-accumulator-dag/mock"
	"testing"

	mdag "github.com/ipfs/go-merkledag"
	"github.com/stretchr/testify/assert"

	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag/traverse"
)

func testAddNode(t *testing.T, ds *accumulator.Service, name string) ipld.Node {
	n := mdag.NodeWithData([]byte("Foo"))
	ctx := context.Background()
	depth, ok := ds.Append(ctx, n)
	t.Logf("Depth: %v OK: %t n: %v", depth, ok, n)
	return n
}

func testBuildDag(t *testing.T, ds *accumulator.Service) (*mdag.ProtoNode, bool) {
	testAddNode(t, ds, "aa")
	testAddNode(t, ds, "ab")
	testAddNode(t, ds, "ac")
	testAddNode(t, ds, "ad")
	testAddNode(t, ds, "ba")
	testAddNode(t, ds, "bb")
	testAddNode(t, ds, "bc")
	testAddNode(t, ds, "bd")
	ctx := context.Background()
	return ds.TruncateRoot(ctx)
}

func TestAccumulatorDag_Append(t *testing.T) {
	ds := accumulator.New(mock.BlockService(), 5)
	n, _ := testBuildDag(t, ds)
	assert.Equal(t, "QmfJb5i379FnVsBEn4X3doh4MBnogmkJqRKSJHWuESQTbU", n.Cid().String())

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

	ds.Traverse(n, func(state traverse.State) error {
		n := state.Node.(*mdag.ProtoNode)
		if n.Data() != nil {
			t.Logf("%s, %s, %v", n.Cid(), n.Data(), show(n.Links()))
		} else {
			t.Logf("weld: %s =>%v", n.Cid(), show(n.Links()))
		}
		return nil
	})
}
