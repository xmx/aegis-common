package muxproto

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"math"
)

// ReadJSON 读取 JSON 认证报文。
//
// 报文格式：
//   - 前 2 字节为大端序 uint16，表示后续 JSON 数据的字节长度
//   - 随后紧跟长度为 n 的 JSON 数据
//
// 约束与约定：
//   - JSON 数据长度大于 0 小于等于 math.MaxUint16(64K)
//   - 当长度字段为 0 时，视为对端无更多数据，返回 io.EOF
//
// 错误说明：
//   - 读取长度字段或数据体失败时，返回底层 I/O 错误
//   - 实际读取字节数不足时，返回 io.ErrShortBuffer
//   - JSON 解析失败时，返回 json.Unmarshal 的错误
func ReadJSON(r io.Reader, v any) error {
	size := make([]byte, 2)
	if n, err := r.Read(size); err != nil {
		return err
	} else if n != 2 {
		return io.ErrShortBuffer
	}

	n := binary.BigEndian.Uint16(size)
	if n == 0 {
		return io.EOF
	}

	raw := make([]byte, n)
	num, err := io.ReadFull(r, raw)
	if err != nil {
		return err
	} else if num != int(n) {
		return io.ErrShortBuffer
	}

	return json.Unmarshal(raw, v)
}

// WriteJSON 写入报文。
//
// 报文格式：
//   - 前 2 字节为大端序 uint16，表示后续 JSON 数据的字节长度
//   - 随后紧跟 JSON 数据本身
//
// 约束与约定：
//   - JSON 数据最大长度为 math.MaxUint16（64KB）
//   - 当 JSON 编码后的长度超过上限时，返回错误
//
// 错误说明：
//   - JSON 序列化失败时，返回 json.Marshal 的错误
//   - 写入长度前缀或数据体失败时，返回底层 I/O 错误
func WriteJSON(w io.Writer, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}

	n := len(raw)
	if n > math.MaxUint16 {
		return io.ErrUnexpectedEOF
	}

	size := make([]byte, 2)
	binary.BigEndian.PutUint16(size, uint16(n))
	if _, err = w.Write(size); err == nil {
		_, err = w.Write(raw)
	}

	return err
}
