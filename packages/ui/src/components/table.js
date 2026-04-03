import { jsx as _jsx } from "react/jsx-runtime";
import * as React from "react";
import { cn } from "../lib/cn";
export function Table({ className, ...props }) {
    return _jsx("table", { className: cn("w-full border-collapse", className), ...props });
}
export function TableHead({ className, ...props }) {
    return _jsx("thead", { className: cn("text-left", className), ...props });
}
export function TableBody({ className, ...props }) {
    return _jsx("tbody", { className: cn(className), ...props });
}
export function TableRow({ className, ...props }) {
    return (_jsx("tr", { className: cn("border-b border-[var(--border)] last:border-none", className), ...props }));
}
export function TableHeaderCell({ className, ...props }) {
    return (_jsx("th", { className: cn("px-4 py-3 text-[11px] font-medium uppercase tracking-[0.16em] text-[var(--muted-foreground)]", className), ...props }));
}
export function TableCell({ className, ...props }) {
    return _jsx("td", { className: cn("px-4 py-4 text-sm text-[var(--foreground)]", className), ...props });
}
