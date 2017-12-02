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

func tailFile(e Exporter) (*tail.Tail, error) {
	return tail.TailFile(e.GetFilePath(), tail.Config{
		Follow:    true,
		MustExist: true,
		Location: &tail.SeekInfo{
			Offset: 0,
			Whence: 2,
		},
	})
}
