import { useState, type FormEvent } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { useI18n } from "../i18n/I18nContext";

export function LoginPage() {
  const { user, error, login } = useAuth();
  const { lang, setLang, t } = useI18n();
  const [email, setEmail] = useState(import.meta.env.DEV ? "kamil@op.pl" : "");
  const [password, setPassword] = useState(import.meta.env.DEV ? "lion12345" : "");
  const [submitting, setSubmitting] = useState(false);

  if (user) return <Navigate to="/listings" replace />;

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    await login(email, password);
    setSubmitting(false);
  }

  return (
    <div className="login-page">
      <form className="card" style={{ width: 320 }} onSubmit={handleSubmit}>
        <div className="lang-switch" style={{ justifyContent: "flex-end", marginBottom: "0.5rem" }}>
          <button
            type="button"
            className={lang === "en" ? "active" : ""}
            onClick={() => setLang("en")}
            aria-pressed={lang === "en"}
          >
            EN
          </button>
          <button
            type="button"
            className={lang === "pl" ? "active" : ""}
            onClick={() => setLang("pl")}
            aria-pressed={lang === "pl"}
          >
            PL
          </button>
        </div>
        <h1 style={{ marginTop: 0, fontSize: "1.25rem" }}>{t("login.title")}</h1>
        {error && <div className="error">{error}</div>}
        <div className="field">
          <label htmlFor="email">{t("login.email")}</label>
          <input
            id="email"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </div>
        <div className="field">
          <label htmlFor="password">{t("login.password")}</label>
          <input
            id="password"
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        <button className="btn-approve" type="submit" disabled={submitting} style={{ width: "100%" }}>
          {submitting ? t("login.signingIn") : t("login.signIn")}
        </button>
      </form>
    </div>
  );
}
