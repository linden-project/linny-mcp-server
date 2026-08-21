## 1. Indexer

- [x] 1.1 loadNotebook returns plural->singular taxonomy map
- [x] 1.2 Graph.Singular field; set in Build
- [x] 1.3 L1 emit looks up L2-CONF by singular taxonomy name

## 2. Spec

- [x] 2.1 docs/linden-index-spec.md §9.1: L1 lookup uses singular; §13 Q7 resolved

## 3. Tests & gate

- [x] 3.1 Hugo round-trip asserts zero drift
- [x] 3.2 nix flake check green (coverage >= 70%; round-trip runs with hugo)
