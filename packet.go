package ffmpeg

// #include <libavcodec/avcodec.h>
// #include <libswscale/swscale.h>
// #cgo pkg-config: libavdevice libavformat libavfilter libavcodec libswscale libavutil
import "C"
import (
	"bytes"
	"io"
	"unsafe"
)

type Packet struct {
	p *C.struct_AVPacket
	r *bytes.Reader
}

func NewPacket() *Packet {
	return &Packet{&C.struct_AVPacket{}, nil}
}

func (p *Packet) Read(b []byte) (n int, err error) {
	if p.p.data == nil {
		return 0, io.EOF
	}

	return p.r.Read(b)
}

func (p *Packet) WriteTo(w io.Writer) (n int64, err error) {
	p.r = bytes.NewReader((*[1 << 30]byte)(unsafe.Pointer(p.p.data))[:int(p.p.size):int(p.p.size)])

	return p.r.WriteTo(w)
}
