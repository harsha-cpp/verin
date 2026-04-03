import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";

import { DocumentDetailPage, LoginPage } from "@/app/pages";

function renderWithProviders(ui: ReactNode, route = "/") {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[route]}>
        <Routes>
          <Route path="/" element={ui} />
          <Route path="/documents/:documentId" element={ui} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("DMS pages", () => {
  it("renders the login surface", () => {
    renderWithProviders(<LoginPage />);
    expect(screen.getByText("Log in to your Verin account")).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toBeInTheDocument();
  });

  it("renders a document detail view from fallback data", async () => {
    renderWithProviders(<DocumentDetailPage />, "/documents/90000000-0000-0000-0000-000000000001");
    expect(await screen.findByText("Quarterly vendor compliance pack")).toBeInTheDocument();
    expect(screen.getByText("Retention class verified against the latest policy.")).toBeInTheDocument();
  });
});
