// join room redirection multiplayer
window.onload = function () {
  const form = document.getElementById("join-room-form");
  if (!form) {
    console.error("join-room-form not found in DOM.");
    return;
  }

  form.addEventListener("submit", function (e) {
    e.preventDefault();
    const roomCode = document.getElementById("roomCode").value.trim();
    if (!roomCode) {
      alert("Please enter a valid room code.");
      return;
    }
    window.location.href = "/bw/ingame/" + encodeURIComponent(roomCode);
  });
};