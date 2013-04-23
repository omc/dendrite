package dendrite

import (
	"fmt"
	"github.com/fizx/logs"
	"encoding/json"
	"github.com/kylelemons/go-gypsy/yaml"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
)

var DefaultPattern = "(?P<line>.*?)\r?\n"

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

type FieldConfig struct {
	Name    string
	Alias   string
	Type    FieldType
	Group   int
	Format  string
	Pattern *regexp.Regexp
}

type SourceConfig struct {
	Glob             string
	Pattern          string
	Fields           []FieldConfig
	Name             string
	OffsetDir        string
	Hostname         string
	MaxBackfillBytes int64
	MaxLineSizeBytes int64
}

type DestinationConfig struct {
	Name string
	Url  *url.URL
}

type Config struct {
	OffsetDir        string
	MaxBackfillBytes int64
	MaxLineSizeBytes int64
	Destinations     []DestinationConfig
	Sources          []SourceConfig
}

func (config *Config) CreateDestinations() Destinations {
	dests := NewDestinations()
	for _, subConfig := range config.Destinations {
		dest, err := NewDestination(subConfig)
		if err != nil {
			logs.Warn("Can't load destination, continuing...: %s", err)
			continue
		}
		dests = append(dests, dest)
	}

	return dests
}

func (config *Config) CreateAllTailGroups(drain chan Record) TailGroups {
	groups := make([]*TailGroup, 0)
	for _, subConfig := range config.Sources {
		groups = append(groups, NewTailGroup(subConfig, drain))
	}
	return groups
}

// Mostly delegate
func NewConfig(configFile string, hostname string) (*Config, error) {
	mapping, err := assembleConfigFiles(configFile)
	if err != nil {
		return nil, err
	}
	return configFromMapping(mapping, hostname)
}

func assembleConfigFiles(configFile string) (map[string]interface{}, error) {
	doc, err := yaml.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	mapping := YamlUnmarshal(doc.Root).(map[string]interface{})

	entries, err := filepath.Glob(path.Dir(configFile) + "/conf.d/*.yaml")
	if err != nil {
		logs.Warn("Can't read relevant conf.d: %s", err)
	} else {
		for _, path := range entries {
			doc, err := yaml.ReadFile(path)
			if err != nil {
				logs.Warn("Can't read relevant conf.d: %s", err)
			} else {
				inner := YamlUnmarshal(doc.Root).(map[string]interface{})
				RecursiveMergeNoConflict(mapping, inner, "")
			}
		}
	}
	return mapping, nil
}

func configFromMapping(mapping map[string]interface{}, hostname string) (*Config, error) {
	b, _ := json.Marshal(mapping)
	logs.Debug("mapping: %s", string(b))
	var err error = nil
	config := new(Config)
	config.Sources = make([]SourceConfig, 0)
	config.Destinations = make([]DestinationConfig, 0)

	global, err := getMap(mapping, "global")
	if err != nil {
		return nil, fmt.Errorf("no global section in the config file")
	}

	config.OffsetDir, err = getString(global, "offset_dir")
	if err != nil {
		logs.Warn("no offset_dir specified")
		config.MaxBackfillBytes = -1
	}

	config.MaxBackfillBytes, err = getInt64(global, "max_backfill_bytes")
	if err != nil {
		logs.Warn("no max_backfill_bytes, continuing with unlimited")
		config.MaxBackfillBytes = -1
	}

	config.MaxLineSizeBytes, err = getInt64(global, "max_linesize_bytes")
	if err != nil {
		logs.Warn("no max_linesize_bytes, continuing with 32768")
		config.MaxLineSizeBytes = 32768
	}

	sources, err := getMap(mapping, "sources")
	if err != nil {
		return nil, fmt.Errorf("no sources section in the config file")
	}

	for name, _ := range sources {
		src, err := getMap(sources, name)
		if err != nil {
			logs.Warn("Invalid source: %s, continuing...", name)
			continue
		}

		var source SourceConfig
		source.Hostname = hostname
		source.Fields = make([]FieldConfig, 0)
		source.OffsetDir = config.OffsetDir
		source.MaxBackfillBytes = config.MaxBackfillBytes
		source.MaxLineSizeBytes = config.MaxLineSizeBytes
		source.Name = name
		source.Glob, err = getString(src, "glob")
		if err != nil {
			return nil, err
		}
		source.Pattern, err = getString(src, "pattern")
		if err != nil {
			source.Pattern = DefaultPattern
		}

		_, err = regexp.Compile(source.Pattern)
		if err != nil {
			logs.Warn("%s is not a valid regexp, continuing... (%s)", source.Pattern, err)
			continue
		}

		fields, err := getMap(src, "fields")
		for name, _ := range fields {
			fld, err := getMap(fields, name)
			if err != nil {
				logs.Warn("%s is not a map, continuing... (%s)", name, err)
				continue
			}

			var field FieldConfig
			field.Alias = name

			field.Name, err = getString(fld, "name")
			if err != nil {
				field.Name = field.Alias
			}

			field.Group, err = getInt(fld, "group")

			s, err := getString(fld, "type")
			if err != nil {
				field.Type = String
			} else {
				field.Type, err = parseField(s)
				if err != nil {
					logs.Warn("Invalid field type: %s, continuing... (error was %s)", s, err)
					continue
				}
			}
			logs.Info("found type %s", field.Type)

			field.Format, err = getString(fld, "format")

			s, err = getString(fld, "pattern")
			field.Pattern, err = regexp.Compile(s)
			if err != nil {
				logs.Warn("Invalid regex: %s, continuing... (error was %s)", s, err)
				continue
			}
			source.Fields = append(source.Fields, field)
		}
		config.Sources = append(config.Sources, source)
	}

	destinations, err := getMap(mapping, "destinations")
	if err != nil {
		return nil, fmt.Errorf("no destinations section in the config file")
	}

	for name, _ := range destinations {
		var dest DestinationConfig
		urlString, err := getString(destinations, name)
		u, err := url.Parse(urlString)
		if err != nil {
			logs.Warn("Invalid URL: %s, continuing... (error was %s)", urlString, err)
			continue
		}
		logs.Info("Found destination: %s", urlString)
		dest.Name = name
		dest.Url = u
		config.Destinations = append(config.Destinations, dest)
	}

	return config, nil
}
