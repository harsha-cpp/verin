import { jsx as _jsx } from "react/jsx-runtime";
import * as React from "react";
import { cn } from "../lib/cn";
export function Select({ className, options, ...props }) {
    return (_jsx("select", { className: cn("h-11 w-full rounded-xl border border-[var(--border)] bg-white px-3 text-sm text-[var(--foreground)] outline-none transition focus:border-[var(--ring)] focus:ring-2 focus:ring-[var(--ring)]/15", className), ...props, children: options.map((option) => (_jsx("option", { value: option.value, children: option.label }, option.value))) }));
}
