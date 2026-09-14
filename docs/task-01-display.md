# Task 1: Display / Hidden flag

> Add a new value to the system to identify if an Event should be displayed or
> hidden. Ensure this new value is able to be returned on the `GetSportEvent` RPC.

## What changed

| File | Change |
|---|---|
| `model/event.proto` | New `OptionalBool` wrapper message; `Event.Display` field 8 |
| `merger/modelmerge.go` | New `MergeOptionalBool`; `Display` wired into `MergeEvent` |
| `core/core.proto` | `SportEvent.Display` field 12; `reserved 10` |
| `core/package.go` | `Display` populated in `ConvertFromModel` |
| `merger/modelmerge_test.go` | Merge-completeness guard + `Display` merge semantics |
| `core/service/implementation_test.go` | End-to-end `Update` to `GetSportEvent` coverage |
| `exampledata/0{3,4}_*.json` | Payloads for hiding and re-showing an event |

## Verifying

```sh
docker-compose up -d
go test ./merger/... ./core/... -v -run 'Merge|Display'
```

Manually, against a running service (`./run_local.sh`):

1. `Update` with `exampledata/01_new_event.json`, then `GetSportEvent` on
   `testEvent`: `Display` is `false`, nobody has set it, and unset defaults to hidden.
2. `Update` with `04_show_event.json`: `Display` becomes `true`.
3. `Update` with `02_open_selection.json` (an unrelated partial update):
   `Display` stays `true`, it is not clobbered.
4. `Update` with `03_hide_event.json`: `Display` becomes `false` again.

## Decisions

**`OptionalBool` in the model, not a bare `bool`.** A proto3 `bool` defaults to
`false` with no way to distinguish "explicitly hidden" from "nobody has said" or
"leave whatever was there." The merge pipeline needs that distinction so a
partial update that omits `Display` doesn't clobber an existing value. The
codebase already solves this with `OptionalString` / `OptionalInt64` /
`OptionalDouble` / `OptionalBettingStatus`, so `Display` follows the same pattern
in the model rather than inventing a new one.

**Named `Display`.** Task 5 asks to search by "date and/or bettingstatus and/or
display", so `Display` is the name the brief already uses.

**Rendered as a plain `bool` on the RPC, with unset defaulting to `false`
(hidden).** This is the opposite of the usual "unset means unchanged" reading of
`Optional*` fields elsewhere in the model, and it is a deliberate product choice:
new events should not be visible to consumers until something explicitly marks
them displayable, matching how most apps treat visibility. The trade-off is that
a consumer can no longer tell "explicitly hidden" apart from "nobody has an
opinion yet" from the RPC response alone, both read as `false`. That distinction
still exists internally (`model.Event.Display` is nil in one case and
`&OptionalBool{Value: false}` in the other) in case a later consumer needs it,
it just is not surfaced on `core.SportEvent` today.

**`reserved 10` in `core.SportEvent`.** Field numbers in that message jump from
9 to 11. The history of 10 is unknown, so rather than assume it is free, it is
reserved and `Display` takes 12. If an old producer still emits field 10, this
avoids silently decoding its bytes as `Display`.

## The merge-completeness test

`merger/modelmerge.go` is hand written. The `//lint:file-ignore ... Generated
code` comments imply a generator, but there is none in the repo. A field added to
the proto but not to its `Merge` function compiles cleanly and fails silently:
it simply never merges, with no error anywhere.

`TestMergeEvent_MergesEveryField` and its siblings turn that into a test failure.
Each merges a fully populated message into an empty one and asserts the result is
`proto.Equal` to the input, so any field the merge function forgets shows up as a
diff. `assertAllFieldsSet` walks the message descriptor first so that an
incomplete *fixture* cannot mask an incomplete *merge function*.

This guards Tasks 2 and 3 as much as this one: both add fields to these same
messages.

## Notes / limitations

- **This is a breaking default for existing events.** Every event already in
  Redis has never set `Display`, so from the moment this ships, `GetSportEvent`
  reports all of them as hidden until something explicitly sets `Display: true`.
  There is no backfill in this change. If existing events need to stay visible,
  that needs either a migration script or a transform that defaults `Display` to
  `true` the first time an event is touched.
- **`Deleted` is inert.** Every `Optional*` wrapper carries a `Deleted` flag, but
  nothing in the codebase reads it: the merge functions copy it from the
  incoming value and no converter acts on it. `Display` therefore expresses
  hiding through `Value`, not `Deleted`. Making `Deleted` meaningful is a
  codebase-wide change, out of scope here.
- **No transform sets `Display`.** The brief asks only that the value exists and
  is returned. If a rule should derive it (for example, hide events whose markets
  are all closed), that belongs in a transform alongside `sporttransform`.
- **`updategotools.sh` was missing `protoc-gen-go-grpc`**, which both
  `gen-proto.sh` scripts require. Added here as a drive-by fix since Task 1
  cannot regenerate without it; happy to split it out if you'd rather.
