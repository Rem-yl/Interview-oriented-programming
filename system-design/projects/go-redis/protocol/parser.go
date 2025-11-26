package protocol

import (
	"bufio"
	"io"
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
