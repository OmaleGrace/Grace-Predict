import { useEffect, useState } from "react";
import { api } from "../services/api";
import type { Model } from "../services/api";

export default function Models() {
  const [models, setModels] = useState<Model[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  async function loadModels() {
    try {
      setLoading(true);
      setError("");

      const response = await api.getModels();

      setModels(response.models ?? []);
    } catch (err) {
      console.error(err);

      setError(
        err instanceof Error
          ? err.message
          : "Unable to load models.",
      );
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadModels();
  }, []);

  return (
    <div>
      <div className="page-header">
        <p className="eyebrow">Machine Learning</p>

        <h1>Models</h1>

        <p>
          View and manage the machine learning models
          available in Grace Predict.
        </p>
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <span>Total Models</span>
          <strong>{loading ? "—" : models.length}</strong>
        </div>

        <div className="stat-card">
          <span>Model Types</span>
          <strong>
            {loading
              ? "—"
              : new Set(models.map((model) => model.model_type))
                  .size}
          </strong>
        </div>

        <div className="stat-card">
          <span>Status</span>
          <strong>
            {loading ? "Loading" : error ? "Error" : "Ready"}
          </strong>
        </div>
      </div>

      <section className="content-section">
        <div className="section-header">
          <h2>Your Models</h2>

          <p>
            Models created and trained in your account.
          </p>
        </div>

        {loading && (
          <div className="empty-state">
            <h3>Loading models...</h3>

            <p>
              Retrieving your models from the server.
            </p>
          </div>
        )}

        {!loading && error && (
          <div className="empty-state">
            <h3>Unable to load models</h3>

            <p>{error}</p>

            <button
              type="button"
              className="primary-button"
              onClick={loadModels}
              style={{ marginTop: "18px" }}
            >
              Try Again
            </button>
          </div>
        )}

        {!loading &&
          !error &&
          models.length === 0 && (
            <div className="empty-state">
              <h3>No models yet</h3>

              <p>
                Create or train a model to see it here.
              </p>
            </div>
          )}

        {!loading &&
          !error &&
          models.length > 0 && (
            <div className="table-wrapper">
              <table>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Model Type</th>
                    <th>Task</th>
                    <th>Created</th>
                  </tr>
                </thead>

                <tbody>
                  {models.map((model) => (
                    <tr key={model.id}>
                      <td>
                        <strong>{model.name}</strong>
                      </td>

                      <td>{model.model_type}</td>

                      <td>{model.task_type}</td>

                      <td>
                        {model.created_at
                          ? new Date(
                              model.created_at,
                            ).toLocaleDateString()
                          : "—"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
      </section>
    </div>
  );
}