export default function Predictions() {
  return (
    <div>
      <div className="page-header">
        <p className="eyebrow">Prediction Engine</p>
        <h1>Predictions</h1>
        <p>Run predictions and review previous results.</p>
      </div>

      <section className="content-section">
        <div className="section-header">
          <h2>Prediction History</h2>
          <p>Your previous prediction requests will appear here.</p>
        </div>

        <div className="empty-state">
          <h3>No predictions yet</h3>
          <p>Train a model and run your first prediction.</p>
        </div>
      </section>
    </div>
  );
}