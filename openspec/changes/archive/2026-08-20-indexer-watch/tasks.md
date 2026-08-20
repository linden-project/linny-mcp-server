## 1. Debounce + watcher

- [x] 1.1 debounce helper coalescing a burst into one fire after the window
- [x] 1.2 Watcher over content/ + lindenConfig via fsnotify -> debounced callback

## 2. CLI

- [x] 2.1 lindexer watch --corpus --state-dir [--index]: build once, then refresh on change

## 3. Tests & gate

- [x] 3.1 debounce fires once per burst (deterministic)
- [x] 3.2 fsnotify integration: writing a record triggers the callback
- [x] 3.3 watch requires --state-dir
- [x] 3.4 vendorHash updated; nix flake check green (coverage >= 70%)
