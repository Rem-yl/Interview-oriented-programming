package protocol

import (
	"bufio"
	"errors"
	"go-redis/logger"
	"io"
	"strconv"
	"strings"
)

type Parser struct {
	reader *bufio.Reader
}

func NewParser(reader io.Reader) *Parser {
	// 如果已经是 bufio.Reader，直接使用；否则包装它
	bufReader, ok := reader.(*bufio.Reader)
	if !ok {
		bufReader = bufio.NewReader(reader)
	}
	return &Parser{
		reader: bufReader,
	}
}

func getFullLine(reader *bufio.Reader) ([]byte, error) {
	var fullLine []byte
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			return nil, err
		}
		fullLine = append(fullLine, line...)
		if !isPrefix {
			break
		}
	}

	return fullLine, nil
}

func (p *Parser) parseSimpleString() (*Value, error) {
	fullLine, err := getFullLine(p.reader)
	if err != nil {
		return nil, err
	}

	logger.Debug("parseSimpleString fullLine: ", string(fullLine))

	value := &Value{
		Type:   StringType,
		Str:    string(fullLine),
		IsNull: false,
	}
	return value, nil
}

func (p *Parser) parseErr() (*Value, error) {
	fullLine, err := getFullLine(p.reader)
	if err != nil {
		return nil, err
	}

	logger.Debug("parseErr fullLine: ", string(fullLine))

	value := &Value{
		Type:   ErrorType,
		Str:    string(fullLine),
		IsNull: false,
	}

	return value, nil
}

func (p *Parser) parseInt() (*Value, error) {
	fullLine, err := getFullLine(p.reader)
	if err != nil {
		return nil, err
	}

	logger.Debug("parseInt fullLine: ", string(fullLine))

	strVal := string(fullLine)
	parts := strings.Fields(strVal)
	if len(parts) != 1 {
		return nil, errors.New("multi int please use array")
	}

	intVal, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, errors.New("invalid int value: " + parts[0])
	}

	value := &Value{
		Type:   IntType,
		Int:    intVal,
		IsNull: false,
	}

	return value, nil
}

func (p *Parser) parseBulkString() (*Value, error) {
	lenByte, err := getFullLine(p.reader)
	if err != nil {
		return nil, err
	}

	length, err := strconv.Atoi(string(lenByte))
	if err != nil {
		return nil, err
	}

	if length == -1 {
		return &Value{
			Type:   NullType,
			IsNull: true,
		}, nil
	}

	logger.Debug("length: ", length)

	buf := make([]byte, length)
	_, err = io.ReadFull(p.reader, buf)
	if err != nil {
		return nil, err
	}

	// 读取并丢弃 \r\n
	_, err = p.reader.ReadByte() // \r
	if err != nil {
		return nil, err
	}
	_, err = p.reader.ReadByte() // \n
	if err != nil {
		return nil, err
	}

	value := &Value{
		Type:   StringType,
		Str:    string(buf),
		IsNull: false,
	}

	return value, nil
}

func (p *Parser) parseArray() (*Value, error) {
	// 解析长度
	lenByte, err := getFullLine(p.reader)
	if err != nil {
		return nil, err
	}

	length, err := strconv.Atoi(string(lenByte))
	if err != nil {
		return nil, err
	}

	if length == -1 {
		return &Value{Type: NullType, IsNull: true}, nil
	}
	if length == 0 {
		return &Value{Type: ArrayType, Array: []Value{}}, nil
	}

	array := make([]Value, length)
	for i := range length {
		value, err := p.Parse()
		if err != nil {
			return nil, err
		}
		array[i] = *value
	}

	return &Value{Type: ArrayType, Array: array}, nil
}

func (p *Parser) Parse() (*Value, error) {
	var value *Value
	var err error

	b, err := p.reader.ReadByte() // 前进了一个字符
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}

	switch b {
	case '+':
		value, err = p.parseSimpleString()
	case '-':
		value, err = p.parseErr()
	case ':':
		value, err = p.parseInt()
	case '$':
		value, err = p.parseBulkString()
	case '*':
		value, err = p.parseArray()
	default:
		value, err = nil, errors.New("unknown RESP type: "+string(b))
	}

	return value, err
}
