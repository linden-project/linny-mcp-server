## 1. Backup/restore

- [x] 1.1 internal/backup: Backup(root, w) tar.gz of content/ + lindenConfig/
- [x] 1.2 Restore(r, root) extract with path sanitization (reject .. / absolute)
- [x] 1.3 Exclude index/state + .git/.jj

## 2. CLI

- [x] 2.1 linny-mcp backup --corpus --out ; restore --in --corpus

## 3. Tests & gate

- [x] 3.1 backup contains content + config, not .git/index
- [x] 3.2 verified round-trip: delete a record, restore, recovered byte-for-byte
- [x] 3.3 mutated record recovered; traversal entry rejected
- [x] 3.4 nix flake check green (coverage >= 70%)
