package watermark

import (
	"context"
	"io"
)

type ProcessInput struct {
	FileName    string
	ContentType string
	Body        io.ReadSeeker
}

type ProcessResult struct {
	Body          io.ReadSeeker
	ContentType   string
	IsWatermarked bool
	Cleanup       func()
}

type Processor interface {
	Process(ctx context.Context, input ProcessInput) (ProcessResult, error)
}
