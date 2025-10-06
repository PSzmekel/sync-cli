// Command-line tool to synchronize two catalogs.
package main

import (
	"context"
	"flag"
	"log"
	"path/filepath"
	"sync-cli/internal/sync"
)

var (
	source        = flag.String("source", "", "source catalog path")
	target        = flag.String("target", "", "target catalog path")
	deleteMissing = flag.Bool("delete-missing", false, "delete missing items in target")
	deepSearch    = flag.Bool("deep-search", false, "perform deep search in directories")
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

	ctx := context.Background()
	err = sync.SynchronizationDirectories(ctx, src, tgt, *deleteMissing, *deepSearch)
	if err != nil {
		log.Fatalf("synchronization failed: %v", err)
	}
}
