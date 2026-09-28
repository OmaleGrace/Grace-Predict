export default function Dashboard() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-bold text-slate-900">
          Dashboard
        </h1>
        <p className="mt-2 text-slate-600">
          Welcome to Grace Predict.
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <div className="rounded-xl border bg-white p-6 shadow-sm">
          <p className="text-sm text-slate-500">Datasets</p>
          <p className="mt-2 text-3xl font-bold text-slate-900">0</p>
        </div>

        <div className="rounded-xl border bg-white p-6 shadow-sm">
          <p className="text-sm text-slate-500">Models</p>
          <p className="mt-2 text-3xl font-bold text-slate-900">0</p>
        </div>

        <div className="rounded-xl border bg-white p-6 shadow-sm">
          <p className="text-sm text-slate-500">Predictions</p>
          <p className="mt-2 text-3xl font-bold text-slate-900">0</p>
        </div>
      </div>

      <div className="rounded-xl border bg-white p-6 shadow-sm">
        <h2 className="text-lg font-semibold text-slate-900">
          Get started
        </h2>

        <p className="mt-2 text-slate-600">
          Upload a dataset, train a machine-learning model,
          and make predictions.
        </p>
      </div>
    </div>
  );
}
