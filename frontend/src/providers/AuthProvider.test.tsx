import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AuthProvider, useAuth } from "./AuthProvider";

const tryRefresh = vi.fn();
const apiMe = vi.fn();
const apiLogin = vi.fn();
const apiLogout = vi.fn();
const setAccessToken = vi.fn();

vi.mock("../services/api", () => ({
  api: {
    me: (...args: unknown[]) => apiMe(...args),
    login: (...args: unknown[]) => apiLogin(...args),
    logout: (...args: unknown[]) => apiLogout(...args),
  },
  tryRefresh: (...args: unknown[]) => tryRefresh(...args),
  setAccessToken: (...args: unknown[]) => setAccessToken(...args),
}));

function Probe() {
  const { user, loading, hasPermission } = useAuth();
  if (loading) return <div>loading</div>;
  if (!user) return <div>guest</div>;
  return (
    <div>
      <span>{user.email}</span>
      <span>{hasPermission("users.read") ? "yes" : "no"}</span>
    </div>
  );
}

describe("AuthProvider", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    tryRefresh.mockResolvedValue(false);
  });

  it("ends loading as guest when refresh fails", async () => {
    render(
      <AuthProvider>
        <Probe />
      </AuthProvider>,
    );

    await waitFor(() => expect(screen.getByText("guest")).toBeInTheDocument());
  });

  it("loads user when refresh and me succeed", async () => {
    tryRefresh.mockResolvedValue(true);
    apiMe.mockResolvedValue({
      user: { id: 1, name: "Admin", email: "admin@test.dev", active: true, roleId: null },
      permissions: ["users.read"],
    });

    render(
      <AuthProvider>
        <Probe />
      </AuthProvider>,
    );

    await waitFor(() => expect(screen.getByText("admin@test.dev")).toBeInTheDocument());
    expect(screen.getByText("yes")).toBeInTheDocument();
  });
});
