import { Link } from "react-router-dom";

export function Logo({ size = "default" }: { size?: "default" | "lg" }) {
  const textClass = size === "lg" ? "text-[28px]" : "text-xl";

  return (
    <Link to="/home" className="group inline-flex items-baseline gap-0.5">
      <span className={`${textClass} font-semibold tracking-tight`} style={{ color: "var(--ink)" }}>
        verin
      </span>
      <span
        className="relative -top-1.5 ml-px h-2 w-2 rounded-full transition-all duration-200 group-hover:scale-110"
        style={{ background: "var(--accent)" }}
      />
    </Link>
  );
}
