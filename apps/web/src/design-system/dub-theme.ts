export const dubTheme = {
  pageTitle: "text-lg font-semibold text-slate-900",
  pageDescription: "mt-0.5 text-sm text-slate-500",
  sectionLabel: "text-xs font-medium text-slate-500",
  panel: "rounded-xl border border-slate-200 bg-white",
  mutedPanel: "rounded-xl border border-slate-100 bg-slate-50",
};

export function initialsFromName(name?: string) {
  if (!name) {
    return "VD";
  }

  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");
}
