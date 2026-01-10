import type {
  LoginRequest,
  SignupRequest,
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
  ListResponsesResponse,
  ListTicketsResponse,
  ListFormQuestionsResponse,
  ListFormStatusesResponse,
  FormStatus,
  CreateFormStatusRequest,
  UpdateFormStatusRequest,
  TicketDetail,
  UpdateTicketRequest,
  ListTicketHistoriesResponse,
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
  private baseUrl: string = "http://localhost:8080";

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

      if (response.status === 204) {
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

  async login(request: LoginRequest): Promise<void> {
    await this.request<void>("/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async signup(request: SignupRequest): Promise<void> {
    await this.request<void>("/v1/auth/signup", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async whoami(): Promise<{ user_id: string }> {
    return this.request<{ user_id: string }>("/v1/whoami");
  }

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

  async checkFormHealth(
    formId: string
  ): Promise<{ form_id: string; title: string }> {
    return this.request<{ form_id: string; title: string }>(
      `/v1/forms/${formId}/health`
    );
  }

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

  async listInvites(formId: string): Promise<ListFormInvitesResponse> {
    return this.request<ListFormInvitesResponse>(`/v1/forms/${formId}/invites`);
  }

  async createInvite(formId: string, request: CreateInviteRequest): Promise<CreateInviteResponse> {
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

  async syncForm(formId: string): Promise<SyncResponse> {
    return this.request<SyncResponse>(`/v1/forms/${formId}/sync`, {
      method: "POST",
    });
  }

  async getResponses(
    formId?: string,
    since?: string
  ): Promise<ListResponsesResponse> {
    const params = new URLSearchParams();
    if (formId) params.append("form_id", formId);
    if (since) params.append("since", since);

    const query = params.toString();
    const endpoint = `/v1/responses${query ? `?${query}` : ""}`;

    return this.request<ListResponsesResponse>(endpoint);
  }

  async getFormQuestions(formId: string): Promise<ListFormQuestionsResponse> {
    return this.request<ListFormQuestionsResponse>(
      `/v1/forms/${formId}/questions`
    );
  }

  async getFormStatuses(formId: string): Promise<ListFormStatusesResponse> {
    return this.request<ListFormStatusesResponse>(
      `/v1/forms/${formId}/statuses`
    );
  }

  async createFormStatus(formId: string, request: CreateFormStatusRequest): Promise<FormStatus> {
    return this.request<FormStatus>(`/v1/forms/${formId}/statuses`, {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async updateFormStatus(formId: string, statusId: string, request: UpdateFormStatusRequest): Promise<FormStatus> {
    return this.request<FormStatus>(`/v1/forms/${formId}/statuses/${statusId}`, {
      method: "PATCH",
      body: JSON.stringify(request),
    });
  }

  async deleteFormStatus(formId: string, statusId: string): Promise<void> {
    await this.request<void>(`/v1/forms/${formId}/statuses/${statusId}`, {
      method: "DELETE",
    });
  }

  async setDefaultFormStatus(formId: string, statusId: string): Promise<FormStatus> {
    return this.request<FormStatus>(`/v1/forms/${formId}/statuses/${statusId}/default`, {
      method: "POST",
    });
  }

  async updateFormTitleQuestion(
    formId: string,
    questionId: string | null
  ): Promise<void> {
    await this.request<void>(`/v1/forms/${formId}/title-question`, {
      method: "PATCH",
      body: JSON.stringify({ title_question_id: questionId }),
    });
  }

  async getTickets(
    formId?: string,
    statusId?: string
  ): Promise<ListTicketsResponse> {
    const params = new URLSearchParams();
    if (formId) params.append("form", formId);
    if (statusId) params.append("status_id", statusId);

    const query = params.toString();
    const endpoint = `/v1/tickets${query ? `?${query}` : ""}`;

    return this.request<ListTicketsResponse>(endpoint);
  }

  async getTicket(ticketId: string): Promise<TicketDetail> {
    return this.request<TicketDetail>(`/v1/tickets/${ticketId}`);
  }

  async updateTicket(
    ticketId: string,
    request: UpdateTicketRequest
  ): Promise<TicketDetail> {
    return this.request<TicketDetail>(`/v1/tickets/${ticketId}`, {
      method: "PATCH",
      body: JSON.stringify(request),
    });
  }

  async getTicketHistories(ticketId: string): Promise<ListTicketHistoriesResponse> {
    return this.request<ListTicketHistoriesResponse>(`/v1/tickets/${ticketId}/histories`);
  }

  async healthz(): Promise<string> {
    return this.request<string>("/healthz");
  }

  async logout(): Promise<void> {
    await this.request<void>("/v1/auth/logout", {
      method: "POST",
    });
  }
}

export const apiClient = new ApiClient();
