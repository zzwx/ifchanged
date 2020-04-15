package ifchanged

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
)

// An implementation of sha256 in one file
// with line #0 being the key and line #1 the value, etc.
// 0  key
// 1  value
// 2  key
// 3  value

type LineTuple struct {
	fileName string
	file     *os.File
}

func NewLineTuple(fileName string) (*LineTuple, error) {
	var l LineTuple
	l.fileName = fileName
	var err error
	l.file, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(fileName), os.ModePerm)
		if err == nil {
			file, err := os.Create(fileName)
			if err != nil {
				return nil, fmt.Errorf("db create file error: %w", err)
			}
			l.file = file
			//defer file.Close()
		} else {
			return nil, err
		}
	}
	return &l, nil
}

func (l *LineTuple) Put(key, value []byte) error {
	l.file.Seek(0, 0)
	lineIndex := 0
	foundKeyLine := -1
	foundValueLine := -1
	newLinesCount := 0
	for {
		i := 0
		line := make([]byte, 0, 64)
		var err error
		by := make([]byte, 1)
	repeat:
		by[0] = 0
		cnt, err := l.file.Read(by)
		if err == io.EOF {
			if len(line) > 0 {
				goto try
			}
			break
		}
		if cnt > 0 {
			if by[0] == '\n' {
				newLinesCount++
				i++
			} else {
				if by[0] != '\r' { // In case it's not a new line feed, we append
					line = append(line, by[0])
				}
				i++
				goto repeat
			}
		}

	try:
		if lineIndex == foundValueLine {
			if bytes.Equal(value, line) {
				// Value equals, we simply break
				return nil
			} else {
				// We take advantage of the fact that values are always the same size
				if len(value) != len(line) {
					return fmt.Errorf("value is of illegal size: %d", len(value))
				} else {
					l.file.Seek(-int64(i), 1)
					l.file.Write(value)
					l.file.Sync()
					return nil
				}
			}
		}

		if lineIndex%2 == 0 && bytes.Equal(key, line) {
			foundKeyLine = lineIndex
			foundValueLine = lineIndex + 1
		}

		lineIndex++
	}

	if foundKeyLine < 0 {
		if newLinesCount%2 != 0 {
			// Keeping key beginning at odd lines
			l.file.Write([]byte{'\n'})
		}
		l.file.Write(key)
		l.file.Write([]byte{'\n'})
		l.file.Write(value)
		l.file.Sync()
	}

	return nil
}

func (l *LineTuple) Has(key []byte) bool {
	l.file.Seek(0, 0)
	scanner := bufio.NewScanner(l.file)
	line := 0
	foundLine := -1
	for scanner.Scan() {
		if foundLine == line {
			// The value can be empty, but we don't consider it as a missing value
			return true
		}
		if line%2 == 0 {
			if bytes.Equal(scanner.Bytes(), key) {
				foundLine = line + 1
			}
		}
		line++
	}
	return false
}

func (l *LineTuple) Get(key []byte) ([]byte, error) {
	l.file.Seek(0, 0)
	scanner := bufio.NewScanner(l.file)
	line := 0
	foundLine := -1
	for scanner.Scan() {
		if foundLine == line {
			return []byte(scanner.Text()), nil
		}
		if line%2 == 0 {
			if bytes.Equal(scanner.Bytes(), key) {
				foundLine = line + 1
			}
		}
		line++
	}
	return []byte{}, nil
}

func (l *LineTuple) Sync() error {
	if l.file == nil {
		return nil
	}
	return l.file.Sync()
}

func (l *LineTuple) Close() error {
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}
