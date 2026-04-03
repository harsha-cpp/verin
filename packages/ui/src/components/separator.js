import { jsx as _jsx } from "react/jsx-runtime";
import { cn } from "../lib/cn";
export function Separator({ className }) {
    return _jsx("div", { className: cn("h-px w-full bg-[var(--border)]", className) });
}
