package dendrite

import (
	"github.com/fizx/logs"
	"os"
	"path/filepath"
	"time"
)

type TailGroups []*TailGroup

type TailGroup struct {
	Glob      string
	Pattern   string
	OffsetDir string
	Name      string
	Hostname  string
	Tails     map[string]*Tail

	output      chan Record
	fields      []FieldConfig
	maxBackfill int64
	maxLineSize int64
}

func (groups *TailGroups) Loop() {
	for i := 0; true; i++ {
		if i%60 == 0 {
			groups.Refresh()
		}
		groups.Poll()
		time.Sleep(1 * time.Second)
	}
}

func (groups *TailGroups) Poll() {
	for _, g := range *groups {
		g.Poll()
	}
}

func (groups *TailGroups) Refresh() {
	for _, g := range *groups {
		g.Refresh()
	}
}

func NewTailGroup(config SourceConfig, output chan Record) *TailGroup {
	group := new(TailGroup)
	group.Hostname = config.Hostname
	group.output = output
	group.Name = config.Name
	group.Glob = config.Glob
	group.Pattern = config.Pattern
	group.OffsetDir = config.OffsetDir
	group.Tails = make(map[string]*Tail)
	group.fields = config.Fields
	group.maxLineSize = config.MaxLineSizeBytes
	group.maxBackfill = config.MaxBackfillBytes
	group.Refresh()
	return group
}

func (group *TailGroup) activate(match string) {
	tail, ok := group.Tails[match]
	if !ok {
		fi, _ := os.Stat(match)
		for _, tail = range group.Tails {
			tfi, _ := tail.Stat()
			if os.SameFile(fi, tfi) {
				tail.Close()
				delete(group.Tails, tail.Path)
				off := tail.Offset()
				tail.SetOffset(0)
				tail.WriteOffset()
				base := filepath.Base(match)
				offset := filepath.Join(group.OffsetDir, base+".ptr")
				tail = NewTail(group.NewParser(base), group.maxBackfill, match, offset, off)
				group.Tails[match] = tail
				return
			}
		}
		base := filepath.Base(match)
		offset := filepath.Join(group.OffsetDir, base+".ptr")
		tail = NewTail(group.NewParser(base), group.maxBackfill, match, offset, 0)
		group.Tails[match] = tail

	}
}

func (group *TailGroup) NewParser(file string) Parser {
	return NewRegexpParser(group.Hostname, group.Name, file, group.output, group.Pattern, group.fields, group.maxLineSize)
}

func (group *TailGroup) deactivate(match string) {
	tail, ok := group.Tails[match]
	if ok {
		delete(group.Tails, match)
		tail.Close()
	}
}

func (group *TailGroup) Refresh() {
	d, _ := os.Getwd()
	logs.Debug("pwd:", d)
	matches, err := filepath.Glob(group.Glob)
	if err != nil {
		logs.Debug("Error in glob: ", err)
	} else if matches == nil {
		logs.Debug("Glob matched zero files: ", group.Glob)
	} else if matches != nil {
		logs.Debug("Glob matched %d files: ", len(matches), group.Glob)
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				logs.Debug("Can't stat: ", err)
			} else if info.IsDir() {
				logs.Debug("Ignoring directory: ", match)
			} else {
				if time.Since(info.ModTime()).Hours() >= 1 {
					logs.Debug("Ignoring idle file: ", match)
					group.deactivate(match)
				} else {
					logs.Debug("Tailing: ", match)
					group.activate(match)
				}
			}
		}
	}
}

func (group *TailGroup) Poll() {
	for _, tail := range group.Tails {
		tail.Poll()
	}
}
