import type {
  LoginRequest,
  SignupRequest,
  VerifyEmailRequest,
  ResendVerificationRequest,
  PasswordResetRequest,
  PasswordResetConfirmRequest,
  RegisterFormRequest,
  RegisterFormResponse,
  ListFormsResponse,
  Form,
  UpdateFormRequest,
  ListMembersResponse,
  AddMemberRequest,
  ChangeMemberRoleRequest,
  ListFormInvitesResponse,
  CreateInviteRequest,
  CreateInviteResponse,
  SyncResponse,
  ListTicketsResponse,
  ListFormQuestionsResponse,
  ListFormStatusesResponse,
  FormStatus,
  CreateFormStatusRequest,
  UpdateFormStatusRequest,
  TicketDetail,
  TicketUpdateResponse,
  UpdateTicketRequest,
  ListTicketHistoriesResponse,
  NotificationSettingsResponse,
  UpdateNotificationSettingsRequest,
  SendNotificationRequest,
  SentNotificationResponse,
  UserProfile,
  UpdateUserProfileRequest,
  ErrorResponse,
  ChangePasswordRequest,
} from "../types";

export class ApiError extends Error {
  status: number;
  error: ErrorResponse;

  constructor(status: number, error: ErrorResponse) {
    super(error.message || `HTTP ${status}`);
    this.name = "ApiError";
    this.status = status;
    this.error = error;
  }

  get isUnauthorized() {
    return this.status === 401;
  }

  get isForbidden() {
    return this.status === 403;
  }

  get isValidationError() {
    return this.status === 400;
  }
}

class ApiClient {
  private baseUrl: string = import.meta.env.VITE_API_URL || "http://localhost:8080";

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };

    if (options.headers) {
      Object.assign(headers, options.headers);
    }

    const config: RequestInit = {
      credentials: "include",
      ...options,
      headers,
    };

    try {
      const response = await fetch(url, config);

      if (!response.ok) {
        const errorData: ErrorResponse = await response.json();
        throw new ApiError(response.status, errorData);
      }

      if (response.status === 204 || response.status === 202) {
        return {} as T;
      }

      const contentType = response.headers.get("content-type");
      if (contentType && contentType.includes("application/json")) {
        return await response.json();
      }

      return response.text() as unknown as T;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }

      throw new ApiError(0, {
        code: "NETWORK_ERROR",
        message: "ネットワークエラーが発生しました",
      });
    }
  }

  // Auth

  async login(request: LoginRequest): Promise<void> {
    await this.request<void>("/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async signup(request: SignupRequest): Promise<{ id: string }> {
    return this.request<{ id: string }>("/v1/auth/signup", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async logout(): Promise<void> {
    await this.request<void>("/v1/auth/logout", {
      method: "POST",
    });
  }

  async verifyEmail(request: VerifyEmailRequest): Promise<void> {
    await this.request<void>("/v1/auth/verify-email", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async resendVerification(
    request: ResendVerificationRequest
  ): Promise<void> {
    await this.request<void>("/v1/auth/verify-email/resend", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async passwordReset(request: PasswordResetRequest): Promise<void> {
    await this.request<void>("/v1/auth/password-reset", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async passwordResetConfirm(
    request: PasswordResetConfirmRequest
  ): Promise<void> {
    await this.request<void>("/v1/auth/password-reset/confirm", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  // User

  async getProfile(): Promise<UserProfile> {
    return this.request<UserProfile>("/v1/me");
  }

  async updateProfile(request: UpdateUserProfileRequest): Promise<UserProfile> {
    return this.request<UserProfile>("/v1/me", {
      method: "PATCH",
      body: JSON.stringify(request),
    });
  }

  async changePassword(request: ChangePasswordRequest): Promise<void> {
    await this.request<void>("/v1/me/password", {
      method: "PATCH",
      body: JSON.stringify(request),
    });
  }

  async deleteProfile(): Promise<void> {
    return this.request<void>("/v1/me", {
      method: "DELETE",
    });
  }

  // Forms

  async getForms(): Promise<ListFormsResponse> {
    return this.request<ListFormsResponse>("/v1/forms");
  }

  async registerForm(
    request: RegisterFormRequest
  ): Promise<RegisterFormResponse> {
    return this.request<RegisterFormResponse>("/v1/forms", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async getForm(formId: string): Promise<Form> {
    return this.request<Form>(`/v1/forms/${formId}`);
  }

  async updateForm(formId: string, request: UpdateFormRequest): Promise<void> {
    await this.request<void>(`/v1/forms/${formId}`, {
      method: "PATCH",
      body: JSON.stringify(request),
    });
  }

  async deleteForm(formId: string): Promise<void> {
    await this.request<void>(`/v1/forms/${formId}`, {
      method: "DELETE",
    });
  }

  // Members

  async getMembers(formId: string): Promise<ListMembersResponse> {
    return this.request<ListMembersResponse>(`/v1/forms/${formId}/members`);
  }

  async addMember(formId: string, request: AddMemberRequest): Promise<void> {
    return this.request<void>(`/v1/forms/${formId}/members`, {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async changeMemberRole(
    formId: string,
    userId: string,
    request: ChangeMemberRoleRequest
  ): Promise<void> {
    return this.request<void>(`/v1/forms/${formId}/members/${userId}`, {
      method: "PUT",
      body: JSON.stringify(request),
    });
  }

  async removeMember(formId: string, userId: string): Promise<void> {
    return this.request<void>(`/v1/forms/${formId}/members/${userId}`, {
      method: "DELETE",
    });
  }

  // Invites

  async listInvites(formId: string): Promise<ListFormInvitesResponse> {
    return this.request<ListFormInvitesResponse>(`/v1/forms/${formId}/invites`);
  }

  async createInvite(
    formId: string,
    request: CreateInviteRequest
  ): Promise<CreateInviteResponse> {
    return this.request<CreateInviteResponse>(`/v1/forms/${formId}/invites`, {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async revokeInvite(formId: string, inviteId: string): Promise<void> {
    return this.request<void>(`/v1/forms/${formId}/invites/${inviteId}`, {
      method: "DELETE",
    });
  }

  async acceptInvite(inviteId: string): Promise<void> {
    return this.request<void>(`/v1/invites/${inviteId}/accept`, {
      method: "POST",
    });
  }

  // Sync

  async syncForm(formId: string): Promise<SyncResponse> {
    return this.request<SyncResponse>(`/v1/forms/${formId}/sync`, {
      method: "POST",
    });
  }

  // Questions

  async getFormQuestions(formId: string): Promise<ListFormQuestionsResponse> {
    return this.request<ListFormQuestionsResponse>(
      `/v1/forms/${formId}/questions`
    );
  }

  // Statuses

  async getFormStatuses(formId: string): Promise<ListFormStatusesResponse> {
    return this.request<ListFormStatusesResponse>(
      `/v1/forms/${formId}/statuses`
    );
  }

  async createFormStatus(
    formId: string,
    request: CreateFormStatusRequest
  ): Promise<FormStatus> {
    return this.request<FormStatus>(`/v1/forms/${formId}/statuses`, {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async updateFormStatus(
    formId: string,
    statusId: string,
    request: UpdateFormStatusRequest
  ): Promise<FormStatus> {
    return this.request<FormStatus>(
      `/v1/forms/${formId}/statuses/${statusId}`,
      {
        method: "PATCH",
        body: JSON.stringify(request),
      }
    );
  }

  async deleteFormStatus(formId: string, statusId: string): Promise<void> {
    await this.request<void>(`/v1/forms/${formId}/statuses/${statusId}`, {
      method: "DELETE",
    });
  }

  // Tickets

  async getTickets(
    formId: string,
    statusId?: string
  ): Promise<ListTicketsResponse> {
    const params = new URLSearchParams();
    params.append("form_id", formId);
    if (statusId) params.append("status_id", statusId);

    return this.request<ListTicketsResponse>(`/v1/tickets?${params.toString()}`);
  }

  async getTicket(ticketId: string): Promise<TicketDetail> {
    return this.request<TicketDetail>(`/v1/tickets/${ticketId}`);
  }

  async updateTicket(
    ticketId: string,
    request: UpdateTicketRequest
  ): Promise<TicketUpdateResponse> {
    return this.request<TicketUpdateResponse>(`/v1/tickets/${ticketId}`, {
      method: "PATCH",
      body: JSON.stringify(request),
    });
  }

  async getTicketHistories(
    ticketId: string
  ): Promise<ListTicketHistoriesResponse> {
    return this.request<ListTicketHistoriesResponse>(
      `/v1/tickets/${ticketId}/histories`
    );
  }

  // Notifications

  async getNotificationSettings(
    formId: string
  ): Promise<NotificationSettingsResponse> {
    return this.request<NotificationSettingsResponse>(
      `/v1/forms/${formId}/notification-settings`
    );
  }

  async updateNotificationSettings(
    formId: string,
    request: UpdateNotificationSettingsRequest
  ): Promise<NotificationSettingsResponse> {
    return this.request<NotificationSettingsResponse>(
      `/v1/forms/${formId}/notification-settings`,
      {
        method: "PATCH",
        body: JSON.stringify(request),
      }
    );
  }

  async sendTicketNotification(
    ticketId: string,
    request: SendNotificationRequest
  ): Promise<SentNotificationResponse> {
    return this.request<SentNotificationResponse>(
      `/v1/tickets/${ticketId}/notifications`,
      {
        method: "POST",
        body: JSON.stringify(request),
      }
    );
  }

  // Health

  async health(): Promise<string> {
    return this.request<string>("/health");
  }
}

export const apiClient = new ApiClient();
