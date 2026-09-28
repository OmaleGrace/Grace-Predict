import {
  NavLink,
  Outlet,
  useNavigate,
} from "react-router-dom";

export default function Layout() {
  const navigate = useNavigate();

  const logout = () => {
    localStorage.removeItem("token");
    localStorage.removeItem("user");

    navigate("/login");
  };

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <div className="brand-mark">GP</div>

          <span>Grace Predict</span>
        </div>

        <nav>
          <NavLink to="/dashboard">
            Dashboard
          </NavLink>

          <NavLink to="/models">
            Models
          </NavLink>

          <NavLink to="/datasets">
            Datasets
          </NavLink>

          <NavLink to="/predictions">
            Predictions
          </NavLink>
        </nav>

        <button
          type="button"
          className="logout-button"
          onClick={logout}
        >
          Logout
        </button>
      </aside>

      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
}