## 1. Store

- [x] 1.1 index.Store.AllDocFilenames()

## 2. Tool

- [x] 2.1 verify_index: rebuild graph, diff doc set vs store, report counts/missing/stale/conflicted/in_sync
- [x] 2.2 Register operational tool; docs/tools.md

## 3. Tests & gate

- [x] 3.1 in-sync -> in_sync true; added-but-unindexed -> missing_from_store
- [x] 3.2 conflicted corpus -> conflicted path listed
- [x] 3.3 nix flake check green (coverage >= 70%)
