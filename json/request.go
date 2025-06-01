package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"nyxze/choco-go"
)

// ToJSON calls json.Marshal() to get the JSON encoding of v then calls SetBody()
func MarshalAsJSON(req *choco.Request, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("error marshalling type %T: %s", v, err)
	}
	r := choco.NopCloser(bytes.NewReader(b))
	return req.SetBody(r, choco.ContentTypeAppJSON)
}

func UnmarshalAsJSON() {

}
