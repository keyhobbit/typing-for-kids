# 🎒 Roadmap: Sân chơi học tập cho học sinh tiểu học (Lớp 1–5)

> Tài liệu nghiên cứu / kế hoạch mở rộng KidTyping VN từ "luyện gõ + bắn cung"
> thành một **nền tảng nhiều mini-game giáo dục** cho học sinh tiểu học.
> Triết lý: **"Chơi là chính, học lồng vào."**

---

## 1. Hiện trạng & tài sản tái dùng được

Dự án đã có sẵn một "động cơ" game tương đối mạnh. Game mới **không cần làm lại từ đầu**.

| Tài sản | Vị trí | Tái dùng cho game mới |
|---|---|---|
| Canvas + game loop, particle, screen-shake, pháo hoa | `templates/index.html` (mỗi game 1 IIFE) | Khung render chung |
| Âm thanh tổng hợp `Sfx` (Web Audio, không cần file nhạc) | `index.html` | Hiệu ứng đúng / sai / thắng / nâng cấp |
| Boss động + phản ứng `window.KTPaintBoss` / `window.KTBossMeta` | shared global | Mọi game có thể có "trùm" |
| Roguelite: level → 1000, nâng cấp 3-chọn-1, prestige (tái sinh +1%) | bow2 IIFE | Khung tiến trình & độ khó |
| Leaderboard ngày/tuần/tháng/năm | `db.go` (`bow_scores`), `ranker.go` (`rebuildBowPeriod`), `store.go` (`GetBowRanking`) | BXH cho từng game |
| `fmtCompact` (1.2K / 3.4M), độ khó từng kỹ năng, bàn phím số mobile, hamburger menu | `index.html` + `static/css/style.css` | UX chung |
| Auth khách + auto-recover token hết hạn | `apiFetch`, `reauthGuest` | Lưu điểm an toàn |

### Nợ kỹ thuật cần xử lý trước khi nhân rộng
Hiện mỗi game mới phải tự thêm **bảng điểm + handler + ranker riêng** (như `bow_scores`).
Thêm 5–10 game theo cách này sẽ rất trùng lặp và khó bảo trì.

---

## 2. Phase 0 — "SDK mini-game" (nền tảng, làm trước khi nhân rộng)

**Mục tiêu:** thêm 1 game mới chỉ tốn ~1 file logic, gần như không đụng backend.

1. **Tổng quát hoá điểm số**
   Gộp `scores` + `bow_scores` → một bảng `game_scores(game_key, user_id, score, meta_json, scored_at)`.
   `ranker.go` lặp theo `game_key` thay vì viết hàm riêng từng game.
   Giữ tương thích ngược: `game_key = 'typing'`, `'bow2'`.
2. **Game registry (JS)**
   Một mảng `GAMES = [{ key, title, icon, page, start, stop, skill, grade }]`
   → tự sinh menu **Games** + điều hướng, thay cho việc hard-code `showBowGame2()`…
3. **Khung chung `KTGame`**
   Helper cho HUD, preset độ khó, prestige, và `submitScore(gameKey, score, meta)`
   (dùng lại `apiFetch` đã có auto-recover token).
4. **BXH đa game**
   Nút chuyển game trên trang Bảng XH **tự sinh từ registry** (hiện đang hard-code Gõ chữ / Bắn Cung).

---

## 3. Nguyên tắc thiết kế cho tiểu học (áp dụng cho mọi game)

- **Chơi là chính, học lồng vào** — vòng lặp ngắn, thưởng liên tục (sao, lên màn, nâng cấp).
- **Phân hoá theo lớp / kỹ năng** như bow2 (Dễ / Vừa / Khó từng phép). Mặc định an toàn cho Lớp 1.
- **Không trừng phạt nặng** — sai thì gợi ý lại, tránh "game over" đột ngột (giống cơ chế phản đòn non-lethal).
- **Đọc ít, biểu tượng nhiều** — bé Lớp 1 chưa đọc trôi chảy: dùng icon + **đọc to** (Web Speech API `speechSynthesis`, giọng `vi-VN`) cho đề bài.
- **Mobile-first** — bàn phím số / cảm ứng trên màn hình, không phụ thuộc bàn phím vật lý.
- **Accessibility** — tương phản cao, vùng chạm ≥ 44px, hỗ trợ điều khiển bằng phím.

---

## 4. Danh mục game đề xuất (ánh xạ Chương trình GDPT 2018 – tiểu học)

Sắp theo độ ưu tiên = **giá trị học tập × mức tái dùng động cơ × độ dễ làm**.

### Nhóm 1 — Làm nhanh, tái dùng cao (ưu tiên)
1. **🫧 Bong Bóng Vần** (Tiếng Việt – đánh vần / ghép vần)
   Bong bóng chứa âm/vần trôi lên, bé chạm để ghép thành tiếng theo đề. *KN: đọc–viết · Lớp 1–2.*
2. **🔤 Dấu Hỏi–Ngã** (chính tả)
   Chọn dấu đúng cho từ; `speechSynthesis` đọc từ. *KN: chính tả · Lớp 2–3.*
3. **🕐 Xem Giờ / ⏱️ Đơn vị đo / 💵 Đếm tiền VND**
   Toán thực tế ngoài số học thuần (bow2 đã lo +−×÷). *KN: Toán đời sống · Lớp 2–4.*
4. **🧠 Lật Hình Trí Nhớ** (memory match)
   Ghép cặp chữ–hình / số–số lượng / Anh–Việt. *KN: từ vựng + trí nhớ · mọi lớp.*

### Nhóm 2 — Cần thêm asset / logic
5. **🏃 Cuộc Đua Đáp Án** (endless runner) — nhân vật chạy, nhảy qua đáp án đúng, né sai (tái dùng vật lý/loop của bắn cung).
6. **🍎 Hái Quả Từ Vựng** (Tiếng Anh) — nghe từ → hái quả có hình đúng. *KN: nghe–từ vựng.*
7. **🔢 Tìm Quy Luật / So Sánh số** — dãy số, lớn–bé, làm tròn. *KN: logic & số học nâng cao.*
8. **🗺️ Bản đồ Việt Nam** (TN&XH) — chỉ đúng tỉnh/miền, nhận biết con vật / cây.

### Nhóm 3 — Tham vọng (về sau)
Sudoku mini · Gõ nhịp (âm nhạc) · Xếp hình hình học.

---

## 5. Lộ trình theo giai đoạn

| Phase | Nội dung | Kết quả |
|---|---|---|
| **0** | SDK mini-game + tổng quát hoá điểm/BXH + registry | Nền tảng để nhân rộng (bắt buộc trước) |
| **1** | Bong Bóng Vần + Lật Hình Trí Nhớ | Validate SDK với 1 game Tiếng Việt + 1 game trí nhớ |
| **2** | Dấu Hỏi–Ngã + Xem Giờ / Đếm tiền | Tích hợp đọc to tiếng Việt |
| **3** | Endless runner + Tiếng Anh nghe–từ vựng | Đa dạng thể loại |
| **4** | Nội dung mở rộng, huy hiệu xuyên game, trang "phụ huynh xem tiến độ" | Giữ chân & giá trị giáo dục |

---

## 6. Mô hình dữ liệu đề xuất (Phase 0)

```sql
-- Thay cho việc mỗi game một bảng
CREATE TABLE game_scores (
    id         TEXT PRIMARY KEY,
    game_key   TEXT NOT NULL,          -- 'typing' | 'bow2' | 'bubble_van' | ...
    user_id    TEXT NOT NULL,
    score      INTEGER NOT NULL DEFAULT 0,
    meta_json  TEXT,                   -- {level, prestige, accuracy, ...} tuỳ game
    scored_at  DATETIME NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_gamescores_key_at ON game_scores(game_key, scored_at);
```

`ranker.go`: lặp `for gameKey in registeredGames { for period { rebuild(gameKey, period) } }`,
cache key dạng `"<gameKey>-<period>"` (đã có tiền lệ `"bow-day"`).

---

## 7. Câu hỏi cần quyết định (để chốt scope)

1. **Môn ưu tiên**: Tiếng Việt / Toán (ngoài số học) / Tiếng Anh / TN&XH — làm môn nào trước?
2. **Độ tuổi trọng tâm**: tập trung Lớp 1–2 (icon + đọc to) hay trải đều Lớp 1–5?
3. **Nội dung**: nhúng bộ câu hỏi/từ vựng trong code, hay tách file dữ liệu (JSON) để dễ mở rộng?
4. **Đọc to tiếng Việt**: dùng `speechSynthesis` của trình duyệt (miễn phí, giọng tuỳ thiết bị) có chấp nhận được không?
5. **Thứ tự**: làm **Phase 0 (refactor nền tảng) trước**, hay làm ngay 1 game mới rồi refactor sau?

---

## 8. Đề xuất khởi đầu

Bắt đầu bằng **Phase 0 + game "🫧 Bong Bóng Vần"** — vừa dựng nền tảng SDK, vừa có sản phẩm chơi được sớm để kiểm chứng.

> _Tài liệu này là bản nháp để nghiên cứu; sẽ tinh chỉnh sau khi chốt câu hỏi mục 7._
