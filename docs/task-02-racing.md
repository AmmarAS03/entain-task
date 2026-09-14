# Task 2: Racing substructure and GetRacingEvent

> The current structure suits Sport Match style events but horse / greyhound /
> harness racing does not really fit into this structure. Add a new Racing
> substructure with some fields you think are appropriate. Add a new
> `GetRacingEvent` RPC that returns a structure useful to consumers wanting a
> racing event. Consider writing a new Transform if there are racing specific
> transformations you want to make.

## What changed

| File | Change |
|---|---|
| `model/event.proto` | `RacingEvent` and `Runner` messages, `RaceStatus` enum + `OptionalRaceStatus`, `Event.RacingData` field 9 |
| `merger/modelmerge.go` | `MergeRacingEvent`, `MergeRunner`, `MergeOptionalRaceStatus`, `MergeRaceStatus`; `RacingData` wired into `MergeEvent` |
| `merger/slices.go` | `MergeRunnerSlice`, merging runners by ID |
| `core/core.proto` | `RacingEvent` + `RacingRunner` response messages, `GetRacingEvent` RPC |
| `core/package.go` | `RacingEvent.ConvertFromModel` with the runner/price join; shared `formatStartTime` helper |
| `core/service/implementation.go` | `GetRacingEvent` implementation |
| `core/transforms/racingtransform/` | New transform deriving the racing code name and `FieldSize` |
| `core/cmd/core/main.go` | Racing transform registered in the pipeline |
| `merger/modelmerge_test.go` | Completeness fixtures for the racing types; runner-merge-by-ID test |
| `core/transforms/racingtransform/racingtransform_test.go` | Transform unit tests |
| `core/service/implementation_test.go` | End-to-end racing test and a sport/racing separation test |
| `exampledata/05,06,07_*.json` | Race card, scratching, and price move payloads |

## Verifying

```sh
docker-compose up -d
go test ./... -v -run 'Racing|Runner|Transform'
```

Manually, against a running service (`./run_local.sh`):

1. `Update` with `exampledata/05_new_race.json`, then `GetRacingEvent` on
   `testRace`. Three runners come back with Win and Place prices attached, and
   `FieldSize` is `3` with `RacingName` set to `Horse Racing` by the transform.
2. `Update` with `07_race_price_move.json`. Runner 1's `WinPrice` moves to 2.10,
   its name and barrier are unchanged, `FieldSize` stays `3`.
3. `Update` with `06_scratch_runner.json`. Runner 2 is flagged scratched,
   `FieldSize` drops to `2`, and runner 2 keeps its name and stays on the card.
4. `GetSportEvent` on `testRace`. Shared fields resolve but `SportName`,
   `League` and `Round` are empty, since this is not a sport event.

## The modelling problem

A sport `Selection` is a **betting option with no independent existence**. "Home"
and "Over" are not entities in the world, they are things you can bet on. They
live inside exactly one market and mean nothing outside it.

A racing **runner is the opposite**. The horse, its jockey, its barrier and its
carried weight exist independently of any betting market, and the same runner is
bet on across every market on the race: Win, Place, Exacta, Quinella, Trifecta.
That is the actual structural mismatch the task is pointing at, and it drives
almost every decision below.

## Decisions

### Runners live on the Event, not inside Markets

The obvious shortcut is to add `Jockey`, `Barrier` and `Weight` to `Selection`
and be done. That is wrong, for reasons that get worse as the system scales:

- **Duplication.** A 16 runner race with 5 markets stores 80 selections. Putting
  runner attributes on the selection stores the jockey's name 5 times per runner.
- **Consistency.** A scratching is one real world fact. On the selection model it
  becomes 5 writes that have to all land, and any partial failure leaves the race
  reporting different field sizes depending on which market you read.
- **Payload weight on the hot path.** Prices move constantly; runner metadata
  barely changes. Coupling them means every price tick carries jockey and trainer
  strings it does not need. Keeping them apart lets a price feed send
  `{market, selection, price}` and nothing else.

So `Runner` hangs off `RacingEvent`, once per event, and markets reference
runners rather than redescribing them.

### Markets reference runners by Selection ID

There is no new link field. For a racing event, `Selection.ID` **is** the
`Runner.ID`.

This is not a new convention, it is the one already in the repo. Look at
`exampledata/01_new_event.json`: `Market.ID` is `"H2H"` and `Selection.ID`s are
`"home"` and `"away"`. These are meaningful business keys, not opaque surrogate
IDs. Racing follows the identical pattern with `Market.ID` of `"WIN"` / `"PLACE"`
and `Selection.ID`s of `"1"`, `"2"`, `"3"`.

The alternative was adding an explicit `RunnerID` to `Selection`. Rejected
because it pushes a racing-only concept into the shared `Selection` type that
every sport also uses. **`Market` and `Selection` are unchanged by this task**,
which is the strongest evidence the substructure is in the right place.

The trade-off is honest: the join is convention-based rather than enforced by the
schema. A selection ID that does not match any runner simply produces no price.
The market IDs are exported as `core.WinMarketID` / `core.PlaceMarketID` so the
convention lives in one place rather than scattered as string literals.

### RacingData sits beside SportData, not in a oneof with it

An event is a match or a race, never both, so a `oneof` looks like the correct
modelling tool. It is actively wrong here.

The entire system is built on merge-preserving partial updates: a field absent
from an update keeps its existing value. Proto3 `oneof` has the opposite
semantic, setting one arm **clears** the other. A partial update that touched
only `RacingData` would silently wipe `SportData`. It would also be a breaking
change to move the existing `SportData` field into a oneof.

So `RacingData` is field 9 on `Event`, directly parallel to `SportData` on field
7. An event carrying both is possible but meaningless, and that costs nothing.

### Reusing existing Event fields rather than duplicating them

Two fields were deliberately *not* added to `RacingEvent`:

- **Race type** (horse / greyhound / harness) uses the existing
  `Event.EventTypeID`, exactly as sport events use it for `soccer` and
  `rugby_league`. The racing transform maps it to a display name the same way
  `sporttransform` does.
- **Scheduled jump time** uses the existing `Event.StartTime`. A race has one
  start time and `Event` already models it.

### RaceStatus is a new enum, separate from BettingStatus

A race has a lifecycle that betting status cannot express: `RaceScheduled`,
`RaceRunning`, `RaceInterim` (results are in but a protest is pending),
`RaceFinal`, `RaceAbandoned`. These are not betting states. Betting typically
closes at the jump, but the race being *run*, *interim* or *final* is a distinct
axis that racing consumers need and sport events have no equivalent of.

### Runner field choices across the three codes

One `Runner` message covers all three codes, with optional fields absorbing the
differences rather than three near-identical messages:

| Field | Horse | Greyhound | Harness |
|---|---|---|---|
| `Number` | saddlecloth | box number | saddlecloth |
| `Barrier` | barrier draw | box draw | barrier/handicap |
| `Weight` | carried weight | unset | unset |
| `Jockey` | jockey | unset | driver |
| `Trainer` | trainer | trainer | trainer |

`Jockey` carrying the harness driver is the one genuinely uncomfortable call. A
neutral name like `Participant` would be more correct but less recognisable, and
industry feeds overwhelmingly use "jockey" and "driver" as separate concepts. The
field is documented in the proto. If harness volume ever justified it, adding a
`Driver` field is additive and non-breaking.

## The response shape

`GetRacingEvent` returns a **race card**, which is what a racing consumer
actually wants to render: race conditions at the top, then a list of runners with
their prices.

The key work is the join. `RacingRunner` flattens a `model.Runner` together with
its current `WinPrice` and `PlacePrice`, pulled from the Win and Place markets.
Without this, every consumer would independently walk `Markets`, find the right
market, index selections by ID, and stitch it to runners. That is the
transformation the RPC exists to do.

Raw `Markets` remain on the response, because exotics (Exacta, Trifecta,
Quinella) price *combinations* of runners and do not fit a per-runner column.
The join covers the common case; the raw markets cover everything else.

`runnerPrices` builds both indexes in a single pass over markets, so the
conversion is O(markets + selections + runners) rather than a nested scan per
runner.

## The transform

`racingtransform` derives two things, and they deliberately have **different
shapes**:

**Racing code name** is *set once*. Same guard as `sporttransform`: skip unless
this update touched `EventTypeID`, skip if a name already exists. Once derived it
never needs recomputing.

**FieldSize** (runners not scratched) *cannot* use that guard. A late scratching
has to move it, so a set-once check would freeze it at the wrong value. Instead
it recomputes whenever an update touches runners, and returns no delta when the
recomputed value matches what is already stored, so an unchanged race produces no
write.

This contrast is the reason the transform is worth having beyond the name
mapping. It shows the two derivation shapes this pipeline supports: values
derived once at birth, and values that track a changing input. Getting the second
one wrong is silent, which is why
`TestTransformEvent_FieldSizeRecomputesOnScratching` exists specifically to pin
it down.

`FieldSize` is stored on the model rather than computed in `ConvertFromModel`
because it is a genuine derived property of the race, not a presentation concern.
Task 5's `SearchEvents` would be able to filter on it; a value computed at
response time would not be queryable.

## Performance, maintainability, scalability

**Performance.** The runner/price join is one pass over markets to build two
maps, then one pass over runners. No nested scanning. Storing runners once rather
than per market keeps the persisted JSON blob materially smaller for a race with
several markets, which matters because the repository reads and writes the whole
event as a single document on every update.

**Maintainability.** `Market` and `Selection` are untouched, so racing adds no
conditional logic to the sport path. The merge-completeness test from Task 1
immediately caught `Event.RacingData` missing from its fixture when the field was
added, which is precisely the silent failure mode it was written for; the same
guard now covers `RacingEvent` and `Runner`.

**Scalability.** The high-frequency path is price updates. Because prices live on
selections and runner metadata lives on runners, a price tick carries only
`{market, selection, price}`. Scratchings are low-frequency and touch one runner
by ID via `MergeRunnerSlice`, leaving the rest of the field untouched. The two
workloads do not interfere.

## Notes / limitations

- **The price join is convention-based.** Market IDs `WIN` and `PLACE` are the
  contract. A producer using different IDs gets runners with zero prices rather
  than an error. Exported constants keep it in one place, but the schema does not
  enforce it.
- **`FieldSize` recomputes only when an update touches runners.** If a producer
  ever expressed a scratching some other way (for example by suspending the
  runner's selections instead of setting `Scratched`), the count would not
  update. That is a deliberate trigger choice, documented rather than defended
  as universal.
- **No results transform.** `FinishPosition` and `RaceStatus` are producer-set.
  Deriving settlement from them is a separate concern and was out of scope.
- **`GetRacingEvent` on a sport event returns empty racing fields** rather than
  an error, matching how `GetSportEvent` already behaves on a missing event. If
  stricter typing were wanted, `EventTypeID` is the natural thing to validate on.
