import {
  FileSearch2,
  FileText,
  FolderOpen,
  Home,
  LogOut,
  Search,
  Share2,
  Upload,
  Users2,
} from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Link, NavLink, Navigate, Outlet, useLocation, useNavigate } from "react-router-dom";
import { Button, cn } from "@verin/ui";

import { DragUploadOverlay } from "@/components/drag-upload-overlay";
import { initialsFromName } from "@/design-system/dub-theme";
import { Logo } from "@/components/logo";
import { useSearch, useSession } from "@/hooks/use-api";

type NavEntry = {
  to: string;
  label: string;
  icon: typeof Home;
  match?: (pathname: string) => boolean;
};

const navItems: NavEntry[] = [
  { to: "/home", label: "Home", icon: Home },
  {
    to: "/documents",
    label: "Documents",
    icon: FileText,
    match: (p) => p.startsWith("/documents") || p.startsWith("/upload"),
  },
  { to: "/spaces", label: "Spaces", icon: FolderOpen },
  { to: "/shared", label: "Shared", icon: Share2 },
  { to: "/search", label: "Search", icon: FileSearch2 },
  { to: "/team", label: "Team", icon: Users2 },
];

function isActive(pathname: string, item: NavEntry) {
  if (item.match) return item.match(pathname);
  return pathname === item.to || pathname.startsWith(`${item.to}/`);
}

export function AppShell() {
  const [commandOpen, setCommandOpen] = useState(false);
  const session = useSession();
  const location = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setCommandOpen((v) => !v);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  if (session.isLoading) return null;

  const sd = session.data as Record<string, any> | undefined;
  if (!sd?.authenticated) {
    return <Navigate to="/login" replace />;
  }

  if (!sd?.hasWorkspace) {
    return <Navigate to="/onboarding" replace />;
  }

  const userName = sd?.user?.fullName ?? "";
  const userEmail = sd?.user?.email ?? "";

  return (
    <div className="flex min-h-screen" style={{ background: "var(--bg)" }}>
      <aside
        className="hidden w-[220px] shrink-0 flex-col lg:flex"
        style={{
          height: "100vh",
          position: "sticky",
          top: 0,
          background: "var(--surface)",
          borderRight: "1px solid var(--line)",
          overflow: "hidden",
        }}
      >
        <div className="px-5 pb-3 pt-5 shrink-0">
          <Logo />
        </div>

        <nav className="min-h-0 flex-1 overflow-y-auto px-3 py-2">
          <div className="space-y-0.5">
            {navItems.map((item) => {
              const Icon = item.icon;
              const active = isActive(location.pathname, item);
              return (
                <NavLink
                  key={item.to}
                  to={item.to}
                  className={cn(
                    "flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm font-medium transition-all duration-150",
                    active
                      ? "font-semibold"
                      : "hover:bg-[var(--surface-soft)]",
                  )}
                  style={
                    active
                      ? { background: "var(--surface-soft)", color: "var(--accent-strong)" }
                      : { color: "var(--ink-soft)" }
                  }
                >
                  <Icon
                    className="h-4 w-4 shrink-0"
                    style={{ color: active ? "var(--accent-strong)" : "var(--ink-muted)" }}
                  />
                  {item.label}
                </NavLink>
              );
            })}
          </div>
        </nav>

        <div
          className="border-t px-3 py-3"
          style={{ borderColor: "var(--line)" }}
        >
          <Link
            to="/home"
            className="flex items-center gap-2.5 rounded-lg px-2 py-2 transition-colors hover:bg-[var(--surface-soft)]"
          >
            <div
              className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold"
              style={{
                background: "var(--surface-soft)",
                color: "var(--ink)",
                border: "1px solid var(--line)",
              }}
            >
              {initialsFromName(userName)}
            </div>
            <div className="min-w-0 flex-1">
              <div
                className="truncate text-sm font-medium"
                style={{ color: "var(--ink)" }}
              >
                {userName}
              </div>
              <div
                className="truncate text-xs"
                style={{ color: "var(--ink-muted)" }}
              >
                {userEmail}
              </div>
            </div>
          </Link>

          <button
            type="button"
            className="mt-1 flex w-full items-center gap-2.5 rounded-lg px-3 py-1.5 text-sm transition-colors hover:bg-[var(--surface-soft)]"
            style={{ color: "var(--ink-soft)" }}
            onClick={() => {
              fetch("/api/v1/auth/logout", { method: "POST", credentials: "include" }).finally(() => {
                window.location.href = "/login";
              });
            }}
          >
            <LogOut className="h-3.5 w-3.5" style={{ color: "var(--ink-muted)" }} />
            Sign out
          </button>
        </div>
      </aside>

      <main className="flex min-w-0 flex-1 flex-col">
        <header
          className="sticky top-0 z-10 flex items-center justify-between gap-4 px-8 py-4"
          style={{
            background: "var(--bg)",
            borderBottom: "1px solid var(--line)",
          }}
        >
          <div />

          <div className="flex items-center gap-2">
            <button
              type="button"
              className="flex h-9 items-center gap-2 rounded-xl px-3 text-sm transition-colors"
              style={{
                background: "var(--surface)",
                border: "1px solid var(--line)",
                color: "var(--ink-muted)",
              }}
              onClick={() => setCommandOpen(true)}
            >
              <Search className="h-3.5 w-3.5" />
              <span className="hidden md:inline">Quick search</span>
              <kbd
                className="ml-1 hidden rounded px-1.5 py-0.5 text-[10px] md:inline"
                style={{
                  background: "var(--surface-soft)",
                  border: "1px solid var(--line)",
                  color: "var(--ink-muted)",
                }}
              >
                ⌘K
              </kbd>
            </button>

            <Link to="/upload">
              <Button className="h-9">
                <Upload className="h-3.5 w-3.5" />
                Upload
              </Button>
            </Link>
          </div>
        </header>

        <div className="flex-1 overflow-y-auto px-8 py-6">
          <Outlet />
        </div>
      </main>

      {commandOpen && (
        <CommandPalette
          onClose={() => setCommandOpen(false)}
          navigate={navigate}
        />
      )}

      <DragUploadOverlay onFileDrop={() => navigate("/upload")} />
    </div>
  );
}

function CommandPalette({
  onClose,
  navigate,
}: {
  onClose: () => void;
  navigate: (to: string) => void;
}) {
  const [query, setQuery] = useState("");
  const search = useSearch(query);
  const results = (search.data as Record<string, any>)?.items ?? [];
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  return (
    <div
      role="dialog"
      aria-modal="true"
      className="fixed inset-0 z-50 flex items-start justify-center px-4 pt-[18vh]"
      style={{ background: "rgba(28,28,30,0.3)", backdropFilter: "blur(4px)" }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      onKeyDown={(e) => {
        if (e.key === "Escape") onClose();
      }}
    >
      <div
        className="animate-slide-up w-full max-w-md overflow-hidden rounded-2xl shadow-2xl"
        style={{
          background: "var(--surface)",
          border: "1px solid var(--line-strong)",
        }}
      >
        <div
          className="flex items-center gap-3 px-4 py-3"
          style={{ borderBottom: "1px solid var(--line)" }}
        >
          <Search className="h-4 w-4 shrink-0" style={{ color: "var(--ink-muted)" }} />
          <input
            ref={inputRef}
            className="flex-1 bg-transparent text-sm outline-none"
            style={{ color: "var(--ink)" }}
            placeholder="Search documents, go to page..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <button
            type="button"
            className="rounded px-1.5 py-0.5 text-xs"
            style={{
              background: "var(--surface-soft)",
              border: "1px solid var(--line)",
              color: "var(--ink-muted)",
            }}
            onClick={onClose}
          >
            Esc
          </button>
        </div>

        <div className="max-h-72 overflow-y-auto p-2">
          {query.length > 0 && results.length > 0 && (
            <div>
              <div
                className="px-2 pb-1 pt-2 text-xs font-medium uppercase tracking-widest"
                style={{ color: "var(--ink-muted)" }}
              >
                Documents
              </div>
              {results.map((item: Record<string, unknown>) => (
                <button
                  key={String(item.documentId)}
                  type="button"
                  className="flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-left text-sm transition-colors hover:bg-[var(--surface-soft)]"
                  style={{ color: "var(--ink)" }}
                  onClick={() => {
                    navigate(`/documents/${String(item.documentId)}`);
                    onClose();
                  }}
                >
                  <FileText className="h-4 w-4 shrink-0" style={{ color: "var(--ink-muted)" }} />
                  <div className="min-w-0 flex-1">
                    <div className="truncate font-medium">{String(item.title)}</div>
                    <div className="truncate text-xs" style={{ color: "var(--ink-muted)" }}>
                      {String(item.originalFilename)}
                    </div>
                  </div>
                </button>
              ))}
            </div>
          )}

          {query.length === 0 && (
            <div>
              <div
                className="px-2 pb-1 pt-2 text-xs font-medium uppercase tracking-widest"
                style={{ color: "var(--ink-muted)" }}
              >
                Go to
              </div>
              {navItems.map((item) => {
                const Icon = item.icon;
                return (
                  <Link
                    key={item.to}
                    to={item.to}
                    className="flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-colors hover:bg-[var(--surface-soft)]"
                    style={{ color: "var(--ink)" }}
                    onClick={onClose}
                  >
                    <Icon className="h-4 w-4 shrink-0" style={{ color: "var(--ink-muted)" }} />
                    {item.label}
                  </Link>
                );
              })}
            </div>
          )}

          {query.length > 0 && results.length === 0 && !search.isLoading && (
            <div className="px-3 py-6 text-center text-sm" style={{ color: "var(--ink-muted)" }}>
              No results for &ldquo;{query}&rdquo;
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
