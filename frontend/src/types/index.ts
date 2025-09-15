export type Role = "admin" | "editor";

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
}

export interface WhoAmIResponse {
  user_id: string;
}

export interface FormSummary {
  form_id: string;
  title: string;
}

export interface ListFormsResponse {
  forms: FormSummary[];
}

export interface Member {
  id: string;
  email: string;
  role: Role;
}

export interface ListMembersResponse {
  members: Member[];
}

export interface AddMemberRequest {
  email: string;
  role: Role;
}

export interface ChangeMemberRoleRequest {
  user_id: string;
  role: Role;
}

export interface SyncResponse {
  synced: number;
  newTickets: number;
  last: string;
}

export interface Response {
  response_id: string;
  form_id: string;
  submitted_at: string;
  payload: Record<string, any>;
  schema_version: number;
  created_at: string;
}

export interface ListResponsesResponse {
  responses: Response[];
}

export type TicketStatus = "new" | "in_progress" | "done";

export interface Ticket {
  id: string;
  form_id: string;
  response_id: string;
  status: TicketStatus;
  assignee_id?: string | null;
  priority: number;
  created_at: string;
  updated_at: string;
}

export interface ListTicketsResponse {
  tickets: Ticket[];
}

export interface UpdateTicketRequest {
  status?: TicketStatus;
  assignee_id?: string | null;
}

export interface ApiError {
  code: string;
  message?: string;
  details?: Record<string, any>;
}

// UI State Types
// API Response Types based on OpenAPI spec
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
}

export interface RegisterFormRequest {
  url: string;
  polling_sec?: number;
}

export interface RegisterFormResponse {
  form_id: string;
}

export interface FormSummary {
  form_id: string;
  title: string;
}

export interface ListFormsResponse {
  forms: FormSummary[];
}

export interface Member {
  id: string;
  email: string;
  role: "admin" | "editor";
}

export interface ListMembersResponse {
  members: Member[];
}

export interface AddMemberRequest {
  email: string;
  role: "admin" | "editor";
}

export interface ChangeMemberRoleRequest {
  user_id: string;
  role: "admin" | "editor";
}

export interface SyncResponse {
  synced: number;
  newTickets: number;
  last: string;
}

export interface Response {
  response_id: string;
  form_id: string;
  submitted_at: string;
  payload: Record<string, any>;
  schema_version: number;
  created_at: string;
}

export interface ListResponsesResponse {
  responses: Response[];
}

export interface Ticket {
  id: string;
  form_id: string;
  response_id: string;
  status: "new" | "in_progress" | "done";
  assignee_id?: string | null;
  priority: number;
  created_at: string;
  updated_at: string;
}

export interface ListTicketsResponse {
  tickets: Ticket[];
}

export interface UpdateTicketRequest {
  status?: "new" | "in_progress" | "done";
  assignee_id?: string | null;
}

export interface ErrorResponse {
  code: string;
  message?: string;
  details?: Record<string, any>;
}

export interface User {
  id: string;
  email?: string;
}

export interface ToastMessage {
  id: string;
  type: "success" | "error" | "info" | "warning";
  title: string;
  message?: string;
  duration?: number;
}
