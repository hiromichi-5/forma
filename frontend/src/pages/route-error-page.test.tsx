import { render, screen } from "@testing-library/react"
import { createMemoryRouter, RouterProvider } from "react-router"
import { describe, expect, it, vi } from "vitest"
import RouteErrorPage from "@/pages/route-error-page"

function Broken(): never {
  throw new Error("boom")
}

describe("RouteErrorPage", () => {
  it("画面の描画中に例外が起きても、アプリ全体を落とさずにエラー画面を表示する", () => {
    vi.spyOn(console, "error").mockImplementation(() => undefined)
    vi.spyOn(console, "warn").mockImplementation(() => undefined)
    const router = createMemoryRouter([
      { path: "/", element: <Broken />, errorElement: <RouteErrorPage /> },
    ])

    render(<RouterProvider router={router} />)

    expect(screen.getByText("問題が発生しました")).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "再読み込み" })).toBeInTheDocument()
  })
})
