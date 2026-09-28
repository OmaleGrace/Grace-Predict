import { useEffect, useState } from "react";
import { api } from "../services/api";
import type { Dataset } from "../services/api";
export default function Datasets() {
  const [datasets, setDatasets] = useState<Dataset[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  async function loadDatasets() {
    try {
      setLoading(true);
      setError("");

      const data = await api.getDatasets();

      setDatasets(data);
    } catch (err) {
      console.error(err);

      setError(
        err instanceof Error
          ? err.message
          : "Unable to load datasets.",
      );
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadDatasets();
  }, []);

  return (
    <div>
      <div className="page-header">
        <p className="eyebrow">Data Management</p>

        <h1>Datasets</h1>

        <p>
          Upload and manage datasets used for model training.
        </p>
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <span>Total Datasets</span>
          <strong>{loading ? "—" : datasets.length}</strong>
        </div>

        <div className="stat-card">
          <span>Available for Training</span>
          <strong>{loading ? "—" : datasets.length}</strong>
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
          <h2>Your Datasets</h2>

          <p>
            Datasets stored in your Grace Predict account.
          </p>
        </div>

        {loading && (
          <div className="empty-state">
            <h3>Loading datasets...</h3>

            <p>
              Retrieving your datasets from the server.
            </p>
          </div>
        )}

        {!loading && error && (
          <div className="empty-state">
            <h3>Unable to load datasets</h3>

            <p>{error}</p>

            <button
              type="button"
              className="primary-button"
              onClick={loadDatasets}
              style={{ marginTop: "18px" }}
            >
              Try Again
            </button>
          </div>
        )}

        {!loading &&
          !error &&
          datasets.length === 0 && (
            <div className="empty-state">
              <h3>No datasets yet</h3>

              <p>
                Upload a dataset to start training models.
              </p>
            </div>
          )}

        {!loading &&
          !error &&
          datasets.length > 0 && (
            <div className="table-wrapper">
              <table>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Description</th>
                    <th>Rows</th>
                    <th>Columns</th>
                    <th>Created</th>
                  </tr>
                </thead>

                <tbody>
                  {datasets.map((dataset) => (
                    <tr key={dataset.id}>
                      <td>
                        <strong>{dataset.name}</strong>
                      </td>

                      <td>
                        {dataset.description || "—"}
                      </td>

                      <td>
                        {dataset.row_count ?? "—"}
                      </td>

                      <td>
                        {dataset.column_count ?? "—"}
                      </td>

                      <td>
                        {dataset.created_at
                          ? new Date(
                              dataset.created_at,
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