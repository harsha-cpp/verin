import {
  startTransition,
  useDeferredValue,
  useState,
  type ChangeEvent,
  type ReactNode,
} from "react";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Clock3,
  Files,
  FolderOpen,
  Link2,
  MessageSquare,
  Plus,
  Search,
  Share2,
  Sparkles,
  Upload,
  Users2,
} from "lucide-react";
import { Link, Navigate, useNavigate, useParams } from "react-router-dom";
import {
  Badge,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Input,
  Label,
  Select,
  Textarea,
  cn,
} from "@verin/ui";

import { DocumentTable } from "@/components/document-table";
import { Logo } from "@/components/logo";
import { DocumentListSkeleton, EmptyState, ErrorState, HomePageSkeleton, LoadingState, SearchSkeleton } from "@/components/states";
import {
  useAddComment,
  useCreateInvite,
  useCreateSpace,
  useCreateTeam,
  useDeleteSpace,
  useDeleteTeam,
  useDocument,
  useDocuments,
  useHome,
  useJoinTeam,
  useRemoveMember,
  useSearch,
  useSession,
  useShareDocument,
  useSharedDocuments,
  useSpaces,
  useTeamInfo,
  useTeamMembers,
  useUpdateMemberRole,
  useUpdateTeam,
  useUploadDocument,
} from "@/hooks/use-api";
import { formatBytes, formatDate, getStatusTone } from "@/lib/utils";

type SessionData = {
  authenticated?: boolean;
  hasWorkspace?: boolean;
  workspace?: { name?: string };
  user?: { fullName?: string };
};

type DocumentListItem = {
  id: string;
  title: string;
  originalFilename: string;
  status: string;
  currentVersionNumber?: number;
  sizeBytes?: number;
  updatedAt?: string;
  summary?: string;
  summaryStatus?: string;
  previewStatus?: string;
  sharedWithCount?: number;
  commentCount?: number;
};

const onboardingSchema = z.object({
  name: z.string().min(1, "Workspace name is required"),
  slug: z.string().min(1, "Workspace slug is required").regex(/^[a-z0-9-]+$/, "Use lowercase letters, numbers, and hyphens"),
});

const joinSchema = z.object({
  code: z.string().min(1, "Invite code is required"),
});

const uploadSchema = z.object({
  title: z.string().optional(),
  collectionId: z.string().optional(),
  department: z.string().optional(),
  tags: z.string().optional(),
  changeSummary: z.string().optional(),
});

const createSpaceSchema = z.object({
  name: z.string().min(1, "Space name is required"),
  description: z.string().optional(),
});

const commentSchema = z.object({
  body: z.string().min(1, "Comment is required"),
});

const shareSchema = z.object({
  userId: z.string().min(1, "Teammate email is required"),
});

const updateTeamSchema = z.object({
  name: z.string().min(1, "Workspace name is required"),
  slug: z.string().min(1, "Workspace slug is required").regex(/^[a-z0-9-]+$/, "Use lowercase letters, numbers, and hyphens"),
});

export function LandingPage() {
  const session = useSession();
  const sd = session.data as SessionData | undefined;

  if (sd?.authenticated) {
    return <Navigate to={sd?.hasWorkspace ? "/home" : "/onboarding"} replace />;
  }

  return (
    <div className="relative min-h-screen overflow-hidden bg-[var(--bg)] text-[var(--ink)]">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(113,131,119,0.18),transparent_34%),radial-gradient(circle_at_bottom_right,rgba(34,46,43,0.08),transparent_30%)]" />
      <div className="mx-auto flex min-h-screen max-w-6xl flex-col px-6 py-8 lg:px-10">
        <header className="flex items-center justify-between">
          <Logo />
          <Button variant="secondary" onClick={() => { window.location.href = "/api/v1/auth/google"; }}>
            Sign in
          </Button>
        </header>

        <main className="grid flex-1 items-center gap-14 py-16 lg:grid-cols-[1.1fr_0.9fr]">
          <section className="animate-rise space-y-8">
            <Badge className="border-[var(--line-strong)] bg-[var(--surface-soft)] text-[var(--ink-soft)]">
              Shared document workspace for small teams
            </Badge>
            <div className="space-y-5">
              <h1 className="font-display text-5xl leading-[0.94] tracking-[-0.04em] text-[var(--ink)] md:text-7xl">
                Clear enough for anyone, fast enough for everyday work.
              </h1>
              <p className="max-w-xl text-lg leading-8 text-[var(--ink-soft)]">
                Verin turns upload, search, sharing, and team coordination into one calm workspace.
                No admin maze. No noisy dashboards. Just a polished place to keep important files moving.
              </p>
            </div>

            <div className="flex flex-wrap gap-3">
              <Button className="h-12 px-5 gap-3" onClick={() => { window.location.href = "/api/v1/auth/google"; }}>
                <svg className="h-4 w-4" viewBox="0 0 24 24" aria-hidden="true">
                  <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                  <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                  <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z"/>
                  <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
                </svg>
                Sign in with Google
              </Button>
            </div>

            <div id="landing-highlights" className="grid gap-3 sm:grid-cols-3">
              {[
                { title: "Fast first-use loop", copy: "Join a workspace, upload a file, and understand it immediately." },
                { title: "Search that feels instant", copy: "Title, OCR, and metadata work together without extra setup." },
                { title: "Light collaboration", copy: "Comments, sharing, and invites stay simple and visible." },
              ].map((item, index) => (
                <Card key={item.title} className={cn("border-[var(--line)] bg-white/80 backdrop-blur-sm animate-rise", `delay-${(index + 1) * 100}`)}>
                  <CardContent className="space-y-2 py-5">
                    <div className="text-sm font-semibold text-[var(--ink)]">{item.title}</div>
                    <p className="text-sm leading-6 text-[var(--ink-soft)]">{item.copy}</p>
                  </CardContent>
                </Card>
              ))}
            </div>
          </section>

          <section className="animate-rise delay-200">
            <Card className="overflow-hidden border-[var(--line-strong)] bg-[linear-gradient(180deg,rgba(255,255,255,0.9),rgba(248,245,239,0.9))] shadow-[0_32px_80px_rgba(22,27,24,0.08)]">
              <CardContent className="space-y-6 p-6">
                <div className="flex items-center justify-between rounded-[22px] border border-[var(--line)] bg-white px-4 py-3">
                  <div>
                    <div className="text-xs uppercase tracking-[0.24em] text-[var(--ink-muted)]">Workspace preview</div>
                    <div className="mt-1 text-base font-semibold text-[var(--ink)]">A calmer daily command center</div>
                  </div>
                  <Badge tone="success">Live</Badge>
                </div>

                <div className="grid gap-4">
                  <PreviewPanel
                    eyebrow="Today"
                    title="Recent work, pending processing, and team updates in one place"
                    icon={<Sparkles className="h-4 w-4" />}
                  />
                  <div className="grid gap-4 md:grid-cols-2">
                    <PreviewPanel
                      eyebrow="Documents"
                      title="Summary-first file list with quiet status and stronger previews"
                      icon={<Files className="h-4 w-4" />}
                    />
                    <PreviewPanel
                      eyebrow="Search"
                      title="OCR snippets and concise summaries instead of blank results"
                      icon={<Search className="h-4 w-4" />}
                    />
                  </div>
                </div>
              </CardContent>
            </Card>
          </section>
        </main>
      </div>
    </div>
  );
}

export function LoginPage() {
  return <LandingPage />;
}

export function OnboardingPage() {
  const navigate = useNavigate();
  const session = useSession();
  const createTeam = useCreateTeam();
  const joinTeam = useJoinTeam();
  const [mode, setMode] = useState<"create" | "join">("create");
  const [error, setError] = useState<string | null>(null);

  const createForm = useForm<z.infer<typeof onboardingSchema>>({
    resolver: zodResolver(onboardingSchema),
    defaultValues: { name: "", slug: "" },
  });
  const joinForm = useForm<z.infer<typeof joinSchema>>({
    resolver: zodResolver(joinSchema),
    defaultValues: { code: "" },
  });

  if (session.isLoading) {
    return <CenteredLoading title="Preparing workspace setup" />;
  }

  const sd2 = session.data as SessionData | undefined;
  if (!sd2?.authenticated) {
    return <Navigate to="/" replace />;
  }

  if (sd2?.hasWorkspace) {
    return <Navigate to="/home" replace />;
  }

  return (
    <CenteredLayout
      title="Create or join your workspace"
      description="Start simple. One shared home for documents, search, and team coordination."
      aside="You can create a fresh space for your team or join an existing one with an invite code."
    >
      <div className="mb-4 flex rounded-full border border-[var(--line)] bg-[var(--surface-soft)] p-1">
        {[
          { value: "create", label: "Create workspace" },
          { value: "join", label: "Join workspace" },
        ].map((item) => (
          <button
            key={item.value}
            type="button"
            className={cn(
              "flex-1 rounded-full px-4 py-2 text-sm font-medium transition",
              mode === item.value ? "bg-[var(--ink)] text-white shadow-sm" : "text-[var(--ink-soft)]",
            )}
            onClick={() => { setMode(item.value as "create" | "join"); setError(null); }}
          >
            {item.label}
          </button>
        ))}
      </div>

      {error ? <ErrorBanner message={error} /> : null}

      {mode === "create" ? (
        <Card className="border-[var(--line-strong)]">
          <CardContent className="pt-6">
            <form
              className="space-y-4"
              onSubmit={createForm.handleSubmit(async (values) => {
                setError(null);
                try {
                  await createTeam.mutateAsync(values);
                  navigate("/home");
                } catch (err) {
                  setError(err instanceof Error ? err.message : "Could not create workspace");
                }
              })}
            >
              <FormField label="Workspace name" htmlFor="workspace-name">
                <Input
                  id="workspace-name"
                  placeholder="Northwind Studio"
                  {...createForm.register("name", {
                    onChange: (event: ChangeEvent<HTMLInputElement>) => {
                      const slug = event.target.value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-|-$/g, "");
                      createForm.setValue("slug", slug);
                    },
                  })}
                />
              </FormField>
              <FormField label="Workspace slug" htmlFor="workspace-slug" helper="Used in links and invite context.">
                <Input id="workspace-slug" placeholder="northwind-studio" {...createForm.register("slug")} />
              </FormField>
              <Button className="w-full" type="submit" disabled={createTeam.isPending}>
                {createTeam.isPending ? "Creating workspace..." : "Create workspace"}
              </Button>
            </form>
          </CardContent>
        </Card>
      ) : (
        <Card className="border-[var(--line-strong)]">
          <CardContent className="pt-6">
            <form
              className="space-y-4"
              onSubmit={joinForm.handleSubmit(async (values) => {
                setError(null);
                try {
                  await joinTeam.mutateAsync({ code: values.code.toUpperCase().trim() });
                  navigate("/home");
                } catch (err) {
                  setError(err instanceof Error ? err.message : "Invite code is not valid");
                }
              })}
            >
              <FormField label="Invite code" htmlFor="invite-code" helper="Use the code shared by your workspace owner.">
                <Input id="invite-code" placeholder="ACME42QW" {...joinForm.register("code")} />
              </FormField>
              <Button className="w-full" type="submit" disabled={joinTeam.isPending}>
                {joinTeam.isPending ? "Joining workspace..." : "Join workspace"}
              </Button>
            </form>
          </CardContent>
        </Card>
      )}
    </CenteredLayout>
  );
}

export function HomePage() {
  const home = useHome();

  if (home.isLoading) return <HomePageSkeleton />;
  if (home.isError) return <ErrorState title="Workspace overview unavailable" />;

  const data = home.data as Record<string, any>;
  const recentDocuments = (data?.recentDocuments ?? []) as DocumentListItem[];
  const pending = (data?.pending ?? []) as Array<Record<string, any>>;
  const activity = (data?.activity ?? []) as Array<Record<string, any>>;
  const teammates = (data?.teammates ?? []) as Array<Record<string, any>>;
  const stats = data?.stats as Record<string, number> | undefined;

  return (
    <div className="space-y-8">
      {(stats?.documentCount ?? 0) === 0 && <GettingStartedStrip />}
      <PageIntro
        eyebrow="Workspace home"
        title={`Welcome back to ${data?.workspace?.name ?? "your workspace"}`}
        description="Recent work, pending processing, and team activity stay visible without forcing you through an admin-heavy dashboard."
        action={<Link to="/upload"><Button><Upload className="h-4 w-4" />Upload a document</Button></Link>}
      />

      <div className="grid gap-4 md:grid-cols-4">
        <MetricCard icon={<Files className="h-4 w-4" />} label="Documents" value={String(stats?.documentCount ?? 0)} />
        <MetricCard icon={<FolderOpen className="h-4 w-4" />} label="Spaces" value={String(stats?.spaceCount ?? 0)} />
        <MetricCard icon={<Users2 className="h-4 w-4" />} label="Teammates" value={String(stats?.memberCount ?? 0)} />
        <MetricCard icon={<Clock3 className="h-4 w-4" />} label="Processing" value={String(stats?.pendingCount ?? 0)} />
      </div>

      <div className="grid gap-6 xl:grid-cols-[1.3fr_0.7fr]">
        <SectionCard
          title="Recent documents"
          description="Keep the latest work visible and searchable."
          action={<Link to="/documents" className="text-sm font-medium text-[var(--accent-strong)]">View all</Link>}
        >
          <DocumentTable data={recentDocuments} />
        </SectionCard>

        <SectionCard title="Pending work" description="Uploads still building preview or OCR.">
          {pending.length ? (
            <div className="space-y-3">
              {pending.map((item) => (
                <div key={String(item.id)} className="rounded-2xl border border-[var(--line)] bg-[var(--surface-soft)] px-4 py-4">
                  <div className="flex items-center justify-between gap-3">
                    <div>
                      <div className="text-sm font-semibold text-[var(--ink)]">{String(item.title)}</div>
                      <div className="mt-1 text-sm text-[var(--ink-soft)]">{String(item.summaryStatus ?? "processing")}</div>
                    </div>
                    <Badge tone="warning">{String(item.previewStatus ?? "processing")}</Badge>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <EmptyState title="Nothing queued right now" description="New uploads will appear here while OCR and previews finish." />
          )}
        </SectionCard>
      </div>

      <div className="grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
        <SectionCard title="Team activity" description="A light view of what changed around you.">
          {activity.length ? (
            <div className="space-y-3">
              {activity.map((item) => (
                <div key={String(item.id)} className="rounded-2xl border border-[var(--line)] px-4 py-4">
                  <div className="text-sm font-semibold text-[var(--ink)]">{String(item.title)}</div>
                  <p className="mt-1 text-sm leading-6 text-[var(--ink-soft)]">{String(item.body)}</p>
                  <div className="mt-3 text-xs uppercase tracking-[0.2em] text-[var(--ink-muted)]">{formatDate(String(item.createdAt))}</div>
                </div>
              ))}
            </div>
          ) : (
            <EmptyState title="No fresh activity" description="Comments, uploads, and shares from your workspace will show up here." />
          )}
        </SectionCard>

        <SectionCard title="Teammates" description="Keep collaboration one step away.">
          <div className="grid gap-3 sm:grid-cols-2">
            {teammates.map((item, index) => (
              <div
                key={String(item.id)}
                className="animate-rise rounded-2xl border border-[var(--line)] bg-white px-4 py-4"
                style={{ animationDelay: `${index * 60}ms` }}
              >
                <div className="text-sm font-semibold text-[var(--ink)]">{String(item.fullName)}</div>
                <div className="mt-1 text-sm text-[var(--ink-soft)]">{String(item.email)}</div>
                <div className="mt-3">
                  <Badge>{String(item.role)}</Badge>
                </div>
              </div>
            ))}
          </div>
        </SectionCard>
      </div>
    </div>
  );
}

export function DocumentsPage() {
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);
  const documents = useDocuments(deferredQuery);

  if (documents.isLoading) return <DocumentListSkeleton />;
  if (documents.isError) return <ErrorState title="Documents unavailable" />;

  const items = ((documents.data as Record<string, any>)?.items ?? []) as DocumentListItem[];

  return (
    <div className="space-y-6">
      <PageIntro
        eyebrow="Documents"
        title="A cleaner shared library"
        description="Search by title, keep version state visible, and scan summaries before opening each file."
        action={<Link to="/upload"><Button><Plus className="h-4 w-4" />New upload</Button></Link>}
      />

      <Card>
        <CardContent className="flex flex-col gap-4 py-5 md:flex-row md:items-center md:justify-between">
          <div className="relative w-full max-w-xl">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--ink-muted)]" />
            <Input
              className="pl-10"
              placeholder="Search title or filename"
              value={query}
              onChange={(event) => startTransition(() => setQuery(event.target.value))}
            />
          </div>
          <div className="text-sm text-[var(--ink-soft)]">{items.length} visible documents</div>
        </CardContent>
      </Card>

      <DocumentTable data={items} />
    </div>
  );
}

export function DocumentDetailPage() {
  const { documentId = "" } = useParams();
  const document = useDocument(documentId);
  const addComment = useAddComment();
  const shareDocument = useShareDocument();
  const commentForm = useForm<z.infer<typeof commentSchema>>({
    resolver: zodResolver(commentSchema),
    defaultValues: { body: "" },
  });
  const shareForm = useForm<z.infer<typeof shareSchema>>({
    resolver: zodResolver(shareSchema),
    defaultValues: { userId: "" },
  });
  const [shareState, setShareState] = useState<string | null>(null);

  if (document.isLoading) return <LoadingState title="Loading document" />;
  if (document.isError) return <ErrorState title="Document unavailable" />;

  const detail = (document.data as Record<string, any>)?.document as Record<string, any>;
  const comments = (detail?.comments ?? []) as Array<Record<string, any>>;
  const metadata = (detail?.metadata ?? []) as Array<Record<string, any>>;
  const versions = (detail?.versions ?? []) as Array<Record<string, any>>;
  const findings = (detail?.findings ?? []) as string[];
  const contacts = (detail?.contacts ?? []) as string[];
  const sensitive = (detail?.sensitive ?? []) as string[];

  return (
    <div className="space-y-6">
      <PageIntro
        eyebrow="Document detail"
        title={String(detail?.title ?? "Untitled document")}
        description={String(detail?.originalFilename ?? "")}
        action={
          <div className="flex gap-2">
            {detail?.downloadUrl ? (
              <a href={String(detail.downloadUrl)} target="_blank" rel="noreferrer">
                <Button variant="secondary">Download</Button>
              </a>
            ) : null}
            <Link to="/documents"><Button variant="secondary">Back</Button></Link>
          </div>
        }
      />

      <div className="grid gap-6 lg:grid-cols-[1.3fr_0.7fr]">
        <div className="space-y-6">
          <Card className="border-[var(--line)]">
            <CardContent className="space-y-4 py-5">
              <div className="flex flex-wrap gap-2">
                <Badge tone={getStatusTone(String(detail?.status ?? "processing"))}>{String(detail?.status ?? "processing")}</Badge>
                <Badge>{String(detail?.summaryStatus ?? "processing")} summary</Badge>
                <Badge>{String(detail?.spaceName ?? "No space")}</Badge>
              </div>

              <div className="grid gap-3 sm:grid-cols-3">
                <InfoBlock label="Version" value={`v${String(detail?.currentVersionNumber ?? 1)}`} />
                <InfoBlock label="Size" value={formatBytes(Number(detail?.sizeBytes ?? 0))} />
                <InfoBlock label="Updated" value={formatDate(String(detail?.updatedAt ?? ""))} />
              </div>

              <div className="space-y-3">
                <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface-soft)] px-4 py-4">
                  <div className="text-xs font-semibold uppercase tracking-widest text-[var(--ink-muted)]">Summary</div>
                  <p className="mt-2 text-sm leading-7 text-[var(--ink)]">
                    {String(detail?.summary || "Summary will appear after OCR completes.")}
                  </p>
                </div>

                {findings.length > 0 && (
                  <div className="rounded-2xl border border-[var(--line)] px-4 py-4">
                    <div className="text-xs font-semibold uppercase tracking-widest text-[var(--ink-muted)]">Key findings</div>
                    <ul className="mt-2 space-y-1.5">
                      {findings.map((f) => (
                        <li key={f} className="flex items-start gap-2 text-sm leading-6 text-[var(--ink)]">
                          <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full" style={{ background: "var(--accent)" }} />
                          {f}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {(contacts.length > 0 || sensitive.length > 0) && (
                  <div className="rounded-2xl border border-[var(--line)] px-4 py-4">
                    <div className="text-xs font-semibold uppercase tracking-widest text-[var(--ink-muted)]">Contacts & sensitive</div>
                    <ul className="mt-2 space-y-1">
                      {[...contacts, ...sensitive].map((item) => (
                        <li key={item} className="text-sm leading-6" style={{ color: "var(--ink-soft)", fontFamily: "var(--font-mono, monospace)" }}>
                          {item}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          <SectionCard title="Comments" description="Keep discussion next to the source file.">
            <form
              className="mb-4 space-y-3"
              onSubmit={commentForm.handleSubmit(async (values) => {
                await addComment.mutateAsync({ documentId, body: values.body });
                commentForm.reset();
              })}
            >
              <Textarea rows={3} placeholder="Leave context, approvals, or questions here" {...commentForm.register("body")} />
              <div className="flex justify-end">
                <Button type="submit" disabled={addComment.isPending}>
                  <MessageSquare className="h-4 w-4" />
                  {addComment.isPending ? "Posting..." : "Add comment"}
                </Button>
              </div>
            </form>

            <div className="space-y-3">
              {comments.length ? comments.map((comment) => (
                <div key={String(comment.id)} className="rounded-2xl border border-[var(--line)] px-4 py-3">
                  <div className="flex items-center justify-between gap-3">
                    <div className="text-sm font-semibold text-[var(--ink)]">{String(comment.authorName)}</div>
                    <div className="text-xs text-[var(--ink-muted)]">{formatDate(String(comment.createdAt))}</div>
                  </div>
                  <p className="mt-2 text-sm leading-6 text-[var(--ink-soft)]">{String(comment.body)}</p>
                </div>
              )) : <EmptyState title="No comments yet" description="Leave review feedback or context for your team." />}
            </div>
          </SectionCard>
        </div>

        <div className="space-y-4">
          <SectionCard title="Share" description="Share with a teammate by user ID or email.">
            <form
              className="space-y-2"
              onSubmit={shareForm.handleSubmit(async (values) => {
                try {
                  await shareDocument.mutateAsync({ documentId, userId: values.userId, accessLevel: "viewer" });
                  shareForm.reset();
                  setShareState("Shared successfully");
                } catch (error) {
                  setShareState(error instanceof Error ? error.message : "Could not share");
                }
              })}
            >
              <Input placeholder="teammate@company.com" {...shareForm.register("userId")} />
              <Button className="w-full" type="submit" disabled={shareDocument.isPending}>
                <Share2 className="h-4 w-4" />
                {shareDocument.isPending ? "Sharing..." : "Share"}
              </Button>
            </form>
            {shareState ? <p className="mt-2 text-sm text-[var(--ink-soft)]">{shareState}</p> : null}
            <p className="mt-3 text-xs text-[var(--ink-muted)]">
              Shared with {String(detail?.sharedWithCount ?? 0)} teammate{Number(detail?.sharedWithCount ?? 0) === 1 ? "" : "s"}
            </p>
          </SectionCard>

          <SectionCard title="Metadata" description="File context from upload.">
            <div className="space-y-2">
              {metadata.length ? metadata.map((item) => (
                <div key={String(item.schemaKey)} className="flex items-start justify-between gap-3 rounded-xl border border-[var(--line)] px-3 py-2.5">
                  <div className="text-xs font-medium capitalize text-[var(--ink-soft)]">{String(item.schemaKey)}</div>
                  <div className="max-w-[55%] text-right text-xs text-[var(--ink)]">
                    {String(item.valueText ?? item.valueDate ?? item.valueBoolean ?? "—")}
                  </div>
                </div>
              )) : <p className="text-sm text-[var(--ink-muted)]">No metadata added during upload.</p>}
            </div>
          </SectionCard>

          <SectionCard title="Version history" description="Every update is tracked.">
            <div className="space-y-2">
              {versions.length ? versions.map((version) => (
                <div key={String(version.id)} className="rounded-xl border border-[var(--line)] px-3 py-3">
                  <div className="flex items-center justify-between gap-2">
                    <div className="text-sm font-semibold text-[var(--ink)]">v{String(version.versionNumber)}</div>
                    <div className="text-xs text-[var(--ink-muted)]">{formatDate(String(version.createdAt))}</div>
                  </div>
                  <div className="mt-1 text-xs text-[var(--ink-soft)]">{String(version.changeSummary || "—")}</div>
                </div>
              )) : <p className="text-sm text-[var(--ink-muted)]">No versions yet.</p>}
            </div>
          </SectionCard>
        </div>
      </div>
    </div>
  );
}

export function UploadPage() {
  const navigate = useNavigate();
  const spaces = useSpaces();
  const uploadDocument = useUploadDocument();
  const [file, setFile] = useState<File | null>(null);
  const [feedback, setFeedback] = useState<string | null>(null);
  const form = useForm<z.infer<typeof uploadSchema>>({
    resolver: zodResolver(uploadSchema),
    defaultValues: {
      title: "",
      collectionId: "",
      department: "",
      tags: "",
      changeSummary: "",
    },
  });

  return (
    <div className="space-y-6">
      <PageIntro
        eyebrow="Upload"
        title="Get a file into the workspace quickly"
        description="Upload once, then let previewing, OCR, and summary generation happen in the background."
      />

      <div className="grid gap-6 xl:grid-cols-[0.92fr_1.08fr]">
        <Card className="border-[var(--line-strong)] bg-[var(--surface-soft)]">
          <CardContent className="py-6">
            <div className="rounded-[30px] border border-dashed border-[var(--line-strong)] bg-white px-6 py-10 text-center">
              <Upload className="mx-auto h-8 w-8 text-[var(--accent-strong)]" />
              <div className="mt-4 text-lg font-semibold text-[var(--ink)]">
                {file ? file.name : "Drop a file here or choose one"}
              </div>
              <p className="mt-2 text-sm leading-6 text-[var(--ink-soft)]">
                PDF, DOCX, image, or plain text. The API stores it immediately and processes the rest asynchronously.
              </p>
              <label className="mt-5 inline-flex cursor-pointer">
                <input
                  className="hidden"
                  type="file"
                  onChange={(event) => setFile(event.target.files?.[0] ?? null)}
                />
                <span className="inline-flex h-11 items-center rounded-xl border border-[var(--line-strong)] bg-[var(--ink)] px-5 text-sm font-medium text-white">
                  Choose file
                </span>
              </label>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="py-6">
            <form
              className="space-y-4"
              onSubmit={form.handleSubmit(async (values) => {
                if (!file) {
                  setFeedback("Choose a file before uploading.");
                  return;
                }
                setFeedback(null);
                try {
                  const response = await uploadDocument.mutateAsync({
                    file,
                    metadata: {
                      title: values.title,
                      collectionId: values.collectionId,
                      changeSummary: values.changeSummary,
                      department: values.department,
                      tags: values.tags,
                    },
                  });
                  const documentId = (response as Record<string, any>)?.document?.id;
                  navigate(documentId ? `/documents/${documentId}` : "/documents");
                } catch (error) {
                  setFeedback(error instanceof Error ? error.message : "Upload failed");
                }
              })}
            >
              <FormField label="Document title" htmlFor="document-title">
                <Input id="document-title" placeholder="Q2 partner agreement" {...form.register("title")} />
              </FormField>

              <FormField label="Space" htmlFor="document-space">
                <Select
                  id="document-space"
                  options={[
                    { label: "No space", value: "" },
                    ...(((spaces.data as Record<string, any> | undefined)?.items ?? []) as Array<{ id: string; name: string }>).map((item) => ({
                      label: item.name,
                      value: item.id,
                    })),
                  ]}
                  {...form.register("collectionId")}
                />
              </FormField>

              <FormField label="Department" htmlFor="document-department">
                <Input id="document-department" placeholder="Operations" {...form.register("department")} />
              </FormField>

              <FormField label="Tags" htmlFor="document-tags" helper="Separate tags with commas.">
                <Input id="document-tags" placeholder="legal, renewal, partner" {...form.register("tags")} />
              </FormField>

              <FormField label="Change summary" htmlFor="document-change-summary">
                <Textarea id="document-change-summary" rows={4} placeholder="Initial upload of the signed draft." {...form.register("changeSummary")} />
              </FormField>

              {feedback ? <p className="text-sm text-[var(--ink-soft)]">{feedback}</p> : null}

              <Button className="w-full" type="submit" disabled={uploadDocument.isPending}>
                {uploadDocument.isPending ? "Uploading..." : "Upload document"}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export function SearchPage() {
  const [query, setQuery] = useState("");
  const deferred = useDeferredValue(query);
  const search = useSearch(deferred);
  const items = ((search.data as Record<string, any> | undefined)?.items ?? []) as Array<Record<string, any>>;

  return (
    <div className="space-y-6">
      <PageIntro
        eyebrow="Search"
        title="Search with more context"
        description="Result cards surface the best snippet, extracted summary, and where the match came from."
      />

      <Card>
        <CardContent className="py-5">
          <div className="relative">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--ink-muted)]" />
            <Input
              className="pl-10"
              placeholder="Search documents, OCR text, or metadata"
              value={query}
              onChange={(event) => startTransition(() => setQuery(event.target.value))}
            />
          </div>
        </CardContent>
      </Card>

      {deferred.length === 0 ? (
        <EmptyState title="Start typing to search" description="Look up filenames, extracted text, or metadata without leaving the workspace." />
      ) : search.isLoading ? (
        <SearchSkeleton />
      ) : search.isError ? (
        <ErrorState title="Search unavailable" />
      ) : items.length ? (
        <div className="space-y-3">
          {items.map((item, index) => (
            <Link key={String(item.documentId)} to={`/documents/${String(item.documentId)}`}>
              <Card
                className="animate-rise transition duration-200 hover:-translate-y-0.5 hover:border-[var(--line-strong)] hover:shadow-[0_14px_40px_rgba(22,27,24,0.06)]"
                style={{ animationDelay: `${index * 50}ms` }}
              >
                <CardContent className="py-5">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <div className="text-base font-semibold text-[var(--ink)]">{String(item.title)}</div>
                      <div className="mt-1 text-sm text-[var(--ink-soft)]">{String(item.originalFilename)}</div>
                    </div>
                    <div className="flex gap-2">
                      <Badge tone={getStatusTone(String(item.status))}>{String(item.status)}</Badge>
                      <Badge>{String(item.matchSource ?? "title")}</Badge>
                    </div>
                  </div>
                  <p className="mt-4 text-sm leading-7 text-[var(--ink-soft)]">{String(item.summary || item.snippet || "No snippet available.")}</p>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      ) : (
        <EmptyState title="No results found" description="Try a filename, department, or phrase from the document body." />
      )}
    </div>
  );
}

export function SpacesPage() {
  const spaces = useSpaces();
  const createSpace = useCreateSpace();
  const deleteSpace = useDeleteSpace();
  const form = useForm<z.infer<typeof createSpaceSchema>>({
    resolver: zodResolver(createSpaceSchema),
    defaultValues: { name: "", description: "" },
  });

  if (spaces.isLoading) return <LoadingState title="Loading spaces" />;
  if (spaces.isError) return <ErrorState title="Spaces unavailable" />;

  const items = ((spaces.data as Record<string, any>)?.items ?? []) as Array<Record<string, any>>;

  return (
    <div className="space-y-6">
      <PageIntro
        eyebrow="Spaces"
        title="Organize work by team context"
        description="Use spaces for a little structure, not a heavy taxonomy. Keep them small and obvious."
      />

      <div className="grid gap-6 xl:grid-cols-[0.88fr_1.12fr]">
        <SectionCard title="Create a space" description="A simple label for a stream of related documents.">
          <form
            className="space-y-4"
            onSubmit={form.handleSubmit(async (values) => {
              await createSpace.mutateAsync(values);
              form.reset();
            })}
          >
            <FormField label="Space name" htmlFor="space-name">
              <Input id="space-name" placeholder="Operations" {...form.register("name")} />
            </FormField>
            <FormField label="Description" htmlFor="space-description">
              <Textarea id="space-description" rows={4} placeholder="Shared runbooks, contracts, and operating records." {...form.register("description")} />
            </FormField>
            <Button type="submit" disabled={createSpace.isPending}>
              <Plus className="h-4 w-4" />
              {createSpace.isPending ? "Creating..." : "Create space"}
            </Button>
          </form>
        </SectionCard>

        <SectionCard title="Existing spaces" description="Keep the list short and legible.">
          {items.length ? (
            <div className="space-y-3">
              {items.map((item) => (
                <div key={String(item.id)} className="flex items-start justify-between gap-4 rounded-2xl border border-[var(--line)] px-4 py-4">
                  <div>
                    <div className="text-sm font-semibold text-[var(--ink)]">{String(item.name)}</div>
                    <p className="mt-1 text-sm leading-6 text-[var(--ink-soft)]">{String(item.description || "No description yet.")}</p>
                    <div className="mt-3 text-xs uppercase tracking-[0.16em] text-[var(--ink-muted)]">Created {formatDate(String(item.createdAt))}</div>
                  </div>
                  <Button variant="ghost" onClick={() => deleteSpace.mutate(String(item.id))}>Remove</Button>
                </div>
              ))}
            </div>
          ) : (
            <EmptyState title="No spaces yet" description="Create a space when your team needs a small amount of structure." />
          )}
        </SectionCard>
      </div>
    </div>
  );
}

export function SharedWithMePage() {
  const shared = useSharedDocuments();

  if (shared.isLoading) return <LoadingState title="Loading shared documents" />;
  if (shared.isError) return <ErrorState title="Shared documents unavailable" />;

  const items = ((shared.data as Record<string, any>)?.items ?? []) as DocumentListItem[];

  return (
    <div className="space-y-6">
      <PageIntro
        eyebrow="Shared with me"
        title="Files others brought into your flow"
        description="Shared work lands here with the same clean list and status context as your own documents."
      />
      <DocumentTable data={items} />
    </div>
  );
}

export function TeamPage() {
  const teamInfo = useTeamInfo();
  const teamMembers = useTeamMembers();
  const createInvite = useCreateInvite();
  const updateTeam = useUpdateTeam();
  const updateMemberRole = useUpdateMemberRole();
  const removeMember = useRemoveMember();
  const deleteTeam = useDeleteTeam();
  const [inviteCode, setInviteCode] = useState<string | null>(null);

  const form = useForm<z.infer<typeof updateTeamSchema>>({
    resolver: zodResolver(updateTeamSchema),
    defaultValues: {
      name: "",
      slug: "",
    },
  });

  if (teamInfo.isLoading || teamMembers.isLoading) return <LoadingState title="Loading team" />;
  if (teamInfo.isError || teamMembers.isError) return <ErrorState title="Team unavailable" />;

  const info = (teamInfo.data ?? {}) as Record<string, any>;
  const members = ((teamMembers.data as Record<string, any>)?.items ?? []) as Array<Record<string, any>>;

  if (!form.getValues("name")) {
    form.reset({ name: String(info.name ?? ""), slug: String(info.slug ?? "") });
  }

  return (
    <div className="space-y-6">
      <PageIntro
        eyebrow="Team"
        title={String(info.name ?? "Workspace team")}
        description="Invites, lightweight roles, and a small settings surface instead of a bulky admin center."
      />

      <div className="grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
        <SectionCard title="Workspace settings" description="Keep the basics easy to adjust.">
          <form
            className="space-y-4"
            onSubmit={form.handleSubmit(async (values) => {
              await updateTeam.mutateAsync(values);
            })}
          >
            <FormField label="Workspace name" htmlFor="team-name">
              <Input id="team-name" {...form.register("name")} />
            </FormField>
            <FormField label="Workspace slug" htmlFor="team-slug">
              <Input id="team-slug" {...form.register("slug")} />
            </FormField>
            <div className="flex flex-wrap gap-3">
              <Button type="submit" disabled={updateTeam.isPending}>
                {updateTeam.isPending ? "Saving..." : "Save changes"}
              </Button>
              <Button
                type="button"
                variant="secondary"
                onClick={async () => {
                  const invite = await createInvite.mutateAsync({ expiresInHours: 72, maxUses: 20 });
                  setInviteCode(String((invite as Record<string, any>).code));
                }}
              >
                <Link2 className="h-4 w-4" />
                Generate invite
              </Button>
            </div>

            {inviteCode ? (
              <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface-soft)] px-4 py-4">
                <div className="text-xs uppercase tracking-[0.18em] text-[var(--ink-muted)]">Invite code</div>
                <div className="mt-2 text-lg font-semibold tracking-[0.18em] text-[var(--ink)]">{inviteCode}</div>
              </div>
            ) : null}

            <Button
              type="button"
              variant="ghost"
              onClick={async () => {
                if (window.confirm("Delete this workspace and remove everyone from it?")) {
                  await deleteTeam.mutateAsync();
                }
              }}
            >
              Delete workspace
            </Button>
          </form>
        </SectionCard>

        <SectionCard title="Members" description="Only three roles: owner, editor, viewer.">
          <div className="space-y-3">
            {members.map((member, index) => (
              <div
                key={String(member.id)}
                className="animate-rise rounded-2xl border border-[var(--line)] px-4 py-4"
                style={{ animationDelay: `${index * 50}ms` }}
              >
                <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                  <div>
                    <div className="text-sm font-semibold text-[var(--ink)]">{String(member.fullName)}</div>
                    <div className="mt-1 text-sm text-[var(--ink-soft)]">{String(member.email)}</div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Select
                      className="w-[150px]"
                      options={[
                        { label: "Owner", value: "owner" },
                        { label: "Editor", value: "editor" },
                        { label: "Viewer", value: "viewer" },
                      ]}
                      value={String(member.role ?? member.roles?.[0] ?? "viewer")}
                      onChange={(event) => updateMemberRole.mutate({ memberId: String(member.id), roleKey: event.target.value })}
                    />
                    <Button variant="ghost" onClick={() => removeMember.mutate(String(member.id))}>
                      Remove
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </SectionCard>
      </div>
    </div>
  );
}

function GettingStartedStrip() {
  const [dismissed, setDismissed] = useState(() => {
    try { return localStorage.getItem("verin:onboarding-dismissed") === "1"; } catch { return false; }
  });

  if (dismissed) return null;

  return (
    <div
      className="animate-rise relative flex flex-col gap-4 rounded-2xl px-5 py-5 sm:flex-row sm:items-center sm:justify-between"
      style={{ background: "var(--accent-soft)", border: "1px solid var(--accent)" }}
    >
      <div>
        <div className="text-xs font-semibold uppercase tracking-widest" style={{ color: "var(--accent-strong)" }}>
          Getting started
        </div>
        <div className="mt-3 flex flex-wrap gap-4">
          <Step done label="Workspace created" />
          <Step done={false} label="Upload your first document" to="/upload" />
          <Step done={false} label="Invite a teammate" to="/team" />
        </div>
      </div>
      <button
        type="button"
        className="absolute right-4 top-4 rounded-lg p-1 transition-colors hover:bg-[var(--accent)]"
        style={{ color: "var(--accent-strong)" }}
        aria-label="Dismiss"
        onClick={() => {
          try { localStorage.setItem("verin:onboarding-dismissed", "1"); } catch { /* ignore */ }
          setDismissed(true);
        }}
      >
        ×
      </button>
    </div>
  );
}

function Step({ done, label, to }: { done: boolean; label: string; to?: string }) {
  const content = (
    <div className="flex items-center gap-2 text-sm font-medium" style={{ color: done ? "var(--ink-muted)" : "var(--accent-strong)" }}>
      <span
        className="flex h-5 w-5 shrink-0 items-center justify-center rounded-full text-xs"
        style={{
          background: done ? "var(--accent)" : "var(--surface)",
          border: `1px solid ${done ? "var(--accent)" : "var(--accent)"}`,
          color: done ? "#fff" : "var(--accent-strong)",
        }}
      >
        {done ? "✓" : "→"}
      </span>
      {label}
    </div>
  );
  if (to) return <Link to={to} className="hover:opacity-80 transition-opacity">{content}</Link>;
  return <div>{content}</div>;
}

function PreviewPanel({ eyebrow, title, icon }: { eyebrow: string; title: string; icon: ReactNode }) {
  return (
    <div className="rounded-[26px] border border-[var(--line)] bg-white px-5 py-5 transition duration-300 hover:-translate-y-1 hover:shadow-[0_16px_40px_rgba(22,27,24,0.06)]">
      <div className="inline-flex h-9 w-9 items-center justify-center rounded-full bg-[var(--surface-soft)] text-[var(--accent-strong)]">
        {icon}
      </div>
      <div className="mt-4 text-xs uppercase tracking-[0.22em] text-[var(--ink-muted)]">{eyebrow}</div>
      <div className="mt-2 text-base font-semibold leading-7 text-[var(--ink)]">{title}</div>
    </div>
  );
}

function CenteredLayout({ title, description, aside, children }: { title: string; description: string; aside: string; children: ReactNode }) {
  return (
    <div className="min-h-screen bg-[var(--bg)] px-6 py-8">
      <div className="mx-auto grid max-w-5xl gap-10 lg:grid-cols-[0.94fr_1.06fr]">
        <div className="space-y-5 pt-8">
          <Logo size="lg" />
          <div className="space-y-3">
            <h1 className="font-display text-4xl tracking-[-0.04em] text-[var(--ink)]">{title}</h1>
            <p className="max-w-md text-base leading-8 text-[var(--ink-soft)]">{description}</p>
          </div>
          <Card className="border-[var(--line)] bg-[var(--surface-soft)]">
            <CardContent className="py-5 text-sm leading-7 text-[var(--ink-soft)]">{aside}</CardContent>
          </Card>
        </div>
        <div className="py-8">{children}</div>
      </div>
    </div>
  );
}

function CenteredLoading({ title }: { title: string }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-[var(--bg)] px-6">
      <LoadingState title={title} />
    </div>
  );
}

function FormField({
  label,
  htmlFor,
  helper,
  children,
}: {
  label: string;
  htmlFor: string;
  helper?: string;
  children: ReactNode;
}) {
  return (
    <div>
      <Label htmlFor={htmlFor}>{label}</Label>
      {children}
      {helper ? <p className="mt-2 text-sm text-[var(--ink-muted)]">{helper}</p> : null}
    </div>
  );
}

function PageIntro({
  eyebrow,
  title,
  description,
  action,
}: {
  eyebrow: string;
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
      <div>
        <div className="text-xs uppercase tracking-[0.22em] text-[var(--ink-muted)]">{eyebrow}</div>
        <h1 className="mt-3 font-display text-4xl tracking-[-0.04em] text-[var(--ink)]">{title}</h1>
        <p className="mt-3 max-w-2xl text-base leading-8 text-[var(--ink-soft)]">{description}</p>
      </div>
      {action}
    </div>
  );
}

function MetricCard({ icon, label, value }: { icon: ReactNode; label: string; value: string }) {
  return (
    <Card className="border-[var(--line)] bg-white/95">
      <CardContent className="py-5">
        <div className="flex items-center justify-between text-[var(--accent-strong)]">{icon}<span className="text-xs uppercase tracking-[0.2em] text-[var(--ink-muted)]">{label}</span></div>
        <div className="mt-4 font-display text-4xl tracking-[-0.05em] text-[var(--ink)]">{value}</div>
      </CardContent>
    </Card>
  );
}

function SectionCard({
  title,
  description,
  action,
  children,
}: {
  title: string;
  description: string;
  action?: ReactNode;
  children: ReactNode;
}) {
  return (
    <Card className="border-[var(--line)] shadow-[0_18px_50px_rgba(22,27,24,0.04)]">
      <CardHeader className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
        <div>
          <CardTitle className="text-base">{title}</CardTitle>
          <CardDescription className="mt-1">{description}</CardDescription>
        </div>
        {action}
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  );
}

function ErrorBanner({ message }: { message: string }) {
  return (
    <div className="mb-4 rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
      {message}
    </div>
  );
}

function InfoBlock({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-[var(--line)] px-4 py-4">
      <div className="text-xs uppercase tracking-[0.18em] text-[var(--ink-muted)]">{label}</div>
      <div className="mt-2 text-base font-semibold text-[var(--ink)]">{value}</div>
    </div>
  );
}
