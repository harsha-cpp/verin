import { jsx as _jsx } from "react/jsx-runtime";
import * as React from "react";
import { cn } from "../lib/cn";
export function Label({ className, ...props }) {
    return (_jsx("label", { className: cn("mb-2 block text-sm font-medium text-[var(--foreground)]", className), ...props }));
}
