# Bow Game v2 – Math Quest Archery

## Product Overview

**Bow Game v2 – Math Quest** is an educational mini-game integrated into the KidTyping VN platform. It combines archery gameplay with elementary arithmetic practice, targeting children aged 6–12. Players solve math problems to fire arrows at a boss enemy; correct answers trigger an automatic, perfectly-aimed shot.

---

## Goals

| Goal | Description |
|------|-------------|
| **Learning** | Reinforce addition, subtraction, multiplication, and division through active gameplay |
| **Engagement** | Motivate repeated practice via score tracking, round progression, and boss defeat animations |
| **Accessibility** | No complex keyboard controls — just type the answer and press Enter |

---

## Access

Bow Game v2 is available from the top navigation bar:
**🎮 Games ▾ → 🧮 Bắn Cung v2**

It is a standalone sub-menu item, separate from the original Bow Game (v1).

---

## Gameplay Mechanics

### Math Questions
Each question is randomly selected from four operation types:

| Operation | Constraint | Example |
|-----------|-----------|---------|
| Addition (`+`) | Both operands ≤ 50, result < 100 | `34 + 47 = ?` |
| Subtraction (`−`) | Minuend ≤ 99, result > 0, no negative answers | `82 − 35 = ?` |
| Multiplication (`×`) | Both operands in range [1, 9] (times table ≤ 9×9) | `7 × 8 = ?` |
| Division (`÷`) | Derived from times table — always a whole number | `56 ÷ 7 = ?` |

All four operation types appear with equal probability. Questions are replaced after each shot (correct) or after a 1.8-second delay (wrong answer).

### Answering
- The player types the numeric answer in the input field.
- Press **Enter** or click **🏹 Bắn!** to submit.

### Correct Answer
1. Input is immediately disabled to prevent double-submission.
2. Feedback `"🎉 Chính xác! Bắn!"` is displayed.
3. After a 260 ms animation delay, the character fires an arrow.
4. The arrow follows a **fixed 25° launch angle** with auto-calculated power.
5. The arrow always hits the boss (guaranteed by physics formula).
6. Boss loses **20 HP** per hit.
7. A new question appears once the arrow lands.

### Wrong Answer
1. Input shakes and turns red.
2. Feedback shows the correct answer: `"❌ Sai rồi! Đáp án đúng là: X"`.
3. After 1.8 seconds, a new question auto-loads.
4. **Boss HP is unchanged** — wrong answers do not cause damage.
5. The wrong attempt is counted toward total questions (affects accuracy %).

---

## Auto-Aim Physics

The character fires at a **fixed angle of 25°**. The launch speed is computed analytically each round using the projectile formula:

$$v_0 = \sqrt{\dfrac{0.5 \cdot g \cdot \Delta x^2}{(\Delta y + \Delta x \cdot \tan\theta) \cdot \cos^2\theta}}$$

Where:
- `g = 0.4 px/frame²` (gravity constant)
- `Δx = targetX − startX` (horizontal distance to boss center)
- `Δy = targetY − startY` (vertical distance, negative = boss above)
- `θ = 25°` (fixed launch angle)

This guarantees the arrow hits the boss regardless of player position.

---

## Boss & Win Condition

- Boss starts with **100 HP**.
- Each correct answer deals **20 HP** damage (5 arrows needed per round).
- At 0 HP: win screen displays with score + fireworks, then the next round begins automatically after 3.3 seconds.
- The boss resets to full HP each new round; rounds are infinite.

---

## Scoring

```
Round Score = max(100, 500 + correctAnswers × 50 − wrongAnswers × 30)
```

| Metric | Value |
|--------|-------|
| Base score | 500 per round |
| Correct answer bonus | +50 per correct answer |
| Wrong answer penalty | −30 per wrong answer |
| Minimum round score | 100 |

- **Session score** accumulates across rounds.
- **High score** is persisted to `localStorage` key `bow2_highscore`.
- **Accuracy %** is displayed on the canvas HUD: correct / total × 100.

---

## HUD & UI

### Canvas HUD (top of canvas)
| Position | Content |
|----------|---------|
| Top-left | `⭐ {score}  🏆 {highscore}  Vòng {round}` |
| Top-right | `✅ {correct}/{total}  ĐCN: {accuracy}%` |

### Stats Panel (below canvas)
| Element | Description |
|---------|-------------|
| ❓ X câu | Total questions attempted this session |
| ✅ X đúng | Correct answers this session |
| 💥 HP Boss | Current boss health (0–100) |
| ⭐ Điểm | Current session score |
| 🏆 Kỷ lục | All-time high score (localStorage) |
| 🔁 Vòng | Current round number |

---

## Game Controls

| Action | Control |
|--------|---------|
| Type answer | Click input field and type number |
| Submit answer | Press **Enter** or click **🏹 Bắn!** |
| Restart game | Click **🔄 Chơi lại** |
| Exit to home | Click **← Quay lại** |

---

## Technical Notes

| Item | Detail |
|------|--------|
| Canvas size | 1000 × 500 px |
| Physics engine | Euler integration, 60 fps via `requestAnimationFrame` |
| Gravity | 0.4 px/frame² |
| Arrow damage | 20 HP per hit |
| Wind | None (v2 is wind-free for simplicity) |
| localStorage key | `bow2_highscore` |
| Character position | Fixed at x=80 (no movement in v2) |
| Boss position | Fixed at x=850 |

---

## QA Test Results

| Test Case | Expected | Result |
|-----------|----------|--------|
| Games dropdown shows both v1 and v2 | `["🏹 Bắn Cung", "🧮 Bắn Cung v2"]` | ✅ PASS |
| Navigate to Bow Game v2 | Page visible, canvas renders, question displayed | ✅ PASS |
| Correct answer fires arrow | Input disabled, "Chính xác!" shown, arrow flies | ✅ PASS |
| Arrow hits boss after correct answer | Boss HP −20 after each correct answer | ✅ PASS |
| 5 correct answers defeat boss | HP 100→80→60→40→20→0, win screen, round++ | ✅ PASS |
| Wrong answer shows correct answer | "❌ Sai rồi! Đáp án đúng là: X" displayed | ✅ PASS |
| Wrong answer generates new question | New question appears after 1.8s, boss HP unchanged | ✅ PASS |
| Back button returns to home | Home page visible, game page hidden, game loop stopped | ✅ PASS |
| Restart resets all state | HP=100, score=0, round=1, qCount=0, correct=0 | ✅ PASS |
| High score persists across restarts | `localStorage["bow2_highscore"]` preserved | ✅ PASS |

---

## Future Enhancements (Backlog)

| Priority | Feature |
|----------|---------|
| High | Difficulty levels (Easy: add/sub only; Medium: ×/÷ included; Hard: larger numbers) |
| High | Timer per question (creates urgency, prevents looking up answers) |
| Medium | Moving boss target (increases engagement for repeat players) |
| Medium | Combo multiplier (consecutive correct answers boost score) |
| Medium | Online leaderboard integration (reuse existing `/api/ranking` endpoint) |
| Low | Sound effects on hit, miss, and win |
| Low | Animated arrow trail |
| Low | Mobile touch input optimization |
