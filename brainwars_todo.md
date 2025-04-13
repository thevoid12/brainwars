- [x] design doc
- [x] set up our stack
- [x] base schema
- [x] create room feature
- [x] users join room feature(web socket)
- [x] static quiz for v1
- [ ] error rendering on screen pop up
- [ ] validate and sanitize all models
- [ ] when a person leaves the room or got disconnected we need to update room member table
- [ ] xss protection

- ## starting quiz 
- [x] all users( including bot) joins the room.
- [x] each question duration should be given while room creation
- [ ] they shouldnt add a bot which takes more time than room each question time
- [x] all users should click start game (bot not needed) room_members table should have a new col status
- [ ] if a user leaves before starting a game we need to remove him( ie if his connection is not active or he leaves explicitly)
- [x] first question comes to all
- [ ] once solved he can click next and live leaderboard gets updated
- [ ] there is another checklist shown with the status of others submitted the question or not
- [x] if all submits before the deadline (other than bots) we move to next question or if the question time gets over
- [ ] finally we show whats right and whats wrong answer after the end of quiz
- [x] start the bots thingy when all other players clicked start and just before loading the first question

## todo
- [x] go through 1 full flow with static data, 
  - [x] room creation
  - [x] starting game
  - [x] single player flow
  - [ ] multp player flow
  - [x] going to all the questions until end answer
- [ ] build the leader board feature
- [ ] start with ui
- [ ] auth
- [ ] session
- [ ] exiting in the middle of the game and rejoining
- [ ] error rendering on screen pop up
- [ ] validate and sanitize all models
- [ ] when a person leaves the room or got disconnected we need to update room member table
- [ ] xss protection
- [ ] uploading pdf's to generate question
- [ ] handle ui errors display appropriately
- [x] room members room_id column should be renamed to room_code
- [x] fix room members 
- [ ] get the bot times drop down from backend and fill the dropdown
-  [ ] use safehtml for all the backend data which we are sending to ui
- [ ] add otp ferature in websocket to prevent csrf 
- [ ] rename answer id to answer option
- [ ] pong not working after some time!
- [ ] store data in db and fetch
- [ ] multiplayer
- [ ] winner user should get confittie
- [ ] go through all fmt.println statements and remove unwanted stuff
- [ ] leaderboard


{"type":"send_message","payload":{"data":"welcome boys!","time":"2025-03-22T13:38:23.65Z","sent":"2025-03-22T19:08:23.675124+05:30"}}
croom:31 {"type":"ready_game","payload":{"data":"User admin is ready","time":"2025-03-22T19:08:23.681868+05:30"}}
croom:31 {"type":"ready_game","payload":{"data":"Bot Sec10 is ready","time":"2025-03-22T19:08:26.826269+05:30"}}
croom:31 {"type":"ready_game","payload":{"data":"Bot Sec15 is ready","time":"2025-03-22T19:08:26.82644+05:30"}}
croom:31 {"type":"new_question","payload":{"questionIndex":1,"totalQuestions":2,"question":{"ID":"c2b15dfd-206c-4118-ac28-5049b4502fce","Question":"this is test question 1","Options":[{"ID":1,"Option":"ans 1"},{"ID":2,"Option":"ans 2"},{"ID":3,"Option":"ans 3"},{"ID":4,"Option":"ans 4"}],"Answer":1},"startTime":"2025-03-22T19:08:36.720943+05:30","timeLimit":2}}
