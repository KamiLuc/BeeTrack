import { useMemo, useState } from "react";
import { useI18n } from "../i18n/I18nContext";

type ReasonPickerProps = {
  options: string[];
  onSelect: (reason: string) => void;
};

export function ReasonPicker({ options, onSelect }: ReasonPickerProps) {
  const { t } = useI18n();
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return options;
    return options.filter((option) => option.toLowerCase().includes(q));
  }, [options, query]);

  return (
    <div className="reason-picker">
      <input
        type="text"
        value={query}
        placeholder={t("common.reasonPickerPlaceholder")}
        onChange={(e) => {
          setQuery(e.target.value);
          setOpen(true);
        }}
        onFocus={() => setOpen(true)}
        onBlur={() => setOpen(false)}
      />
      {open && (
        <ul className="reason-picker-list">
          {filtered.length === 0 ? (
            <li className="reason-picker-empty">{t("common.reasonPickerNoMatch")}</li>
          ) : (
            filtered.map((option) => (
              <li
                key={option}
                onMouseDown={() => {
                  onSelect(option);
                  setQuery("");
                  setOpen(false);
                }}
              >
                {option}
              </li>
            ))
          )}
        </ul>
      )}
    </div>
  );
}
