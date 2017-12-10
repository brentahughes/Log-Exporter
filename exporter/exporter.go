package exporter

import (
	"github.com/hpcloud/tail"
)

type Exporter interface {
	StartExport() error
	SetupMetrics()
	Close()
	GetFilePath() string
}

func tailFile(e Exporter, debug bool) (*tail.Tail, error) {
	var location *tail.SeekInfo
	if !debug {
		location = &tail.SeekInfo{
			Offset: 0,
			Whence: 2,
		}
	}

	return tail.TailFile(e.GetFilePath(), tail.Config{
		Follow:    true,
		Location:  location,
	})
}
