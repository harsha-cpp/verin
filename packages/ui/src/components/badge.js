import { jsx as _jsx } from "react/jsx-runtime";
import { cva } from "class-variance-authority";
import { cn } from "../lib/cn";
const badgeVariants = cva("inline-flex items-center rounded-full border px-2.5 py-1 text-[11px] font-medium uppercase tracking-[0.12em]", {
    variants: {
        tone: {
            neutral: "border-[var(--border)] bg-[var(--surface-strong)] text-[var(--muted-foreground)]",
            success: "border-emerald-200 bg-emerald-50 text-emerald-700",
            warning: "border-amber-200 bg-amber-50 text-amber-700",
            danger: "border-rose-200 bg-rose-50 text-rose-700",
        },
    },
    defaultVariants: {
        tone: "neutral",
    },
});
export function Badge({ className, tone, children, }) {
    return _jsx("span", { className: cn(badgeVariants({ tone }), className), children: children });
}
