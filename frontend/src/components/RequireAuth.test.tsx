import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { RequireAuth } from "./RequireAuth";

const useAuth = vi.fn();

vi.mock("../providers/AuthProvider", () => ({
  useAuth: () => useAuth(),
}));

describe("RequireAuth", () => {
  it("shows loading state", () => {
    useAuth.mockReturnValue({ user: null, loading: true });
    render(
      <MemoryRouter>
        <RequireAuth>
          <div>protected</div>
        </RequireAuth>
      </MemoryRouter>,
    );
    expect(screen.getByText("Carregando...")).toBeInTheDocument();
  });

  it("redirects guest to login", () => {
    useAuth.mockReturnValue({ user: null, loading: false });
    render(
      <MemoryRouter initialEntries={["/"]}>
        <RequireAuth>
          <div>protected</div>
        </RequireAuth>
      </MemoryRouter>,
    );
    expect(screen.queryByText("protected")).not.toBeInTheDocument();
  });

  it("renders children for authenticated user", () => {
    useAuth.mockReturnValue({
      user: { id: 1, name: "U", email: "u@test.dev", active: true, roleId: null },
      loading: false,
    });
    render(
      <MemoryRouter>
        <RequireAuth>
          <div>protected</div>
        </RequireAuth>
      </MemoryRouter>,
    );
    expect(screen.getByText("protected")).toBeInTheDocument();
  });
});
