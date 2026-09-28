const API_URL =
  import.meta.env.VITE_API_URL || "http://localhost:8080";

async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const token = localStorage.getItem("token");

  const headers = new Headers(options.headers);

  if (!headers.has("Content-Type") && options.body) {
    headers.set("Content-Type", "application/json");
  }

  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  });

  const text = await response.text();

  let data: unknown = null;

  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = text;
    }
  }

  if (!response.ok) {
    const message =
      typeof data === "object" &&
      data !== null &&
      "message" in data
        ? String((data as { message: unknown }).message)
        : typeof data === "string"
          ? data
          : `Request failed with status ${response.status}`;

    throw new Error(message);
  }

  return data as T;
}

export interface LoginResponse {
  token: string;
  user?: {
    id: string;
    email: string;
    name?: string;
  };
}

export interface Dataset {
  id: string;
  name: string;
  description?: string | null;
  file_path?: string | null;
  row_count?: number | null;
  column_count?: number | null;
  created_at?: string;
}

export interface Model {
  id: string;
  dataset_id?: string | null;
  name: string;
  model_type: string;
  task_type: string;
  metrics?: Record<string, unknown> | null;
  parameters?: Record<string, unknown> | null;
  created_at?: string;
}

export interface User {
  id: string;
  email: string;
  name?: string | null;
}

export interface PredictionHistory {
  id: string;
  user_id: string;
  model_id?: string | null;
  prediction_type: string;
  input_data?: Record<string, unknown> | null;
  prediction: unknown;
  confidence?: number | null;
  created_at?: string;
}

export const api = {
  login(email: string, password: string) {
    return request<LoginResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({
        email,
        password,
      }),
    });
  },

  register(
    email: string,
    password: string,
    name: string,
  ) {
    return request<LoginResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify({
        email,
        password,
        name,
      }),
    });
  },

  getDatasets() {
    return request<Dataset[]>("/api/datasets");
  },

  createDataset(
    name: string,
    description: string,
  ) {
    return request<Dataset>("/api/datasets", {
      method: "POST",
      body: JSON.stringify({
        name,
        description,
      }),
    });
  },

  getModels() {
    return request<{ models: Model[] }>("/api/models");
  },

  createModel(data: {
    id?: string;
    dataset_id?: string;
    name: string;
    model_type: string;
    task_type: string;
    metrics?: Record<string, unknown>;
    parameters?: Record<string, unknown>;
  }) {
    return request<{ id: string; message: string }>(
      "/api/models",
      {
        method: "POST",
        body: JSON.stringify(data),
      },
    );
  },

    predict(data: {
    model_id: string;
    input_data: Record<string, unknown>;
  }) {
    return request<{
      prediction: unknown;
      confidence?: number | null;
    }>("/api/predict", {
      method: "POST",
      body: JSON.stringify(data),
    });
  },
};