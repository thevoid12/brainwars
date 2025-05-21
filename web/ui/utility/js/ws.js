window.onload = function () {
  if (window["WebSocket"]) {
    let roomcode = document.getElementById("ws-container").dataset.roomcode;
    let gameType = document.getElementById("ws-container").dataset.gametype;
    let lobbyPlayers = {};
    let playerListEl = document.getElementById("player-list");
    let readyGameBtn = document.getElementById("ready-game-btn");
    let startGameBtn = document.getElementById("start-game-btn");
    let leaveRoomBtn = document.getElementById("leave-room-btn");

    // Chat UI Elements
    let chatMessagesEl = document.getElementById("chat-messages");
    let chatInputEl = document.getElementById("chat-input");
    let sendChatBtn = document.getElementById("send-chat-btn");
    let chatErrorEl = document.getElementById("chat-error");

    console.log("WebSocket is supported");
    let protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
    let conn = new WebSocket(protocol + window.location.host + "/bw/ws?roomCode=" + encodeURIComponent(roomcode));

    conn.onopen = function (e) {
      console.log("Connection established!");
      // This initial message might be displayed by the server as a system message in chat
      var payload = { data: "Welcome All! The game is about to begin.", time: new Date().toISOString() }
      // The server should decide if this "send_message" type is broadcasted as a chat message
      // or if there's a specific system message type.
      // For this example, we'll assume the server might convert this to a chat message.
      let data = JSON.stringify({ type: "send_message", payload: payload });
      conn.send(data);
    };

    conn.onmessage = function (e) {
      const data = JSON.parse(e.data);

      if (data.type === "lobby_state" && gameType === "MULTI_PLAYER") {
        lobbyPlayers = {}; // reset
        data.payload.forEach(player => {
          lobbyPlayers[player.username] = player.data;
        });
        renderLobbyPlayers();
      } else if (data.type === "joined_game" && gameType === "MULTI_PLAYER") {
        const username = data.payload.username;
        lobbyPlayers[username] = "joined";
        renderLobbyPlayers();
      } else if (data.type === "ready_game" && gameType === "MULTI_PLAYER") {
        const username = data.payload.username;
        lobbyPlayers[username] = "ready";
        renderLobbyPlayers();
      } else if (data.type === "start_game") {
        if (gameType === "MULTI_PLAYER") {
          const lobbyContainer = document.getElementById("lobby-container");
          if (lobbyContainer) {
            lobbyContainer.classList.add("hidden");
          }
        }
      } else if (data.type === "new_question") {
        renderQuestion(data.payload);
      } else if (data.type === "end_game") {
        renderEndGame(data.payload);
      } else if (data.type === "leaderboard") {
        renderLeaderboard(data.payload.scores);
      } else if (data.type === "game_error") {
        renderGameError(data.payload.errorMessage);
      } else if (data.type === "leave_room") {
        conn.close();
        window.location.href = "/bw/home/";
        return;
      } else if (data.type === "chat_message") { // Handle incoming chat messages
        renderChatMessage(data.payload); // payload should be { username: "user", message: "text" }
      } else if (data.type === "send_message") { 
        // This handles the initial "welcome All!" if server broadcasts it as is,
        // and treats it as a system message.
        // Assumes payload is { data: "message text" }
        if (data.payload && data.payload.data) {
             renderChatMessage({ username: "System", message: data.payload.data });
        }
      }
    };

    conn.onclose = function () {
      console.log("Connection closed!");
      // Avoid redirect if modal is handling it or if it's an unexpected close.
      // For now, keeping the original behavior.
      // Consider showing a message like "Connection lost. Redirecting..."
      renderGameError("Connection to server lost. Redirecting to homepage.");
      setTimeout(() => {
        window.location.href = "/bw/home/";
      }, 3000);
      return;
    };

    // Debounce helpers
    function debounceClick(callback, delay = 500) {
      let lastClick = 0;
      return () => {
        const now = Date.now();
        if (now - lastClick > delay) {
          lastClick = now;
          callback();
        }
      };
    }

    if (gameType === "MULTI_PLAYER" && readyGameBtn) {
      readyGameBtn.classList.remove("hidden");
      readyGameBtn.onclick = debounceClick(() => {
        conn.send(JSON.stringify({ type: "ready_game" }));
      });
    }

    if (gameType === "MULTI_PLAYER" && startGameBtn) {
      startGameBtn.classList.remove("hidden");
      startGameBtn.onclick = debounceClick(() => {
        openModal({ url: '/bw/home/', method: 'ws', body: JSON.stringify({ type: "start_game" }), wsconnection: conn, message: 'Clicking Yes will force start game. Are you sure?' });
      });
    }

    if (gameType === "MULTI_PLAYER" && leaveRoomBtn) {
      leaveRoomBtn.classList.remove("hidden");
      leaveRoomBtn.onclick = () => {
        openModal({ url: '/bw/home/', method: 'ws', body: JSON.stringify({ type: "leave_room" }), wsconnection: conn, message: 'Clicking Yes will redirect you to the homepage. Are you sure?' });
      };
    }

    // Chat Functionality
    if (sendChatBtn && chatInputEl && chatMessagesEl && chatErrorEl) {
      sendChatBtn.addEventListener("click", sendChatMessage);
      chatInputEl.addEventListener("keypress", function(event) {
        if (event.key === "Enter") {
          event.preventDefault();
          sendChatMessage();
        }
      });
    }

    function sendChatMessage() {
      const message = chatInputEl.value.trim();
      chatErrorEl.classList.add("hidden");
      chatErrorEl.textContent = "";

      if (message.length === 0) {
        return; // Don't send empty messages
      }

      if (message.length > 200) {
        chatErrorEl.textContent = "Message is too long (max 200 characters).";
        chatErrorEl.classList.remove("hidden");
        return;
      }

      const chatPayload = {
        type: "chat_message", // Client sends this type to the server
        payload: {
          message: message
        }
      };
      conn.send(JSON.stringify(chatPayload));
      chatInputEl.value = ""; // Clear input field
    }

    function renderChatMessage(payload) {
      if (!chatMessagesEl || !payload || !payload.username || !payload.message) {
        console.error("Invalid payload for renderChatMessage:", payload);
        return;
      }
      const { username, message } = payload;

      const messageWrapper = document.createElement('div');
      messageWrapper.classList.add('flex', 'items-start'); // items-start for better alignment with multi-line messages

      const avatar = document.createElement('div');
      avatar.classList.add('h-7', 'w-7', 'rounded-full', 'bg-gray-300', 'flex', 'items-center', 'justify-center', 'text-xs', 'font-semibold', 'text-gray-700', 'mr-2', 'flex-shrink-0');
      avatar.textContent = username.slice(0, 2).toUpperCase();

      const messageContent = document.createElement('div');
      messageContent.classList.add('bg-gray-100', 'rounded-md', 'p-2', 'max-w-xs', 'sm:max-w-sm', 'md:max-w-md'); // Responsive max width

      const usernameSpan = document.createElement('span');
      usernameSpan.classList.add('font-semibold', 'text-sm', 'block', 'mb-0.5');
      // Differentiate user's own messages if username is available globally, e.g. myUsername
      // if (username === myUsername) {
      //  usernameSpan.classList.add('text-blue-600'); // Example for own messages
      // } else {
      usernameSpan.classList.add(username === "System" ? 'text-purple-600' : 'text-primary-600');
      // }
      usernameSpan.textContent = username;

      const messageText = document.createElement('span');
      messageText.classList.add('text-sm', 'text-gray-800', 'break-words'); // break-words for long messages
      messageText.textContent = message;

      messageContent.appendChild(usernameSpan);
      messageContent.appendChild(messageText);

      messageWrapper.appendChild(avatar);
      messageWrapper.appendChild(messageContent);

      chatMessagesEl.appendChild(messageWrapper);
      // Scroll to the bottom of the chat messages
      chatMessagesEl.scrollTop = chatMessagesEl.scrollHeight;
    }


   function renderGameError(errorMessage) {
      const popup = document.createElement('div');
      popup.id = 'errorPopup';
      popup.className = 'fixed top-5 left-1/2 transform -translate-x-1/2 bg-red-100 text-red-700 border border-red-300 rounded-md shadow-md p-4 max-w-md w-full z-50 flex justify-between items-start';
      popup.innerHTML = `
        <div class="flex items-start space-x-2">
          <svg class="w-6 h-6 text-red-500 mt-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-sm font-medium">${errorMessage}</span>
        </div>
        <button id="closeError" class="ml-4 text-sm text-gray-500 hover:text-gray-700">Dismiss</button>
      `;
      
      document.body.appendChild(popup);

      const closeBtn = popup.querySelector('#closeError');
      closeBtn.addEventListener('click', () => {
        popup.remove();
      });

      setTimeout(() => {
            popup.remove();
      }, 10000);
    }

    function renderLobbyPlayers() {
      if (!playerListEl) return;
      playerListEl.innerHTML = ""; // Clear previous entries

      Object.entries(lobbyPlayers).forEach(([username, status]) => {
        const li = document.createElement("li");
        li.className = "flex items-center justify-between bg-gray-100 px-4 py-2 rounded-md shadow-sm"; // Added shadow-sm

        const playerInfoDiv = document.createElement("div");
        playerInfoDiv.className = "flex items-center";

        const avatarDiv = document.createElement("div");
        avatarDiv.className = "h-8 w-8 rounded-full bg-gray-300 flex items-center justify-center text-gray-700 mr-3 font-semibold"; // Added font-semibold
        avatarDiv.textContent = username.slice(0, 2).toUpperCase();

        const usernameSpan = document.createElement("span");
        usernameSpan.className = "text-gray-800"; // Darker text for username
        usernameSpan.textContent = username;

        playerInfoDiv.appendChild(avatarDiv);
        playerInfoDiv.appendChild(usernameSpan);

        const statusSpan = document.createElement("span");
        statusSpan.className = `text-sm font-semibold px-2 py-0.5 rounded-full ${status === 'ready' ? 'text-green-700 bg-green-100' : 'text-yellow-700 bg-yellow-100'}`; // Pill-like status
        statusSpan.textContent = status === 'ready' ? 'Ready' : 'Joined';

        li.appendChild(playerInfoDiv);
        li.appendChild(statusSpan);
        playerListEl.appendChild(li);
      });
    }

    function renderQuestion(payload) {
      const questionBlock = document.getElementById("question-block");
      if (!questionBlock) return;
      const { questionIndex, totalQuestions, question, timeLimit } = payload;
      const { ID: questionID, Question: questionText, Options } = question;

      questionBlock.dataset.questionid = questionID;

      // Calculate completion percent
      const percentComplete = Math.round((questionIndex / totalQuestions) * 100);

      let html = `
        <div class="flex-1 flex flex-col overflow-hidden">
          <div class="bg-white border-b border-gray-200 p-4">
            <div class="flex justify-between items-center">
            <h1 class="text-xl font-bold text-gray-800">Quiz</h1>
              <div class="flex items-center">
              <span class="text-sm text-gray-600 mr-2">Completion</span>
              <div class="w-48 h-2 bg-gray-200 rounded-full overflow-hidden">
                <div class="h-full bg-primary-500 rounded-full" style="width: ${percentComplete}%"></div>
                </div>
              </div>
        <button id="leave-game-btn" class="bg-red-500 hover:bg-red-600 text-white px-6 py-2 rounded-lg transition-colors">End Game</button>
            </div>
          </div>
      
        <div class="flex-1 overflow-y-auto p-6">
            <div class="max-w-3xl mx-auto">
            <div class="bg-white rounded-lg shadow-md p-6 mb-6">
                <div class="flex items-start mb-4">
                <div class="w-8 h-8 rounded-full bg-primary-500 flex items-center justify-center text-white flex-shrink-0 mr-3">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <div>
                  <h3 class="text-lg font-semibold text-gray-800">QuizMaster AI</h3>
                  <p class="text-gray-700 mt-1">${questionText}</p>
                  </div>
                </div>
      
              <div class="ml-11 space-y-3" id="options-container">
      `;

      Options.forEach((opt, index) => {
        const letter = ['A', 'B', 'C', 'D'][index] || '';
        html += `
          <div class="border rounded-md p-3 cursor-pointer transition-colors option-item"
              data-optionid="${opt.ID}">
              <div class="flex items-center">
              <div class="w-6 h-6 rounded-md border border-gray-300 flex items-center justify-center mr-3 text-xs font-medium option-box">
                  ${letter}
                </div>
              <span>${opt.Option}</span>
              </div>
          </div>
        `;
      });

      html += `
                </div>
              </div>
      
            <div class="bg-white rounded-lg shadow-md p-6 mb-6" id="user-response">
                <div class="flex items-start">
                <div class="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center text-gray-700 flex-shrink-0 mr-3">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24"
                      stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                    </svg>
                  </div>
                  <div>
                  <h3 class="text-lg font-semibold text-gray-800">You</h3>
                    <div class="mt-2 flex items-center">
                    <div class="text-2xl font-bold text-primary-600" id="timer-display">${timeLimit * 60}</div>
                    <svg class="animate-spin ml-2 h-5 w-5 text-primary-500"
                        xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor"
                          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                        </path>
                      </svg>
                    </div>
                    <div class="mt-4">
                    <button class="bg-primary-500 hover:bg-primary-600 text-white py-2 px-4 rounded-md transition-colors" id="next-question-btn">
                        Next Question
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>`;
      questionBlock.innerHTML = html;

      let timeLeft = timeLimit * 60;
      const timerDisplay = document.getElementById("timer-display");
      const timerInterval = setInterval(() => {
        if (timeLeft > 0) {
          timeLeft--;
          if (timerDisplay) {
            timerDisplay.textContent = timeLeft;
          }
        } else {
          clearInterval(timerInterval);
        }
      }, 1000);

      // Option click handlers
      document.querySelectorAll(".option-item").forEach(el => {
        el.addEventListener("click", function () {
          const selectedOptionID = this.dataset.optionid;

          // Style updates
          document.querySelectorAll(".option-item").forEach(opt => {
            opt.classList.remove("border-primary-500", "bg-primary-50");
            opt.querySelector(".option-box").classList.remove("bg-primary-500", "text-white", "border-primary-500");
          });
          this.classList.add("border-primary-500", "bg-primary-50");
          this.querySelector(".option-box").classList.add("bg-primary-500", "text-white", "border-primary-500");
          
          // Show user response area
          // document.getElementById("user-response").style.display = "block";

          const answerPayload = {
            type: "submit_answer",
            payload: {
              questionDataID: questionID,
              answerOption: parseInt(selectedOptionID),
            }
          };
          conn.send(JSON.stringify(answerPayload));
        });
      });

      // Next button handler
      const nextBtn = document.getElementById("next-question-btn");
      if (nextBtn) {
        nextBtn.addEventListener("click", function () {
          clearInterval(timerInterval);
          const questionID = parseInt(questionBlock.dataset.questionid);
          const nextQuestionPayload = {
            type: "next_question",
            payload: {
              questionID: questionID
            }
          };
          conn.send(JSON.stringify(nextQuestionPayload));
        });
      }

      // leave game inbetween
      const leaveGameBtn = document.getElementById("leave-game-btn");
      if (leaveGameBtn) {
        leaveGameBtn.addEventListener("click", () => {
          console.log("Leave game button clicked from renderQuestion");
          openModal({
            url: '/bw/home/',
            method: 'ws',
            body: JSON.stringify({ type: 'leave_room' }),
            wsconnection: conn,
            message: "Are you sure you want to leave the game? You won't be able to rejoin."
          });
        });
      }
    }


    function renderLeaderboard(scoreList) {
      const leaderboardList = document.getElementById("leaderboard-list");
      leaderboardList.innerHTML = ""; // Clear previous entries

      let tableHTML = `
        <div class="bg-white border-t border-gray-200 p-4">
          <h2 class="text-lg font-semibold mb-3">Live Leaderboard</h2>
          <div class="overflow-x-auto">
            <table class="min-w-full">
              <thead>
                <tr class="bg-gray-50">
                  <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Rank</th>
                  <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">User</th>
                  <th class="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Score</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
      `;

        scoreList.forEach((entry, index) => {
          tableHTML += `
            <tr>
            <td class="px-4 py-2 whitespace-nowrap text-sm font-medium text-gray-900">${index + 1}</td>
            <td class="px-4 py-2 whitespace-nowrap text-sm text-gray-700">
                <div class="flex items-center">
                <div class="h-8 w-8 rounded-full bg-gray-200 flex items-center justify-center text-gray-700 mr-2">
                    <span>${entry.username.slice(0, 2).toUpperCase()}</span>
                  </div>
                  <span>${entry.username}</span>
                </div>
              </td>
            <td class="px-4 py-2 whitespace-nowrap text-sm text-gray-700">${entry.score}</td>
          </tr>
        `;
        });

      tableHTML += `
              </tbody>
            </table>
          </div>
        </div>
      `;

      leaderboardList.innerHTML = tableHTML;
    }

    // Add the canvas-confetti script
    const confettiScript = document.createElement('script');
    confettiScript.src = 'https://cdn.jsdelivr.net/npm/canvas-confetti@1.5.1/dist/confetti.browser.min.js';
    // confettiScript.async = true;
    document.head.appendChild(confettiScript);

    function renderEndGame(payload) {
      const questionBlock = document.getElementById("question-block");
      if (!questionBlock) return;
      const { message, scores, finishTime } = payload;

    // Start confetti
        startConfetti();

    let html = `
      <div class="flex-1 flex flex-col items-center justify-center p-8">
    
        <div class="bg-white rounded-lg shadow-xl p-8 max-w-2xl w-full">
          <div class="text-center mb-8">
            <h2 class="text-3xl font-bold text-primary-600 mb-2">${message}</h2>
            <p class="text-gray-600">Game finished on ${new Date(finishTime).toLocaleString()}</p>
          </div>

          <div class="space-y-6">
            ${scores.length > 0 ? `
              <div class="flex flex-col">
                ${scores.map((score, index) => `
                  <div class="flex items-center p-4 ${index === 0 ? 'bg-yellow-50' : ''} rounded-lg mb-2">
                    <div class="w-12 h-12 rounded-full bg-primary-100 flex items-center justify-center mr-4">
                      <span class="text-2xl font-bold ${index === 0 ? 'text-yellow-500' : 'text-primary-600'}">
                  ${index + 1}
                </span>
              </div>
                    <div class="flex-1">
                      <h3 class="font-semibold text-lg">${score.username}</h3>
                      <p class="text-gray-600">Score: ${score.score}</p>
                    </div>
                    ${index === 0 ? `
                      <div class="ml-4">
                        <svg class="w-8 h-8 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"/>
                        </svg>
                      </div>
                    ` : ''}
                  </div>
                `).join('')}
              </div>
            ` : `
              <div class="text-center text-gray-500">
                No scores available
              </div>
            `}
          </div>
              <div class="mt-8 text-center">
            <button onclick="openModal({ url: '/bw/home/', method: 'GET',  message: 'Clicking Yes redirect you to homePage. Are you sure?' })" class="bg-red-500 hover:bg-red-600 text-white px-6 py-2 rounded-lg transition-colors">
             End Game
            </button>
          </div>
          <div class="mt-8 text-center">
            <button onclick="openModal({ url: '/bw/analyze/', method: 'GET',body: { roomCode: '${roomcode}' },  message: 'Clicking Yes will move you out of the game room. Are you sure?' })" class="bg-primary-500 hover:bg-primary-600 text-white px-6 py-2 rounded-lg transition-colors">
             Analyze Results 
            </button>
          </div>
        </div>
      </div>
    `;

    html+=`<!-- Modal Backdrop -->
  <div id="modal-backdrop"
     class="fixed inset-0 bg-black bg-opacity-50 hidden z-50 flex items-center justify-center">
              <!-- Modal Box -->
      <div class="bg-white p-6 rounded-2xl shadow-2xl text-center w-80">
        <p id="modal-message" class="mb-4 text-lg font-semibold">Are you sure?</p>
        <div class="flex justify-center gap-4">
          <button onclick="confirmYes()" class="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700">Yes</button>
          <button onclick="closeModal()" class="px-4 py-2 bg-gray-300 text-black rounded hover:bg-gray-400">Cancel</button>
        </div>
      </div>
</div>`
    questionBlock.innerHTML = html;
    }

  // Add this confetti function
    function startConfetti() {
      const duration = 4000; // Increased from 3000 to 5000ms
      const end = Date.now() + duration;

      (function frame() {
        confetti({
          particleCount: 5, // Increased from 2 to 5
          angle: 60,
          spread: 55,
          origin: { x: 0 },
          colors: ['#ff0000', '#00ff00', '#0000ff']
        });
        confetti({
          particleCount: 5, // Increased from 2 to 5
          angle: 120,
          spread: 55,
          origin: { x: 1 },
          colors: ['#ff0000', '#00ff00', '#0000ff']
        });

        if (Date.now() < end) {
          requestAnimationFrame(frame);
        }
      }());
    }
  } else {
    // Fallback for browsers that don't support WebSockets
    const errorDiv = document.createElement('div');
    errorDiv.style.padding = '20px';
    errorDiv.style.backgroundColor = '#ffdddd';
    errorDiv.style.border = '1px solid #ff0000';
    errorDiv.style.textAlign = 'center';
    errorDiv.style.fontSize = '16px';
    errorDiv.textContent = "Sorry, your browser does not support WebSockets, which are required for this game to function. Please try a different browser.";
    document.body.innerHTML = ''; // Clear the body
    document.body.appendChild(errorDiv);
    // alert("WebSockets are not supported in this browser."); // Avoid alert
  }
}
