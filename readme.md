# .env format brainwars
```bash
DATABASE_URL=postgres://postgres:postgres@db:5432/brainwars
ENV=dev
PG_USER=postgres
PG_PASSWORD=postgres
PG_HOST=db
PG_PORT=5432
PG_DB=brainwars
PG_SSLMODE=disable
```
# brainwars v1 design:

- user signs in
- can create a room
- can invite at max 10 people for v1
- can add bots as player 
- bots are 1 min,2 min,3 min, 5 min. randomly picks answer once in selected min
- each correct answer +100
- each wrong answer -50
- live leader board between friends
- owner can create questions or pick from questions pack
- lock room:once game has started you cant join the room
- show countdown when will bots answer next
- show live score update
- use llms to generate questions

#### v2:
- randomly join room
- ability to share the score
- chat between friends
- redis for in meomory instead of go in memory
How Redis Solves the Problem
Imagine you're running two app instances:

Instance A
Instance B
If Player 1 connects to Instance A and Player 2 connects to Instance B, they both belong to the same quiz room.

With Go in-memory maps:

Instance A wouldn't know about Player 2.
Instance B wouldn't know about Player 1.
The game would break.
With Redis:

Both instances would subscribe to the same Redis Pub/Sub channel for the room.
Each message would be broadcasted via Redis and delivered to both instances.
Both players would stay in sync — as if they were on the same server.
