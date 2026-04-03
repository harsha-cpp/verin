import { useState, useRef, type ChangeEvent, type ReactNode, type DragEvent } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  FileArchive,
  FileSearch2,
  FileText,
  FolderOpen,
  HardDrive,
  Plus,
  Search,
  Trash2,
  Upload as UploadIcon,
  Users2,
} from "lucide-react";
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
  Separator,
  Textarea,
  cn,
} from "@verin/ui";
import { useNavigate, useParams } from "react-router-dom";

import { DocumentTable } from "@/components/document-table";
import { Logo } from "@/components/logo";
import { EmptyState, ErrorState, LoadingState } from "@/components/states";
import { dubTheme, initialsFromName } from "@/design-system/dub-theme";
import {
  useAdminData,
  useAudit,
  useCollections,
  useCreateCollection,
  useDeleteCollection,
  useDocument,
  useDocuments,
  useLogin,
  useSearch,
  useSession,
  useSharedDocuments,
  useUploadDocument,
} from "@/hooks/use-api";
import { formatBytes, formatDate, getStatusTone } from "@/lib/utils";

const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
});

export function LoginPage() {
  const navigate = useNavigate();
  const login = useLogin();
  const [submitError, setSubmitError] = useState<string | null>(null);
  const form = useForm<z.infer<typeof loginSchema>>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "admin@verin.local", password: "verin123!" },
  });

  return (
    <div className="flex min-h-screen bg-gray-50">
      <div className="flex flex-1 flex-col items-center justify-center px-4 py-12">
        <div className="w-full max-w-sm">
          <div className="mb-8 flex flex-col items-center">
            <Logo size="lg" />
            <p className="mt-3 text-center text-sm text-slate-500">
              Document management for teams that care about compliance.
            </p>
          </div>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Log in to your account</CardTitle>
              <CardDescription>Enter your credentials to access the workspace.</CardDescription>
            </CardHeader>
            <CardContent>
              <form
                className="space-y-4"
                onSubmit={form.handleSubmit(async (values) => {
                  setSubmitError(null);
                  try {
                    await login.mutateAsync(values);
                    navigate("/dashboard");
                  } catch {
                    setSubmitError("Sign-in failed. Check that the API is reachable and the credentials are correct.");
                  }
                })}
              >
                <FormField label="Email" htmlFor="email">
                  <Input id="email" {...form.register("email")} />
                </FormField>

                <FormField label="Password" htmlFor="password">
                  <Input id="password" type="password" {...form.register("password")} />
                </FormField>

                <Button className="mt-1 w-full" type="submit">
                  Log in
                </Button>

                {submitError && (
                  <p className="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-600">
                    {submitError}
                  </p>
                )}
              </form>
            </CardContent>
          </Card>

          <div className="mt-6 rounded-xl border border-slate-200 bg-white px-5 py-4">
            <div className="text-xs font-medium text-slate-500 mb-2">Demo accounts</div>
            <div className="space-y-1.5">
              {[
                { email: "admin@verin.local", role: "Admin" },
                { email: "editor@verin.local", role: "Editor" },
                { email: "auditor@verin.local", role: "Auditor" },
              ].map((account) => (
                <div key={account.email} className="flex items-center justify-between text-sm">
                  <span className="font-medium text-slate-700">{account.email}</span>
                  <Badge>{account.role}</Badge>
                </div>
              ))}
            </div>
            <p className="mt-3 text-xs text-slate-400">Password for all accounts: verin123!</p>
          </div>
        </div>
      </div>

      <div className="hidden w-[480px] flex-col justify-between border-l border-slate-200 bg-white px-10 py-12 lg:flex">
        <div>
          <div className="text-xs font-medium text-slate-400">Verin DMS</div>
          <h2 className="mt-4 text-2xl font-semibold tracking-tight text-slate-900">
            Quiet control for documents that matter.
          </h2>
          <p className="mt-3 text-sm leading-6 text-slate-500">
            Upload, search, retain, audit, and govern critical records through a workspace built for compliance-first teams.
          </p>
        </div>

        <div className="space-y-3">
          {[
            { label: "Full-text search", desc: "OCR and metadata indexed together" },
            { label: "Retention policies", desc: "Archive and deletion rules per collection" },
            { label: "Audit trail", desc: "Every sensitive action tracked and exportable" },
          ].map((item) => (
            <div key={item.label} className="rounded-xl border border-slate-100 bg-slate-50 px-4 py-3">
              <div className="text-sm font-medium text-slate-900">{item.label}</div>
              <div className="mt-0.5 text-xs text-slate-500">{item.desc}</div>
            </div>
          ))}
        </div>

        <p className="text-xs text-slate-400">
          By continuing, you agree to Verin's terms and privacy notice.
        </p>
      </div>
    </div>
  );
}

export function DashboardPage() {
  const documents = useDocuments();
  const admin = useAdminData();
  const usage = admin.usage.data;

  if (documents.isLoading) return <LoadingState title="Loading dashboard" />;
  if (documents.isError) return <ErrorState title="Dashboard is unavailable" />;

  return (
    <div className="space-y-5">
      <div className="grid gap-4 md:grid-cols-3">
        <MetricCard icon={FileText} label="Documents" value={`${usage?.documentCount ?? documents.data?.total ?? 0}`} detail="Indexed records" />
        <MetricCard icon={HardDrive} label="Storage" value={formatBytes(usage?.storageBytes ?? 0)} detail="Across current versions" />
        <MetricCard icon={Users2} label="Users" value={`${usage?.userCount ?? 0}`} detail="Provisioned accounts" />
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Recent documents</CardTitle>
              <CardDescription>Latest records in the workspace.</CardDescription>
            </div>
            <Badge>{documents.data?.total ?? 0} total</Badge>
          </div>
        </CardHeader>
        <CardContent>
          <DocumentTable data={documents.data?.items ?? []} />
        </CardContent>
      </Card>
    </div>
  );
}

export function DocumentsPage() {
  const [query, setQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string | null>(null);
  const documents = useDocuments(query);

  const allItems = documents.data?.items ?? [];
  const items = statusFilter
    ? allItems.filter((item: Record<string, unknown>) => item.status === statusFilter)
    : allItems;

  const statusCounts = allItems.reduce<Record<string, number>>((acc, item: Record<string, unknown>) => {
    const s = String(item.status ?? "unknown");
    acc[s] = (acc[s] ?? 0) + 1;
    return acc;
  }, {});

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
          <Input
            className="pl-9"
            value={query}
            onChange={(event: ChangeEvent<HTMLInputElement>) => setQuery(event.target.value)}
            placeholder="Search by title or filename"
          />
        </div>
        <div className="flex gap-1.5">
          <button
            type="button"
            onClick={() => setStatusFilter(null)}
            className={cn(
              "rounded-md border px-2.5 py-1 text-xs font-medium transition",
              !statusFilter ? "border-slate-900 bg-slate-900 text-white" : "border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
            )}
          >
            All ({allItems.length})
          </button>
          {Object.entries(statusCounts).map(([status, count]) => (
            <button
              key={status}
              type="button"
              onClick={() => setStatusFilter(statusFilter === status ? null : status)}
              className={cn(
                "rounded-md border px-2.5 py-1 text-xs font-medium transition",
                statusFilter === status ? "border-slate-900 bg-slate-900 text-white" : "border-slate-200 bg-white text-slate-600 hover:bg-slate-50",
              )}
            >
              {status} ({count})
            </button>
          ))}
        </div>
      </div>

      {documents.isLoading ? (
        <LoadingState title="Loading documents" />
      ) : documents.isError ? (
        <ErrorState title="Documents could not be loaded" />
      ) : (
        <DocumentTable data={items} />
      )}
    </div>
  );
}

export function DocumentDetailPage() {
  const { documentId = "" } = useParams();
  const document = useDocument(documentId);
  const detail = document.data?.document;

  if (document.isLoading) return <LoadingState title="Loading document" />;
  if (document.isError || !detail) return <ErrorState title="Document detail is unavailable" />;

  return (
    <div className="space-y-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-xl font-semibold text-slate-900">{detail.title}</h2>
            <Badge tone={getStatusTone(detail.status)}>{detail.status}</Badge>
          </div>
          <p className="mt-1 text-sm text-slate-500">
            {detail.originalFilename} · {formatBytes(detail.sizeBytes)}
          </p>
        </div>
        <div className="flex gap-2">
          <Button variant="secondary" size="sm">Version history</Button>
          <Button size="sm">Download</Button>
        </div>
      </div>

      <div className="grid gap-4 xl:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Overview</CardTitle>
            <CardDescription>Collection, extraction state, and metadata.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <DetailLine label="Collection" value={detail.collectionName ?? "Unassigned"} />
            <DetailLine label="OCR status" value={detail.ocrStatus ?? "pending"} />
            <Separator />
            <div className="grid gap-3 sm:grid-cols-2">
              {detail.metadata?.map((item: Record<string, unknown>) => (
                <div key={String(item.schemaKey)} className={cn(dubTheme.mutedPanel, "p-3")}>
                  <div className="text-xs font-medium text-slate-500">{String(item.schemaKey)}</div>
                  <div className="mt-1 text-sm font-medium text-slate-900">
                    {String(item.valueText ?? item.valueDate ?? item.valueBoolean ?? "")}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Activity</CardTitle>
            <CardDescription>Versions and comments for this record.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <section>
              <div className="mb-2 text-xs font-medium text-slate-500">Versions</div>
              <div className="space-y-2">
                {detail.versions?.map((version: Record<string, unknown>) => (
                  <div key={String(version.id)} className={cn(dubTheme.mutedPanel, "p-3")}>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-slate-900">v{String(version.versionNumber)}</span>
                      <span className="text-xs text-slate-500">{formatDate(String(version.createdAt))}</span>
                    </div>
                    {String(version.changeSummary ?? "") !== "" ? <p className="mt-1 text-sm text-slate-500">{String(version.changeSummary)}</p> : null}
                  </div>
                ))}
              </div>
            </section>

            <Separator />

            <section>
              <div className="mb-2 text-xs font-medium text-slate-500">Comments</div>
              <div className="space-y-2">
                {detail.comments?.map((comment: Record<string, unknown>) => (
                  <div key={String(comment.id)} className={cn(dubTheme.mutedPanel, "p-3")}>
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-slate-900">{String(comment.authorName)}</span>
                      <span className="text-xs text-slate-500">{formatDate(String(comment.createdAt))}</span>
                    </div>
                    <p className="mt-1 text-sm text-slate-500">{String(comment.body)}</p>
                  </div>
                ))}
              </div>
            </section>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

const uploadSchema = z.object({
  title: z.string().min(1),
  collectionId: z.string().optional(),
  changeSummary: z.string().optional(),
  department: z.string().optional(),
  tags: z.string().optional(),
});

export function UploadPage() {
  const uploadDocument = useUploadDocument();
  const navigate = useNavigate();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [dragging, setDragging] = useState(false);
  const [progress, setProgress] = useState<"idle" | "uploading" | "done" | "error">("idle");
  const [errorMsg, setErrorMsg] = useState("");

  const form = useForm<z.infer<typeof uploadSchema>>({
    resolver: zodResolver(uploadSchema),
    defaultValues: { title: "", department: "", tags: "" },
  });

  function handleDrop(e: DragEvent<HTMLDivElement>) {
    e.preventDefault();
    setDragging(false);
    const dropped = e.dataTransfer.files[0];
    if (dropped) selectFile(dropped);
  }

  function selectFile(f: File) {
    setFile(f);
    if (!form.getValues("title")) {
      form.setValue("title", f.name.replace(/\.[^.]+$/, ""));
    }
  }

  async function handleSubmit(values: z.infer<typeof uploadSchema>) {
    if (!file) return;
    try {
      setProgress("uploading");
      setErrorMsg("");
      await uploadDocument.mutateAsync({
        file,
        metadata: {
          title: values.title,
          collectionId: values.collectionId,
          changeSummary: values.changeSummary,
          department: values.department,
          tags: values.tags,
        },
      });
      setProgress("done");
      setTimeout(() => navigate("/documents"), 1500);
    } catch (err) {
      setProgress("error");
      setErrorMsg(err instanceof Error ? err.message : "Upload failed");
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-5">
      <div
        className={cn(
          "relative flex min-h-[200px] cursor-pointer flex-col items-center justify-center rounded-xl border-2 border-dashed px-6 py-10 text-center transition",
          dragging ? "border-slate-400 bg-slate-100" : "border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50",
          file && "border-emerald-300 bg-emerald-50",
        )}
        onClick={() => fileInputRef.current?.click()}
        onDragOver={(e) => { e.preventDefault(); setDragging(true); }}
        onDragLeave={() => setDragging(false)}
        onDrop={handleDrop}
      >
        <input
          ref={fileInputRef}
          type="file"
          className="hidden"
          onChange={(e) => { if (e.target.files?.[0]) selectFile(e.target.files[0]); }}
        />
        {file ? (
          <>
            <FileText className="h-8 w-8 text-emerald-500" />
            <div className="mt-3 text-sm font-medium text-slate-900">{file.name}</div>
            <div className="text-xs text-slate-500">{formatBytes(file.size)} · {file.type || "unknown"}</div>
            <button
              type="button"
              className="mt-2 text-xs text-slate-400 hover:text-rose-500"
              onClick={(e) => { e.stopPropagation(); setFile(null); setProgress("idle"); }}
            >
              Remove
            </button>
          </>
        ) : (
          <>
            <UploadIcon className="h-8 w-8 text-slate-300" />
            <div className="mt-3 text-sm font-medium text-slate-700">Drop a file here or click to browse</div>
            <div className="mt-1 text-xs text-slate-400">PDF, DOCX, images up to 50 MB</div>
          </>
        )}
      </div>

      {progress !== "idle" && progress !== "error" && (
        <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
          <div
            className={cn(
              "h-1.5 transition-all duration-500",
              progress === "done" ? "w-full bg-emerald-500" : "w-2/3 bg-blue-400",
            )}
          />
          <div className="px-4 py-2.5 text-xs text-slate-500">
            {progress === "uploading" && "Uploading and processing..."}
            {progress === "done" && "Upload complete — redirecting to documents."}
          </div>
        </div>
      )}

      {progress === "error" && (
        <div className="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-600">
          {errorMsg || "Upload failed. Check the API and try again."}
        </div>
      )}

      <Card>
        <CardContent className="pt-5">
          <form className="grid gap-3" onSubmit={form.handleSubmit(handleSubmit)}>
            <FormField label="Title">
              <Input {...form.register("title")} placeholder="Document title" />
            </FormField>

            <div className="grid gap-3 md:grid-cols-2">
              <FormField label="Collection">
                <Input {...form.register("collectionId")} placeholder="Optional collection UUID" />
              </FormField>
              <FormField label="Department">
                <Input {...form.register("department")} placeholder="e.g. Compliance" />
              </FormField>
            </div>

            <FormField label="Tags">
              <Input {...form.register("tags")} placeholder="Comma-separated tags" />
            </FormField>

            <FormField label="Change summary">
              <Textarea {...form.register("changeSummary")} placeholder="Brief description of this version" />
            </FormField>

            <Button type="submit" disabled={!file || progress === "uploading"}>
              {progress === "uploading" ? "Uploading..." : "Upload document"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

export function SearchPage() {
  const [query, setQuery] = useState("compliance");
  const search = useSearch(query);
  const items = search.data?.items ?? [];

  return (
    <div className="space-y-4">
      <Card>
        <CardContent className="pt-6">
          <div className="flex flex-col gap-3 lg:flex-row">
            <div className="relative flex-1">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                className="pl-9"
                value={query}
                onChange={(event: ChangeEvent<HTMLInputElement>) => setQuery(event.target.value)}
                placeholder="Search OCR text or metadata"
              />
            </div>
            <Select
              options={[
                { label: "Any status", value: "all" },
                { label: "Ready", value: "ready" },
                { label: "Processing", value: "processing" },
              ]}
            />
            <Select
              options={[
                { label: "All collections", value: "all" },
                { label: "Compliance", value: "compliance" },
                { label: "Operations", value: "operations" },
              ]}
            />
          </div>
        </CardContent>
      </Card>

      {search.isLoading ? (
        <LoadingState title="Searching documents" />
      ) : items.length === 0 ? (
        <EmptyState title="No matching records" description="Try broader metadata or OCR keywords." />
      ) : (
        <div className="space-y-2">
          {items.map((item: Record<string, unknown>) => (
            <Card key={String(item.documentId)}>
              <CardContent className="flex items-center justify-between py-3">
                <div className="flex items-center gap-3">
                  <div className="flex h-8 w-8 items-center justify-center rounded-md bg-slate-100 text-slate-500">
                    <FileSearch2 className="h-4 w-4" />
                  </div>
                  <div>
                    <div className="text-sm font-medium text-slate-900">{String(item.title)}</div>
                    <div className="text-xs text-slate-500">{String(item.originalFilename)}</div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge tone={getStatusTone(String(item.status))}>{String(item.status)}</Badge>
                  <span className="text-xs text-slate-400">Rank {String(item.rank)}</span>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}

export function AuditPage() {
  const audit = useAudit();
  if (audit.isLoading) return <LoadingState title="Loading audit events" />;
  if (audit.isError) return <ErrorState title="Audit data is unavailable" />;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Audit log</CardTitle>
        <CardDescription>Chronological record of sensitive actions.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        {audit.data?.items?.map((event: Record<string, unknown>) => (
          <div key={String(event.id)} className={cn(dubTheme.mutedPanel, "flex items-center justify-between p-3")}>
            <div>
              <div className="text-sm font-medium text-slate-900">{String(event.action)}</div>
              <div className="text-xs text-slate-500">{String(event.resourceType)} · {String(event.actorRole)}</div>
            </div>
            <div className="text-xs text-slate-400">{formatDate(String(event.createdAt))}</div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}

export function AdminUsersPage() {
  const admin = useAdminData();
  if (admin.users.isLoading) return <LoadingState title="Loading users" />;

  return (
    <AdminScaffold title="Users" description="Accounts, roles, and access state.">
      <div className="grid gap-3 xl:grid-cols-2">
        {admin.users.data?.items?.map((user: Record<string, unknown>) => (
          <Card key={String(user.id)}>
            <CardContent className="py-4">
              <div className="flex items-center justify-between gap-3">
                <div className="flex items-center gap-3">
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-slate-200 text-xs font-medium text-slate-700">
                    {initialsFromName(String(user.fullName))}
                  </div>
                  <div>
                    <div className="text-sm font-medium text-slate-900">{String(user.fullName)}</div>
                    <div className="text-xs text-slate-500">{String(user.email)}</div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge>{Array.isArray(user.roles) ? user.roles.length : 0} roles</Badge>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </AdminScaffold>
  );
}

export function AdminRolesPage() {
  const admin = useAdminData();

  return (
    <AdminScaffold title="Roles" description="Role definitions and permission grouping.">
      <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
        {admin.roles.data?.items?.map((role: Record<string, unknown>) => (
          <Card key={String(role.id)}>
            <CardHeader>
              <CardTitle>{String(role.name)}</CardTitle>
              <CardDescription>{String(role.description ?? "")}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className={cn(dubTheme.mutedPanel, "p-3")}>
                <div className="text-xs text-slate-500">Key</div>
                <div className="mt-1 text-sm font-medium text-slate-900">{String(role.key)}</div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </AdminScaffold>
  );
}

export function AdminQuotasPage() {
  const admin = useAdminData();

  return (
    <AdminScaffold title="Quotas" description="Storage and document limits.">
      <div className="space-y-3">
        {admin.quotas.data?.items?.map((quota: Record<string, unknown>) => (
          <Card key={String(quota.id)}>
            <CardContent className="flex items-center gap-5 py-4">
              <div className="min-w-[100px]">
                <div className="text-xs text-slate-500">Target</div>
                <div className="mt-1 text-sm font-medium capitalize text-slate-900">{String(quota.targetType)}</div>
              </div>
              <InfoCompact icon={HardDrive} label="Max storage" value={formatBytes(Number(quota.maxStorageBytes))} />
              <InfoCompact icon={FileArchive} label="Max documents" value={`${String(quota.maxDocumentCount)}`} />
            </CardContent>
          </Card>
        ))}
      </div>
    </AdminScaffold>
  );
}

export function AdminRetentionPage() {
  const admin = useAdminData();

  return (
    <AdminScaffold title="Retention" description="Policy windows and archive thresholds.">
      <div className="space-y-3">
        {admin.retention.data?.items?.map((policy: Record<string, unknown>) => (
          <Card key={String(policy.id)}>
            <CardContent className="flex items-center justify-between py-4">
              <div>
                <div className="text-sm font-medium text-slate-900">{String(policy.name)}</div>
                <div className="mt-0.5 text-xs text-slate-500">{String(policy.retentionDays)} day retention</div>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-xs text-slate-500">
                  {policy.archiveAfterDays ? `${String(policy.archiveAfterDays)}d archive` : "No auto-archive"}
                </span>
                <Badge tone={Boolean(policy.enabled) ? "success" : "warning"}>
                  {Boolean(policy.enabled) ? "Enabled" : "Paused"}
                </Badge>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </AdminScaffold>
  );
}

export function AdminSettingsPage() {
  const admin = useAdminData();

  return (
    <AdminScaffold title="Settings" description="Workspace configuration, jobs, and health.">
      <div className="space-y-4">
        {admin.settings.data?.items?.map((setting: Record<string, unknown>) => (
          <Card key={String(setting.settingKey)}>
            <CardHeader>
              <CardTitle>{String(setting.settingKey)}</CardTitle>
              <CardDescription>{JSON.stringify(setting.settingValue)}</CardDescription>
            </CardHeader>
          </Card>
        ))}

        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Jobs</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {admin.jobs.data?.items?.map((job: Record<string, unknown>) => (
                <div key={String(job.id)} className="flex items-center justify-between text-sm">
                  <span className="text-slate-700">{String(job.jobType)}</span>
                  <Badge tone={getStatusTone(String(job.status))}>{String(job.status)}</Badge>
                </div>
              ))}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Health</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {Object.entries(admin.health.data ?? {}).map(([key, value]) => (
                <div key={key} className="flex items-center justify-between text-sm">
                  <span className="capitalize text-slate-500">{key}</span>
                  <span className="font-medium text-slate-900">{String(value)}</span>
                </div>
              ))}
            </CardContent>
          </Card>
        </div>
      </div>
    </AdminScaffold>
  );
}

const collectionSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
});

export function CollectionsPage() {
  const collections = useCollections();
  const createCollection = useCreateCollection();
  const deleteCollection = useDeleteCollection();
  const [showForm, setShowForm] = useState(false);
  const form = useForm<z.infer<typeof collectionSchema>>({
    resolver: zodResolver(collectionSchema),
    defaultValues: { name: "", description: "" },
  });

  if (collections.isLoading) return <LoadingState title="Loading collections" />;
  if (collections.isError) return <ErrorState title="Collections unavailable" />;

  const items = collections.data?.items ?? [];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="text-sm text-slate-500">{items.length} collections</div>
        <Button size="sm" onClick={() => setShowForm(!showForm)}>
          <Plus className="h-4 w-4" />
          New collection
        </Button>
      </div>

      {showForm && (
        <Card>
          <CardContent className="pt-5">
            <form
              className="flex items-end gap-3"
              onSubmit={form.handleSubmit(async (values) => {
                await createCollection.mutateAsync(values);
                form.reset();
                setShowForm(false);
              })}
            >
              <div className="flex-1 space-y-1.5">
                <Label>Name</Label>
                <Input {...form.register("name")} placeholder="Collection name" />
              </div>
              <div className="flex-1 space-y-1.5">
                <Label>Description</Label>
                <Input {...form.register("description")} placeholder="Optional" />
              </div>
              <Button type="submit" size="sm">Create</Button>
            </form>
          </CardContent>
        </Card>
      )}

      {items.length === 0 ? (
        <EmptyState title="No collections" description="Create a collection to organize documents." />
      ) : (
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          {items.map((item: Record<string, unknown>) => (
            <Card key={String(item.id)}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <FolderOpen className="h-4 w-4 text-slate-400" />
                    <CardTitle>{String(item.name)}</CardTitle>
                  </div>
                  <button
                    type="button"
                    className="rounded-md p-1 text-slate-400 hover:bg-slate-100 hover:text-rose-500"
                    onClick={() => deleteCollection.mutate(String(item.id))}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
                {String(item.description ?? "") && (
                  <CardDescription>{String(item.description)}</CardDescription>
                )}
              </CardHeader>
              <CardContent>
                <div className="text-xs text-slate-400">{formatDate(String(item.createdAt))}</div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}

export function SharedWithMePage() {
  const shared = useSharedDocuments();

  if (shared.isLoading) return <LoadingState title="Loading shared documents" />;
  if (shared.isError) return <ErrorState title="Shared documents unavailable" />;

  const items = shared.data?.items ?? [];

  return (
    <div className="space-y-4">
      <div className="text-sm text-slate-500">{items.length} documents shared with you</div>

      {items.length === 0 ? (
        <EmptyState title="Nothing shared yet" description="When someone shares a document with you, it will appear here." />
      ) : (
        <DocumentTable data={items} />
      )}
    </div>
  );
}

export function ProfilePage() {
  const session = useSession();
  const user = session.data?.user;

  if (session.isLoading) return <LoadingState title="Loading profile" />;

  return (
    <div className="max-w-2xl space-y-5">
      <Card>
        <CardHeader>
          <CardTitle>Profile</CardTitle>
          <CardDescription>Your account information.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-4">
            <div className="flex h-14 w-14 items-center justify-center rounded-full bg-slate-200 text-lg font-semibold text-slate-700">
              {initialsFromName(String(user?.fullName ?? ""))}
            </div>
            <div>
              <div className="text-base font-semibold text-slate-900">{String(user?.fullName ?? "")}</div>
              <div className="text-sm text-slate-500">{String(user?.email ?? "")}</div>
            </div>
          </div>

          <Separator />

          <div className="grid gap-3 sm:grid-cols-2">
            <DetailLine label="Status" value={String(user?.status ?? "active")} />
          </div>

          {Array.isArray(user?.roles) && user.roles.length > 0 && (
            <div>
              <div className="mb-2 text-xs font-medium text-slate-500">Roles</div>
              <div className="flex flex-wrap gap-1.5">
                {user.roles.map((role: Record<string, unknown>) => (
                  <Badge key={String(role.id ?? role.name)}>{String(role.name)}</Badge>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function MetricCard({ icon: Icon, label, value, detail }: { icon: typeof FileText; label: string; value: string; detail: string }) {
  return (
    <Card>
      <CardContent className="pt-5">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium text-slate-500">{label}</span>
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-slate-100 text-slate-500">
            <Icon className="h-4 w-4" />
          </div>
        </div>
        <div className="mt-3 text-2xl font-semibold text-slate-900">{value}</div>
        <div className="mt-0.5 text-xs text-slate-500">{detail}</div>
      </CardContent>
    </Card>
  );
}

function DetailLine({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between rounded-xl border border-slate-100 bg-slate-50 px-3 py-2">
      <span className="text-sm text-slate-500">{label}</span>
      <span className="text-sm font-medium text-slate-900">{value}</span>
    </div>
  );
}

function AdminScaffold({ title, description, children }: { title: string; description: string; children: ReactNode }) {
  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-sm font-semibold text-slate-900">{title}</h2>
        <p className="mt-0.5 text-sm text-slate-500">{description}</p>
      </div>
      {children}
    </div>
  );
}

function FormField({ label, htmlFor, helper, children }: { label: string; htmlFor?: string; helper?: string; children: ReactNode }) {
  return (
    <div className="space-y-1.5">
      <Label htmlFor={htmlFor}>{label}</Label>
      {helper && <p className="text-xs text-slate-400">{helper}</p>}
      {children}
    </div>
  );
}

function InfoCompact({ icon: Icon, label, value }: { icon: typeof FileText; label: string; value: string }) {
  return (
    <div className="flex items-center gap-2">
      <Icon className="h-4 w-4 text-slate-400" />
      <div>
        <div className="text-xs text-slate-500">{label}</div>
        <div className="text-sm font-medium text-slate-900">{value}</div>
      </div>
    </div>
  );
}
