# QuestionnaireWithBackend

An online career anchor assessment system with full-stack capabilities. Answer 10 questions, get your dominant career anchor(s), and submit data to a persistent backend.

Based on Edgar Schein's Career Anchor theory: 8 core drives that shape career choices. Forked from [Spandan-Bhattarai/Personality-Traits-Tester](https://github.com/Spandan-Bhattarai/Personality-Traits-Tester) (React + TypeScript + Tailwind CSS frontend). Replaced the MBTI model with a Career Anchor model and added a Go backend with SQLite for data collection and export.

---

## Features

- **Questionnaire Flow**: 10-step questions with auto-advance and progress bar
- **Auto Scoring**: Counts A-H selections, determines dominant career anchor(s)
- **Result Page**: Anchor name, description, key traits, career suggestions, and score distribution across all 8 anchors
- **Dual Anchor Support**: If two anchors tie for highest count, both are displayed (e.g. "TF+SV")
- **Data Persistence**: Every submission saved to SQLite with UUIDs
- **CSV Export**: Download all submissions as CSV
- **Statistics**: Total count and per-type distribution
- **Version Check**: Frontend polls backend version every 30s; prompts refresh on mismatch
- **Duplicate Protection**: Prevents double-submit on the final question

---

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React 18 + TypeScript + Tailwind CSS + Vite |
| Backend | Go 1.25 + Gin |
| Database | SQLite (modernc.org/sqlite, pure Go, no CGO) |

### Environment Variables

| Variable | Required | Description |
|---|---|---|
| `PORT` | No | Server port (default: 3000) |
| `ADMIN_TOKEN` | No | Protects `/submissions`, `/export`, `/stats`. If unset, these endpoints return 401. |

---

## Quick Start

### Prerequisites

- Node.js 18+
- Go 1.23+
- npm

### 1. Clone

```bash
git clone https://github.com/chorious/QuestionnaireWithBackend.git
cd QuestionnaireWithBackend
```

### 2. Start Backend

```bash
cd backend-go
# Optional: set admin token to protect data endpoints
# If not set, /submissions, /export, /stats return 401
set ADMIN_TOKEN=your-secret-token
go run main.go
```

Backend runs at `http://localhost:3000`

### 3. Start Frontend

In another terminal:

```bash
npm install
npm run dev
```

Frontend runs at `http://localhost:5174`

The Vite dev server proxies `/api` requests to `http://localhost:3000`.

### 4. Build for Production

Frontend:
```bash
npm run build
```

Backend (single binary):
```bash
cd backend-go
go build -o questionnaire-backend
```

---

## API Reference

### GET /api/version

Response:
```json
{"version": "1.0.0"}
```

### POST /api/submit

Request body:
```json
{
  "answers": ["A", "B", "C", "D", "E", "F", "G", "H", "A", "B"],
  "scores": {"TF": 2, "GM": 2, "AU": 1, "SE": 1, "EC": 1, "SV": 1, "CH": 1, "LS": 1},
  "result": "TF+GM",
  "source": "",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

- `user_id` is optional. If provided, it is stored as-is. If omitted, the server generates one.
- The frontend persists `user_id` in localStorage so the same user always submits the same ID.

Response:
```json
{
  "success": true,
  "id": "c8e06fa3-d3ef-4116-be03-7a08c5805c64",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### GET /api/submissions

**Protected.** Requires `X-Admin-Token` header matching `ADMIN_TOKEN` env var.

```bash
curl -H "X-Admin-Token: your-secret-token" http://localhost:3000/api/submissions
```

Response:
```json
{
  "count": 5,
  "submissions": [
    {
      "id": "...",
      "user_id": "...",
      "answers": "[\"3\",\"5\",\"1\"]",
      "scores": "{\"TF\":3,\"GM\":2}",
      "result": "TF+GM",
      "created_at": 1777567977647,
      "source": ""
    }
  ]
}
```

### GET /api/submissions/export

**Protected.** Requires `X-Admin-Token` header.

Returns a CSV file with BOM for Excel compatibility.

### GET /api/stats

**Protected.** Requires `X-Admin-Token` header.

Response:
```json
{
  "total": 5,
  "byResult": [
    {"result": "TF", "count": 3},
    {"result": "GM", "count": 2},
    {"result": "TF+SV", "count": 1}
  ]
}
```

---

## Project Structure

```
QuestionnaireWithBackend/
|
|-- src/                           # Frontend
|   |-- components/
|   |   |-- WelcomeScreen.tsx
|   |   |-- QuestionnaireScreen.tsx
|   |   |-- ResultsScreen.tsx
|   |   |-- VersionCheck.tsx
|   |   |-- BackgroundAnimation.tsx
|   |   |-- ProgressBar.tsx
|   |-- api/
|   |   |-- client.ts              # HTTP client for backend API
|   |-- config/
|   |   |-- version.ts             # Frontend version constant
|   |-- data/
|   |   |-- questions.ts           # Question definitions
|   |   |-- personalityTypes.ts    # Result type definitions
|   |-- types/
|   |   |-- personality.ts
|   |-- utils/
|   |   |-- personalityCalculator.ts
|   |-- App.tsx
|   |-- main.tsx
|   |-- index.css
|
|-- backend-go/                    # Go backend
|   |-- main.go                    # HTTP server + API handlers + DB init
|   |-- go.mod
|   |-- go.sum
|   |-- data.db                    # SQLite database (gitignored)
|
|-- package.json
|-- vite.config.ts
|-- tailwind.config.js
|-- README.md
```

---

## Roadmap / TODO

- [x] Replace MBTI 32-question model with Career Anchor 10-question A-H scoring
- [ ] Add result share image generation (Canvas / html2canvas)
- [ ] Add nickname and optional fields to submission
- [ ] Deploy to production (cloud server or VPS)
- [ ] Admin data dashboard (simple HTML page)
- [ ] Multi-language support

---

## Acknowledgments

- Original frontend UI by [Spandan Bhattarai](https://github.com/Spandan-Bhattarai) / [Personality-Traits-Tester](https://github.com/Spandan-Bhattarai/Personality-Traits-Tester)
- Icons from [Lucide](https://lucide.dev/)
- Styling powered by [Tailwind CSS](https://tailwindcss.com/)

---

## License

MIT License - see the original repository for details.
