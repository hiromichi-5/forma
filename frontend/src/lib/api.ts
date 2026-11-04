import createClient from "openapi-fetch"
import type {
  AddMemberRequest,
  ChangeMemberRoleRequest,
  ChangePasswordRequest,
  CreateFormStatusRequest,
  CreateInviteRequest,
  Error as ApiErrorBody,
  ErrorCode as ServerErrorCode,
  LoginRequest,
  PasswordResetConfirmRequest,
  PasswordResetRequest,
  RegisterFormRequest,
  ResendVerificationRequest,
  SendNotificationRequest,
  SignupRequest,
  UpdateFormRequest,
  UpdateFormStatusRequest,
  UpdateNotificationSettingsRequest,
  UpdateTicketRequest,
  UpdateUserProfileRequest,
  VerifyEmailRequest,
} from "@/types"
import type { paths } from "@/types/api"

export const API_BASE_URL: string = import.meta.env.VITE_API_URL || "http://localhost:8080"

/**
 * サーバーが返すコードに、クライアント側で補うコードを加えたもの。
 * - NETWORK_ERROR: サーバーに到達できず、レスポンス自体が無い
 * - UNEXPECTED_RESPONSE: プロキシのエラーページなど、API のエラー形式でない応答
 */
export type ErrorCode = ServerErrorCode | "NETWORK_ERROR" | "UNEXPECTED_RESPONSE"

export type ErrorResponse = Omit<ApiErrorBody, "code"> & { code: ErrorCode }

export class ApiError extends Error {
  status: number
  error: ErrorResponse

  constructor(status: number, error: ErrorResponse) {
    super(error.message || `HTTP ${status}`)
    this.name = "ApiError"
    this.status = status
    this.error = error
  }
}

const isErrorBody = (value: unknown): value is ApiErrorBody =>
  typeof value === "object" && value !== null && typeof (value as ApiErrorBody).code === "string"

let sessionExpiredHandler: (() => void) | null = null

const client = createClient<paths>({
  baseUrl: API_BASE_URL,
  credentials: "include",
  // テストで差し替えられるよう、呼び出しのたびに現在の fetch を参照する。
  fetch: (request) => globalThis.fetch(request),
})

type FetchResult<T> = { data?: T; error?: unknown; response: Response }

/**
 * openapi-fetch の結果を、成功時はデータ、失敗時は ApiError の例外に揃える。
 * 呼び出し側は try/catch だけでエラーを扱える。
 */
async function call<T>(request: Promise<FetchResult<T>>): Promise<T> {
  let result: FetchResult<T>
  try {
    result = await request
  } catch {
    throw new ApiError(0, { code: "NETWORK_ERROR", message: "ネットワークエラーが発生しました" })
  }

  const { data, error, response } = result
  if (!response.ok) {
    if (!isErrorBody(error)) {
      throw new ApiError(response.status, { code: "UNEXPECTED_RESPONSE" })
    }
    // ログインの失敗は INVALID_CREDENTIALS なので、セッション失効とは区別できる。
    if (error.code === "INVALID_SESSION") sessionExpiredHandler?.()
    throw new ApiError(response.status, error)
  }
  // 204 などボディの無い成功応答では data が undefined になる。戻り値の型が void の呼び出しでのみ起こる。
  return data as T
}

const formPath = (form_id: string) => ({ params: { path: { form_id } } })
const ticketPath = (ticket_id: string) => ({ params: { path: { ticket_id } } })

export const apiClient = {
  setSessionExpiredHandler(handler: (() => void) | null) {
    sessionExpiredHandler = handler
  },

  // Auth

  async login(body: LoginRequest): Promise<void> {
    await call(client.POST("/v1/auth/login", { body }))
  },
  signup: (body: SignupRequest) => call(client.POST("/v1/auth/signup", { body })),
  async logout(): Promise<void> {
    await call(client.POST("/v1/auth/logout"))
  },
  async verifyEmail(body: VerifyEmailRequest): Promise<void> {
    await call(client.POST("/v1/auth/verify-email", { body }))
  },
  async resendVerification(body: ResendVerificationRequest): Promise<void> {
    await call(client.POST("/v1/auth/verify-email/resend", { body }))
  },
  async passwordReset(body: PasswordResetRequest): Promise<void> {
    await call(client.POST("/v1/auth/password-reset", { body }))
  },
  async passwordResetConfirm(body: PasswordResetConfirmRequest): Promise<void> {
    await call(client.POST("/v1/auth/password-reset/confirm", { body }))
  },

  // User

  getProfile: () => call(client.GET("/v1/me")),
  updateProfile: (body: UpdateUserProfileRequest) => call(client.PATCH("/v1/me", { body })),
  async changePassword(body: ChangePasswordRequest): Promise<void> {
    await call(client.PATCH("/v1/me/password", { body }))
  },
  async deleteProfile(): Promise<void> {
    await call(client.DELETE("/v1/me"))
  },

  // Forms

  getForms: () => call(client.GET("/v1/forms")),
  registerForm: (body: RegisterFormRequest) => call(client.POST("/v1/forms", { body })),
  getForm: (formId: string) => call(client.GET("/v1/forms/{form_id}", formPath(formId))),
  async updateForm(formId: string, body: UpdateFormRequest): Promise<void> {
    await call(client.PATCH("/v1/forms/{form_id}", { ...formPath(formId), body }))
  },
  async deleteForm(formId: string): Promise<void> {
    await call(client.DELETE("/v1/forms/{form_id}", formPath(formId)))
  },
  syncForm: (formId: string) => call(client.POST("/v1/forms/{form_id}/sync", formPath(formId))),

  // Members

  getMembers: (formId: string) => call(client.GET("/v1/forms/{form_id}/members", formPath(formId))),
  async addMember(formId: string, body: AddMemberRequest): Promise<void> {
    await call(client.POST("/v1/forms/{form_id}/members", { ...formPath(formId), body }))
  },
  async changeMemberRole(
    formId: string,
    userId: string,
    body: ChangeMemberRoleRequest,
  ): Promise<void> {
    await call(
      client.PUT("/v1/forms/{form_id}/members/{user_id}", {
        params: { path: { form_id: formId, user_id: userId } },
        body,
      }),
    )
  },
  async removeMember(formId: string, userId: string): Promise<void> {
    await call(
      client.DELETE("/v1/forms/{form_id}/members/{user_id}", {
        params: { path: { form_id: formId, user_id: userId } },
      }),
    )
  },

  // Invites

  listInvites: (formId: string) =>
    call(client.GET("/v1/forms/{form_id}/invites", formPath(formId))),
  createInvite: (formId: string, body: CreateInviteRequest) =>
    call(client.POST("/v1/forms/{form_id}/invites", { ...formPath(formId), body })),
  async revokeInvite(formId: string, inviteId: string): Promise<void> {
    await call(
      client.DELETE("/v1/forms/{form_id}/invites/{invite_id}", {
        params: { path: { form_id: formId, invite_id: inviteId } },
      }),
    )
  },
  async acceptInvite(inviteId: string): Promise<void> {
    await call(
      client.POST("/v1/invites/{invite_id}/accept", {
        params: { path: { invite_id: inviteId } },
      }),
    )
  },

  // Questions / Statuses

  getFormQuestions: (formId: string) =>
    call(client.GET("/v1/forms/{form_id}/questions", formPath(formId))),
  getFormStatuses: (formId: string) =>
    call(client.GET("/v1/forms/{form_id}/statuses", formPath(formId))),
  createFormStatus: (formId: string, body: CreateFormStatusRequest) =>
    call(client.POST("/v1/forms/{form_id}/statuses", { ...formPath(formId), body })),
  updateFormStatus: (formId: string, statusId: string, body: UpdateFormStatusRequest) =>
    call(
      client.PATCH("/v1/forms/{form_id}/statuses/{status_id}", {
        params: { path: { form_id: formId, status_id: statusId } },
        body,
      }),
    ),
  async deleteFormStatus(formId: string, statusId: string): Promise<void> {
    await call(
      client.DELETE("/v1/forms/{form_id}/statuses/{status_id}", {
        params: { path: { form_id: formId, status_id: statusId } },
      }),
    )
  },

  // Tickets

  getTickets: (
    formId: string,
    options: { statusIds?: string[]; query?: string; limit?: number; cursor?: string | null } = {},
  ) =>
    call(
      client.GET("/v1/tickets", {
        params: {
          query: {
            form_id: formId,
            status_id: options.statusIds?.length ? options.statusIds : undefined,
            q: options.query || undefined,
            limit: options.limit,
            cursor: options.cursor ?? undefined,
          },
        },
      }),
    ),
  getTicketCounts: (formId: string, query?: string) =>
    call(
      client.GET("/v1/tickets/counts", {
        params: { query: { form_id: formId, q: query || undefined } },
      }),
    ),
  getTicket: (ticketId: string) =>
    call(client.GET("/v1/tickets/{ticket_id}", ticketPath(ticketId))),
  updateTicket: (ticketId: string, body: UpdateTicketRequest) =>
    call(client.PATCH("/v1/tickets/{ticket_id}", { ...ticketPath(ticketId), body })),
  getTicketHistories: (ticketId: string) =>
    call(client.GET("/v1/tickets/{ticket_id}/histories", ticketPath(ticketId))),

  // Notifications

  getNotificationSettings: (formId: string) =>
    call(client.GET("/v1/forms/{form_id}/notification-settings", formPath(formId))),
  updateNotificationSettings: (formId: string, body: UpdateNotificationSettingsRequest) =>
    call(client.PATCH("/v1/forms/{form_id}/notification-settings", { ...formPath(formId), body })),
  sendTicketNotification: (ticketId: string, body: SendNotificationRequest) =>
    call(client.POST("/v1/tickets/{ticket_id}/notifications", { ...ticketPath(ticketId), body })),
}
