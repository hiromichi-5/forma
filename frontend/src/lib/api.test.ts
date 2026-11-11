import { beforeEach, describe, expect, it, vi } from "vitest"
import { ApiError, apiClient } from "@/lib/api"

const fetchMock = vi.fn<(request: Request) => Promise<Response>>()

const jsonResponse = (status: number, body: unknown) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  })

beforeEach(() => {
  vi.stubGlobal("fetch", fetchMock)
  fetchMock.mockReset()
})

describe("apiClient", () => {
  it("Cookie を送るため credentials: include でリクエストする", async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { forms: [] }))

    await apiClient.getForms()

    const [request] = fetchMock.mock.calls[0]
    expect(request.url).toMatch(/\/v1\/forms$/)
    expect(request.credentials).toBe("include")
  })

  it("チケット一覧の絞り込み条件をクエリ文字列にする", async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { tickets: [], next_cursor: null }))

    await apiClient.getTickets("form-1", {
      statusIds: ["s1", "s2"],
      query: "foo",
      limit: 50,
      cursor: "c1",
    })

    const url = new URL(fetchMock.mock.calls[0][0].url)
    expect(url.searchParams.get("form_id")).toBe("form-1")
    expect(url.searchParams.getAll("status_id")).toEqual(["s1", "s2"])
    expect(url.searchParams.get("q")).toBe("foo")
    expect(url.searchParams.get("cursor")).toBe("c1")
  })

  it("エラー応答はコード付きの ApiError にする", async () => {
    fetchMock.mockResolvedValue(jsonResponse(403, { code: "FORBIDDEN", message: "forbidden" }))

    const error = await apiClient.getForm("form-1").catch((e: unknown) => e)

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).status).toBe(403)
    expect((error as ApiError).error.code).toBe("FORBIDDEN")
  })

  it("INVALID_SESSION を受け取るとセッション失効のハンドラを呼ぶ", async () => {
    const handler = vi.fn()
    apiClient.setSessionExpiredHandler(handler)
    fetchMock.mockResolvedValue(jsonResponse(401, { code: "INVALID_SESSION", message: "x" }))

    await apiClient.getProfile().catch(() => undefined)

    expect(handler).toHaveBeenCalledOnce()
    apiClient.setSessionExpiredHandler(null)
  })

  it("ログイン失敗（INVALID_CREDENTIALS）ではセッション失効のハンドラを呼ばない", async () => {
    const handler = vi.fn()
    apiClient.setSessionExpiredHandler(handler)
    fetchMock.mockResolvedValue(jsonResponse(401, { code: "INVALID_CREDENTIALS", message: "x" }))

    await apiClient.login({ email: "a@example.com", password: "x" }).catch(() => undefined)

    expect(handler).not.toHaveBeenCalled()
    apiClient.setSessionExpiredHandler(null)
  })

  it("API のエラー形式でない応答（プロキシのエラーページなど）は UNEXPECTED_RESPONSE にする", async () => {
    fetchMock.mockResolvedValue(
      new Response("<html>502 Bad Gateway</html>", {
        status: 502,
        headers: { "Content-Type": "text/html" },
      }),
    )

    const error = await apiClient.getForms().catch((e: unknown) => e)

    expect((error as ApiError).status).toBe(502)
    expect((error as ApiError).error.code).toBe("UNEXPECTED_RESPONSE")
  })

  it("本文の無い成功応答（204）を扱える", async () => {
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }))

    await expect(apiClient.deleteForm("form-1")).resolves.toBeUndefined()
  })

  it("通信自体に失敗した場合は NETWORK_ERROR にする", async () => {
    fetchMock.mockRejectedValue(new TypeError("Failed to fetch"))

    const error = await apiClient.getForms().catch((e: unknown) => e)

    expect((error as ApiError).status).toBe(0)
    expect((error as ApiError).error.code).toBe("NETWORK_ERROR")
  })
})
