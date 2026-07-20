import { NavLink, Outlet } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { useI18n } from "../i18n/I18nContext";

export function Layout() {
  const { user, logout } = useAuth();
  const { lang, setLang, t } = useI18n();

  return (
    <div className="layout">
      <nav className="nav">
        <NavLink to="/listings" className={({ isActive }) => (isActive ? "active" : "")}>
          {t("nav.listings")}
        </NavLink>
        <NavLink to="/certifications" className={({ isActive }) => (isActive ? "active" : "")}>
          {t("nav.certifications")}
        </NavLink>
        <div className="nav-spacer" />
        <span>{user?.email}</span>
        <div className="lang-switch">
          <button
            className={lang === "en" ? "active" : ""}
            onClick={() => setLang("en")}
            aria-pressed={lang === "en"}
          >
            EN
          </button>
          <button
            className={lang === "pl" ? "active" : ""}
            onClick={() => setLang("pl")}
            aria-pressed={lang === "pl"}
          >
            PL
          </button>
        </div>
        <button onClick={logout}>{t("nav.logout")}</button>
      </nav>
      <main className="page">
        <Outlet />
      </main>
    </div>
  );
}
