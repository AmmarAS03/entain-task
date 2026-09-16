# Task 5: `SearchEvents` RPC

> Write a new `SearchEvents` RPC in Core that allows a user to search by date
> and/or bettingstatus and/or display. Return a slice of all events that match
> the criteria.

## What changed

| File | Change |
|---|---|
| `core/core.proto` | New `SearchEventsRequest`, `SearchEventsResponse`, `EventSummary` messages; new `SearchEvents` rpc on `service Service` |
| `core/core.pb.go`, `core/core_grpc.pb.go` | Regenerated only |
| `core/repository/types.go` | New `EventFilter` struct; `Repository` gains `SearchEvents(ctx, EventFilter) ([]*model.Event, error)` |
| `core/repository/mongo.go` | `mongoRepo.SearchEvents` runs the filter as a server-side Mongo query; new `ensureIndexes`, called from `NewMongoRepository` |
| `core/service/implementation.go` | New `SearchEvents` RPC: translates the request into an `EventFilter`, calls the repository, converts each result |
| `core/package.go` | New `(*EventSummary).ConvertFromModel` |
| `core/repository/mongo_test.go` | 9 new subtests covering every filter combination, boundary inclusivity/exclusivity, the `Display` null case, and ordering |
| `core/service/implementation_test.go` | New end-to-end integration test through the real RPC |
| `core/package_test.go` | `EventSummary.ConvertFromModel` unit tests: fully populated and sparse |
| `exampledata/10_search_event_two.json`, `11_search_event_three.json` | Two more events, on different dates/statuses/display values, so the existing `testEvent` plus these three give the demo below something to filter |

No `model.Event` field is added or changed by this task, so the 3-places rule
(proto / merger / read-out) does not apply to anything here — see the plan for
the full breakdown. `merger/`, `core/transforms/`, `core/cmd/core/main.go`, and
`model/event.proto` are untouched.

## Decisions

### Result shape: a new `core.EventSummary`, not the existing sport/racing views

Events are now sport *or* racing (Task 2). Reusing `core.SportEvent` would
silently report empty `SportName`/`Region`/`League`/`Round` for a matched
race; `core.RacingEvent` would do the reverse for a matched sport event.
Returning `model.Event` directly would leak the `Optional*` wrapper types to
consumers — the exact thing the `core.*` response views exist to hide.
`EventSummary` is sport-agnostic, carries just enough to identify and triage a
result, and points the consumer at `GetSportEvent`/`GetRacingEvent` for detail.

### Date filter: `StartTimeFrom`/`StartTimeTo`, epoch nanoseconds, from inclusive, to exclusive

Maps directly onto Mongo's `$gte`/`$lt`. A single day is `[start of day, start
of next day)`. Timezone conversion is left to the caller — a race meeting can
span two UTC days, and the service has no basis for guessing which timezone
the caller means.

### `Display: false` matches explicitly-false **and** never-set

Task 1 established "unset `Display` means hidden", and `GetSportEvent`/
`GetRacingEvent` already report a never-set event's `Display` as `false`. If
`SearchEvents` disagreed with those RPCs about the same stored event, that
would be a bug, not a nuance worth defending. Because Task 4 stores unset
`Optional*` fields as BSON `null` rather than omitting them, "hidden" is
implemented as "not explicitly `true`":

```go
if *filter.Display {
    query["display.value"] = true
} else {
    query["display.value"] = bson.M{"$ne": true}
}
```

`$ne: true` matches `false`, `null`, and a missing field alike. The one cost:
this branch cannot use the `display.value` index efficiently — negation and
null-matching are poor index citizens. The index still fully serves the
`Display: true` branch.

### Multiple filters combine with AND; zero filters return every event

The README's literal reading of "date and/or bettingstatus and/or display".
Each filter is independently optional, so "none supplied" is the degenerate
case of that, not an error — `SearchEvents({})` returns the whole collection.
See Notes/limitations below for the size implication.

### No pagination

The README asks for "all events that match". Adding a cursor nobody asked for
would be scope nobody asked for either. Documented as a limitation, not
silently deferred.

### Sort: `starttime.value` ascending, `_id` as a tiebreak

Mongo's natural order is not stable across reads. A deterministic order is
both a prerequisite for asserting exact results in tests and the order a
consumer most likely wants a result set in.

### Three single-field indexes, not one compound index

`ensureIndexes` (called from `NewMongoRepository`, idempotent, safe on every
startup) creates one index each on `starttime.value`, `bettingstatus.value`,
and `display.value`. The three filters are independently optional, giving 2³
possible query shapes; a compound index only serves queries that use its
prefix, while three single-field indexes let Mongo pick the most selective one
(or intersect several) for any combination. Task 4 shipped with none — there
was exactly one access pattern (`_id`) until this task introduced the first
query that can table-scan.

**Index creation failure is a warning, not a boot failure.** `CreateMany`
needs write permission and is a round trip on every startup; if it fails,
`SearchEvents` still returns correct results, just slower. Refusing to boot
over a missing index would be strictly worse than booting slow, so the error
is logged at `Warn` and the service continues.

## Verifying

```sh
docker compose up -d
go build ./...
go vet ./...
go test ./... -short -cover
./core/check.sh
```

(`go test ./... -short -cover` — and therefore `./core/check.sh`, which runs it —
prints `go: no such tool "covdata"` for packages with no test files, such as
`core/cmd/core` and `model`. This is a pre-existing toolchain gap unrelated to
this task: it reproduces identically on `main`. `go test ./... -short` without
`-cover` passes cleanly.)

Manual pass through the real service:

```sh
./run_local.sh
# Seed a distinguishable set:
#   01_new_event.json           -> testEvent        (BettingOpen,  Display unset -> hidden)
#   10_search_event_two.json    -> testEventTwo      (BettingClosed, Display true)
#   11_search_event_three.json  -> testEventThree     (BettingOpen,  Display false)

# No filters: all three, ascending by StartTime
grpcurl -plaintext -import-path . -import-path ./core -proto core.proto \
  -d '{}' localhost:50051 core.Service/SearchEvents

# Display: true -> only testEventTwo
grpcurl -plaintext -import-path . -import-path ./core -proto core.proto \
  -d '{"Display":{"Value":true}}' localhost:50051 core.Service/SearchEvents

# Display: false -> testEvent (never set) and testEventThree (explicit false)
grpcurl -plaintext -import-path . -import-path ./core -proto core.proto \
  -d '{"Display":{"Value":false}}' localhost:50051 core.Service/SearchEvents

# BettingStatus: BettingOpen -> testEvent and testEventThree
grpcurl -plaintext -import-path . -import-path ./core -proto core.proto \
  -d '{"BettingStatus":{"Value":"BettingOpen"}}' localhost:50051 core.Service/SearchEvents

# StartTimeFrom narrows to events on/after that instant
grpcurl -plaintext -import-path . -import-path ./core -proto core.proto \
  -d '{"StartTimeFrom":{"Value":"1758300000000000000"}}' localhost:50051 core.Service/SearchEvents

# A filter matching nothing returns {} (an empty Events slice), not an error
grpcurl -plaintext -import-path . -import-path ./core -proto core.proto \
  -d '{"StartTimeFrom":{"Value":"9223372036854775807"}}' localhost:50051 core.Service/SearchEvents
```

Confirm the indexes were actually created:

```sh
docker compose exec mongo mongosh codetest --eval "db.events.getIndexes()"
```

This was run against a live `mongo:8` for this task: all six queries above
returned exactly the expected `EventSummary` sets, and `getIndexes()` showed
`starttime.value_1`, `bettingstatus.value_1`, and `display.value_1` alongside
the default `_id_` index.

## Notes / limitations

- **Unbounded result set.** `SearchEvents({})` decodes every event in the
  collection into memory in one response. Fine at code-test scale; a
  production hazard at real scale. The README explicitly asks for "all events
  that match", so no pagination was added — but this is the first thing that
  would need to change before this endpoint saw real traffic, and it would hit
  the gRPC `MaxSendMsgSize` of 64MB already set in `core/service/service.go`
  before anything else did.
- **Search tests use their own database, one per package**, not the shared
  `codetest` dev database every other test in this repo uses. Every prior test
  fetches by a unique ID, so a stray document left behind by `./run_local.sh`
  couldn't affect the result. `SearchEvents` scans the whole collection, so a
  test asserting "this filter returns 2 events" would fail intermittently the
  moment anyone left dev data behind. This is a targeted fix for the one test
  class that needs isolation, not a reversal of Task 4's "tests share the dev
  database" decision. It has to be *one database per package* rather than one
  shared "search" database: `go test ./...` runs each package's test binary
  concurrently, so `core/repository` (`codetest_search_test`) and
  `core/service` (`codetest_search_service_test`) would otherwise scan each
  other's documents mid-run and fail on the exact same "someone else's
  machine, looks like a real bug" flakiness this isolation exists to prevent.
- **`.claude/CLAUDE.md` is stale**, describing Redis as the repository and a
  four-method `Repository` interface — both changed in Task 4, and this task
  adds a fifth method on top. Out of scope here (and `README.md`/existing task
  docs must stay untouched), but worth a follow-up commit so it isn't
  rediscovered as a surprise.
