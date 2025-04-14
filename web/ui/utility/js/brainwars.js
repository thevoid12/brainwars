function handleQuizStart(event, mode) {
  event.preventDefault();

  const quizSetupSection = document.getElementById('quizSetupSection');
  const form = quizSetupSection.querySelector('form');
  const gameTypeSelect = quizSetupSection.querySelector('#game-type');
  // const roomNameInput = quizSetupSection.querySelector('#roomName');
  const titleHeading = quizSetupSection.querySelector('#quizTitleHeading');

  // Reset form
  if (form) {
    form.reset();
  }

  if (mode === 'random') {
    window.location.href = '/bw/quick-start'; //todo: fix this
  } else if (mode === 'solo' || mode === 'multiplayer') {
    // Set game type
    if (mode === 'solo') {
      gameTypeSelect.value = '1';
      //roomNameInput.style.display = 'none';
      if (titleHeading) {
        titleHeading.textContent = 'Set up your Singleplayer Quiz';
      }
    } else {
      gameTypeSelect.value = '2';
      //  roomNameInput.style.display = 'block';
      if (titleHeading) {
        titleHeading.textContent = 'Set up your Multiplayer Quiz';
      }
    }

    quizSetupSection.style.display = 'block';
  }
}
