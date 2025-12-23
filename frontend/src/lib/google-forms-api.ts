// Google Forms API連携用のユーティリティ

interface GoogleFormsConfig {
  serviceAccountEmail: string
  privateKey: string
  formId: string
}

export async function fetchFormResponses(config: GoogleFormsConfig) {
  // Service Accountを使用してGoogle Forms APIにアクセス
  const credentials = {
    client_email: config.serviceAccountEmail,
    private_key: config.privateKey,
  }

  try {
    // Google Forms API v1を使用して回答を取得
    const response = await fetch(`https://forms.googleapis.com/v1/forms/${config.formId}/responses`, {
      headers: {
        Authorization: `Bearer ${await getAccessToken(credentials)}`,
        "Content-Type": "application/json",
      },
    })

    if (!response.ok) {
      throw new Error(`Google Forms API error: ${response.status}`)
    }

    const data = await response.json()
    return data.responses || []
  } catch (error) {
    console.error("[v0] Error fetching form responses:", error)
    throw error
  }
}

async function getAccessToken(credentials: { client_email: string; private_key: string }) {
  // JWT生成とトークン取得のロジック
  // 実際の実装ではgoogle-auth-libraryなどを使用
  console.log("[v0] Getting access token for:", credentials.client_email)

  // 環境変数から認証情報を取得することを推奨
  return "mock_access_token"
}

export function transformGoogleFormResponse(googleResponse: any, formTitle: string) {
  // Google FormsのレスポンスをアプリケーションのFormResponse型に変換
  const responses: Record<string, string> = {}

  if (googleResponse.answers) {
    Object.entries(googleResponse.answers).forEach(([questionId, answer]: [string, any]) => {
      responses[questionId] = answer.textAnswers?.answers?.[0]?.value || ""
    })
  }

  return {
    id: googleResponse.responseId,
    formId: googleResponse.formId,
    formTitle,
    respondentEmail: googleResponse.respondentEmail || "",
    respondentName: responses["name"] || "名無し",
    submittedAt: new Date(googleResponse.createTime),
    status: "new" as const,
    assignedTo: null,
    priority: "medium" as const,
    responses,
  }
}
