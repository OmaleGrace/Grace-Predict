import uuid
from pathlib import Path

import joblib
import pandas as pd
from sklearn.ensemble import (
    RandomForestClassifier,
    RandomForestRegressor,
)
from sklearn.metrics import (
    accuracy_score,
    mean_absolute_error,
    r2_score,
)
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import LabelEncoder


MODEL_DIR = Path(__file__).resolve().parent.parent / "models"
MODEL_DIR.mkdir(parents=True, exist_ok=True)

MIN_ROWS = 10


def _validate_dataset(
    df: pd.DataFrame,
    target_column: str,
):
    if df.empty:
        raise ValueError("dataset is empty")

    if target_column not in df.columns:
        raise ValueError(
            f"target column '{target_column}' was not found"
        )

    if len(df) < MIN_ROWS:
        raise ValueError(
            f"dataset must contain at least {MIN_ROWS} rows "
            f"for reliable model evaluation; received {len(df)}"
        )

    if df[target_column].isna().all():
        raise ValueError(
            "target column contains no usable values"
        )


def _prepare_features(
    df: pd.DataFrame,
    target_column: str,
):
    X = df.drop(columns=[target_column])
    y = df[target_column]

    X = X.select_dtypes(include="number")

    if X.empty:
        raise ValueError(
            "dataset must contain at least one numeric feature"
        )

    valid_rows = X.notna().all(axis=1) & y.notna()

    X = X.loc[valid_rows]
    y = y.loc[valid_rows]

    if len(X) < MIN_ROWS:
        raise ValueError(
            "not enough usable rows after removing missing values"
        )

    return X, y


def _detect_task(y: pd.Series):
    if (
        pd.api.types.is_object_dtype(y)
        or pd.api.types.is_categorical_dtype(y)
        or pd.api.types.is_bool_dtype(y)
    ):
        return "classification"

    return "regression"


def train_model(
    file_path: str,
    target_column: str,
):
    try:
        df = pd.read_csv(file_path)
    except FileNotFoundError:
        raise ValueError("dataset file was not found")
    except Exception as exc:
        raise ValueError(
            f"unable to read dataset: {exc}"
        )

    _validate_dataset(df, target_column)

    X, y = _prepare_features(
        df,
        target_column,
    )

    task_type = _detect_task(y)

    label_encoder = None

    if task_type == "classification":
        label_encoder = LabelEncoder()
        y = label_encoder.fit_transform(y)

        class_counts = pd.Series(y).value_counts()

        if len(class_counts) < 2:
            raise ValueError(
                "classification target must contain at least "
                "two classes"
            )

        stratify = (
            y
            if class_counts.min() >= 2
            else None
        )

        model = RandomForestClassifier(
            n_estimators=100,
            random_state=42,
            n_jobs=-1,
        )

    else:
        if y.nunique() < 2:
            raise ValueError(
                "regression target must contain at least "
                "two distinct values"
            )

        stratify = None

        model = RandomForestRegressor(
            n_estimators=100,
            random_state=42,
            n_jobs=-1,
        )

    test_size = max(
        2,
        round(len(X) * 0.2),
    )

    if test_size >= len(X):
        test_size = len(X) - 1

    X_train, X_test, y_train, y_test = train_test_split(
        X,
        y,
        test_size=test_size,
        random_state=42,
        stratify=stratify,
    )

    if len(X_train) < 2:
        raise ValueError(
            "not enough training rows after splitting dataset"
        )

    model.fit(X_train, y_train)

    predictions = model.predict(X_test)

    if task_type == "classification":
        metrics = {
            "accuracy": float(
                accuracy_score(
                    y_test,
                    predictions,
                )
            )
        }

    else:
        metrics = {
            "mae": float(
                mean_absolute_error(
                    y_test,
                    predictions,
                )
            ),
            "r2": (
                float(
                    r2_score(
                        y_test,
                        predictions,
                    )
                )
                if len(y_test) >= 2
                else None
            ),
        }

    model_id = str(uuid.uuid4())

    model_path = MODEL_DIR / f"{model_id}.joblib"

    joblib.dump(
        {
            "model": model,
            "features": list(X.columns),
            "task_type": task_type,
            "target_column": target_column,
            "label_encoder": label_encoder,
        },
        model_path,
    )

    return {
        "model_id": model_id,
        "task_type": task_type,
        "target_column": target_column,
        "features": list(X.columns),
        "rows": len(X),
        "training_rows": len(X_train),
        "test_rows": len(X_test),
        "metrics": metrics,
    }


def predict(
    model_id: str,
    input_data: dict,
):
    model_path = MODEL_DIR / f"{model_id}.joblib"

    if not model_path.exists():
        raise ValueError("model not found")

    try:
        model_info = joblib.load(model_path)
    except Exception:
        raise ValueError("unable to load model")

    model = model_info["model"]
    features = model_info["features"]
    task_type = model_info["task_type"]
    label_encoder = model_info.get("label_encoder")

    missing_features = [
        feature
        for feature in features
        if feature not in input_data
    ]

    if missing_features:
        raise ValueError(
            f"missing features: {', '.join(missing_features)}"
        )

    X = pd.DataFrame(
        [[input_data[feature] for feature in features]],
        columns=features,
    )

    prediction = model.predict(X)[0]

    if (
        task_type == "classification"
        and label_encoder is not None
    ):
        prediction = label_encoder.inverse_transform(
            [prediction]
        )[0]

    result = {
        "model_id": model_id,
        "prediction": (
            prediction.item()
            if hasattr(prediction, "item")
            else prediction
        ),
        "task_type": task_type,
    }

    if (
        task_type == "classification"
        and hasattr(model, "predict_proba")
    ):
        probabilities = model.predict_proba(X)[0]

        result["confidence"] = float(
            max(probabilities)
        )

    return result 
