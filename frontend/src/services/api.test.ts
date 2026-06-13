import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError, getAccessToken, request, setAccessToken, tryRefresh } from "./api";

describe("api", () => {
  beforeEach(() => {
    setAccessToken(null);
    vi.stubGlobal("fetch", vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("stores access token in memory", () => {
    setAccessToken("tok-123");
    expect(getAccessToken()).toBe("tok-123");
    setAccessToken(null);
    expect(getAccessToken()).toBeNull();
  });

  it("tryRefresh stores token on success", async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      new Response(JSON.stringify({ accessToken: "refreshed" }), { status: 200 }),
    );

    const ok = await tryRefresh();
    expect(ok).toBe(true);
    expect(getAccessToken()).toBe("refreshed");
  });

  it("tryRefresh clears token on failure", async () => {
    setAccessToken("old");
    vi.mocked(fetch).mockResolvedValueOnce(new Response(null, { status: 401 }));

    const ok = await tryRefresh();
    expect(ok).toBe(false);
    expect(getAccessToken()).toBeNull();
  });

  it("request retries once after 401 refresh", async () => {
    setAccessToken("expired");

    vi.mocked(fetch)
      .mockResolvedValueOnce(new Response("unauthorized", { status: 401 }))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ accessToken: "new-token" }), { status: 200 }),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ ok: true }), { status: 200 }));

    const data = await request<{ ok: boolean }>("/me");
    expect(data.ok).toBe(true);
    expect(getAccessToken()).toBe("new-token");
    expect(fetch).toHaveBeenCalledTimes(3);
  });

  it("request throws ApiError when response fails", async () => {
    setAccessToken("valid");
    vi.mocked(fetch).mockResolvedValueOnce(new Response("forbidden", { status: 403 }));

    await expect(request("/users")).rejects.toSatisfy((err: unknown) => {
      return err instanceof ApiError && err.status === 403;
    });
  });
});
