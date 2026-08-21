# Local testing — kick the tires

A guided local run against a **throwaway test corpus**. Never point this at the real
`secondbrain` repo while experimenting.

## 0. Tools

Everything you need is in the dev shell:

```sh
nix develop        # Go, golangci-lint, hugo (for verify)
```

(Or use a host Go ≥ 1.24 and `hugo` for the `verify --hugo` step.)

Build the binaries once:

```sh
go build -o bin/linny-mcp ./cmd/linny-mcp
go build -o bin/lindexer  ./cmd/lindexer
go build -o bin/gen-corpus ./cmd/gen-corpus
```

## 1. Make a test notebook

```sh
DEMO=/tmp/linny-demo; rm -rf $DEMO
./bin/gen-corpus --dir $DEMO --count 30 --seed 1        # synthetic, deterministic
( cd $DEMO && git init -q && git add -A \
    && git -c user.email=you@example.com -c user.name=you commit -qm "test corpus" )
```

`--edge-cases=true` (default) adds a committed conflict marker + a malformed record so
you can see degraded mode and error handling. Use `--edge-cases=false` for a clean
corpus (needed for the Hugo `verify` step — Hugo aborts on malformed front matter).

## 2. Explore the indexer (no MCP client needed)

```sh
# Build the index: linny.vim-compatible JSON + a SQLite/FTS5 store in the state dir
./bin/lindexer build --corpus $DEMO --index $DEMO/lindenIndex --state-dir $DEMO/state

# Full-text search the store
./bin/lindexer search --state-dir $DEMO/state --limit 5 "backup"

# Prove our JSON matches the Hugo reference (needs hugo + a clean corpus)
./bin/lindexer verify --corpus $DEMO --hugo

# Watch: rebuild automatically on change (Ctrl-C to stop)
./bin/lindexer watch --corpus $DEMO --state-dir $DEMO/state
```

Poke around `$DEMO/lindenIndex/*.json` to see the emitted index.

## 3. Run the server

```sh
# Mint a token — the secret is printed ONCE; the record line goes in the tokens file.
OUT=$(./bin/linny-mcp gen-token --name local --scopes 'read:*,write:inbox')
echo "$OUT"
TOKEN=$(echo "$OUT" | sed -n 's/^token: //p')
echo "$OUT" | grep '^{' > $DEMO/tokens.jsonl

# Serve on loopback (writes enabled because a state dir is set and not --read-only).
./bin/linny-mcp serve --corpus $DEMO --state-dir $DEMO/state \
    --tokens-file $DEMO/tokens.jsonl --listen 127.0.0.1 --port 8765
```

Health check (no auth) and the auth gate:

```sh
curl -s http://127.0.0.1:8765/healthz | jq .          # {"status":"ok",...}
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:8765/mcp   # 401 (no token)
```

## 4. Connect an MCP client

The tool endpoint is **`POST http://127.0.0.1:8765/mcp`** (MCP streamable-HTTP) behind
`Authorization: Bearer <token>`.

- **Claude Code / Claude Desktop:** add it as an HTTP MCP server with that URL and an
  `Authorization: Bearer <token>` header (see `claude mcp add --help`). Then ask it to
  `search`, `get_doc`, `list_taxonomies`, etc.
- **MCP Inspector (GUI):** `npx @modelcontextprotocol/inspector` → Transport
  "Streamable HTTP" → the `/mcp` URL → add the bearer header → browse/call tools.

Raw sanity check (an authenticated `initialize` returns 200; the streamable transport
needs the `Accept` header):

```sh
curl -s -o /dev/null -w '%{http_code}\n' \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -X POST http://127.0.0.1:8765/mcp \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"curl","version":"0"}}}'
```

The tool surface (arguments, semantics, scopes) is documented in `docs/tools.md`.

## 5. Things worth trying

- **Degraded mode:** while the server runs, add a conflict marker to a note
  (`printf '<<<<<<< HEAD\n' >> $DEMO/content/<some>.md`), then call `sync_status` /
  `GET /healthz` — it flips to `degraded`, and writes are refused. Remove it → recovers
  automatically (no restart).
- **Scopes:** mint a second token with `--scopes 'read:*,deny:taxonomy:tags:health'`
  and confirm `docs_by_term tags health` returns nothing and `get_doc` on a
  health-tagged note reports not-found.
- **Redaction:** the generated corpus plants fake secrets; `get_doc fake_secrets.md`
  returns them `[REDACTED:...]`.
- **Quarantine + audit:** `create_doc` lands in `status: agent-draft`; every write is
  appended to `$DEMO/state/audit.log`.
- **Disposable cache / backup:** delete `$DEMO/state` and re-run `lindexer build` — a
  valid recovery. `linny-mcp backup --corpus $DEMO --out snap.tar.gz` then
  `restore --in snap.tar.gz --corpus <dir>` round-trips the content.

## Reset

```sh
rm -rf $DEMO           # nothing here is precious; regenerate anytime
```
