package cdbpool

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
	"github.com/k81/kate/knet"
)

type Header struct {
	Command   uint32
	Magic     uint32
	ContextId uint32
	Reserved1 uint32
	Reserved2 uint32
	BodyLen   uint32
}

func (h *Header) Id() uint64 {
	return uint64(h.ContextId)
}

const (
	CmdPing  = 0x1
	CmdQuery = 0x88888888
)

type Packet struct {
	Header
	proto.Message
}

var (
	nextId = uint32(0)
)

func nextRequestId() uint32 {
	return atomic.AddUint32(&nextId, 1)
}

func newPacket(id uint32, cmd uint32, m proto.Message) *Packet {
	pkt := &Packet{
		Header: Header{
			Command:   cmd,
			Magic:     0x8756457F,
			ContextId: id,
		},
		Message: m,
	}
	return pkt
}

func newPingPacket() *Packet {
	return newPacket(0, CmdPing, nil)
}

func newQueryPacket(id uint32, m proto.Message) *Packet {
	return newPacket(id, CmdQuery, m)
}

const (
	KeyReadBuf   = "r_buf"
	KeyWriteBuf  = "w_buf"
	KeyDecodeBuf = "d_buf"
	KeyEncodeBuf = "e_buf"
)

type Protocol struct{}

func (p *Protocol) Decode(session *knet.IoSession, reader io.Reader) (m knet.Message, err error) {
	var (
		pkt  = &Packet{}
		resp = &CdbPoolResponse{}
	)

	if err = binary.Read(reader, binary.BigEndian, &pkt.Header); err != nil {
		return
	}

	if pkt.BodyLen == 0 {
		m = pkt
		return
	}

	var (
		decodeBuf *proto.Buffer
		readBuf   *bytes.Buffer
		ok        bool
	)

	if readBuf, ok = session.GetAttr(KeyReadBuf).(*bytes.Buffer); !ok {
		readBuf = &bytes.Buffer{}
		session.SetAttr(KeyReadBuf, readBuf)
	}
	readBuf.Reset()

	if _, err = io.CopyN(readBuf, reader, int64(pkt.BodyLen)); err != nil {
		return
	}

	if decodeBuf, ok = session.GetAttr(KeyDecodeBuf).(*proto.Buffer); !ok {
		decodeBuf = &proto.Buffer{}
		session.SetAttr(KeyDecodeBuf, decodeBuf)
	}
	decodeBuf.SetBuf(readBuf.Bytes())

	if err = decodeBuf.Unmarshal(resp); err != nil {
		return
	}

	pkt.Message = resp
	m = pkt
	return
}

func (p *Protocol) Encode(session *knet.IoSession, m knet.Message) (data []byte, err error) {
	var (
		pkt       = m.(*Packet)
		writeBuf  *bytes.Buffer
		encodeBuf *proto.Buffer
		ok        bool
	)

	if writeBuf, ok = session.GetAttr(KeyWriteBuf).(*bytes.Buffer); !ok {
		writeBuf = &bytes.Buffer{}
		session.SetAttr(KeyWriteBuf, writeBuf)
	}
	writeBuf.Reset()

	if pkt.Message != nil {
		if encodeBuf, ok = session.GetAttr(KeyEncodeBuf).(*proto.Buffer); !ok {
			encodeBuf = &proto.Buffer{}
			session.SetAttr(KeyEncodeBuf, encodeBuf)
		}
		encodeBuf.Reset()

		if err = encodeBuf.Marshal(pkt.Message); err != nil {
			return
		}

		pkt.Header.BodyLen = uint32(len(encodeBuf.Bytes()))
	}

	binary.Write(writeBuf, binary.BigEndian, &pkt.Header)

	if encodeBuf != nil {
		writeBuf.Write(encodeBuf.Bytes())
	}

	data = writeBuf.Bytes()
	return
}
