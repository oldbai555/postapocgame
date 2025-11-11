package network

import (
	"sync"
)

// BufferPool 字节缓冲池（使用分级策略）
type BufferPool struct {
	pools []*sync.Pool // 不同大小的池
	sizes []int        // 对应的大小
}

var defaultBufferPool *BufferPool

// NewBufferPool 创建缓冲池
func NewBufferPool(sizes []int) *BufferPool {
	bp := &BufferPool{
		sizes: sizes,
		pools: make([]*sync.Pool, len(sizes)),
	}

	for i, size := range sizes {
		sz := size // 捕获循环变量
		bp.pools[i] = &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, sz)
				return &buf
			},
		}
	}

	return bp
}

// Get 获取缓冲区（返回至少 size 大小的 buffer）
func (bp *BufferPool) Get(size int) []byte {
	// 找到最小的满足 size 的池
	for i, poolSize := range bp.sizes {
		if size <= poolSize {
			bufPtr := bp.pools[i].Get().(*[]byte)
			buf := *bufPtr
			return buf[:size] // 返回精确大小
		}
	}

	// 超过最大池大小，直接分配
	return make([]byte, size)
}

// Put 归还缓冲区
// ✅ 改进：使用范围匹配而非严格相等
func (bp *BufferPool) Put(buf []byte) {
	if buf == nil {
		return
	}

	capacity := cap(buf)

	// ✅ 改进：找到最接近的池大小
	for i, poolSize := range bp.sizes {
		// 如果容量在合理范围内（50%-100%），可以放入该池
		if capacity >= poolSize/2 && capacity <= poolSize {
			fullBuf := buf[:capacity]
			bp.pools[i].Put(&fullBuf)
			return
		}
	}

	// 不在池范围内，让 GC 回收
}

// GetBuffer 全局获取缓冲区
func GetBuffer(size int) []byte {
	return defaultBufferPool.Get(size)
}

// PutBuffer 全局归还缓冲区
func PutBuffer(buf []byte) {
	defaultBufferPool.Put(buf)
}

// BufferWriter 带缓冲的写入器
type BufferWriter struct {
	buf []byte
	pos int
}

// NewBufferWriter 创建写入器
func NewBufferWriter(size int) *BufferWriter {
	return &BufferWriter{
		buf: GetBuffer(size),
		pos: 0,
	}
}

// Write 写入数据
func (bw *BufferWriter) Write(p []byte) (int, error) {
	n := copy(bw.buf[bw.pos:], p)
	bw.pos += n
	return n, nil
}

// Bytes 返回已写入的数据（不复制）
func (bw *BufferWriter) Bytes() []byte {
	return bw.buf[:bw.pos]
}

// Reset 重置写入器
func (bw *BufferWriter) Reset() {
	bw.pos = 0
}

// Release 释放缓冲区
func (bw *BufferWriter) Release() {
	if bw.buf != nil {
		PutBuffer(bw.buf)
		bw.buf = nil
	}
}

// MessagePool 消息对象池（优化版）
var messagePool = sync.Pool{
	New: func() interface{} {
		return &Message{}
	},
}

// GetMessage 获取消息对象
func GetMessage() *Message {
	return messagePool.Get().(*Message)
}

// PutMessage 归还消息对象
func PutMessage(msg *Message) {
	if msg == nil {
		return
	}
	msg.Reset()
	messagePool.Put(msg)
}

// RPCRequestPool RPC请求池
var rpcRequestPool = sync.Pool{
	New: func() interface{} {
		return &RPCRequest{}
	},
}

// GetRPCRequest 获取RPC请求
func GetRPCRequest() *RPCRequest {
	return rpcRequestPool.Get().(*RPCRequest)
}

// PutRPCRequest 归还RPC请求
func PutRPCRequest(req *RPCRequest) {
	if req == nil {
		return
	}
	req.Reset()
	rpcRequestPool.Put(req)
}

// RPCResponsePool RPC响应池
var rpcResponsePool = sync.Pool{
	New: func() interface{} {
		return &RPCResponse{}
	},
}

// GetRPCResponse 获取RPC响应
func GetRPCResponse() *RPCResponse {
	return rpcResponsePool.Get().(*RPCResponse)
}

// PutRPCResponse 归还RPC响应
func PutRPCResponse(resp *RPCResponse) {
	if resp == nil {
		return
	}
	resp.Reset()
	rpcResponsePool.Put(resp)
}

var forwardMessagePool = sync.Pool{
	New: func() interface{} {
		return &ForwardMessage{}
	},
}

func GetForwardMessage() *ForwardMessage {
	return forwardMessagePool.Get().(*ForwardMessage)
}

func PutForwardMessage(msg *ForwardMessage) {
	if msg == nil {
		return
	}
	msg.Reset()
	forwardMessagePool.Put(msg)
}

var clientMessagePool = sync.Pool{
	New: func() interface{} {
		return &ClientMessage{}
	},
}

func GetClientMessage() *ClientMessage {
	return clientMessagePool.Get().(*ClientMessage)
}

func PutClientMessage(msg *ClientMessage) {
	if msg == nil {
		return
	}
	msg.Reset()
	clientMessagePool.Put(msg)
}

func init() {
	// 预定义常用大小: 256B, 1KB, 4KB, 16KB, 64KB, 256KB, 1MB
	sizes := []int{256, 1024, 4096, 16384, 65536, 262144, 1048576}
	defaultBufferPool = NewBufferPool(sizes)
}
