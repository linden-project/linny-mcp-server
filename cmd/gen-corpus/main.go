// Command gen-corpus materializes a synthetic Linny notebook on disk for manual
// inspection and for the indexer `verify` path. It never touches real data.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/linden-project/linny-mcp-server/internal/corpus"
)

func main() {
	var (
		dir       = flag.String("dir", "testdata-gen", "target directory")
		seed      = flag.Int64("seed", 1, "PRNG seed (same seed => identical corpus)")
		count     = flag.Int("count", 200, "number of normal records")
		edgeCases = flag.Bool("edge-cases", true, "include deliberate edge-case records")
	)
	flag.Parse()

	opts := corpus.Options{Seed: *seed, Count: *count, EnableEdgeCases: *edgeCases}
	if err := corpus.Generate(*dir, opts); err != nil {
		fmt.Fprintf(os.Stderr, "gen-corpus: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("generated %d records (+edge cases=%t, seed=%d) in %s\n", *count, *edgeCases, *seed, *dir)
}
