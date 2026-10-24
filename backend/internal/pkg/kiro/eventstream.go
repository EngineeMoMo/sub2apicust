package kiro

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

const maxFrameSize = 4 << 20
const MaxResponseSize = 32 << 20

type Event struct {
	Type    string
	Payload []byte
}

func ReadEvents(ctx context.Context, reader io.Reader, consume func(Event) error) error {
	total := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		var prelude [12]byte
		readCount, err := io.ReadFull(reader, prelude[:])
		if err == io.EOF && readCount == 0 {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Kiro EventStream prelude: %w", err)
		}
		if crc32.ChecksumIEEE(prelude[:8]) != binary.BigEndian.Uint32(prelude[8:]) {
			return fmt.Errorf("Kiro EventStream prelude CRC 无效")
		}
		length := int(binary.BigEndian.Uint32(prelude[:4]))
		headerLength := int(binary.BigEndian.Uint32(prelude[4:8]))
		if length < 16 || length > maxFrameSize || headerLength > length-16 {
			return fmt.Errorf("Kiro EventStream 帧长度无效")
		}
		total += length
		if total > MaxResponseSize {
			return fmt.Errorf("Kiro 响应超过大小上限")
		}
		frame := make([]byte, length)
		copy(frame, prelude[:])
		if _, err := io.ReadFull(reader, frame[12:]); err != nil {
			return fmt.Errorf("Kiro EventStream 截断: %w", err)
		}
		if crc32.ChecksumIEEE(frame[:length-4]) != binary.BigEndian.Uint32(frame[length-4:]) {
			return fmt.Errorf("Kiro EventStream message CRC 无效")
		}
		headers, err := decodeHeaders(frame[12 : 12+headerLength])
		if err != nil {
			return err
		}
		if headers[":message-type"] == "exception" || headers[":message-type"] == "error" {
			return fmt.Errorf("Kiro 上游事件异常: %s", headers[":exception-type"]+headers[":error-code"])
		}
		eventType := headers[":event-type"]
		if eventType == "" {
			return fmt.Errorf("Kiro EventStream 缺少事件类型")
		}
		if err := consume(Event{Type: eventType, Payload: frame[12+headerLength : length-4]}); err != nil {
			return err
		}
	}
}

func decodeHeaders(input []byte) (map[string]string, error) {
	values := make(map[string]string)
	for len(input) > 0 {
		nameLength := int(input[0])
		input = input[1:]
		if nameLength == 0 || len(input) < nameLength+1 {
			return nil, fmt.Errorf("Kiro EventStream 头截断")
		}
		name := string(input[:nameLength])
		valueType := input[nameLength]
		input = input[nameLength+1:]
		size := 0
		switch valueType {
		case 0, 1:
			size = 0
		case 2:
			size = 1
		case 3:
			size = 2
		case 4:
			size = 4
		case 5, 8:
			size = 8
		case 9:
			size = 16
		case 6, 7:
			if len(input) < 2 {
				return nil, fmt.Errorf("Kiro EventStream 头长度截断")
			}
			size = int(binary.BigEndian.Uint16(input[:2]))
			input = input[2:]
		default:
			return nil, fmt.Errorf("Kiro EventStream 未知头类型")
		}
		if len(input) < size {
			return nil, fmt.Errorf("Kiro EventStream 头值截断")
		}
		if valueType == 7 {
			if _, exists := values[name]; exists {
				return nil, fmt.Errorf("Kiro EventStream 重复头")
			}
			values[name] = string(input[:size])
		}
		input = input[size:]
	}
	return values, nil
}
