package sse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"nyxze/choco-go/seqio"
	"reflect"
	"strings"
)

type SseIter[T any] struct {
	rc       io.ReadCloser
	reader   *bufio.Reader
	prefix   map[string]struct{}
	endToken string
	current  T
	err      error
}

func (s *SseIter[T]) Err() error {
	return s.err
}
func (s *SseIter[T]) Curr() T {
	return s.current
}

func (s *SseIter[T]) Next() bool {
	if s.err != nil {
		return false
	}
	fields := make(map[string][]byte)

	for {
		txt, err := line(s.reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.err = err
			break
		}

		// Dispatch event on evet boundary
		if len(txt) == 0 {
			break
		}

		// Split line into name and value
		name, value, _ := bytes.Cut(txt, []byte(":"))
		if len(value) > 0 && value[0] == ' ' {
			value = value[1:]
		}

		// If prefix map is set, check it; otherwise allow all fields
		if s.prefix != nil {
			if _, ok := s.prefix[string(name)]; !ok {
				continue
			}
		}

		// Check for end token after passing the prefix filter
		if s.endToken != "" && bytes.Equal(value, []byte(s.endToken)) {
			// hit end token - stop iteration immediately
			break
		}

		// Write
		if prev, ok := fields[string(name)]; ok {
			fields[string(name)] = append(prev, value...)
		} else {
			fields[string(name)] = value
		}
	}
	// Nothing written
	if len(fields) == 0 {
		return false
	}
	s.current = applyFieldsToStruct[T](fields)
	return true
}

type SSEOptions func(s *SseIter[any])

func NewSSEIter[T any](r io.ReadCloser, endToken string) seqio.Iterator[T] {

	prefix := make(map[string]struct{})

	// Get type of T
	var t T
	typ := reflect.TypeOf(t)

	// If T is a pointer, get the element type
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// Iterate over fields
	for i := range typ.NumField() {
		field := typ.Field(i)
		name := field.Name

		// Optional: check for `sse` tag override
		if tag := field.Tag.Get("sse"); tag != "" {
			name = tag
		}

		prefix[name] = struct{}{}
	}

	return &SseIter[T]{
		endToken: endToken,
		rc:       r,
		prefix:   prefix,
		reader:   bufio.NewReader(r),
	}
}

func applyFieldsToStruct[T any](fields map[string][]byte) T {
	var result T
	val := reflect.ValueOf(&result).Elem()
	typ := val.Type()

	for k, v := range fields {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldName := field.Name

			// Optional: support sse:"fieldname" tag
			if tag := field.Tag.Get("sse"); tag != "" {
				fieldName = tag
			}

			if strings.EqualFold(fieldName, k) {
				fv := val.Field(i)
				if !fv.CanSet() {
					continue
				}

				switch fv.Kind() {
				case reflect.String:
					fv.SetString(string(v))
				case reflect.Slice:
					if fv.Type().Elem().Kind() == reflect.Uint8 { // []byte
						fv.SetBytes(v)
					}
				default:
					fmt.Println("unsupported field type: only string and []byte are allowed. Skipping fields")
					break
				}
			}
		}
	}
	return result
}

func line(r *bufio.Reader) ([]byte, error) {
	var overflow bytes.Buffer

	// To prevent infinite loops, the failsafe stops when a line is
	// 100 times longer than the [io.Reader] default buffer size,
	// or after 20 failed attempts to find an end of line.
	for f := 0; f < 100; f++ {
		part, isPrefix, err := r.ReadLine()
		if err != nil {
			return nil, err
		}

		// Happy case, the line fits in the default buffer.
		if !isPrefix && overflow.Len() == 0 {
			return part, nil
		}

		// Overflow case, append to the buffer.
		if isPrefix || overflow.Len() > 0 {
			n, err := overflow.Write(part)
			if err != nil {
				return nil, err
			}

			// Didn't find an end of line, heavily increment the failsafe.
			if n != r.Size() {
				f += 5
			}
		}

		if !isPrefix {
			return overflow.Bytes(), nil
		}
	}

	return nil, fmt.Errorf("ssestream: too many attempts to read a line")
}
