# 🏹 BowBound Mini — Simple Arrow Shooting Game

## 1. Overview
A single‑player 2D artillery game inspired by *Gunbound*.  
The player controls an archer (left side), adjusts angle and power, and shoots arrows at a large stationary boss (right side).  
The game runs entirely in the browser using **HTML5 Canvas & vanilla JavaScript** (single `.html` file).

- **Theme:** Simple archery with ballistic trajectory
- **View:** Side‑view, flat terrain
- **Goal:** Hit the boss to reduce its HP to 0

---

## 2. Gameplay & Controls

| Action                | Key / Input                | Description                                                                 |
|-----------------------|----------------------------|-----------------------------------------------------------------------------|
| Move left/right       | `ArrowLeft` / `ArrowRight` | Move player horizontally on flat ground (limited range, left area only)     |
| Adjust angle          | `ArrowUp` / `ArrowDown`    | Increase / decrease firing angle (0° – 50°, default = 30°)                  |
| Charge power          | `Space` (hold)             | Fill the power bar while holding                                            |
| Shoot                 | Release `Space`            | Fire arrow with current angle and power                                     |

### Aiming Assistance (new feature)
- **Trajectory preview:** A dotted parabolic line is drawn from the player to the estimated impact point (based on real‑time angle and power).
- **Power bar:** A visual gauge below the canvas (red‑green bar + percentage text) shows current charge level.

---

## 3. Technical Specifications

### 3.1 Physics (simplified frame‑based)
- **Gravity:** `g = 0.4` (px/frame²)
- **Angle:** 0° = straight right, 50° = max upward tilt (input is in **degrees**)
- **Power:** `minPower = 0`, `maxPower = 100`
- **Initial velocity:** `v0 = power * 0.15`
- **Motion per frame (Euler integration):**
  ```javascript
  x += vx
  y += vy
  vy += g

where vx = cos(angleRad) * v0, vy = -sin(angleRad) * v0
3.2 Canvas & Layout

    Canvas size: 1000 × 500 px

    Ground: Flat horizontal line at y = 450

    UI (HTML overlays): angle display, power bar & percentage

3.3 Game States

    aiming – Player can move, change angle, charge power; aiming line is shown

    flying – Arrow is in the air; no input allowed until arrow lands

    (optional) hit / win – brief pause before resetting

4. File Structure (single HTML file)
text

index.html
├── <style>  (minimal CSS for body, canvas, power bar)
├── <canvas id="gameCanvas">
├── <div id="ui-panel">  (angle, power bar, percentage)
└── <script>
    ├── Constants & global variables
    ├── Player, Boss, Arrow objects
    ├── Input handling (keydown/keyup flags)
    ├── Update loop (charging, physics, collision)
    ├── Drawing functions (player, boss, arrow, aim line, ground)
    └── Main game loop (requestAnimationFrame)

5. Detailed Component Design
5.1 Data Structures
javascript

player = {
  x, y,          // bottom‑left anchor
  width, height,
  moveSpeed: 3,
  minX: 30, maxX: 300
}

boss = {
  x, y,          // bottom‑left anchor (large hitbox)
  width: 80, height: 120,
  hp: 100
}

arrow = {
  x, y,          // current position
  vx, vy,        // velocity components
  active: false
}

gameState = "aiming"  // "aiming" | "flying"
currentAngle = 30
currentPower = 0
isCharging = false

5.2 Input System (polling, not event‑repeat dependent)
javascript

const keys = {};
window.addEventListener('keydown', e => {
  keys[e.code] = true;
  if (e.code === 'Space') e.preventDefault();
});
window.addEventListener('keyup', e => {
  keys[e.code] = false;
  // trigger fire when Space is released during aiming
});

In the game loop, read keys to move player / adjust angle / set isCharging.
5.3 Aiming Line (trajectory preview)
javascript

function drawAimLine(ctx, startX, startY, angle, power) {
  if (power <= 0) return;
  let v0 = power * 0.15;
  let rad = angle * Math.PI / 180;
  let vx = Math.cos(rad) * v0;
  let vy = -Math.sin(rad) * v0;
  let simX = startX, simY = startY;
  const g = 0.4;

  ctx.beginPath();
  ctx.strokeStyle = "rgba(255,255,0,0.5)";
  ctx.setLineDash([6, 8]);
  ctx.moveTo(simX, simY);

  for (let i = 0; i < 80; i++) {
    simX += vx;
    simY += vy;
    vy += g;
    ctx.lineTo(simX, simY);
    if (simY > groundY) break;   // stop at ground
  }
  ctx.stroke();
  ctx.setLineDash([]);
}

Called every frame when gameState === "aiming" and currentPower > 0.
5.4 Collision Detection

    Arrow vs Boss: simple AABB (axis‑aligned bounding box) check
    javascript

    if (arrow.x >= boss.x && arrow.x <= boss.x + boss.width &&
        arrow.y >= boss.y && arrow.y <= boss.y + boss.height) {
        // hit!
    }

    Arrow vs ground: arrow.y >= groundY → land & reset

5.5 Win / Lose Condition

    Boss HP ≤ 0 → display “WIN”, lock input

    (Future) Boss could shoot back after a timer

6. UI Updates (sync every frame)
javascript

document.getElementById('angle-display').textContent = currentAngle;
document.getElementById('power-display').textContent = currentPower;
const fill = document.getElementById('power-fill');
fill.style.width = currentPower + '%';
fill.style.background = currentPower > 80 ? '#ff4444' : '#44ff44';

7. Development Steps (Checklist)

    Setup Canvas & basic drawing

        Draw background, ground line, player (stick figure) and boss (red rectangle)

    Player movement & angle control

        Implement LEFT/RIGHT movement and UP/DOWN angle adjustment

        Display angle on UI

    Power charging & aiming line

        Handle SPACE hold/release, fill power bar

        Draw dotted parabolic trajectory preview

    Arrow physics

        Fire arrow with vx, vy, g, update position each frame

        Draw arrow (simple triangle or line)

    Collision & state management

        Detect ground hit → wait 1s → reset to aiming

        Detect boss hit → reduce HP → reset or show win

    Polish

        Smooth animations, screen shake on hit, particle effects (optional)

        Sound effects (optional)

8. Future Enhancements (Out of Scope for MVP)

    Multiple players (turn‑based hot‑seat)

    Moving boss / AI shooting back

    Wind system (affects trajectory)

    Items, different arrow types

    Mobile touch controls