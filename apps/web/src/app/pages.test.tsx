import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";

import { LoginPage } from "@/app/pages";

function wrap(ui: ReactNode) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={["/"]}>
        <Routes>
          <Route path="/" element={ui} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("Verin pages", () => {
  it("landing page shows Google sign-in", () => {
    wrap(<LoginPage />);
    expect(screen.getByText(/sign in with google/i)).toBeInTheDocument();
  });

  it("landing page shows product value proposition", () => {
    wrap(<LoginPage />);
    expect(screen.getByText(/shared document workspace/i)).toBeInTheDocument();
  });

  it("landing page shows how-it-works features", () => {
    wrap(<LoginPage />);
    expect(screen.getByText(/fast first-use loop/i)).toBeInTheDocument();
  });
});
