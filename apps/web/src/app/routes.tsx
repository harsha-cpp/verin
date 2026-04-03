import { createBrowserRouter, Navigate } from "react-router-dom";

import { AppShell } from "@/components/app-shell";
import {
  AdminQuotasPage,
  AdminRetentionPage,
  AdminRolesPage,
  AdminSettingsPage,
  AdminUsersPage,
  AuditPage,
  CollectionsPage,
  DashboardPage,
  DocumentDetailPage,
  DocumentsPage,
  LoginPage,
  ProfilePage,
  SearchPage,
  SharedWithMePage,
  UploadPage,
} from "@/app/pages";

export const router = createBrowserRouter([
  { path: "/", element: <Navigate to="/login" replace /> },
  { path: "/login", element: <LoginPage /> },
  {
    path: "/",
    element: <AppShell />,
    children: [
      { path: "/dashboard", element: <DashboardPage /> },
      { path: "/documents", element: <DocumentsPage /> },
      { path: "/documents/:documentId", element: <DocumentDetailPage /> },
      { path: "/upload", element: <UploadPage /> },
      { path: "/collections", element: <CollectionsPage /> },
      { path: "/shared", element: <SharedWithMePage /> },
      { path: "/search", element: <SearchPage /> },
      { path: "/profile", element: <ProfilePage /> },
      { path: "/audit", element: <AuditPage /> },
      { path: "/admin/users", element: <AdminUsersPage /> },
      { path: "/admin/roles", element: <AdminRolesPage /> },
      { path: "/admin/quotas", element: <AdminQuotasPage /> },
      { path: "/admin/retention", element: <AdminRetentionPage /> },
      { path: "/admin/settings", element: <AdminSettingsPage /> },
    ],
  },
]);
