import type { NotificationType } from "@/types"

/**
 * 通知メールの件名と HTML 本文。
 * `backend/internal/infra/resend/templates/<name>/{subject.txt,body.html}` の写しである。
 *
 *
 * バックエンドのテンプレートを変更したときは、こちらもあわせて更新すること。
 */
const TEMPLATES: Record<string, { subject: string; html: string }> = {
  "ticket-status-changed": {
    subject: "【{{form_title}}】フォームの回答の対応状況が更新されました",
    html: `<!DOCTYPE html>
<html>

<head>
  <meta charset="UTF-8">
</head>

<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px">
  <h2>フォームの回答の対応状況が更新されました</h2>
  <p><strong>{{form_title}}</strong> に提出した回答について、対応状況が更新されました。</p>
  <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0">
  <p style="color: #9ca3af; font-size: 12px">
    このメールは送信専用です。ご返信いただいてもお答えできませんのでご了承ください。
  </p>
</body>

</html>`,
  },
  "ticket-status-changed-detailed": {
    subject: "【{{form_title}}】フォームの回答の対応状況が「{{status_name}}」になりました",
    html: `<!DOCTYPE html>
<html>

<head>
  <meta charset="UTF-8">
</head>

<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px">
  <h2>フォームの回答の対応状況が更新されました</h2>
  <p><strong>{{form_title}}</strong> に提出した回答について、対応状況が更新されました。</p>
  <p style="margin: 24px 0">
    現在の対応状況:
    <strong style="background: #f3f4f6; padding: 6px 12px; border-radius: 6px">{{status_name}}</strong>
  </p>
  <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0">
  <p style="color: #9ca3af; font-size: 12px">
    このメールは送信専用です。ご返信いただいてもお答えできませんのでご了承ください。
  </p>
</body>

</html>`,
  },
  "ticket-assigned": {
    subject: "【{{form_title}}】フォームの回答の担当者が決まりました",
    html: `<!DOCTYPE html>
<html>

<head>
  <meta charset="UTF-8">
</head>

<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px">
  <h2>フォームの回答の担当者が決まりました</h2>
  <p><strong>{{form_title}}</strong> に提出した回答について、担当者が割り当てられました。</p>
  <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0">
  <p style="color: #9ca3af; font-size: 12px">
    このメールは送信専用です。ご返信いただいてもお答えできませんのでご了承ください。
  </p>
</body>

</html>`,
  },
  "ticket-assigned-detailed": {
    subject: "【{{form_title}}】フォームの回答の担当者が決まりました",
    html: `<!DOCTYPE html>
<html>

<head>
  <meta charset="UTF-8">
</head>

<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 24px">
  <h2>フォームの回答の担当者が決まりました</h2>
  <p><strong>{{form_title}}</strong> に提出した回答について、担当者が割り当てられました。</p>
  <p style="margin: 24px 0">
    担当者:
    <strong style="background: #f3f4f6; padding: 6px 12px; border-radius: 6px">{{assignee_name}}</strong>
  </p>
  <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0">
  <p style="color: #9ca3af; font-size: 12px">
    このメールは送信専用です。ご返信いただいてもお答えできませんのでご了承ください。
  </p>
</body>

</html>`,
  },
}

/** フォームにステータスが1件もない場合にプレビューで使う名前。 */
export const FALLBACK_STATUS_NAME = "対応中"

/** 閲覧者の表示名が取れない場合にプレビューで使う名前。 */
export const FALLBACK_ASSIGNEE_NAME = "担当者"

/** プレビューに差し込むサンプル値。実際の送信では各チケットの値に置き換わる。 */
export type NotificationEmailSample = {
  formTitle: string
  statusName: string
  assigneeName: string
}

export type NotificationEmailPreview = {
  subject: string
  html: string
}

/** バックエンドの `repository.TemplateTicket*` と同じ命名規則。 */
function templateName(notificationType: NotificationType, includeDetail: boolean): string {
  const base = notificationType === "status_change" ? "ticket-status-changed" : "ticket-assigned"
  return includeDetail ? `${base}-detailed` : base
}

function replaceVars(template: string, vars: Record<string, string>): string {
  return Object.entries(vars).reduce(
    (acc, [key, value]) => acc.replaceAll(`{{${key}}}`, value),
    template
  )
}

function escapeHtml(value: string): string {
  return value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
}


export function renderNotificationEmailPreview(
  notificationType: NotificationType,
  includeDetail: boolean,
  sample: NotificationEmailSample
): NotificationEmailPreview {
  const template = TEMPLATES[templateName(notificationType, includeDetail)]
  const vars = {
    form_title: sample.formTitle,
    status_name: sample.statusName,
    assignee_name: sample.assigneeName,
  }
  const htmlVars = Object.fromEntries(
    Object.entries(vars).map(([key, value]) => [key, escapeHtml(value)])
  )

  return {
    subject: replaceVars(template.subject, vars),
    html: replaceVars(template.html, htmlVars),
  }
}
