# Deployment wiring (mipnix side) — documentation only

This describes how `linny-mcp` is wired into the fleet. It is **documentation**: the
actual host configuration lives in `github.com/mipmip/mipnix`, not in this repo.

## Topology

```
Claude mobile / claude.ai connector
        │  HTTPS (public hostname, e.g. secondbrain.pimsnel.com)
        ▼
   Hetzner box  ── reverse proxy, terminates TLS ── Nebula lighthouse + mesh member
        │  Nebula overlay (mesh IP)
        ▼
   Old MacBook running Linux (NixOS)
        │  loopback / mesh bind, TLS terminated upstream
        ▼
   linny-mcp serve  ──►  /mcp (bearer auth) + /healthz
```

The server binds loopback or the Nebula mesh only; it refuses a public bind without
an explicit override. TLS is terminated by the Hetzner proxy, which forwards over the
mesh.

## Flake input

```nix
# mipnix flake.nix
inputs.linny-mcp.url = "github:linden-project/linny-mcp-server";
# in the MacBook host's modules:
imports = [ inputs.linny-mcp.nixosModules.linny-mcp ];
nixpkgs.overlays = [ inputs.linny-mcp.overlays.default ];
```

## Host config (sketch)

```nix
services.linny-mcp = {
  enable = true;
  publicHostname = "secondbrain.pimsnel.com";   # configuration, never a constant
  listenAddress = "127.0.0.1";                  # or the mesh IP
  port = 8765;
  tokensFile = config.age.secrets.linny-mcp-tokens.path;   # a PATH, never a token value
  ntfyTopicURL = "https://ntfy.<host>/linny-mcp";
  notebooks = [
    { name = "personal"; corpusPath = "/home/pim/secondbrain"; stateDir = "/var/lib/linny-mcp/personal"; }
    # add a "business" notebook here later — the config already supports N.
  ];
};
```

## Secrets (agenix)

```nix
age.secrets.linny-mcp-tokens = {
  file = ../secrets/linny-mcp-tokens.age;
  owner = "linny-mcp";            # the service user
};
```

The token file holds hashed records (`linny-mcp gen-token` prints them). **No token
value ever appears in a Nix option** — options land world-readable in `/nix/store`.

## Nebula

The MacBook is a Nebula mesh member; the Hetzner box is the lighthouse and reaches it
over the overlay. Reverse-proxy `secondbrain.pimsnel.com` → the MacBook's mesh IP:port.
(Nebula certs/config live in mipnix, out of scope here.)

## ntfy

A self-hosted ntfy runs on the Hetzner box; `ntfyTopicURL` points the server at a
topic. Degraded-mode transitions POST an alert to the phone — **never** into the
corpus.

## git-sync

The existing external git-sync keeps each `corpusPath` in sync on all machines,
including this host. `linny-mcp` never takes ownership of git; it only inspects the
working tree and degrades to read-only when it is conflicted.
