import { jsx as _jsx } from "react/jsx-runtime";
import * as React from "react";
import { cva } from "class-variance-authority";
import { cn } from "../lib/cn";
const buttonVariants = cva("inline-flex items-center justify-center rounded-xl text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--ring)] focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-60", {
    variants: {
        variant: {
            primary: "bg-[var(--foreground)] px-4 py-2.5 text-[var(--background)] hover:bg-black/85",
            secondary: "border border-[var(--border)] bg-[var(--surface)] px-4 py-2.5 text-[var(--foreground)] hover:bg-[var(--surface-strong)]",
            ghost: "px-3 py-2 text-[var(--muted-foreground)] hover:bg-[var(--surface)] hover:text-[var(--foreground)]",
        },
        size: {
            default: "h-11",
            sm: "h-9 rounded-lg px-3 text-xs",
            icon: "h-10 w-10 rounded-full",
        },
    },
    defaultVariants: {
        variant: "primary",
        size: "default",
    },
});
export const Button = React.forwardRef(({ className, variant, size, ...props }, ref) => (_jsx("button", { ref: ref, className: cn(buttonVariants({ variant, size }), className), ...props })));
Button.displayName = "Button";
