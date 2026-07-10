const BACKEND_URL = 'https://website-sandbox-1nba.onrender.com/';

function handleAuth(endpoint) {
    const isReg = endpoint === '/register';
    const user = document.getElementById(isReg ? 'reg-user' : 'login-user').value;
    const pass = document.getElementById(isReg ? 'reg-pass' : 'login-pass').value;
    const msgBox = document.getElementById(isReg ? 'reg-msg' : 'login-msg');

    msgBox.style.color = "blue";
    msgBox.innerText = "Processing...";

    fetch(`${BACKEND_URL}${endpoint}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: user, password: pass })
    })
    .then(res => res.json())
    .then(data => {
        msgBox.style.color = data.status === "success" ? "green" : "red";
        msgBox.innerText = data.message;
    })
    .catch(err => {
        msgBox.style.color = "red";
        msgBox.innerText = "Error: Cannot connect to server";
    });
}