# Myfitnesspal

> Disclaimer: I dont care about this looking nice on github. I'm just testing things. Have my API keys bro

Roughly, I want a grpc/golang service that will let me 

1. Create/Delete foods
2. Log foods
3. Let me see my macros for the day

Out of scope

1. Users, I will be the only user
2. Meals, food is daily, not split up
3. Other consumption/workouts i.e. water, steps, supplements

## Milestones

I will first list them out with a basic description and later sections will go into more detail

- [x] Scaffolding: repo structure, docker compose, setup buf, basic go stuff for my service and cli, some make commands even
- [ ] Proto: setup the protobufs and generate stubs
- [ ] Server Skeleton: setup in memory endpoints and print a bit
- [ ] Cli Skeleton: Some basic commands that just call the service
- [ ] Add Postgres: Wire up simple sqlc postgres calls
- [ ] Polish Service: make entire calls make sense and beautiful
- [ ] Polish Cli: make end to end make sense

### Scaffolding

repo structure, docker compose, setup buf, basic go stuff for my service and cli, some make commands even

Things I want:

- docker compose (service, postgres)
- make file with some basic commands (building/running/compose things)
- service and cli directories and basic main functions

What I did and learned:

The common "main" directory where your executeables go is the `cmd` folder.
In this your main package and NOTHING else should live. You can instantiate configs, setup logging, setup your app really.
But no other logic should live here.

I setup a basic docker compose and dockerfile that will build/spin up my service and a postgres database (for later)

I setup a basic makefile that will run my cli, service, and build/run my compose

### Proto

