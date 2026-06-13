import { describe, expect, it } from "vitest";
import { matchPermission } from "./permission";

describe("matchPermission", () => {
  const granted = ["users.*", "audit_logs.read"];

  it("matches exact permission", () => {
    expect(matchPermission(granted, "audit_logs.read")).toBe(true);
  });

  it("matches module wildcard", () => {
    expect(matchPermission(granted, "users.manage")).toBe(true);
    expect(matchPermission(granted, "users.read")).toBe(true);
  });

  it("denies unrelated permission", () => {
    expect(matchPermission(granted, "permissions.manage")).toBe(false);
  });

  it("matches global wildcard", () => {
    expect(matchPermission(["*"], "anything.here")).toBe(true);
  });

  it("does not treat legacy manage codes as module wildcards", () => {
    expect(matchPermission(["users.manage"], "users.read")).toBe(false);
    expect(matchPermission(["permissions.manage"], "permissions.grant")).toBe(false);
  });
});
