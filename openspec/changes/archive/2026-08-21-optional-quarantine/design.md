## Context

Quarantine-on-create is a hostile-corpus defense: agent-authored documents land in a
quarantine term (`status: agent-draft`) so promotion is a deliberate, separate act.
It was unconditional, which is correct for an untrusted agent but obstructive when
driving the server from a trusted client (local testing, single-user use). This design
records the trade-off of making it optional.

_(Backfilled after the change shipped, to match the `opsx:propose` artifact set.)_

## Goals / Non-Goals

**Goals:** an explicit, discoverable, loudly-warned opt-out; default unchanged
(quarantine on); no weakening of the other defenses.

**Non-Goals:** no per-tool or per-scope granularity; it is all-or-nothing per server.

## Decisions

- **Opt-out, not opt-in.** `DisableQuarantine` / `--no-quarantine` default to false so
  the safe behaviour is the default; you must actively disable it.
- **A `Disabled` flag on `defense.Policy`, not deletion of the term config.**
  `ApplyQuarantine` early-returns when disabled, so `IsQuarantined`, the quarantine
  term, and confirmation rules stay intact and re-enabling is instant.
- **Loud startup warning.** Disabling a defense must be visible in the logs, so
  `serve` prints `WARNING quarantine disabled …`.
- **Threaded through all three config surfaces** (flag, JSON config, NixOS option)
  so it behaves identically however the server is launched; the NixOS option is
  `quarantine = true` by default (disabling requires setting it false).

## Risks / Trade-offs

- [Weaker prompt-injection posture] → a compromised note could get an agent to write
  outside quarantine. Mitigation: default on; warned; other defenses (auth, scopes,
  redaction, git-safety, append-only audit) unaffected, so writes are still authorized
  and recorded. Recommend keeping it on in production (`dapperehaan`).

## Open Questions

- Whether promotion-out-of-quarantine should itself be a tool (so quarantine can stay
  on while still allowing deliberate promotion). Deferred to the write-tools backlog.
