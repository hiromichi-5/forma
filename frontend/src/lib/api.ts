import type {
  LoginRequest,
  LoginResponse,
  SignupRequest,
  SignupResponse,
  RegisterFormRequest,
  RegisterFormResponse,
  ListFormsResponse,
  ListMembersResponse,
  AddMemberRequest,
  ChangeMemberRoleRequest,
  ListFormInvitesResponse,
  AcceptInviteRequest,
  IssueInviteResponse,
  SyncResponse,
  ListResponsesResponse,
  ListTicketsResponse,
  ListFormQuestionsResponse,
  TicketDetail,
  UpdateTicketRequest,
  UserProfile,
  UpdateUserProfileRequest,
  ErrorResponse,
  TicketStatus,
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
  private token: string | null = null;

  constructor() {
    this.token = localStorage.getItem("forma_token");
  }

  setToken(token: string) {
    this.token = token;
    localStorage.setItem("forma_token", token);
  }

  clearToken() {
    this.token = null;
    localStorage.removeItem("forma_token");
  }

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

    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }

    const config: RequestInit = {
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

  async login(request: LoginRequest): Promise<LoginResponse> {
    return this.request<LoginResponse>("/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(request),
    });
  }

  async signup(request: SignupRequest): Promise<SignupResponse> {
    return this.request<SignupResponse>("/v1/auth/signup", {
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
      method: "PUT",
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
    request: ChangeMemberRoleRequest
  ): Promise<void> {
    return this.request<void>(`/v1/forms/${formId}/members`, {
      method: "PATCH",
      body: JSON.stringify(request),
    });
  }

  async removeMember(formId: string, userId: string): Promise<void> {
    return this.request<void>(`/v1/forms/${formId}/members?user_id=${userId}`, {
      method: "DELETE",
    });
  }

  async listInvites(formId: string): Promise<ListFormInvitesResponse> {
    return this.request<ListFormInvitesResponse>(
      `/v1/forms/${formId}/invites`
    );
  }

  async createInvite(formId: string): Promise<IssueInviteResponse> {
    return this.request<IssueInviteResponse>(`/v1/forms/${formId}/invites`, {
      method: "POST",
    });
  }

  async revokeInvite(formId: string, code: string): Promise<void> {
    return this.request<void>(`/v1/forms/${formId}/invites/${code}`, {
      method: "DELETE",
    });
  }

  async acceptInvite(request: AcceptInviteRequest): Promise<void> {
    return this.request<void>(`/v1/invites/accept`, {
      method: "POST",
      body: JSON.stringify(request),
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
    status?: TicketStatus
  ): Promise<ListTicketsResponse> {
    const params = new URLSearchParams();
    if (formId) params.append("form_id", formId);
    if (status) params.append("status", status);

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

  async healthz(): Promise<string> {
    return this.request<string>("/healthz");
  }

  logout() {
    this.clearToken();
  }
}

export const apiClient = new ApiClient();
