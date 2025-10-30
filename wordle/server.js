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

// Create games directory if it doesn't exist
const GAMES_DIR = path.join(__dirname, 'games');
if (!fs.existsSync(GAMES_DIR)) {
  fs.mkdirSync(GAMES_DIR);
}

// Load words from file
const WORD_LIST = fs.readFileSync(path.join(__dirname, 'words'), 'utf-8')
  .split('\n')
  .map(word => word.trim().toUpperCase())
  .filter(word => word.length === 5);

// In-memory game state per IP
const activeGames = {};

// Helper to get client IP
const getClientIp = (req) => {
  return req.headers['x-forwarded-for']?.split(',')[0].trim() || 
         req.socket.remoteAddress || 
         'unknown';
};

// Helper to get or create game state for IP
const getGameState = (ip) => {
  if (!activeGames[ip]) {
    const randomWord = WORD_LIST[Math.floor(Math.random() * WORD_LIST.length)];
    activeGames[ip] = {
      targetWord: randomWord,
      attempts: [],
      gameOver: false,
      won: false,
      startTime: new Date().toISOString()
    };
  }
  return activeGames[ip];
};

// Helper to load games for IP from disk
const loadGamesForIp = (ip) => {
  const file = path.join(GAMES_DIR, `${ip}.json`);
  if (fs.existsSync(file)) {
    return JSON.parse(fs.readFileSync(file, 'utf-8'));
  }
  return [];
};

// Helper to save game to disk
const saveGame = (ip, game) => {
  const file = path.join(GAMES_DIR, `${ip}.json`);
  const games = loadGamesForIp(ip);
  games.push({
    ...game,
    endTime: new Date().toISOString()
  });
  fs.writeFileSync(file, JSON.stringify(games, null, 2));
};

console.log(`✓ Loaded ${WORD_LIST.length} words`);

// Routes
app.get('/api/game', (req, res) => {
  const ip = getClientIp(req);
  const game = getGameState(ip);
  res.json({
    attempts: game.attempts,
    gameOver: game.gameOver,
    won: game.won
  });
});

app.post('/api/guess', (req, res) => {
  const ip = getClientIp(req);
  const { guess } = req.body;
  const game = getGameState(ip);
  
  if (!guess || typeof guess !== 'string' || guess.length !== 5) {
    return res.status(400).json({ error: 'Guess must be a 5-letter word' });
  }
  
  const uppercaseGuess = guess.toUpperCase();
  
  if (!WORD_LIST.includes(uppercaseGuess)) {
    return res.status(400).json({ error: 'Word not in dictionary' });
  }
  
  if (game.gameOver) {
    return res.status(400).json({ error: 'Game is already over' });
  }
  
  // Process the guess
  const feedback = [];
  const targetLetters = game.targetWord.split('');
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
  
  game.attempts.push(feedback);
  
  // Check win/lose conditions
  if (uppercaseGuess === game.targetWord) {
    game.gameOver = true;
    game.won = true;
  } else if (game.attempts.length >= 6) {
    game.gameOver = true;
    game.won = false;
  }
  
  const response = {
    feedback,
    attempts: game.attempts,
    gameOver: game.gameOver,
    won: game.won
  };
  
  // Only reveal targetWord when game is over
  if (game.gameOver) {
    response.targetWord = game.targetWord;
    // Save completed game to disk
    saveGame(ip, {
      targetWord: game.targetWord,
      attempts: game.attempts,
      won: game.won,
      startTime: game.startTime
    });
    // Clear the active game so next request gets a new one
    delete activeGames[ip];
  }
  
  res.json(response);
});

app.post('/api/reset', (req, res) => {
  const ip = getClientIp(req);
  // Clear the active game for this IP
  delete activeGames[ip];
  res.json({ message: 'Game reset' });
});

// Get game history for current IP
app.get('/api/history', (req, res) => {
  const ip = getClientIp(req);
  const games = loadGamesForIp(ip);
  res.json({ ip, games });
});

// Serve the frontend
app.get('/', (req, res) => {
  res.sendFile(path.join(__dirname, 'public', 'index.html'));
});

app.listen(PORT, () => {
  console.log(`Wordle server running on port ${PORT}`);
});