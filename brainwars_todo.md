- [x] design doc
- [x] set up our stack
- [x] base schema
- [x] create room feature
- [x] users join room feature(web socket)
- [x] static quiz for v1

- [x] when a person leaves the room or got disconnected we need to update room member table

 ## starting quiz 
- [x] all users( including bot) joins the room.
- [x] each question duration should be given while room creation

- [x] all users should click start game (bot not needed) room_members table should have a new col status

- [x] first question comes to all
- [x] once solved he can click next and live leaderboard gets updated

- [x] if all submits before the deadline (other than bots) we move to next question or if the question time gets over
- [x] finally we show whats right and whats wrong answer after the end of quiz
- [x] start the bots thingy when all other players clicked start and just before loading the first question

## todo
- [x] go through 1 full flow with static data, 
  - [x] room creation
  - [x] starting game
  - [x] single player flow
  - [x] going to all the questions until end answer
- [x] build the leader board feature
- [x] start with ui

~~- [ ] exiting in the middle of the game and rejoining~~

~~- [ ] when a person leaves the room or got disconnected we need to update room member table~~

- [x] room members room_id column should be renamed to room_code
- [x] fix room members 
- [x] get the bot times drop down from backend and fill the dropdown

- [x] rename answer id to answer option
- [x] pong not working after some time!
- [x] store data in db and fetch
- [x] leaderboard
- [x] free up memory from the map after things are done both user and bots
that even bot can use the same client whenever the want to send
- [x] display the answers user selected when they click analyze button
- [x] points system needs to change, every other attempts you need to reduce points
- [x] fix leaderboard
- [x] find a way to store all inmemory data to pgsql
- [x] store all the answers in db
- [x] i selected only 1 question but all the questions were visible
- [x] if you update answer the existing answer should be updated not added
- [x] fix final leaderboard
- [x] fix game completion
- [x] fix memory issue (how can we store it in db) and clear inmemory data once used 

- [x] other page displays
- [x] scoring needs to be fixed, if he is trying for the second time we need to do -50
game
- [x] update room meta with the final leader board scores as a score json
- [x] need detailed analytics including time it took to answer each question in analytics page based on the picked answer we need to give final analytics comments 

- [x] add exit game button which will exit you out of the game remove ws
- [x] write a timer which will delete the inmemory cache once in an hour
- [x] find a way on how to check for memory leak
- [ ] check for deadlock and race conditions
- [ ] productionize tailwind remove tailwind.config from layout
- [ ] error rendering on screen pop up
- [ ] validate and sanitize all models
- [ ] tips to imporve using gpt in analysis
- [ ] if i do hard reload after game is over it again goes to the first question
- [ ] close the ws connection when he clicks yes to the modal for game analysis
- [ ] there is some bug with updating state of the room fix it check all the logic and make sure we are updating the db
- [ ] disable refresh in game or atleast get a pop up. upon refreshjing he shouldf be kicked out of the ws and sent to home
- [ ] close all open channels
- [ ] confetti needs to come only for the winning user
- [ ] shouldnt be able to exit without ticking a modal in between 
- [ ] fix multiplayer
- [ ] authentication and session
- [ ] need to figure out websocket reconnection if something is messed up
- [ ] remove egress channel and change it into connection map so 
- [ ] multiplayer
- [ ] go through all fmt.println statements and remove unwanted stuff
- [ ] use safehtml for all the backend data which we are sending to ui
- [ ] add otp ferature in websocket to prevent csrf 
- [ ] xss protection
- [ ] uploading pdf's to generate question
- [ ] handle ui errors display appropriately
- [ ] error rendering on screen pop up
- [ ] validate and sanitize all models
- [ ] auth
- [ ] session
- [ ] if a user leaves before starting a game we need to remove him( ie if his connection is not active or he leaves explicitly)
- [ ] they shouldnt add a bot which takes more time than room each question time
- [ ] there is another checklist shown with the status of others submitted the question or not

## v2
- [ ] move to redis