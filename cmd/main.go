// Command-line tool to synchronize two catalogs.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"sync-cli/internal/sync"
)

var (
	source        = flag.String("source", "", "source catalog path")
	target        = flag.String("target", "", "target catalog path")
	deleteMissing = flag.Bool("delete-missing", false, "delete missing items in target")
)

func main() {
	flag.Parse()

	if *source == "" || *target == "" {
		log.Fatal("source and target flags are required")
	}

	src, err := filepath.Abs(*source)
	if err != nil {
		log.Fatalf("failed to get absolute path of source: %v", err)
	}
	tgt, err := filepath.Abs(*target)
	if err != nil {
		log.Fatalf("failed to get absolute path of target: %v", err)
	}

	fmt.Printf("Source: %s\n", src)
	fmt.Printf("Target: %s\n", tgt)
	fmt.Printf("Delete Missing: %v\n", *deleteMissing)

	ctx := context.Background()
	err = sync.SynchronizationDirectories(ctx, src, tgt, deleteMissing)
	if err != nil {
		log.Fatalf("synchronization failed: %v", err)
	}
}
