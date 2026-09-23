-- Users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Uploaded datasets
CREATE TABLE IF NOT EXISTS datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    file_path TEXT,
    target_column TEXT,
    row_count INTEGER,
    column_count INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ML models
CREATE TABLE IF NOT EXISTS models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID REFERENCES datasets(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    model_type TEXT NOT NULL,
    task_type TEXT NOT NULL,
    metrics JSONB,
    parameters JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Prediction history
CREATE TABLE IF NOT EXISTS prediction_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    model_id UUID REFERENCES models(id) ON DELETE SET NULL,
    prediction_type TEXT NOT NULL,
    input_data JSONB,
    prediction JSONB NOT NULL,
    confidence DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Saved predictions
CREATE TABLE IF NOT EXISTS saved_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    prediction_id UUID NOT NULL REFERENCES prediction_history(id) ON DELETE CASCADE,
    title TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Useful indexes
CREATE INDEX IF NOT EXISTS idx_datasets_user_id
    ON datasets(user_id);

CREATE INDEX IF NOT EXISTS idx_prediction_history_user_id
    ON prediction_history(user_id);

CREATE INDEX IF NOT EXISTS idx_prediction_history_type
    ON prediction_history(prediction_type);

CREATE INDEX IF NOT EXISTS idx_saved_predictions_user_id
    ON saved_predictions(user_id);