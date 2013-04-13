package dendrite

import (
	"github.com/fizx/logs"
	"github.com/kylelemons/go-gypsy/yaml"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
)

type FieldType int

const (
	String = iota
	Tokens
	Integer
	Gauge
	Metric
	Counter
	Timestamp
)

type FieldSpec struct {
	Name    string
	Alias   string
	Type    FieldType
	Group   int
	Format  string
	Pattern *regexp.Regexp
}

type SourceConfig struct {
	Glob      string
	Pattern   string
	Fields    []FieldSpec
	Name      string
	OffsetDir string
}

type DestinationConfig struct {
	Url *url.URL
}

type Config struct {
	OffsetDir    string
	Destinations []DestinationConfig
	Sources      []SourceConfig
}

func (config *Config) CreateReadWriter() io.ReadWriter {
	return nil
}

func (config *Config) CreateAllGroups(rw io.ReadWriter) TailGroups {
	groups := make([]*TailGroup, 0)
	// for _, cg := range config.Groups {
	//     // groups = append(groups, NewTailGroup(cg, out))
	// }
	return groups
}

func NewConfig(configFile string) (*Config, error) {
	doc, err := yaml.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	config := new(Config)
	config.Sources = make([]SourceConfig, 0)
	config.Destinations = make([]DestinationConfig, 0)

	config.Populate(doc)

	entries, err := filepath.Glob(path.Dir(configFile) + "/conf.d/*.yaml")
	if err != nil {
		logs.Warn("Can't read relevant conf.d: %s", err)
	} else {
		for _, path := range entries {
			doc, err := yaml.ReadFile(path)
			if err != nil {
				logs.Warn("Can't read relevant conf.d: %s", err)
			} else {
				config.Populate(doc)
			}
		}
	}
	return config, nil
}

func (config *Config) Populate(doc *yaml.File) {
	// root := doc.Root.(yaml.Map)
	off, err := doc.Get(".global.offset_dir")
	if err == nil {
		config.OffsetDir = off
	}

	// groups := root.Key("groups")
	// for name, group := range groups.(yaml.Map) {
	//  config.AddGroup(name, group)
	// }

	// p, err := doc.Get(".uri")
	// if err != nil {
	//   p = "tcp"
	// }
	// config.Url, err = url.Parse(p)
	//   tmp, _ = yaml.Child(fieldDetails, ".format")
	// if tmp != nil {
	//  fieldSpec.Format = tmp.(yaml.Scalar).String()
	// }
	// config.Groups = make([]ConfigGroup, 0)
	// 
	// root := doc.Root.(yaml.Map)
	// 
	// p, err := doc.Get(".uri")
	// if err != nil {
	//  p = "tcp"
	// }
	// config.Url, err = url.Parse(p)
	// 
	// groups := root.Key("groups")
	// for name, group := range groups.(yaml.Map) {
	//  config.AddGroup(name, group)
	// }
}

func (config *Config) AddGroup(name string, group yaml.Node) {
	// logs.Info("Adding group: %s", name)
	// groupMap := group.(yaml.Map)
	// var groupStruct ConfigGroup
	// groupStruct.Name = name
	// groupStruct.Glob = groupMap.Key("glob").(yaml.Scalar).String()
	// groupStruct.Pattern = groupMap.Key("pattern").(yaml.Scalar).String()
	// groupStruct.Fields = make([]FieldSpec, 0)
	// groupStruct.OffsetDir = config.OffsetDir
	// groupStruct.Encoder = config.Encoder
	// 
	// for alias, v := range groupMap.Key("fields").(yaml.Map) {
	//  var fieldDetails = v.(yaml.Map)
	//  var fieldSpec FieldSpec
	//  fieldSpec.Alias = alias
	//  fieldSpec.Name = alias
	// 
	//  tmp, _ := yaml.Child(fieldDetails, ".name")
	//  if tmp != nil {
	//    fieldSpec.Name = tmp.(yaml.Scalar).String()
	//  }
	// 
	//  fieldSpec.Group = -1
	//  tmp, _ = yaml.Child(fieldDetails, ".group")
	//  if tmp != nil {
	//    fieldSpec.Name = ""
	//    i, err := strconv.ParseInt(tmp.(yaml.Scalar).String(), 10, 64)
	//    if err != nil {
	//      logs.Error("error in parsing int", err)
	//    }
	// 
	//    fieldSpec.Group = int(i)
	//  }
	// 
	//  tmp, _ = yaml.Child(fieldDetails, ".pattern")
	//  if tmp != nil {
	//    p, err := regexp.Compile(tmp.(yaml.Scalar).String())
	//    if err != nil {
	//      logs.Error("error in compiling regexp", err)
	//    } else {
	//      fieldSpec.Pattern = p
	//    }
	//  }
	// 
	//  tmp, _ = yaml.Child(fieldDetails, ".format")
	//  if tmp != nil {
	//    fieldSpec.Format = tmp.(yaml.Scalar).String()
	//  }
	// 
	//  tmp, _ = yaml.Child(fieldDetails, ".type")
	//  if tmp == nil {
	//    fieldSpec.Type = String
	//  } else {
	//    switch tmp.(yaml.Scalar).String() {
	//    case "int":
	//      fieldSpec.Type = Integer
	//    case "gauge":
	//      fieldSpec.Type = Gauge
	//    case "metric":
	//      fieldSpec.Type = Metric
	//    case "counter":
	//      fieldSpec.Type = Counter
	//    case "string":
	//      fieldSpec.Type = String
	//    case "tokenized":
	//      fieldSpec.Type = Tokens
	//    case "timestamp", "date":
	//      fieldSpec.Type = Timestamp
	//    default:
	//      logs.Error("Can't recognize field type")
	//      panic(nil)
	//    }
	//  }
	// 
	//  groupStruct.Fields = append(groupStruct.Fields, fieldSpec)
	// }
	// config.Groups = append(config.Groups, groupStruct)
}
