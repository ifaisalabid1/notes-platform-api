package watermark

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	pdfcpuapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type PDFCPUProcessor struct {
	enabled   bool
	brandText string
}

func NewPDFCPUProcessor(enabled bool, brandText string) *PDFCPUProcessor {
	return &PDFCPUProcessor{
		enabled:   enabled,
		brandText: strings.TrimSpace(brandText),
	}
}

func (p *PDFCPUProcessor) Process(ctx context.Context, input ProcessInput) (ProcessResult, error) {
	if !p.enabled || input.ContentType != "application/pdf" {
		return ProcessResult{
			Body:          input.Body,
			ContentType:   input.ContentType,
			IsWatermarked: false,
			Cleanup:       func() {},
		}, nil
	}

	if input.Body == nil {
		return ProcessResult{}, errors.New("watermark input body is required")
	}

	select {
	case <-ctx.Done():
		return ProcessResult{}, ctx.Err()
	default:
	}

	originalFile, err := os.CreateTemp("", "notes-original-*.pdf")
	if err != nil {
		return ProcessResult{}, fmt.Errorf("create original temp pdf: %w", err)
	}

	originalPath := originalFile.Name()

	if _, err := input.Body.Seek(0, io.SeekStart); err != nil {
		_ = originalFile.Close()
		_ = os.Remove(originalPath)
		return ProcessResult{}, fmt.Errorf("rewind input pdf: %w", err)
	}

	if _, err := io.Copy(originalFile, input.Body); err != nil {
		_ = originalFile.Close()
		_ = os.Remove(originalPath)
		return ProcessResult{}, fmt.Errorf("copy input pdf: %w", err)
	}

	if err := originalFile.Close(); err != nil {
		_ = os.Remove(originalPath)
		return ProcessResult{}, fmt.Errorf("close original temp pdf: %w", err)
	}

	watermarkedFile, err := os.CreateTemp("", "notes-watermarked-*.pdf")
	if err != nil {
		_ = os.Remove(originalPath)
		return ProcessResult{}, fmt.Errorf("create watermarked temp pdf: %w", err)
	}

	watermarkedPath := watermarkedFile.Name()

	if err := watermarkedFile.Close(); err != nil {
		_ = os.Remove(originalPath)
		_ = os.Remove(watermarkedPath)
		return ProcessResult{}, fmt.Errorf("close watermarked temp pdf: %w", err)
	}

	watermarkDescription := "rot:45, scale:0.6 abs, op:0.12, pos:c, points:48"

	if err := pdfcpuapi.AddTextWatermarksFile(
		originalPath,
		watermarkedPath,
		nil,
		true,
		p.brandText,
		watermarkDescription,
		model.NewDefaultConfiguration(),
	); err != nil {
		_ = os.Remove(originalPath)
		_ = os.Remove(watermarkedPath)
		return ProcessResult{}, fmt.Errorf("add pdf watermark: %w", err)
	}

	resultFile, err := os.Open(watermarkedPath)
	if err != nil {
		_ = os.Remove(originalPath)
		_ = os.Remove(watermarkedPath)
		return ProcessResult{}, fmt.Errorf("open watermarked pdf: %w", err)
	}

	cleanup := func() {
		_ = resultFile.Close()
		_ = os.Remove(originalPath)
		_ = os.Remove(watermarkedPath)
	}

	return ProcessResult{
		Body:          resultFile,
		ContentType:   "application/pdf",
		IsWatermarked: true,
		Cleanup:       cleanup,
	}, nil
}
