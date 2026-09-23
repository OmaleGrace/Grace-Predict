from fastapi import FastAPI

app = FastAPI(
    title="Grace Predict ML Service",
    description="AI and machine learning service for Grace Predict",
    version="1.0.0",
)


@app.get("/health")
def health_check():
    return { 
        "status": "ok",
        "service": "grace-predict-ml",
    }


@app.get("/predict/test")
def test_prediction():
    return {
        "prediction": 42,
        "model": "test-model",
        "message": "Python ML service is working",
    }