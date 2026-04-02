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

export interface VerifyEmailRequest {
  token: string;
}

export interface ResendVerificationRequest {
  email: string;
}

export interface PasswordResetRequest {
  email: string;
}

export interface PasswordResetConfirmRequest {
  token: string;
  new_password: string;
}

export interface RegisterFormRequest {
  url: string;
}

export interface RegisterFormResponse {
  id: string;
}

export interface FormSummary {
  id: string;
  form_id: string;
  title: string;
}

export interface ListFormsResponse {
  forms: FormSummary[];
}

export interface Form {
  id: string;
  form_id: string;
  title: string;
  description?: string | null;
  title_question_id?: string | null;
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

export interface SyncResponse {
  synced: boolean;
  new_tickets: number;
  last?: string;
}

export interface FormQuestion {
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
  form_id: string;
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
  is_default?: boolean;
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

export interface ErrorField {
  field: string;
  code: string;
}

export interface ErrorResponse {
  code: string;
  message?: string;
  fields?: ErrorField[];
}

export interface User {
  id: string;
  email?: string;
}

export interface UserProfile {
  id: string;
  email: string;
  display_name: string;
  verified_at?: string | null;
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
