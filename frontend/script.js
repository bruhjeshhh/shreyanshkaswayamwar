let currentUser = null;

// Register a new user
async function register() {
  const username = document.getElementById('username').value;
  const password = document.getElementById('password').value;

  const response = await fetch('/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });

  const data = await response.json();
  if (data.id) {
    currentUser = data.id;
    document.getElementById('register').style.display = 'none';
    document.getElementById('betting').style.display = 'block';
    loadGirls();
  } else {
    alert('Registration failed');
  }
}

// Load girls to bet on
async function loadGirls() {
  const response = await fetch('/girls');
  const girls = await response.json();
  const girlsList = document.getElementById('girls-list');
  girlsList.innerHTML = girls
    .map(
      (girl) => `
      <li>
        ${girl.name}
        <button onclick="placeBet(${girl.id})">Bet 10 Points</button>
      </li>
    `
    )
    .join('');
}

// Place a bet
async function placeBet(girlId) {
  const response = await fetch('/place-bet', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ userId: currentUser, girlId, betAmount: 10 }),
  });

  const data = await response.json();
  if (data.success) {
    alert('Bet placed successfully!');
  } else {
    alert('Failed to place bet');
  }
}