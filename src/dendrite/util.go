package dendrite

import (
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"io"
	"os"
	"os/exec"
	"strconv"
)

func bash(str string) {
	run("bash", "-c", str)
}

func run(str ...string) {
	cmd := exec.Command(str[0], str[1:]...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	cmd.Start()
	cmd.Wait()
}

func Unescape(in string) string {
	bytes := []byte(in)
	out := make([]byte, 0, len(bytes)-2)
	for i := 1; i < len(bytes)-1; i++ {
		c := bytes[i]
		if c == '\\' {
			i++
			c = bytes[i]
		}
		out = append(out, c)
	}
	return string(out)
}

func YamlUnmarshal(node yaml.Node) interface{} {
	switch node := node.(type) {
	case yaml.Map:
		out := make(map[string]interface{})
		for k, v := range node {
			out[k] = YamlUnmarshal(v)
		}
		return out
	case yaml.List:
		out := make([]interface{}, 0)
		for _, v := range node {
			out = append(out, YamlUnmarshal(v))
		}
		return out
	case yaml.Scalar:
		if len(node) > 0 && node[0] == '"' {
			return Unescape(string(node))
		} else {
			return node
		}
	default:
	}
	return nil
}

type anyReader struct {
	readers []io.Reader
}

func NewAnyReader(r []io.Reader) io.Reader {
	return &anyReader{r}
}

func (any *anyReader) Read(buf []byte) (int, error) {
	for _, r := range any.readers {
		i, e := r.Read(buf)
		if e == io.EOF {
			continue
		} else {
			return i, e
		}
	}
	return 0, io.EOF
}

func RecursiveMergeNoConflict(a map[string]interface{}, b map[string]interface{}, path string) error {
	for k, v := range b {
		old, found := a[k]
		if found {
			switch old := old.(type) {
			case map[string]interface{}:
				switch v := v.(type) {
				case map[string]interface{}:
					err := RecursiveMergeNoConflict(old, v, path+"/"+k)
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("Overwriting map with scalar value at %s/%s", path, k)
				}
			default:
				return fmt.Errorf("Duplicate value at %s/%s", path, k)
			}
		} else {
			a[k] = b[k]
		}
	}
	return nil
}

func parseFieldType(str string) (FieldType, error) {
	switch str {
	case "int":
		return Integer, nil
	case "double":
		return Double, nil
	case "string", "":
		return String, nil
	case "timestamp", "date":
		return Timestamp, nil
	}
	return -1, fmt.Errorf("Can't recognize field type: %s", str)
}

func parseFieldTreatment(str string) (FieldTreatment, error) {
	switch str {
	case "simple", "":
		return Simple, nil
	case "gauge":
		return Gauge, nil
	case "metric":
		return Metric, nil
	case "counter":
		return Counter, nil
	case "tokenized":
		return Tokens, nil
	}
	return -1, fmt.Errorf("Can't recognize field treatment: %s", str)
}

func getMap(mapping map[string]interface{}, key string) (map[string]interface{}, error) {
	val, ok := mapping[key]
	if ok {
		switch val := val.(type) {
		case map[string]interface{}:
			return val, nil
		default:
			return nil, fmt.Errorf("key is not a map: %s", key)
		}
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

func getString(mapping map[string]interface{}, key string) (string, error) {
	val, ok := mapping[key]
	if ok {
		return fmt.Sprintf("%s", val), nil
	}
	return "", fmt.Errorf("key not found: %s", key)
}

func getInt64(mapping map[string]interface{}, key string) (int64, error) {
	val, err := getString(mapping, key)
	if err != nil {
		return -1, err
	}
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return -1, err
	}
	return i, nil
}

func getInt(mapping map[string]interface{}, key string) (int, error) {
	a, b := getInt64(mapping, key)
	return int(a), b
}
