export function formatBytes(value = 0) {
  if (value === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const exponent = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return `${(value / 1024 ** exponent).toFixed(exponent === 0 ? 0 : 1)} ${units[exponent]}`;
}

export function formatDate(value?: string) {
  if (!value) return "Not available";
  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function getStatusTone(status: string) {
  switch (status) {
    case "ready":
    case "completed":
      return "success" as const;
    case "processing":
      return "warning" as const;
    case "archived":
    case "failed":
      return "danger" as const;
    default:
      return "neutral" as const;
  }
}
