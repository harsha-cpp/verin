import { AlertCircle, LoaderCircle, Sparkles } from "lucide-react";
import { Card, CardContent, CardDescription, CardTitle } from "@verin/ui";

export function EmptyState({ title, description }: { title: string; description: string }) {
  return (
    <Card>
      <CardContent className="flex min-h-[200px] flex-col items-center justify-center py-10 text-center">
        <div className="flex h-10 w-10 items-center justify-center rounded-md bg-slate-100 text-slate-400">
          <Sparkles className="h-5 w-5" />
        </div>
        <CardTitle className="mt-3">{title}</CardTitle>
        <CardDescription className="mt-1 max-w-sm">{description}</CardDescription>
      </CardContent>
    </Card>
  );
}

export function LoadingState({ title }: { title: string }) {
  return (
    <Card>
      <CardContent className="py-5">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-md bg-slate-100 text-slate-400">
            <LoaderCircle className="h-4 w-4 animate-spin" />
          </div>
          <div>
            <CardTitle>{title}</CardTitle>
            <CardDescription>Loading the latest view.</CardDescription>
          </div>
        </div>
        <div className="mt-4 space-y-2">
          <div className="h-20 animate-pulse rounded-md bg-slate-100" />
          <div className="grid gap-2 md:grid-cols-3">
            <div className="h-14 animate-pulse rounded-md bg-slate-100" />
            <div className="h-14 animate-pulse rounded-md bg-slate-100" />
            <div className="h-14 animate-pulse rounded-md bg-slate-100" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function ErrorState({ title }: { title: string }) {
  return (
    <Card className="border-rose-200 bg-rose-50">
      <CardContent className="py-5">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-md border border-rose-200 bg-white text-rose-500">
            <AlertCircle className="h-4 w-4" />
          </div>
          <div>
            <CardTitle>{title}</CardTitle>
            <CardDescription>Could not load this view. Check the API and try again.</CardDescription>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
