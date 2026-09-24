# Grace-Predict

**Grace-Predict** is an AI-powered prediction platform built with Go, Python, PostgreSQL, and Next.js.

It is designed to support:

* 📊 Custom user-uploaded dataset
* 🏟️ Sports data and predictions
* 📈 Stock market and predictions
* 🤖 Machine learning model training and evaluation
* 👤 User accounts, authentication and prediction history
* 📜 Prediction history and saved predictions

## Architecture

```text
Next.js Frontend
       ↓
    Go API
       ↓
 PostgreSQL
       ↕
 Python ML Service
```

### Tech Stack

* **Frontend:** Next.js, TypeScript
* **Backend:** Go
* **ML:** Python, FastAPI, scikit-learn
* **Database:** PostgreSQL
* **Authentication:** JWT + bcrypt


## Current Features

* Go REST API
* Python ML service
* PostgreSQL database
* User registration and login
* Password hashing
* JWT authentication
* Protected API routes
* User profiles
* Prediction history database structure

## Running Locally

### Go API

```bash
cd backend/go-api
go run ./cmd/server
```

Runs on:

```text
http://localhost:8080
```

### Python ML Service

```bash
cd backend/python-ml
python -m uvicorn app.main:app --reload --port 8000
```

Runs on:

```text
http://localhost:8000
```

### Environment

Create a local `.env` file using `.env.example`:

```env
ML_SERVICE_URL=http://localhost:8000
GO_PORT=8080
DATABASE_URL=
JWT_SECRET=
```

**Never commit `.env` or other secrets to Git.**

## Development Roadmap

* [x] Go API
* [x] PostgreSQL
* [x] Authentication
* [x] JWT authorization
* [x] Python ML service
* [ ] CSV upload
* [ ] Dataset inspection
* [ ] Model training
* [ ] Predictions
* [ ] Next.js frontend
* [ ] Sports prediction system
* [ ] Stock prediction system
* [ ] Production deployment

## Testing

```bash
cd backend/go-api
go test ./...
```

## Status

🚧 **Grace-Predict is currently under active development.**

The next major milestone is building the **CSV upload and custom dataset prediction system**.
