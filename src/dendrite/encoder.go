package dendrite

import (
	"encoding/json"
	"fmt"
	"github.com/fizx/logs"
	"io"
	"net/url"
	"strings"
)

type Encoder interface {
	Encode(out map[string]Column, writer io.Writer)
}

type JsonEncoder struct{}
type StatsdEncoder struct{}
type RawStringEncoder struct{}
type LibratoEncoder struct{}

func NewEncoder(u *url.URL) (Encoder, error) {
	a := strings.Split(u.Scheme, "+")
	switch a[len(a)-1] {
	case "json":
		return new(JsonEncoder), nil
	case "statsd":
		return new(StatsdEncoder), nil
	case "librato":
		return new(LibratoEncoder), nil
	}
	return new(RawStringEncoder), nil
}

func (*RawStringEncoder) Encode(out map[string]Column, writer io.Writer) {
	for _, v := range out {
		if v.Type == String {
			writer.Write([]byte(v.Value.(string)))
		}
	}
}

func (*JsonEncoder) Encode(out map[string]Column, writer io.Writer) {
	stripped := make(map[string]interface{})
	for k, v := range out {
		stripped[k] = v.Value
	}
	bytes, err := json.Marshal(stripped)
	if err != nil {
		panic(err)
	}
	bytes = append(bytes, '\n')
	writer.Write(bytes)
}

func (*LibratoEncoder) Encode(out map[string]Column, writer io.Writer) {
	m := make(map[string]interface{})
	m["source"] = out["_hostname"].Value
	for k, v := range out {
		switch v.Type {
		case Gauge:
			m["name"] = k
			m["value"] = v.Value
			bytes, err := json.Marshal(m)
			if err != nil {
				panic(err)
			}
			writer.Write(bytes)
		}
	}
}

func (*StatsdEncoder) Encode(out map[string]Column, writer io.Writer) {
	for k, v := range out {
		switch v.Type {
		case Gauge:
			writer.Write([]byte(fmt.Sprintf("%s:%d|g", k, v.Value)))
		case Metric:
			writer.Write([]byte(fmt.Sprintf("%s:%d|m", k, v.Value)))
		case Counter:
			writer.Write([]byte(fmt.Sprintf("%s:%d|c", k, v.Value)))
		}
	}
}
