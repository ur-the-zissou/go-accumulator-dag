package codec

import (
	"encoding/json"

	"github.com/google/uuid"
	mdag "github.com/ipfs/go-merkledag"
)

// coding for IPLD link, labels, and any.Any dataURL
const (
	WeldLeft   = "←" // https://en.wikipedia.org/wiki/List_of_Unicode_characters#Arrows
	WeldRight  = "→"
	WeldAttach = "↑"
	WeldAnchor = "↓"
	Payload    = "Ⓟ" // https://en.wikipedia.org/wiki/Enclosed_Alphanumerics
	Action     = "Ⓐ"
	Signature  = "Ⓢ"
	Epoch      = "Ⓔ"
	Txt        = "⒯"
)

type ProtoNode = mdag.ProtoNode
type Schema = string

// reference to an object or object-hierarchy
// Oid should not allow any inbound arcs
type ObjectID = uuid.UUID

// link between parent<-event nodes
type LinkID = uuid.UUID

// application specific data
// uses proto3 Any.any data
type Raw = []byte

type ProtoEventFactory interface {
	ProtoEvent() *ProtoEvent
}

// reference mdag func
var NodeWithData = mdag.NodeWithData

// transitional datatype for converting between protobuf <-> IPLD Node
type ProtoEvent struct {
	Schema  Schema      `json:"schema"`
	Oid     ObjectID    `json:"oid"`
	Parent  LinkID      `json:"parent"`
	Event   LinkID      `json:"eventId"`
	Multi   uint64      `json:"multi"`
	Command Raw         `json:"-"`
	Payload Raw         `json:"-"`
	State   StateVector `json:"-"`
}

// state transformation delta
type Vector []int64

// allow conversion to []uint64
func (v Vector) ToStateVector() StateVector {
	x := make([]uint64, len(v))
	for k, s := range v {
		x[k] = uint64(s)
	}
	return x
}

// persisted state
type StateVector []uint64

// allow conversion to []int64
func (v StateVector) ToVector() Vector {
	x := make([]int64, len(v))
	for k, s := range v {
		x[k] = int64(s)
	}
	return x
}

// construct new UUID with single return
func NewUuid() uuid.UUID {
	id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	return id
}

// generate new LinkID
func NewLinkID() LinkID { return NewUuid() }

// generate new ObjectID
func NewObjectID() ObjectID { return NewUuid() }

// load UUID from string with single return
func ParseUuid(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

// IPLD node implementation of an event
type EventTree struct {
	Msg     *ProtoNode
	Action  *ProtoNode
	Payload *ProtoNode
}

func (t *EventTree) Node() (n *ProtoNode) {
	return Ref{Cid: t.Msg.Cid().String()}.Node()
}

// convert Event to IPLD nodes
func (pe *ProtoEvent) EventTree() (sub *EventTree, err error) {
	var msg []byte
	msg, err = json.Marshal(pe)
	sub = new(EventTree)
	sub.Msg = mdag.NodeWithData(msg)
	sub.Action = mdag.NodeWithData(pe.Command)
	sub.Payload = mdag.NodeWithData(pe.Payload)
	_ = sub.Msg.AddNodeLink(Action, sub.Action)
	_ = sub.Msg.AddNodeLink(Payload, sub.Payload)
	return sub, err
}

func (a Ref) Node() *ProtoNode {
	data, _ := json.Marshal(a)
	return mdag.NodeWithData(data)
}

type Ref struct {
	Cid   string `json:"cid"`
	Depth uint64 `json:"depth,omitempty"`
	Epoch uint64 `json:"epoch,omitempty"`
}

// convert anchor to IPLD Node
func (a *Ref) ToNode() *ProtoNode {
	data, _ := json.Marshal(a)
	return mdag.NodeWithData(data)
}
