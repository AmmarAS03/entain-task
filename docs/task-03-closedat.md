# Task 3: ClosedAt on Market

> Add a ClosedAt time to markets and set this value in the transform the first
> time a market is closed.

## What changed

| File | Change |
|---|---|
| `model/event.proto` | `OptionalInt64 ClosedAt = 6` on `Market` |
| `merger/modelmerge.go` | `ClosedAt` wired into `MergeMarket` |
| `merger/modelmerge_test.go` | `ClosedAt` added to the `populatedMarket` fixture |
| `core/transforms/marketclosetransform/` | New transform stamping `ClosedAt` the first time a market's `BettingStatus` becomes `BettingClosed` |
| `core/cmd/core/main.go` | Market close transform registered in the pipeline |
| `core/service/implementation_test.go` | Registered in the test suite's own `Transforms` slices; new `TestService_IntegrationTest_MarketClose` end-to-end test |
| `exampledata/08,09_*.json` | Close and reopen a market, for demonstrating the frozen timestamp |

No changes were needed to `core/core.proto`, `core/package.go` or
`core/repository/redis.go` — `ClosedAt` rides through both `GetSportEvent` and
`GetRacingEvent` for free because both responses embed `model.Market` directly,
and the repository persists the whole event as one JSON blob.

## Decisions

### Epoch nanoseconds, reusing `OptionalInt64`

`Market.StartTime` already uses `OptionalInt64` in epoch nanoseconds
(`core/package.go` renders it with `time.Unix(0, v)`). `ClosedAt` follows the
same convention rather than introducing a new type or a different unit — one
fewer thing for a consumer to get wrong.

### Freeze at first close

`open → close(t1) → open → close` leaves `ClosedAt = t1`. This is the literal
reading of the README ("set this value ... the first time a market is
closed"), and it is a deliberate call, not an oversight: `ClosedAt` answers
"when did this market first close", not "is it closed now" — that second
question is what `BettingStatus` is for.

The alternative (reset `ClosedAt` on reopen) was rejected because it needs
merger support that does not exist today: `MergeOptionalInt64` treats a nil
right-hand side as "keep left", and nothing in this codebase honours the
`Deleted` flag on read-out. Adding that would be new merger behaviour, not a
market-close transform.

### Trigger: only when the partial update touched markets

The transform's guard mirrors the existing shape (`sporttransform`,
`racingtransform`): skip unless *this specific update* touched the relevant
subtree, then scan the merged model rather than just the partial. Scanning the
full model means a market that was already closed before this transform
shipped gets backfilled the next time any market on that event is touched,
rather than needing a one-off migration.

## The transform

`marketclosetransform` is its own package rather than folded into
`sporttransform` or `racingtransform`, because market closure applies equally
to soccer, rugby league and every racing code — gating it on `EventTypeID` or
`RacingData`, like the other two transforms do, would misrepresent its scope.

Guard shape, same as every other transform in this pipeline:

1. **Trigger guard** — `len(partialUpdate.GetMarkets()) == 0` → no delta. This
   update did not touch any market, so there is nothing to check.
2. **Per-market idempotency guard** — for each market on the merged model,
   skip anything not `BettingClosed`, and skip anything that already has a
   `ClosedAt`.
3. **Minimal delta** — only `{ID, ClosedAt}` per newly-closed market. Merging
   by ID means nothing else on the market (`Name`, `Selections`, ...) is
   disturbed.

## Verifying

```sh
docker-compose up -d
./model/gen-proto.sh
go build ./...
go test ./merger/... ./core/... -cover
go test ./core/transforms/marketclosetransform/... -v
```

Manually, against a running service (`./run_local.sh`):

1. `Update` with `exampledata/01_new_event.json`, then `GetSportEvent` on
   `testEvent` — the `H2H` market is present, `ClosedAt` is absent.
2. `Update` with `08_close_market.json` — `GetSportEvent` now shows `ClosedAt`
   as an epoch-nanosecond timestamp. Note the value.
3. `Update` with `02_open_selection.json` (an unrelated partial touching a
   selection, not the market's status) — `ClosedAt` is unchanged.
4. `Update` with `09_reopen_market.json`, then `08_close_market.json` again —
   `ClosedAt` is still the value from step 2, not a new one (frozen at first
   close).

## Notes / limitations

- **`ClosedAt` is client-writable, not access-controlled.** `Update` is a
  public RPC and `MergeMarket` merges `ClosedAt` like any other field, so a
  client can set it directly rather than letting the transform derive it. The
  idempotency guard keys off `Value != 0` rather than pointer nil-ness
  specifically so a client-supplied empty `ClosedAt` (`{}`, which unmarshals
  to a non-nil zero-value message) cannot permanently block the transform
  from later stamping a real close time — but a client sending a
  non-zero value still overrides the transform outright. Stopping that
  would need stripping `ClosedAt` from incoming partial updates before the
  merge, which was judged out of scope here since no field on this model is
  currently access-controlled this way.
- **Transform errors are swallowed**, matching the existing pipeline semantics
  (`core/service/implementation.go`): a transform failure logs and the write
  still succeeds without `ClosedAt`. This implementation returns a nil error on
  every path, so it is a theoretical concern here, not an active one — kept
  the existing best-effort semantics deliberately rather than special-casing
  this transform.
- **Frozen re-close is deliberate**, not a bug: a reopened market keeps its
  original `ClosedAt`. Consumers that need "is it closed *right now*" should
  read `BettingStatus`, not `ClosedAt`.
- **Backfill is lazy, and the backfilled value is not the real close time.**
  A market closed before this transform shipped is stamped on its *next*
  market-bearing update — with the timestamp of *that update*, not the
  market's actual historical close time, which this system has no record of.
  A market that closed last week but is only touched again today gets
  `ClosedAt = today`. Consumers doing settlement or "closed N minutes before
  jump" checks should treat `ClosedAt` as trustworthy only for markets that
  closed after this transform shipped; a suspiciously recent `ClosedAt` on an
  old market is this backfill, not a re-close (which is impossible per the
  freeze-at-first-close rule above).
