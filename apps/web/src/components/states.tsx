import { AlertCircle } from "lucide-react";
import { Card, CardContent, CardDescription, CardTitle } from "@verin/ui";

export function EmptyState({ title, description }: { title: string; description: string }) {
  return (
    <Card>
      <CardContent className="flex min-h-[200px] flex-col items-center justify-center py-10 text-center">
        <div
          className="flex h-10 w-10 items-center justify-center rounded-xl"
          style={{ background: "var(--surface-soft)", border: "1px solid var(--line)" }}
        >
          <span style={{ color: "var(--ink-muted)", fontSize: 18 }}>·</span>
        </div>
        <CardTitle className="mt-4" style={{ color: "var(--ink)" }}>{title}</CardTitle>
        <CardDescription className="mt-1 max-w-sm" style={{ color: "var(--ink-soft)" }}>{description}</CardDescription>
      </CardContent>
    </Card>
  );
}

export function ErrorState({ title }: { title: string }) {
  return (
    <Card style={{ borderColor: "rgb(254 202 202)", background: "rgb(255 241 242)" }}>
      <CardContent className="py-5">
        <div className="flex items-center gap-3">
          <div
            className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl"
            style={{ background: "#fff", border: "1px solid rgb(254 202 202)", color: "#e11d48" }}
          >
            <AlertCircle className="h-4 w-4" />
          </div>
          <div>
            <CardTitle style={{ color: "var(--ink)" }}>{title}</CardTitle>
            <CardDescription style={{ color: "var(--ink-soft)" }}>Could not load this view. Check the API and try again.</CardDescription>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function LoadingState({ title }: { title: string }) {
  return (
    <Card>
      <CardContent className="py-5">
        <div className="flex items-center gap-3 mb-4">
          <div className="skeleton h-9 w-9 rounded-xl" />
          <div className="space-y-2 flex-1">
            <div className="skeleton h-4 w-40 rounded" />
            <div className="skeleton h-3 w-24 rounded" />
          </div>
        </div>
        <div className="sr-only">{title}</div>
        <div className="space-y-2">
          <div className="skeleton h-16 w-full" />
          <div className="grid gap-2 md:grid-cols-3">
            <div className="skeleton h-12 w-full" />
            <div className="skeleton h-12 w-full" />
            <div className="skeleton h-12 w-full" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function DocumentListSkeleton() {
  return (
    <div
      className="overflow-hidden rounded-2xl"
      style={{ border: "1px solid var(--line)", background: "var(--surface)" }}
    >
      {Array.from({ length: 5 }).map((_, i) => (
        <div
          key={i}
          className="flex items-center gap-4 px-5 py-4"
          style={{
            borderBottom: i < 4 ? "1px solid var(--line)" : undefined,
            animationDelay: `${i * 60}ms`,
          }}
        >
          <div className="skeleton h-8 w-8 rounded-lg shrink-0" />
          <div className="flex-1 space-y-1.5">
            <div className="skeleton h-3.5 rounded" style={{ width: `${55 + (i % 3) * 15}%` }} />
            <div className="skeleton h-3 rounded w-32" />
          </div>
          <div className="hidden md:flex gap-2">
            <div className="skeleton h-5 w-16 rounded-md" />
            <div className="skeleton h-5 w-8 rounded-md" />
            <div className="skeleton h-5 w-20 rounded-md" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function HomePageSkeleton() {
  return (
    <div className="space-y-8">
      <div className="space-y-3">
        <div className="skeleton h-3 w-24 rounded" />
        <div className="skeleton h-9 w-72 rounded" />
        <div className="skeleton h-4 w-96 rounded" />
      </div>
      <div className="grid gap-4 md:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="rounded-2xl p-5 space-y-3" style={{ border: "1px solid var(--line)", background: "var(--surface)" }}>
            <div className="skeleton h-4 w-4 rounded" />
            <div className="skeleton h-8 w-12 rounded" />
          </div>
        ))}
      </div>
      <div className="grid gap-6 xl:grid-cols-[1.3fr_0.7fr]">
        <div className="rounded-2xl p-6 space-y-4" style={{ border: "1px solid var(--line)", background: "var(--surface)" }}>
          <div className="skeleton h-4 w-36 rounded" />
          <DocumentListSkeleton />
        </div>
        <div className="rounded-2xl p-6 space-y-3" style={{ border: "1px solid var(--line)", background: "var(--surface)" }}>
          <div className="skeleton h-4 w-28 rounded" />
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="skeleton h-16 w-full rounded-xl" />
          ))}
        </div>
      </div>
    </div>
  );
}

export function SearchSkeleton() {
  return (
    <div className="space-y-3">
      {Array.from({ length: 3 }).map((_, i) => (
        <div
          key={i}
          className="rounded-2xl p-5 space-y-3"
          style={{ border: "1px solid var(--line)", background: "var(--surface)" }}
        >
          <div className="flex items-start justify-between gap-3">
            <div className="space-y-1.5 flex-1">
              <div className="skeleton h-4 rounded" style={{ width: `${50 + i * 12}%` }} />
              <div className="skeleton h-3 w-40 rounded" />
            </div>
            <div className="flex gap-2">
              <div className="skeleton h-5 w-14 rounded-md" />
              <div className="skeleton h-5 w-14 rounded-md" />
            </div>
          </div>
          <div className="skeleton h-3 w-full rounded" />
          <div className="skeleton h-3 rounded" style={{ width: "70%" }} />
        </div>
      ))}
    </div>
  );
}
