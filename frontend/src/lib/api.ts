import type {
  LoginRequest,
  LoginResponse,
  WhoAmIResponse,
  ListFormsResponse,
  ListMembersResponse,
  AddMemberRequest,
  ChangeMemberRoleRequest,
  SyncResponse,
  ListTicketsResponse,
  Ticket,
  UpdateTicketRequest,
  ApiError as ApiErrorType,
} from "@/types";

export class ApiError extends Error {
  status: number;
  error: ApiErrorType;

  constructor(status: number, error: ApiErrorType) {
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
        const errorData: ApiErrorType = await response.json();
        throw new ApiError(response.status, errorData);
      }

      if (response.status === 204) {
        return {} as T;
      }

      return await response.json();
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

  async whoami(): Promise<WhoAmIResponse> {
    return this.request<WhoAmIResponse>("/v1/whoami");
  }

  async getForms(): Promise<ListFormsResponse> {
    return this.request<ListFormsResponse>("/v1/forms");
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

  async syncForm(formId: string): Promise<SyncResponse> {
    return this.request<SyncResponse>(`/v1/forms/${formId}/sync`, {
      method: "POST",
    });
  }

  async getTickets(
    formId?: string,
    status?: string
  ): Promise<ListTicketsResponse> {
    const params = new URLSearchParams();
    if (formId) params.append("form_id", formId);
    if (status) params.append("status", status);

    const query = params.toString();
    const endpoint = `/v1/tickets${query ? `?${query}` : ""}`;

    return this.request<ListTicketsResponse>(endpoint);
  }

  async getTicket(ticketId: string): Promise<Ticket> {
    return this.request<Ticket>(`/v1/tickets/${ticketId}`);
  }

  async updateTicket(
    ticketId: string,
    request: UpdateTicketRequest
  ): Promise<Ticket> {
    return this.request<Ticket>(`/v1/tickets/${ticketId}`, {
      method: "PATCH",
      body: JSON.stringify(request),
    });
  }
}

export const apiClient = new ApiClient();
