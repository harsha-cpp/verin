import { createBrowserRouter, Navigate } from "react-router-dom";

import { AppShell } from "@/components/app-shell";
import {
  DocumentDetailPage,
  DocumentsPage,
  HomePage,
  LoginPage,
  OnboardingPage,
  SearchPage,
  SharedWithMePage,
  SpacesPage,
  TeamPage,
  UploadPage,
} from "@/app/pages";

export const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  { path: "/onboarding", element: <OnboardingPage /> },
  {
    path: "/",
    element: <AppShell />,
    children: [
      { index: true, element: <Navigate to="/home" replace /> },
      { path: "home", element: <HomePage /> },
      { path: "documents", element: <DocumentsPage /> },
      { path: "documents/:documentId", element: <DocumentDetailPage /> },
      { path: "upload", element: <UploadPage /> },
      { path: "spaces", element: <SpacesPage /> },
      { path: "shared", element: <SharedWithMePage /> },
      { path: "search", element: <SearchPage /> },
      { path: "team", element: <TeamPage /> },
    ],
  },
]);
