let currentScore = 0;
let displayedScore = 0;
let animationFrame;

async function updateScore() {
    const response = await fetch('/score');
    const data = await response.json();
    animateScore(data.score);
    updateUI(data);
}

function animateScore(newScore) {
    currentScore = newScore;
    if (!animationFrame) {
        requestAnimationFrame(animate);
    }
}

function animate() {
    if (displayedScore < currentScore) {
        displayedScore += Math.ceil((currentScore - displayedScore) / 10);
        document.getElementById('score').innerText = displayedScore;
        animationFrame = requestAnimationFrame(animate);
    } else {
        animationFrame = null;
    }
}

function updateUI(data) {
    document.getElementById('score').innerText = data.score;
    document.getElementById('autoClicks').innerText = data.autoClicker1 + data.autoClicker10*10 + data.autoClicker1000*1000 + data.autoClicker120*120 + data.autoClicker5000*5000;
    document.getElementById('autoClicker1').innerText = data.autoClicker1;
    document.getElementById('autoClicker120').innerText = data.autoClicker120;
    document.getElementById('autoClicker1000').innerText = data.autoClicker1000;
    document.getElementById('autoClicker10').innerText = data.autoClicker10;
    document.getElementById('autoClicker5000').innerText = data.autoClicker5000;
    document.getElementById('clickPowerUpgrades').innerText = data.clickPowerUpgrades;
    document.getElementById('buyAutoClicker1').innerText = `CAMPFIRE +1 ENERGY/S (BUY, ${data.autoClicker1Price})`;
    document.getElementById('buyAutoClicker1').disabled = data.score < data.autoClicker1Price;
    document.getElementById('buyAutoClicker10').innerText = `FARM +10 ENERGY/S (BUY, ${data.autoClicker10Price})`;
    document.getElementById('buyAutoClicker10').disabled = data.score < data.autoClicker10Price;
    document.getElementById('buyAutoClicker120').innerText = `ANIMAL FARM +120 ENERGY/S (BUY, ${data.autoClicker120Price})`;
    document.getElementById('buyAutoClicker120').disabled = data.score < data.autoClicker120Price;
    document.getElementById('buyAutoClicker1000').innerText = `WINDMILL +1000 ENERGY/S (BUY, ${data.autoClicker1000Price})`;
    document.getElementById('buyAutoClicker1000').disabled = data.score < data.autoClicker1000Price;
    document.getElementById('buyAutoClicker5000').innerText = `FACTORY +5000 ENERGY/S (BUY, ${data.autoClicker5000Price})`;
    document.getElementById('buyAutoClicker5000').disabled = data.score < data.autoClicker5000Price;
    let progress1 = (data.score / data.autoClicker1Price) * 100;
    let progress10 = (data.score / data.autoClicker10Price) * 100;
    let progress120 = (data.score / data.autoClicker120Price) * 100;
    let progress1000 = (data.score / data.autoClicker1000Price) * 100;
    let progress5000 = (data.score / data.autoClicker5000Price) * 100;
    document.getElementById('progress-fill1').style.width = `${progress1}%`;
    document.getElementById('progress-fill10').style.width = `${progress10}%`;
    document.getElementById('progress-fill120').style.width = `${progress120}%`;
    document.getElementById('progress-fill1000').style.width = `${progress1000}%`;
    document.getElementById('progress-fill5000').style.width = `${progress5000}%`;
    document.getElementById('buyClickPowerUpgrade').innerText = `EARTH CLICKER 2X CLICK POWER (BUY, ${data.clickPowerUpgradePrice})`;
    document.getElementById('buyClickPowerUpgrade').disabled = data.score < data.clickPowerUpgradePrice;
    let clickPowerUpgradeProgress = (data.score / data.clickPowerUpgradePrice) * 100;
    document.getElementById('clickPowerUpgradeProgress-fill').style.width = `${clickPowerUpgradeProgress}%`;

    // Добавляем обновление достижений
    updateAchievements(data.achievements);
}

function updateAchievements(achievements) {
    const list = document.getElementById('achievementsList');
    list.innerHTML = ''; // Очищаем список перед обновлением

    for (const [achievement, unlocked] of Object.entries(achievements)) {
        if (unlocked) { // Проверяем, получено ли достижение
            const li = document.createElement('li');
            li.textContent = achievement;
            list.appendChild(li);
        }
    }
}

async function buyAutoClicker1() {
    await fetch('/buy-autoclicker1', { method: 'POST' });
}

async function buyAutoClicker10() {
    await fetch('/buy-autoclicker10', { method: 'POST' });
}

async function buyAutoClicker120() {
    await fetch('/buy-autoclicker120', { method: 'POST' });
}

async function buyAutoClicker1000() {
    await fetch('/buy-autoclicker1000', { method: 'POST' });
}

async function buyAutoClicker5000() {
    await fetch('/buy-autoclicker5000', { method: 'POST' });
}

async function buyClickPowerUpgrade() {
    await fetch('/buy-click-power-upgrade', { method: 'POST' });
}

document.getElementById('clickButton').addEventListener('click', async function(event) {
    const response = await fetch('/click', { method: 'POST' });
    const data = await response.json();

    let floatingScore = data.clickPower;
    // console.log("Критический:", data.critical);
    
    if (data.critical) {
        floatingScore *= 2;
    }

    createFloatingNumber(event.clientX, event.clientY, floatingScore);
});

function createFloatingNumber(x, y, amount) {
    const number = document.createElement('div');
    number.className = 'floating-number';
    number.innerText = `+${amount}`;
    number.style.left = `${x}px`;
    number.style.top = `${y}px`;
    document.body.appendChild(number);
    setTimeout(() => { number.style.opacity = '0'; number.style.transform = 'translateY(-50px)'; }, 50);
    setTimeout(() => number.remove(), 1000);
}

async function showCritical() {
    const crit = document.getElementById('critical');
    crit.style.display = 'block';
    setTimeout(() => { crit.style.display = 'none'; }, 500);
}

// Подключаем SSE
const eventSource = new EventSource('/events');
eventSource.onmessage = function(event) {
    
    console.log("Получено событие:", event.data); // Проверим, что передается
    const data = JSON.parse(event.data);
    console.log("Achievements received:", data.achievements);
    updateUI(data);
    if (data.critical) {
        showCritical();
    }
};

updateScore(); // Первичное обновление UI
function toggleAchievements() {
    const achievements = document.getElementById('achievements');
    achievements.style.display = achievements.style.display === 'none' || achievements.style.display === '' ? 'block' : 'none';
}

function generateStars() {
    const starContainer = document.getElementById('stars');
    for (let i = 0; i < 100; i++) {
        const star = document.createElement('div');
        star.className = 'star';
        star.style.top = `${Math.random() * 100}vh`;
        star.style.left = `${Math.random() * 100}vw`;
        star.style.animationDuration = `${Math.random() * 3 + 1}s`;
        starContainer.appendChild(star);
    }
}
generateStars();

let statsVisible = false;

async function toggleStats() {
    const response = await fetch('/score');
    const data = await response.json();

    const statsElement = document.getElementById("stats");
    if (!statsElement) {
        console.error("Элемент #stats не найден!");
        return;
    }

    if (statsVisible) {
        statsElement.style.display = "none";
    } else {
        statsElement.innerHTML = `
            <p><strong>Общее количество очков:</strong> ${data.totalScore}</p>
            <p><strong>Потраченные очки:</strong> ${data.totalSpentScore}</p>
            <p><strong>Количество кликов:</strong> ${data.clicks}</p>
            <p><strong>Время в игре:</strong> ${data.playedTime} сек</p>
        `;
        statsElement.style.display = "block";
    }

    statsVisible = !statsVisible;
}

