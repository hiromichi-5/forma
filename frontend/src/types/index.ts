export type Role = "admin" | "editor";

export type ErrorCode =
  | "INVALID_CREDENTIALS"
  | "INVALID_SESSION"
  | "EMAIL_NOT_VERIFIED"
  | "FORBIDDEN"
  | "RESOURCE_HIDDEN"
  | "USER_NOT_FOUND"
  | "FORM_NOT_FOUND"
  | "FORM_NOT_SHARED"
  | "TOKEN_NOT_FOUND"
  | "INVITE_NOT_FOUND"
  | "INVITE_EXPIRED"
  | "ALREADY_MEMBER"
  | "INCORRECT_PASSWORD"
  | "LAST_ADMIN"
  | "CONFLICT"
  | "FORM_ALREADY_REGISTERED"
  | "ACTIVE_INVITE_ALREADY_EXISTS"
  | "STATUS_CONFLICT"
  | "NOTIFICATION_DISABLED"
  | "RESPONDENT_EMAIL_MISSING"
  | "NOTIFICATION_RATE_LIMITED"
  | "VALIDATION_ERROR"
  | "NETWORK_ERROR"
  | "INTERNAL";

export type FieldErrorCode =
  | "REQUIRED"
  | "TOO_SHORT"
  | "INVALID_FORMAT"
  | "INVALID_VALUE";

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

export type NotificationType = "status_change" | "assignee_assigned";

export type NotificationMode = "always" | "confirm" | "off";

export interface TicketNotification {
  notification_type: NotificationType;
  last_sent_at?: string | null;
}

export interface NotificationResult {
  notification_type: NotificationType;
  result: "sent" | "failed";
}

export interface TicketDetail extends TicketSummary {
  answers: TicketAnswer[];
  notifications: TicketNotification[];
}

export interface TicketUpdateResponse extends TicketDetail {
  notification_results: NotificationResult[];
}

export interface UpdateTicketRequest {
  status_id?: string;
  assignee_id?: string | null;
  priority?: TicketPriority;
}

export interface NotificationSetting {
  notification_type: NotificationType;
  mode: NotificationMode;
  include_detail: boolean;
}

export interface NotificationSettingsResponse {
  email_collection_type?: string | null;
  settings: NotificationSetting[];
}

export interface UpdateNotificationSettingsRequest {
  settings: NotificationSetting[];
}

export interface SendNotificationRequest {
  notification_type: NotificationType;
}

export interface SentNotificationResponse {
  notification_type: NotificationType;
  sent_at: string;
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
  code: FieldErrorCode;
}

export interface ErrorResponse {
  code: ErrorCode;
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
