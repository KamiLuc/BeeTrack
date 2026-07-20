import { createContext, useContext, useMemo, useState, type ReactNode } from "react";
import { translations, type Lang, type TranslationKey } from "./translations";

const LANG_KEY = "beetrack_admin_lang";

function detectDefaultLang(): Lang {
  const stored = localStorage.getItem(LANG_KEY);
  if (stored === "en" || stored === "pl") return stored;
  return "pl";
}

type I18nState = {
  lang: Lang;
  setLang: (lang: Lang) => void;
  t: (key: TranslationKey, params?: Record<string, string | number>) => string;
};

const I18nContext = createContext<I18nState | null>(null);

export function I18nProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(detectDefaultLang);

  const setLang = (next: Lang) => {
    localStorage.setItem(LANG_KEY, next);
    setLangState(next);
  };

  const t = useMemo(() => {
    return (key: TranslationKey, params?: Record<string, string | number>) => {
      const [ns, field] = key.split(".");
      const dict = translations[lang] as Record<string, Record<string, string>>;
      let value = dict[ns]?.[field] ?? key;
      if (params) {
        for (const [name, val] of Object.entries(params)) {
          value = value.replace(`{{${name}}}`, String(val));
        }
      }
      return value;
    };
  }, [lang]);

  return <I18nContext.Provider value={{ lang, setLang, t }}>{children}</I18nContext.Provider>;
}

export function useI18n(): I18nState {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error("useI18n must be used within an I18nProvider");
  return ctx;
}
