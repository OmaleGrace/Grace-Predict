from io import BytesIO

import pandas as pd
from fastapi import FastAPI, File, HTTPException, UploadFile
from pydantic import BaseModel

from app.services.training import predict, train_model


app = FastAPI(
    title="Grace Predict ML Service",
    description="AI and machine learning service for Grace Predict",
    version="1.0.0",
)


class TrainRequest(BaseModel):
    file_path: str
    target_column: str


class PredictRequest(BaseModel):
    model_id: str
    input_data: dict


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
async def inspect_dataset(
    file: UploadFile = File(...),
):
    if (
        not file.filename
        or not file.filename.lower().endswith(".csv")
    ):
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


@app.post("/models/train")
def train(request: TrainRequest):
    try:
        return train_model(
            request.file_path,
            request.target_column,
        )
    except ValueError as e:
        raise HTTPException(
            status_code=400,
            detail=str(e),
        )
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"model training failed: {str(e)}",
        )


@app.post("/models/predict")
def make_prediction(request: PredictRequest):
    try:
        return predict(
            request.model_id,
            request.input_data,
        )
    except ValueError as e:
        raise HTTPException(
            status_code=400,
            detail=str(e),
        )
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"prediction failed: {str(e)}",
        )
    