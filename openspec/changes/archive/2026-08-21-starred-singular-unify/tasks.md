## 1. Generator (singular config filenames)

- [x] 1.1 taxonomy.singular field (tag/project); singularOf helper
- [x] 1.2 writeLindenConfig names L1/L2 files by singular
- [x] 1.3 writeHugoConfig derives singular:plural from the taxonomy table

## 2. Indexer

- [x] 2.1 loadLindenConfig scans + keys L1/L2 by filename taxonomy
- [x] 2.2 finalize derives starred from config maps (singular), no occurrence filter

## 3. Spec + tests + gate

- [x] 3.1 Update the corpus config-consistency test to the singular filenames
- [x] 3.2 Spec note: config filenames canonical singular; §13-Q7 fully resolved
- [x] 3.3 verify --hugo zero drift; L1 resolves; starred singular
- [x] 3.4 nix flake check green (coverage >= 70%)
