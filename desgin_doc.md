- user signs in
- creates a room or joins existing room
- **create a room:**
- create a room with a room name 
- enable/disable request for join (v2)
- room will have unique room code
- chat app with members (v2)
- owner can pick existing pack
- create customized question
- eliminator contest (v3)
- can create their own questions for quiz
- can use gpt generated quiz
- once the round is over we will delete the room but the questions and answers
    can be saved (v2)
- can kick out people from the room 
- ability to rename and assign bots 
- **bot design**
- 2 min 
- 3 min
- 5 min
- 7 min
- they are system users 
- once the round gets over we delete them

- **questions**
- he can choose single player quiz or multiplayer quiz
- if he chooses single player quiz, he can choose a topic,difficulty and number of questions. llm will generate questions for him when he clicks gets started. then after the answers he gets the analysis for mark
- surprise me chooses topic and difficutly and generate quiz for me using llm for that room
- questions are not room bound but user bound. so eventhough he creates question in the room, the same questions can be reused
- if a room owner creates his own questions, then he cant take part in answering the quiz so he becomes a viewer of that quiz
- after the quiz everyone should know what are the right and wrong as well as each guy's time
1. Quiz Modes
Single Player:

User selects topic, difficulty, and number of questions.
LLM generates fresh questions dynamically.
Analysis at the end with score, time taken, and correct/incorrect breakdown.
Optional "Surprise Me" feature where the app selects the topic and difficulty.
Multiplayer (Room-Based)

Private rooms with up to 20 participants.
Room owner creates custom questions or selects auto-generated questions from the LLM.
If the owner creates questions manually, they become a viewer during the quiz.
Results summary at the end shows correct answers, individual performance, and time taken.
2. Question Management
Questions are user-bound (not room-bound), so once a user creates a question, it becomes part of their pool and can be reused across multiple rooms.
LLM-generated questions are temporary and not stored in the user's question pool.
3. Room Flow
Room creation with:
Topic
Number of questions
Difficulty
Question source: LLM-generated or user-created
Live quiz experience where participants answer simultaneously.
Results page showing:
Each question with the right answer
Each participant’s selected option
Time taken per question
Overall leaderboard
4. Additional Ideas
Bot Players: Automatically join empty rooms to make multiplayer fun without waiting.
Friend Invitations: Share room links for quick multiplayer games.
Streak System: Reward players for consecutive correct answers or wins.
Performance Insights: Graph showing how quickly a player answered compared to others.

**bots**:
Dynamic Speed for Realism
Bots should:

Take different times for each question (not constant delay).
Sometimes skip 1-2 questions.
Randomly make mistakes to feel more human.
Add slight delay for the first 2-3 questions (to simulate thinking time).
3. Human-like Names and Avatars
Instead of Bot 1, Bot 2, generate random names like:

Rajesh Kumar
Anita Sharma
Karthik Vasanth
Meera Rajan
Add simple cartoon avatars like animals, emoji faces, or initials.

4. Chatterbots
Bots can send small auto messages in the chat like:

"That was easy!"
"I totally guessed this one!"
"Too tough for me 😓"
"Leaderboard paakalam!" (in Tamil context 😉)
5. Performance-Based Tuning
Let the Room Owner choose bot difficulty when creating the quiz:

Random Bots (Mixed)
Easy Bots
Competitive Bots (Medium + Hard only)
Genius Bots (Only Hard)
6. Reaction Bots
Bots can auto-react if:

They answer fast 🔥
They come to the leaderboard top
They lose streak 😢
7. Bot Customization (Future Scope)
Let room owners choose bot personalities:

Calm
Aggressive
Funny
Lucky Guess Master
8. Bot Stats (After Game)
Show bots in leaderboard like users with:

Right/Wrong answers
Time Taken
Random comment like "I almost won!"
