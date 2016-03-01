package ffmpeg

// #include <libavcodec/avcodec.h>
// #include <libswscale/swscale.h>
//
// // ... yes. Don't ask.
// typedef struct SwsContext SwsContext;
//
// #ifndef PIX_FMT_RGB0
// #define PIX_FMT_RGB0 AV_PIX_FMT_RGBA
// #endif
//
// #cgo pkg-config: libavdevice libavformat libavfilter libavcodec libswscale libavutil
import "C"

import (
	"fmt"
	"image"
	"io"
	"log"
	"reflect"
	"unsafe"
)

type Codec uint32

const (
	CODEC_ID_MPEG4 Codec = C.AV_CODEC_ID_MPEG4
	CODEC_ID_VP8         = C.AV_CODEC_ID_VP8
)

type Encoder struct {
	codec         Codec
	im            image.Image
	underlying_im image.Image
	Output        io.Writer

	_codec      *C.AVCodec
	_context    *C.AVCodecContext
	_swscontext *C.SwsContext
	_frame      *C.AVFrame
	_packet     *Packet
}

func init() {
	C.avcodec_register_all()
}

func ptr(buf []byte) *C.uint8_t {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	return (*C.uint8_t)(unsafe.Pointer(h.Data))
}

/*
type EncoderOptions struct {
    BitRate uint32
    W, H int
    TimeBase
} */

/*
var DefaultEncoderOptions = EncoderOptions{
    BitRate:400000,
    W: 0, H: 0,
    c.time_base = C.AVRational{1,25}
    c.gop_size = 10
    c.max_b_frames = 1
    c.pix_fmt = C.PIX_FMT_RGB
} */

func NewEncoder(codec Codec, in image.Image, out io.Writer) (*Encoder, error) {
	_codec := C.avcodec_find_encoder(uint32(codec))
	if _codec == nil {
		return nil, fmt.Errorf("could not find codec")
	}

	c := C.avcodec_alloc_context3(_codec)
	f := C.av_frame_alloc()

	c.bit_rate = 400000

	// resolution must be a multiple of two
	w, h := C.int(in.Bounds().Dx()), C.int(in.Bounds().Dy())
	if w%2 == 1 || h%2 == 1 {
		return nil, fmt.Errorf("Bad image dimensions (%d, %d), must be even", w, h)
	}

	log.Printf("Encoder dimensions: %d, %d", w, h)

	c.width = w
	c.height = h
	c.time_base = C.AVRational{1, 25} // FPS
	c.gop_size = 10                   // emit one intra frame every ten frames
	c.max_b_frames = 1

	f.width = w
	f.height = h
	f.format = C.PIX_FMT_RGB0

	underlying_im := image.NewYCbCr(in.Bounds(), image.YCbCrSubsampleRatio420)
	c.pix_fmt = C.AV_PIX_FMT_YUV420P
	f.data[0] = ptr(underlying_im.Y)
	f.data[1] = ptr(underlying_im.Cb)
	f.data[2] = ptr(underlying_im.Cr)
	f.linesize[0] = w
	f.linesize[1] = w / 2
	f.linesize[2] = w / 2

	if C.avcodec_open2(c, _codec, nil) < 0 {
		return nil, fmt.Errorf("could not open codec")
	}

	_swscontext := C.sws_getContext(w, h, C.PIX_FMT_RGB0, w, h, C.AV_PIX_FMT_YUV420P,
		C.SWS_BICUBIC, nil, nil, nil)

	e := &Encoder{codec, in, underlying_im, out, _codec, c, _swscontext, f, NewPacket()}
	return e, nil
}

func (e *Encoder) WriteFrame(m image.Image) error {
	e._frame.pts = C.int64_t(e._context.frame_number)
	C.av_init_packet(e._packet.p)
	e._packet.p.data = nil
	e._packet.p.size = 0

	var input_data [3]*C.uint8_t
	var input_linesize [3]C.int

	switch im := m.(type) {
	case *image.RGBA:
		bpp := 4
		input_data = [3]*C.uint8_t{ptr(im.Pix)}
		input_linesize = [3]C.int{C.int(im.Bounds().Dx() * bpp)}
	case *image.NRGBA:
		bpp := 4
		input_data = [3]*C.uint8_t{ptr(im.Pix)}
		input_linesize = [3]C.int{C.int(im.Bounds().Dx() * bpp)}
	default:
		panic("Unknown input image type")
	}

	// Perform scaling from input type to output type
	C.sws_scale(e._swscontext, &input_data[0], &input_linesize[0],
		0, e._context.height,
		&e._frame.data[0], &e._frame.linesize[0])

	filled := C.int(0)
	ret := C.avcodec_encode_video2(e._context, e._packet.p, e._frame, &filled)

	if ret < 0 {
		return nil
	}

	if filled > 0 {
		n, err := io.Copy(e.Output, e._packet)
		C.av_free_packet(e._packet.p)

		if err != nil {
			return err
		}
		if n < int64(e._packet.p.size) {
			return fmt.Errorf("Short write, expected %d, wrote %d", e._packet.p.size, n)
		}
	}

	return nil
}

func (e *Encoder) Close() {

	// Process "delayed" frames
	for filled := C.int(1); filled > 0; {
		e._frame.pts = C.int64_t(e._context.frame_number)
		C.av_init_packet(e._packet.p)
		e._packet.p.data = nil
		e._packet.p.size = 0
		ret := C.avcodec_encode_video2(e._context, e._packet.p, nil, &filled)

		if ret < 0 {
			break
		}

		if filled > 0 {
			n, err := io.Copy(e.Output, e._packet)
			C.av_free_packet(e._packet.p)
			if err != nil {
				panic(err)
			}
			if n < int64(e._packet.p.size) {
				panic(fmt.Errorf("Short write, expected %d, wrote %d", e._packet.p.size, n))
			}
		}
	}

	n, err := e.Output.Write([]byte{0, 0, 1, 0xb7})
	if err != nil || n != 4 {
		log.Panicf("Error finishing mpeg file: %q; n = %d", err, n)
	}

	C.avcodec_close((*C.AVCodecContext)(unsafe.Pointer(e._context)))
	C.av_free(unsafe.Pointer(e._context))
	C.av_free(unsafe.Pointer(e._frame))
	e._frame, e._codec = nil, nil
	e._packet.p.data, e._packet.p = nil, nil
}
