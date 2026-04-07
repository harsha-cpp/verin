import React from "react";
import ReactDOM from "react-dom/client";
import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "react-router-dom";
import { Toaster } from "sonner";
import "@fontsource-variable/geist";
import "@fontsource-variable/geist-mono";

import { queryClient } from "@/lib/query-client";
import { router } from "@/app/routes";
import "@/styles/global.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <Toaster richColors position="bottom-right" />
    </QueryClientProvider>
  </React.StrictMode>,
);
