function handleQuizStart(event, mode) {
  event.preventDefault();

  const quizSetupSection = document.getElementById('quizSetupSection');
  const form = quizSetupSection.querySelector('form');
  const gameTypeSelect = quizSetupSection.querySelector('#game-type');
  const titleHeading = quizSetupSection.querySelector('#quizTitleHeading');
  const startG = document.getElementById('start-quiz')
  const createG = document.getElementById('create-game-room')

  // Reset form
  if (form) {
    form.reset();
  }

    // Set game type
    if (mode === 'solo') {
      gameTypeSelect.value = '1';
      //roomNameInput.style.display = 'none';
      if (titleHeading) {
        titleHeading.textContent = 'Set up your Singleplayer Quiz';
        createG.classList.add('hidden');
        startG.classList.remove('hidden');
      }
    } else {
      gameTypeSelect.value = '2';
      //  roomNameInput.style.display = 'block';
      if (titleHeading) {
        titleHeading.textContent = 'Set up your Multiplayer Quiz';
        startG.classList.add('hidden');
        createG.classList.remove('hidden');
      }
    }

    quizSetupSection.style.display = 'block';
}

