import {
  Bell,
  FileSearch2,
  FileText,
  FolderOpen,
  Gauge,
  LayoutDashboard,
  LogOut,
  Search,
  Settings2,
  Share2,
  Shield,
  Upload,
  Users2,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
import { Button, cn } from "@verin/ui";

import { initialsFromName } from "@/design-system/dub-theme";
import { Logo } from "@/components/logo";
import { useNotifications, useSearch, useSession } from "@/hooks/use-api";

type NavEntry = {
  to: string;
  label: string;
  icon: typeof LayoutDashboard;
  match?: (pathname: string) => boolean;
};

const workspaceItems: NavEntry[] = [
  { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  {
    to: "/documents",
    label: "Documents",
    icon: FileText,
    match: (pathname) => pathname.startsWith("/documents"),
  },
  { to: "/upload", label: "Upload", icon: Upload },
  { to: "/collections", label: "Collections", icon: FolderOpen },
  { to: "/shared", label: "Shared with me", icon: Share2 },
  { to: "/search", label: "Search", icon: FileSearch2 },
];

const adminItems: NavEntry[] = [
  { to: "/audit", label: "Audit", icon: Shield },
  { to: "/admin/users", label: "Users", icon: Users2 },
  { to: "/admin/roles", label: "Roles", icon: Shield },
  { to: "/admin/quotas", label: "Quotas", icon: Gauge },
  { to: "/admin/retention", label: "Retention", icon: Shield },
  { to: "/admin/settings", label: "Settings", icon: Settings2 },
];

const pageMeta = [
  { match: (p: string) => p.startsWith("/dashboard"), title: "Dashboard", description: "Operational summary and document activity." },
  { match: (p: string) => p.startsWith("/documents/"), title: "Document detail", description: "Versions, metadata, comments, and audit context." },
  { match: (p: string) => p.startsWith("/documents"), title: "Documents", description: "Browse, filter, and inspect records." },
  { match: (p: string) => p.startsWith("/upload"), title: "Upload", description: "Signed upload flow with metadata and async processing." },
  { match: (p: string) => p.startsWith("/collections"), title: "Collections", description: "Organize documents into folders." },
  { match: (p: string) => p.startsWith("/shared"), title: "Shared with me", description: "Documents others have shared with you." },
  { match: (p: string) => p.startsWith("/search"), title: "Search", description: "Metadata and OCR retrieval." },
  { match: (p: string) => p.startsWith("/profile"), title: "Profile", description: "Your account and preferences." },
  { match: (p: string) => p.startsWith("/audit"), title: "Audit", description: "Sensitive actions, chronological and exportable." },
  { match: (p: string) => p.startsWith("/admin/users"), title: "Users", description: "Accounts, roles, and access state." },
  { match: (p: string) => p.startsWith("/admin/roles"), title: "Roles", description: "Role definitions and permission grouping." },
  { match: (p: string) => p.startsWith("/admin/quotas"), title: "Quotas", description: "Storage ceilings and document limits." },
  { match: (p: string) => p.startsWith("/admin/retention"), title: "Retention", description: "Archive timing and policy windows." },
  { match: (p: string) => p.startsWith("/admin/settings"), title: "Settings", description: "Health, jobs, and workspace configuration." },
];

function isActive(pathname: string, item: NavEntry) {
  if (item.match) return item.match(pathname);
  return pathname === item.to || pathname.startsWith(`${item.to}/`);
}

export function AppShell() {
  const [commandOpen, setCommandOpen] = useState(false);
  const session = useSession();
  const notifications = useNotifications();
  const location = useLocation();
  const navigate = useNavigate();

  const currentMeta = useMemo(
    () => pageMeta.find((entry) => entry.match(location.pathname)) ?? pageMeta[0],
    [location.pathname],
  );

  const notificationCount = notifications.data?.items?.length ?? 0;
  const userName = session.data?.user?.fullName ?? "Asha Rao";
  const userEmail = session.data?.user?.email ?? "admin@verin.local";

  useEffect(() => {
    const listener = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        setCommandOpen((v) => !v);
      }
    };
    window.addEventListener("keydown", listener);
    return () => window.removeEventListener("keydown", listener);
  }, []);

  return (
    <div className="flex min-h-screen bg-gray-50">
      <aside className="hidden w-[240px] shrink-0 flex-col border-r border-slate-200 bg-white lg:flex">
        <div className="px-5 pt-5 pb-4">
          <Logo />
        </div>

        <nav className="flex-1 overflow-y-auto px-3">
          <NavSection label="Workspace" items={workspaceItems} pathname={location.pathname} />
          <NavSection label="Administration" items={adminItems} pathname={location.pathname} />
        </nav>

        <div className="border-t border-slate-200 px-3 py-3">
          <Link to="/profile" className="flex items-center gap-3 rounded-md px-2 py-2 hover:bg-slate-50">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-slate-200 text-xs font-medium text-slate-700">
              {initialsFromName(userName)}
            </div>
            <div className="min-w-0 flex-1">
              <div className="truncate text-sm font-medium text-slate-900">{userName}</div>
              <div className="truncate text-xs text-slate-500">{userEmail}</div>
            </div>
            <button
              type="button"
              onClick={(e) => { e.preventDefault(); navigate("/login"); }}
              className="rounded-md p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
              title="Log out"
            >
              <LogOut className="h-3.5 w-3.5" />
            </button>
          </Link>
        </div>
      </aside>

      <main className="flex min-w-0 flex-1 flex-col">
        <header className="border-b border-slate-200 bg-white px-8 py-5">
          <div className="flex items-center justify-between gap-4">
            <div className="min-w-0">
              <h1 className="text-lg font-semibold text-slate-900">{currentMeta.title}</h1>
              <p className="text-sm text-slate-500">{currentMeta.description}</p>
            </div>

            <div className="flex items-center gap-2.5">
              <button
                type="button"
                className="flex h-10 items-center gap-2 rounded-xl border border-slate-200 bg-white px-3.5 text-sm text-slate-500 shadow-sm hover:bg-slate-50"
                onClick={() => setCommandOpen(true)}
              >
                <Search className="h-4 w-4" />
                <span className="hidden md:inline">Search</span>
                <kbd className="ml-1 hidden rounded border border-slate-200 bg-slate-50 px-1.5 py-0.5 text-[10px] font-medium text-slate-400 md:inline">
                  ⌘K
                </kbd>
              </button>

              <Link to="/upload">
                <Button>
                  <Upload className="h-4 w-4" />
                  Upload
                </Button>
              </Link>

              <button
                type="button"
                className="relative flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-500 shadow-sm hover:bg-slate-50"
              >
                <Bell className="h-4 w-4" />
                {notificationCount > 0 && (
                  <span className="absolute -right-1 -top-1 flex h-4 w-4 items-center justify-center rounded-full bg-slate-900 text-[10px] font-medium text-white">
                    {notificationCount}
                  </span>
                )}
              </button>
            </div>
          </div>
        </header>

        <div className="flex-1 overflow-y-auto px-8 py-6">
          <Outlet />
        </div>
      </main>

      {commandOpen && <CommandPalette onClose={() => setCommandOpen(false)} navigate={navigate} />}
    </div>
  );
}

function CommandPalette({ onClose, navigate }: { onClose: () => void; navigate: (to: string) => void }) {
  const [query, setQuery] = useState("");
  const search = useSearch(query);
  const results = search.data?.items ?? [];

  return (
    <div
      role="dialog"
      className="fixed inset-0 z-50 flex items-start justify-center bg-black/20 px-4 pt-[20vh] backdrop-blur-sm"
      onClick={(e) => { if (e.target === e.currentTarget) onClose(); }}
      onKeyDown={(e) => { if (e.key === "Escape") onClose(); }}
    >
      <div className="w-full max-w-lg rounded-xl border border-slate-200 bg-white shadow-lg">
        <div className="flex items-center gap-2 border-b border-slate-200 px-4 py-3">
          <Search className="h-4 w-4 text-slate-400" />
          <input
            autoFocus
            className="flex-1 bg-transparent text-sm text-slate-900 outline-none placeholder:text-slate-400"
            placeholder="Search documents, pages..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <button
            type="button"
            className="rounded border border-slate-200 px-2 py-0.5 text-xs text-slate-400 hover:bg-slate-50"
            onClick={onClose}
          >
            Esc
          </button>
        </div>

        <div className="max-h-80 overflow-y-auto p-2">
          {query.length > 0 && results.length > 0 && (
            <div className="mb-2">
              <div className="px-2 py-1 text-xs font-medium text-slate-400">Documents</div>
              {results.map((item: Record<string, unknown>) => (
                <button
                  key={String(item.documentId)}
                  type="button"
                  className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-left text-sm text-slate-700 transition-colors hover:bg-slate-100"
                  onClick={() => { navigate(`/documents/${String(item.documentId)}`); onClose(); }}
                >
                  <FileText className="h-4 w-4 text-slate-400" />
                  <div className="min-w-0 flex-1">
                    <div className="truncate font-medium">{String(item.title)}</div>
                    <div className="truncate text-xs text-slate-500">{String(item.originalFilename)}</div>
                  </div>
                </button>
              ))}
            </div>
          )}

          {query.length === 0 && (
            <>
              <CommandGroup title="Workspace" items={workspaceItems} onSelect={onClose} />
              <CommandGroup title="Administration" items={adminItems} onSelect={onClose} />
            </>
          )}

          {query.length > 0 && results.length === 0 && !search.isLoading && (
            <div className="px-3 py-6 text-center text-sm text-slate-400">No results found</div>
          )}
        </div>
      </div>
    </div>
  );
}

function NavSection({
  label,
  items,
  pathname,
}: {
  label: string;
  items: NavEntry[];
  pathname: string;
}) {
  return (
    <div className="mb-5">
      <div className="mb-1 px-3 text-xs font-medium text-slate-400">{label}</div>
      <div className="space-y-0.5">
        {items.map((item) => {
          const Icon = item.icon;
          const active = isActive(pathname, item);
          return (
            <NavLink
              key={item.to}
              to={item.to}
              className={cn(
                "flex items-center gap-3 rounded-md px-3 py-1.5 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-100 hover:text-slate-900",
                active && "bg-slate-100 font-semibold text-slate-900",
              )}
            >
              <Icon className="h-4 w-4" />
              {item.label}
            </NavLink>
          );
        })}
      </div>
    </div>
  );
}

function CommandGroup({
  title,
  items,
  onSelect,
}: {
  title: string;
  items: NavEntry[];
  onSelect: () => void;
}) {
  return (
    <div className="mb-2">
      <div className="px-2 py-1 text-xs font-medium text-slate-400">{title}</div>
      {items.map((item) => {
        const Icon = item.icon;
        return (
          <Link
            key={item.to}
            to={item.to}
            className="flex items-center gap-3 rounded-md px-3 py-2 text-sm text-slate-700 transition-colors hover:bg-slate-100"
            onClick={onSelect}
          >
            <Icon className="h-4 w-4 text-slate-400" />
            {item.label}
          </Link>
        );
      })}
    </div>
  );
}
