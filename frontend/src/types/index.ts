export type Role = "admin" | "editor";

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
}

export interface SignupRequest {
  email: string;
  password: string;
  display_name: string;
}

export interface SignupResponse {
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
  display_name: string;
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

export interface FormInvite {
  code: string;
  form_id: string;
  role: "editor";
  expires_at: string;
  created_by: string;
  created_at: string;
  revoked: boolean;
}

export interface ListFormInvitesResponse {
  invites: FormInvite[];
}

export interface AcceptInviteRequest {
  code: string;
}

export interface IssueInviteResponse {
  code: string;
}

export interface SyncResponse {
  synced: number;
  newTickets: number;
  last: string;
}

export interface FormQuestion {
  form_id: string;
  question_id: string;
  title: string;
  question_type: string;
  options?: string[];
}

export interface ListFormQuestionsResponse {
  questions: FormQuestion[];
}

export interface Response {
  response_id: string;
  form_id: string;
  submitted_at: string;
  payload: Record<string, unknown>;
  schema_version: number;
  created_at: string;
}

export interface ListResponsesResponse {
  responses: Response[];
}

export type TicketStatus = "new" | "in_progress" | "done";

export interface TicketAssignee {
  id: string;
  display_name: string;
  email: string;
}

export interface TicketSummary {
  id: string;
  form_id: string;
  form_title: string;
  response_id: string;
  status: TicketStatus;
  priority: number;
  title_question_id?: string | null;
  title: string;
  assignee?: TicketAssignee | null;
  submitted_at: string;
  updated_at: string;
}

export interface ListTicketsResponse {
  tickets: TicketSummary[];
}

export interface TicketAnswer {
  question_id: string;
  question_title: string;
  question_type: string;
  values: string[];
  display_value: string;
}

export interface TicketDetail extends TicketSummary {
  answers: TicketAnswer[];
}

export interface UpdateTicketRequest {
  status?: TicketStatus;
  assignee_id?: string | null;
  priority?: number;
}

export interface ErrorResponse {
  code: string;
  message?: string;
  details?: Record<string, unknown>;
}

export interface User {
  id: string;
  email?: string;
}

export interface UserProfile {
  id: string;
  email: string;
  display_name: string;
}

export interface UpdateUserProfileRequest {
  display_name: string;
}

export interface ToastMessage {
  id: string;
  type: "success" | "error" | "info" | "warning";
  title: string;
  message?: string;
  duration?: number;
}
