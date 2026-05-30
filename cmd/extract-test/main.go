package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/daax-dev/nanofuse/internal/builder"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: extract-test <image> <output-dir> <fallback-kernel>")
		os.Exit(1)
	}

	imageRef := os.Args[1]
	outputDir := os.Args[2]
	fallbackKernel := os.Args[3]

	b := builder.NewDockerBuilder("/tmp", true)
	if err := b.Available(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	opts := builder.ExtractOptions{
		OutputDir:          outputDir,
		RootfsSizeMB:       1024, // 1GB for test
		FallbackKernelPath: fallbackKernel,
		Verbose:            true,
		OnProgress: func(stage string, percent int) {
			fmt.Printf("[%3d%%] %s\n", percent, stage)
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := b.Extract(ctx, imageRef, opts)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Println("=== Extraction Complete ===")
	fmt.Printf("Kernel:       %s\n", result.KernelPath)
	fmt.Printf("Rootfs:       %s\n", result.RootfsPath)
	fmt.Printf("Digest:       %s\n", result.Digest)
	fmt.Printf("Architecture: %s\n", result.Architecture)
	fmt.Printf("Kernel Ver:   %s\n", result.KernelVersion)
	fmt.Printf("Size:         %d bytes\n", result.SizeBytes)
	fmt.Printf("Duration:     %v\n", result.Duration)
}
