package network

import (
	"bytes"
	"io"
	"sync"

	kpgzip "github.com/klauspost/compress/gzip"
)

const (
	FlagNone       byte = 0x00
	FlagCompressed byte = 0x01
	FlagEncrypted  byte = 0x02
)

var defaultCompressionCodec = NewCompressionCodec(nil)

type CompressionConfig struct {
	Enabled          bool
	MinSize          int
	CompressionLevel int     // default level for larger payloads
	CompressionRatio float64 // 压缩后/原始大小阈值
	FastThreshold    int     // 小于该阈值使用 BestSpeed 池
}

func DefaultCompressionConfig() *CompressionConfig {
	return &CompressionConfig{
		Enabled:          true,
		MinSize:          1024,
		CompressionLevel: kpgzip.DefaultCompression,
		CompressionRatio: 0.9,
		FastThreshold:    8 * 1024,
	}
}

type CompressionCodec struct {
	codec  *Codec
	config *CompressionConfig

	// 两个 writer 池：fast（BestSpeed）与 default（config.CompressionLevel）
	writerPoolFast    sync.Pool
	writerPoolDefault sync.Pool

	bufferPool sync.Pool
}

func NewCompressionCodec(config *CompressionConfig) *CompressionCodec {
	if config == nil {
		config = DefaultCompressionConfig()
	}

	cc := &CompressionCodec{
		codec:  NewCodec(),
		config: config,
	}

	// fast pool: BestSpeed writers
	cc.writerPoolFast = sync.Pool{
		New: func() interface{} {
			w, _ := kpgzip.NewWriterLevel(nil, kpgzip.BestSpeed)
			return w
		},
	}
	// default pool: writers created with config.CompressionLevel
	cc.writerPoolDefault = sync.Pool{
		New: func() interface{} {
			level := config.CompressionLevel
			// ensure level valid (kpgzip accepts constants similar to stdlib)
			w, _ := kpgzip.NewWriterLevel(nil, level)
			return w
		},
	}

	cc.bufferPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	return cc
}

func (cc *CompressionCodec) EncodeMessageWithCompression(msg *Message) ([]byte, error) {
	if msg == nil {
		return nil, ErrInvalidMessage
	}

	flags := byte(0)
	payload := msg.Payload

	if cc.config.Enabled && len(payload) >= cc.config.MinSize {
		compressed, err := cc.compress(payload)
		if err == nil {
			ratio := float64(len(compressed)) / float64(len(payload))
			if ratio < cc.config.CompressionRatio {
				payload = compressed
				flags |= FlagCompressed
			}
		}
	}

	totalLen := 2 + len(payload)
	if totalLen > cc.codec.maxFrameSize {
		return nil, ErrFrameTooLarge
	}

	frameSize := 4 + totalLen
	buf := GetBuffer(frameSize)
	cc.codec.byteOrder.PutUint32(buf[:4], uint32(totalLen))
	buf[4] = msg.Type
	buf[5] = flags
	copy(buf[6:], payload)

	return buf, nil
}

func (cc *CompressionCodec) DecodeMessageWithCompression(reader io.Reader) (*Message, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(reader, header); err != nil {
		return nil, err
	}

	totalLen := cc.codec.byteOrder.Uint32(header)
	if totalLen > uint32(cc.codec.maxFrameSize) {
		return nil, ErrFrameTooLarge
	}
	if totalLen < 2 {
		return nil, ErrInvalidMessage
	}

	body := make([]byte, totalLen)
	if _, err := io.ReadFull(reader, body); err != nil {
		return nil, err
	}

	msgType := body[0]
	flags := body[1]
	payload := body[2:]

	if flags&FlagCompressed != 0 {
		decompressed, err := cc.decompress(payload)
		if err != nil {
			return nil, err
		}
		payload = decompressed
	}

	return &Message{Type: msgType, Payload: payload}, nil
}

// compress 使用不同的 writer 池，根据数据大小选择 best speed 或 default
func (cc *CompressionCodec) compress(data []byte) ([]byte, error) {
	// 从 buffer 池获取并 reset
	buf := cc.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	// 不要 defer Put 直接，因为我们要在返回前放回池
	// defer cc.bufferPool.Put(buf)

	var w *kpgzip.Writer
	useFast := len(data) <= cc.config.FastThreshold

	if useFast {
		w = cc.writerPoolFast.Get().(*kpgzip.Writer)
	} else {
		w = cc.writerPoolDefault.Get().(*kpgzip.Writer)
	}

	// Reset writer to write into buf
	w.Reset(buf)

	// 写入数据
	if _, err := w.Write(data); err != nil {
		// 将 writer 放回池前尽量 Close
		_ = w.Close()
		if useFast {
			cc.writerPoolFast.Put(w)
		} else {
			cc.writerPoolDefault.Put(w)
		}
		cc.bufferPool.Put(buf)
		return nil, err
	}

	// Close 刷新输出（注意：Close 不会丢弃 writer，使其仍可 Reset 使用）
	if err := w.Close(); err != nil {
		// push back and free buffer
		if useFast {
			cc.writerPoolFast.Put(w)
		} else {
			cc.writerPoolDefault.Put(w)
		}
		cc.bufferPool.Put(buf)
		return nil, err
	}

	// 复制结果到新切片（避免 buf 被复用后数据被篡改）
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())

	// 把 writer 和 buffer 放回池
	if useFast {
		cc.writerPoolFast.Put(w)
	} else {
		cc.writerPoolDefault.Put(w)
	}
	cc.bufferPool.Put(buf)

	return out, nil
}

// decompress 每次创建一个 reader（安全且简单），性能通常足够好。
// 如果你在解压非常频繁且成为瓶颈，可以再做 reader 池化，但要处理 Reset/Close 的细节。
func (cc *CompressionCodec) decompress(data []byte) ([]byte, error) {
	r, err := kpgzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	out, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return out, nil
}
