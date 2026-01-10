export type Role = "admin" | "editor";

export interface LoginRequest {
  email: string;
  password: string;
}

export interface SignupRequest {
  email: string;
  password: string;
  display_name: string;
}

export interface RegisterFormRequest {
  url: string;
  polling_sec?: number;
}

export interface RegisterFormResponse {
  id: string;
}

export interface FormSummary {
  id: string;
  title: string;
  synced_at?: string | null;
}

export interface ListFormsResponse {
  forms: FormSummary[];
}

export interface Form {
  id: string;
  title: string;
  description?: string | null;
  title_question_id?: string | null;
  email_collection_type?: string | null;
  synced_at?: string | null;
  created_at: string;
}

export interface UpdateFormRequest {
  title_question_id?: string | null;
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

export type ChangeMemberRoleRequest = {
  role: "admin" | "editor";
};

export interface CreateInviteRequest {
  email: string;
  role: "admin" | "editor";
}

export interface CreateInviteResponse {
  invite_id: string;
  expires_at: string;
}

export interface FormInvite {
  id: string;
  email: string;
  role: "admin" | "editor";
  invited_by: string;
  expires_at: string;
  created_at: string;
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
  new_tickets: number;
  last?: string | null;
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

export interface FormStatus {
  id: string;
  name: string;
  color?: string | null;
  display_order: number;
  is_default: boolean;
}

export interface ListFormStatusesResponse {
  statuses: FormStatus[];
}

export interface CreateFormStatusRequest {
  name: string;
  color?: string | null;
  display_order: number;
  is_default?: boolean;
}

export interface UpdateFormStatusRequest {
  name?: string;
  color?: string | null;
  display_order?: number;
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

export type TicketPriority = "high" | "medium" | "low";

export interface TicketStatus {
  id: string;
  name: string;
  color?: string | null;
}

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
  respondent_email?: string | null;
  status: TicketStatus;
  priority: TicketPriority;
  title_question_id?: string | null;
  title: string;
  assignee?: TicketAssignee | null;
  submitted_at: string;
  created_at: string;
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
  status_id?: string;
  assignee_id?: string | null;
  priority?: TicketPriority;
}

export interface TicketHistory {
  id: string;
  ticket_id: string;
  changed_by?: string | null;
  changed_by_name: string;
  field_name: string;
  old_value?: string | null;
  new_value?: string | null;
  created_at: string;
}

export interface ListTicketHistoriesResponse {
  histories: TicketHistory[];
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

export type ChangePasswordRequest = {
  current_password: string;
  new_password: string;
};
