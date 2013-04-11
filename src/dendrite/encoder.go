package dendrite

import (
	"encoding/json"
	"fmt"
)

type Encoder interface {
	Encode(out map[string]Column, ch chan string)
}

type JsonEncoder struct{}
type StatsdEncoder struct{}

func (*JsonEncoder) Encode(out map[string]Column, ch chan string) {
	stripped := make(map[string]interface{})
	for k, v := range out {
		stripped[k] = v.Value
	}
	bytes, err := json.Marshal(stripped)
	if err != nil {
		panic(err)
	}
	ch <- string(bytes)
}

func (*StatsdEncoder) Encode(out map[string]Column, ch chan string) {
	for k, v := range out {
		switch v.Type {
		case Gauge:
			ch <- fmt.Sprintf("%s:%d|g", k, v.Value)
		case Metric:
			ch <- fmt.Sprintf("%s:%d|m", k, v.Value)
		case Counter:
			ch <- fmt.Sprintf("%s:%d|c", k, v.Value)
		}
	}
}
