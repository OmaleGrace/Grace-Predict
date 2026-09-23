from fastapi import FastAPI, UploadFile, File, HTTPException
import pandas as pd
from io import BytesIO

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


@app.post("/datasets/inspect")
async def inspect_dataset(file: UploadFile = File(...)):
    if not file.filename or not file.filename.lower().endswith(".csv"):
        raise HTTPException(
            status_code=400,
            detail="Only CSV files are supported",
        )

    try:
        contents = await file.read()
        df = pd.read_csv(BytesIO(contents))
    except Exception:
        raise HTTPException(
            status_code=400,
            detail="Unable to read CSV file",
        )

    columns = []

    for column in df.columns:
        columns.append({
            "name": str(column),
            "dtype": str(df[column].dtype),
            "missing": int(df[column].isna().sum()),
            "unique": int(df[column].nunique()),
        })

    return {
        "filename": file.filename,
        "rows": int(len(df)),
        "columns": int(len(df.columns)),
        "column_details": columns,
    }