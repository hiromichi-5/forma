import { readFileSync } from "node:fs"
import { resolve } from "node:path"
import { describe, expect, it } from "vitest"
import { renderNotificationEmailPreview } from "@/lib/notification-email-preview"
import type { NotificationType } from "@/types"

const TEMPLATE_DIR = resolve(
  import.meta.dirname,
  "../../../backend/internal/infra/resend/templates",
)

const readTemplate = (name: string) => ({
  subject: readFileSync(resolve(TEMPLATE_DIR, name, "subject.txt"), "utf8").trim(),
  html: readFileSync(resolve(TEMPLATE_DIR, name, "body.html"), "utf8").trim(),
})

// プレースホルダをそのまま差し込むと、置換前のテンプレートが得られる。
const placeholders = {
  formTitle: "{{form_title}}",
  statusName: "{{status_name}}",
  assigneeName: "{{assignee_name}}",
}

const cases: [NotificationType, boolean, string][] = [
  ["status_change", false, "ticket-status-changed"],
  ["status_change", true, "ticket-status-changed-detailed"],
  ["assignee_assigned", false, "ticket-assigned"],
  ["assignee_assigned", true, "ticket-assigned-detailed"],
]

describe("renderNotificationEmailPreview", () => {
  // プレビュー用のテンプレートはバックエンドの写しなので、片方だけ変更されると表示と実際の送信内容が食い違う。
  it.each(cases)("%s (detail=%s) はバックエンドの %s と一致する", (type, detail, name) => {
    const preview = renderNotificationEmailPreview(type, detail, placeholders)
    const backend = readTemplate(name)

    expect(preview.subject.trim()).toBe(backend.subject)
    expect(preview.html.trim()).toBe(backend.html)
  })

  it("HTML 本文に差し込む値はエスケープする", () => {
    const preview = renderNotificationEmailPreview("status_change", true, {
      formTitle: "<b>A&B</b>",
      statusName: "完了",
      assigneeName: "担当",
    })

    expect(preview.html).toContain("&lt;b&gt;A&amp;B&lt;/b&gt;")
    expect(preview.subject).toContain("<b>A&B</b>")
  })
})
