const express = require('express');
const cors = require('cors');
const path = require('path');
const fs = require('fs');

const app = express();
const PORT = process.env.PORT || 3000;

// Middleware
app.use(cors());
app.use(express.json());
app.use(express.static(path.join(__dirname, 'public')));

// Load words from file
const WORD_LIST = fs.readFileSync(path.join(__dirname, 'words'), 'utf-8')
  .split('\n')
  .map(word => word.trim().toUpperCase())
  .filter(word => word.length === 5);

// Select a random word on server startup (changes when backend restarts)
const TARGET_WORD = WORD_LIST[Math.floor(Math.random() * WORD_LIST.length)];
console.log(`✓ Loaded ${WORD_LIST.length} words`);
console.log(`Target word for this session: ${TARGET_WORD}`);

// Game state
let gameState = {
  targetWord: TARGET_WORD,
  attempts: [],
  currentAttempt: '',
  gameOver: false,
  won: false
};

// Routes
app.get('/api/game', (req, res) => {
  res.json({
    targetWord: gameState.targetWord,
    attempts: gameState.attempts,
    currentAttempt: gameState.currentAttempt,
    gameOver: gameState.gameOver,
    won: gameState.won
  });
});

app.post('/api/guess', (req, res) => {
  const { guess } = req.body;
  
  if (!guess || typeof guess !== 'string' || guess.length !== 5) {
    return res.status(400).json({ error: 'Guess must be a 5-letter word' });
  }
  
  const uppercaseGuess = guess.toUpperCase();
  
  if (!WORD_LIST.includes(uppercaseGuess)) {
    return res.status(400).json({ error: 'Word not in dictionary' });
  }
  
  if (gameState.gameOver) {
    return res.status(400).json({ error: 'Game is already over' });
  }
  
  // Process the guess
  const feedback = [];
  const targetLetters = gameState.targetWord.split('');
  const guessLetters = uppercaseGuess.split('');
  const letterCounts = {};
  
  // Count letters in target word
  for (let letter of targetLetters) {
    letterCounts[letter] = (letterCounts[letter] || 0) + 1;
  }
  
  // First pass: mark correct positions (green)
  for (let i = 0; i < 5; i++) {
    if (guessLetters[i] === targetLetters[i]) {
      feedback[i] = { letter: guessLetters[i], status: 'correct' };
      letterCounts[guessLetters[i]]--;
    }
  }
  
  // Second pass: mark present but wrong position (yellow) and absent (gray)
  for (let i = 0; i < 5; i++) {
    if (!feedback[i]) {
      if (letterCounts[guessLetters[i]] > 0) {
        feedback[i] = { letter: guessLetters[i], status: 'present' };
        letterCounts[guessLetters[i]]--;
      } else {
        feedback[i] = { letter: guessLetters[i], status: 'absent' };
      }
    }
  }
  
  gameState.attempts.push(feedback);
  gameState.currentAttempt = '';
  
  // Check win/lose conditions
  if (uppercaseGuess === gameState.targetWord) {
    gameState.gameOver = true;
    gameState.won = true;
  } else if (gameState.attempts.length >= 6) {
    gameState.gameOver = true;
    gameState.won = false;
  }
  
  res.json({
    feedback,
    attempts: gameState.attempts,
    gameOver: gameState.gameOver,
    won: gameState.won,
    targetWord: gameState.targetWord
  });
});

app.post('/api/reset', (req, res) => {
  // Reset game state with a new random word
  const newTargetWord = WORD_LIST[Math.floor(Math.random() * WORD_LIST.length)];
  console.log(`New target word: ${newTargetWord}`);
  
  gameState = {
    targetWord: newTargetWord,
    attempts: [],
    currentAttempt: '',
    gameOver: false,
    won: false
  };
  
  res.json({ message: 'Game reset', targetWord: newTargetWord });
});

// Serve the frontend
app.get('/', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

app.listen(PORT, () => {
  console.log(`Wordle clone server running on port ${PORT}`);
  console.log(`Target word: ${TARGET_WORD}`);
});