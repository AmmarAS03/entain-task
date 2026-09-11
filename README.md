# Code Test 

This Application is intended to test a developers skills against some common problems faced by the Trading Solutions team. It contains a Model, a merger and a Core service. 

The Tasks in this list are intentionally vague to give the developer freedom to pick the solution they think is best.

With each task take into consideration performance, maintainability Documentation and testing.

To make it easier to review each task could you please add a merge request for each task with any documentation you think appropriate.

## Task 1

Add a New value to the system to identify if an Event should be displayed or hidden.

Ensure this new value is able to be returned on the `GetSportEvent` RPC

## Task 2

The current structure sports Sport Match style events but horse / greyhound / harness racing does not really fit into this structure.

Add a new Racing substructure to the system with some fields you think are appropriate.

Add a new `GetRacingEvent` RPC that returns a structure that would be useful to consumers wanting a racing event.

Consider writing a new Transform if there are racing specific transformations you want to make.

## Task 3

Add a ClosedAt time to markets and set this value in the transform the first time a market is closed.

## Task 4

Swap out the Redis database technology for MongoDB. Remember we are using vendoring so you will need to run `go mod tidy` and `go mod vendor` in the root folder after adding any new dependencies


## Task 5

Write a new `SearchEvents` RPC in Core that allows a user to search by date and/or bettingstatus and/or display.

Return a slice of all events that match the criteria.
 
# Apendix

## Prerequisites
* Go 1.25+ installe
* Protoc > 32 installed
* Docker (you could use another container system but might need to make some minor edits to the scripts)
* run `./updategotools.sh` to install some of the linting and formatting tools used

## Layout
### Model

This contains a proto definition of the Event Model used by the code test. You can run `gen-proto.sh` to regenerate the model.

### Merger

This package merges 2 partial events together, when adding new fields you will need to make sure you have updated the code in this package to merge the new fields correctly.

### Core

The main service of the code test. This spins up a GRPC server and exposes an RPC to `Update` and another more user friendly API to retrieve the event `GetSportEvent`

It persists the data in redis and merges a partial update to an event with the existing copy of the event in the database, runs some transformations on the event and saves back to the database.

## Data Flow
This section will just describe how data flows through this codetest service
1. `Update` RPC called the first time
2. Respository Package called to find existing event - does not exist
3. `transforms.sporttransform.TransformEvent` called to mutate the event
4. Respository Package called to save the updated version of the Event
5. `Update` RPC called with a partial update to the existing event
6. Respository Package called to find existing event - the event is retrieved
7. `merger.MergeEvent` called to merge in the new changes with the existing Event
8. `transforms.sporttransform.TransformEvent` called to mutate the event
9. Respository Package called to save the updated version of the Event

## Running the code

you will need to fix up the import paths to match your local repository structure.

`./run_local.sh` will spin up the docker contains in the docker compose file an start running the core service. It will run a rRPC server on port 50051 and http on 8080 (nothing in the test uses the httpserver)

You can then use postman to hit `localhost:50051` with a grpc request, load in the `./core/core.proto` file and then hit the`update` RPC with a payload like `{"Event": {"ID": "testEvent"}}` 

Switch to the GetSportEvent rpc and look up the event you just created `{"EventID": "testEvent"}`

There are some example payloads in the `./exampledata` folder that you can use to send to the `update` RPC