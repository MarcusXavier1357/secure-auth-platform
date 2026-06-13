import { describe, expect, it } from "vitest";
import { adminRoutes } from "../router/admin-routes";

describe("admin-routes", () => {
  it("uses granular permissions for admin pages", () => {
    const users = adminRoutes.find((r) => r.id === "users");
    const permissions = adminRoutes.find((r) => r.id === "permissions");
    const audit = adminRoutes.find((r) => r.id === "audit");

    expect(users?.permission).toBe("users.read");
    expect(permissions?.permission).toBe("permissions.read");
    expect(audit?.permission).toBe("audit_logs.read");
  });
});
