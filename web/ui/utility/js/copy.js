          function copyRoomCode(code) {
            const text = document.getElementById("code-" + code).innerText;
            navigator.clipboard.writeText(text).then(() => {
              const msg = document.getElementById("copied-msg-" + code);
              msg.classList.remove("hidden");
              setTimeout(() => {
                msg.classList.add("hidden");
              }, 1500);
            });
          }
   