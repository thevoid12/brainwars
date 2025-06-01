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

 document.addEventListener("DOMContentLoaded", function () {
    const topicInput = document.getElementById("topic");
    if (!topicInput) return; // Defensive: exit if input not found
    const checkbox = document.getElementById("random-topic");

    const topicInputWrapper = document.getElementById("topic").closest("div");

    function toggleTopicInput() {
      const shouldHide = checkbox.checked;
      topicInputWrapper.style.display = shouldHide ? "none" : "block";
      topicInput.disabled = shouldHide; // Prevent submission errors
    }

    toggleTopicInput(); // Initial toggle on load
    checkbox.addEventListener("change", toggleTopicInput);
  });