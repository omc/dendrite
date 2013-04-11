package dendrite

import (
	"github.com/fizx/logs"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"
)

type Column struct {
	Type  FieldType
	Value interface{}
}

type Parser interface {
	Consume(bytes []byte, counter *int64)
}

type RegexpParser struct {
	group    string
	compiled *regexp.Regexp
	output   chan string
	buffer   []byte
	fields   []FieldSpec
	file     string
	encoder  Encoder
}

func NewRegexpParser(group string, file string, output chan string, pattern string, fields []FieldSpec, encoder Encoder) Parser {
	parser := new(RegexpParser)
	parser.encoder = encoder
	parser.file = file
	parser.group = group
	parser.output = output
	parser.buffer = make([]byte, 0)
	re, err := regexp.Compile(pattern)
	if err != nil {
		logs.Debug("Cannot parse regexp:", err)
	} else {
		parser.compiled = re
		for i, name := range re.SubexpNames() {
			if name != "" {
				found := false
				for n, spec := range fields {
					if spec.Name == "" {
						spec.Name = spec.Alias
					}
					if name == spec.Name {
						found = true
						fields[n].Group = i
						logs.Debug("setting group alias: %s, name: %s, group: %d", spec.Alias, spec.Name, spec.Group)
					}
				}
				if !found {
					var spec FieldSpec
					spec.Group = i
					spec.Alias = name
					spec.Type = String
					fields = append(fields, spec)
				}
			}
		}
	}
	parser.fields = fields
	for _, f := range parser.fields {
		logs.Debug("p.f: alias: %s, name: %s, group: %d, type: %d", f.Alias, f.Name, f.Group, f.Type)
	}
	return parser
}

func (parser *RegexpParser) Consume(bytes []byte, counter *int64) {
	parser.buffer = append(parser.buffer, bytes...)
	logs.Debug("consuming %d bytes", len(bytes))
	for {
		m := parser.compiled.FindSubmatchIndex(parser.buffer)
		if m == nil {
			return
		}

		out := make(map[string]Column)
		out["_offset"] = Column{Integer, atomic.LoadInt64(counter)}
		out["_file"] = Column{String, parser.file}
		out["_time"] = Column{Timestamp, StandardTimeProvider.Now().Unix()}
		out["_group"] = Column{String, parser.group}
		for _, spec := range parser.fields {
			g := spec.Group
			if g < 0 || g > len(m)/2 {
				logs.Error("spec group out of range: alias: %s, name: %s, g: %d", spec.Alias, spec.Name, g)
				panic(-1)
			}
			s := string(parser.buffer[m[g*2]:m[g*2+1]])
			switch spec.Type {
			case Timestamp:
				t, err := time.Parse(spec.Format, s)
				if err != nil {
					logs.Warn("date parse error: %s", err)
				} else {
					if t.Year() == 0 {
						now := StandardTimeProvider.Now()
						adjusted := time.Date(now.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
						if adjusted.After(now) {
							adjusted = time.Date(now.Year()-1, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
						}
						t = adjusted
					}
					out[spec.Alias] = Column{Timestamp, t.Unix()}
				}
			case String:
				out[spec.Alias] = Column{String, s}
			case Gauge, Metric, Integer, Counter:
				n, err := strconv.ParseInt(s, 10, 64)
				if err == nil {
					out[spec.Alias] = Column{spec.Type, n}
				}
			case Tokens:
				out[spec.Alias] = Column{Tokens, spec.Pattern.FindAllString(s, -1)}
			default:
				panic(nil)
			}
		}
		o := parser.output
		parser.encoder.Encode(out, o)
		atomic.AddInt64(counter, int64(m[1]))

		parser.buffer = parser.buffer[m[1]:]
	}
}
