import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { RequirePermission } from "./RequirePermission";

const usePermission = vi.fn();

vi.mock("../hooks/usePermission", () => ({
  usePermission: (code: string) => usePermission(code),
}));

function Home() {
  return <div>home</div>;
}

function Secret() {
  return <div>secret</div>;
}

describe("RequirePermission", () => {
  it("redirects when permission is missing", () => {
    usePermission.mockReturnValue(false);
    render(
      <MemoryRouter initialEntries={["/secret"]}>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route element={<RequirePermission code="users.manage" />}>
            <Route path="/secret" element={<Secret />} />
          </Route>
        </Routes>
      </MemoryRouter>,
    );
    expect(screen.getByText("home")).toBeInTheDocument();
    expect(screen.queryByText("secret")).not.toBeInTheDocument();
  });

  it("renders outlet when permission is granted", () => {
    usePermission.mockReturnValue(true);
    render(
      <MemoryRouter initialEntries={["/secret"]}>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route element={<RequirePermission code="users.manage" />}>
            <Route path="/secret" element={<Secret />} />
          </Route>
        </Routes>
      </MemoryRouter>,
    );
    expect(screen.getByText("secret")).toBeInTheDocument();
  });
});
