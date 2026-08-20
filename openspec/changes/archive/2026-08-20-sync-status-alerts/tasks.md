## 1. Alerter + monitor

- [x] 1.1 internal/alert: Alerter interface; NtfyAlerter (HTTP POST); NopAlerter
- [x] 1.2 DegradedMonitor: alert on clean->degraded transition once; recovery notice

## 2. sync_status tool

- [x] 2.1 SyncStatus type (degraded/conflicted/paths/in-progress/detached/read-only/reason)
- [x] 2.2 Register sync_status tool driven by the guard state

## 3. Wiring

- [x] 3.1 config: ntfyTopicURL
- [x] 3.2 serve: background poller ticks the monitor; ntfy vs nop alerter

## 4. Tests & gate

- [x] 4.1 sync_status returns degraded+paths on a conflicted tree; healthy on clean
- [x] 4.2 monitor: one alert per degradation; recovery notice (fake alerter)
- [x] 4.3 ntfy alerter issues an HTTP POST (httptest)
- [x] 4.4 nix flake check green (coverage >= 70%)
