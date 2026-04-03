import { jsx as _jsx } from "react/jsx-runtime";
import * as React from "react";
import { cn } from "../lib/cn";
export function Card({ className, ...props }) {
    return (_jsx("div", { className: cn("rounded-[22px] border border-[var(--border)] bg-[var(--surface)]", className), ...props }));
}
export function CardHeader({ className, ...props }) {
    return _jsx("div", { className: cn("space-y-1 px-6 py-5", className), ...props });
}
export function CardTitle({ className, ...props }) {
    return (_jsx("h3", { className: cn("text-base font-medium tracking-[-0.02em] text-[var(--foreground)]", className), ...props }));
}
export function CardDescription({ className, ...props }) {
    return (_jsx("p", { className: cn("text-sm text-[var(--muted-foreground)]", className), ...props }));
}
export function CardContent({ className, ...props }) {
    return _jsx("div", { className: cn("px-6 pb-6", className), ...props });
}
