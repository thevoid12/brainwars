
window.onload = function () {
  if (window["WebSocket"]) {
    let roomcode = document.getElementById("ws-container").dataset.roomcode;
    console.log("WebSocket is supported");
    let protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
    let conn = new WebSocket(protocol + window.location.host + "/bw/ws?roomCode=" + encodeURIComponent(roomcode));

    conn.onopen = function (e) {
      console.log("Connection established!");
      var payload = { data: "welcome boys!", time: new Date().toISOString() }
      let data = JSON.stringify({ type: "send_message", payload: payload });
      conn.send(data);
    };
    conn.onmessage = function (e) {
      console.log(e.data);
      const data = JSON.parse(e.data);
      if (data.type === "new_question") {
        renderQuestion(data.payload);
      } else if (data.type === "end_game") {
        renderEndGame(data.payload);
      } else if (data.type === "leaderboard") {
        renderLeaderboard(data.payload.scores);
      }
      else if (data.type === "game_error") {
        // display the error as pop up
      }
    };
    conn.onclose = function (e) {
      console.log("Connection closed!");
    };

    function renderQuestion(payload) {
      const questionBlock = document.getElementById("question-block");
      const { questionIndex, totalQuestions, question, timeLimit } = payload;
      const { ID: questionID, Question: questionText, Options } = question;

      // Store question ID in data attribute for later
      questionBlock.dataset.questionid = questionID;

      // Calculate completion percent
      const percentComplete = Math.round((questionIndex / totalQuestions) * 100);

      // Build HTML
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
    
            <div class="bg-white rounded-lg shadow-md p-6 mb-6" id="user-response" style="display: none;">
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
                  <p class="text-gray-700 mt-1">Your answer is:</p>
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

      // Timer
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
          document.getElementById("user-response").style.display = "block";

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

    function renderEndGame(payload) {
      const questionBlock = document.getElementById("question-block");
      const { message, scores, finishTime } = payload;

      let html = `
  <div class="endgame-box">
    <h2>${message}</h2>
    <p><strong>Game Finished At:</strong> ${new Date(finishTime).toLocaleString()}</p>
    `;

      if (scores.length === 0) {
        html += `<p>No scores available.</p>`;
      } else {
        html += `<ul>`;
        scores.forEach(score => {
          html += `<li>${score.username}: ${score.score}</li>`;
        });
        html += `</ul>`;
      }

      html += `</div>`;
      questionBlock.innerHTML = html;
    }
  } else {
    alert("WebSockets are not supported in this browser.");
  }
};
