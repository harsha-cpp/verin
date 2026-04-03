import { jsx as _jsx } from "react/jsx-runtime";
import * as React from "react";
import { cn } from "../lib/cn";
export const Textarea = React.forwardRef(({ className, ...props }, ref) => (_jsx("textarea", { ref: ref, className: cn("min-h-28 w-full rounded-xl border border-[var(--border)] bg-white px-3 py-3 text-sm text-[var(--foreground)] outline-none transition focus:border-[var(--ring)] focus:ring-2 focus:ring-[var(--ring)]/15 disabled:cursor-not-allowed disabled:opacity-60", className), ...props })));
Textarea.displayName = "Textarea";
