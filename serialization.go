package choco

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type nopCloser struct {
	io.ReadSeeker
}

func MarshalAsJSON(req *Request, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("error marshalling type %T: %s", v, err)
	}
	return req.SetBody(NopCloser(bytes.NewReader(b)), ContentTypeAppJSON)
}

// NopCloser returns a ReadSeekCloser with a no-op close method wrapping the provided io.ReadSeeker.
func NopCloser(rs io.ReadSeeker) io.ReadSeekCloser {
	return nopCloser{rs}
}

func (nopCloser) Close() error {
	return nil
}
