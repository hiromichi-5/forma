import type { FormQuestion, FormStatus, Member, TicketPriority } from "@/types"

type StatusKey = "open" | "progress" | "waiting" | "done"

/** 担当者（DEMO_MEMBERS の添字）か回答者が送ったメッセージ。 */
export type DemoMessageSeed = {
  minutesAgo: number
  by: number | "respondent"
  text: string
}

/** 担当者による変更。値は本番の変更履歴と同じく、ステータス名・担当者名・優先度の値で持つ。 */
export type DemoChangeSeed = {
  minutesAgo: number
  by: number
  field: "status" | "assignee" | "priority"
  from: string | null
  to: string | null
}

export type DemoTicketSeed = {
  email: string | null
  /** 設問の並びと同じ順の回答。 */
  answers: string[]
  status: StatusKey
  priority: TicketPriority
  /** DEMO_MEMBERS の添字。 */
  assignee?: number
  minutesAgo: number
  messages?: DemoMessageSeed[]
  changes?: DemoChangeSeed[]
  memo?: string
}

export type DemoFormSeed = {
  id: string
  title: string
  questions: FormQuestion[]
  titleQuestionId: string | null
  tickets: DemoTicketSeed[]
  /** 「同期する」を押したときに Google フォームから届く回答。 */
  incoming: DemoTicketSeed[]
}

const STATUS_SEEDS: { key: StatusKey; name: string; color: string }[] = [
  { key: "open", name: "未対応", color: "#EF4444" },
  { key: "progress", name: "対応中", color: "#3B82F6" },
  { key: "waiting", name: "確認待ち", color: "#F59E0B" },
  { key: "done", name: "完了", color: "#10B981" },
]

export const demoStatusId = (formId: string, key: StatusKey) => `${formId}-status-${key}`

export const demoStatuses = (formId: string): FormStatus[] =>
  STATUS_SEEDS.map((status, index) => ({
    id: demoStatusId(formId, status.key),
    form_id: formId,
    name: status.name,
    color: status.color,
    display_order: index,
    is_default: index === 0,
  }))

export const DEMO_MEMBERS: Member[] = [
  { id: "demo-member-1", email: "misaki@example.com", display_name: "佐藤 美咲", role: "admin" },
  { id: "demo-member-2", email: "ken@example.com", display_name: "田中 健", role: "editor" },
  { id: "demo-member-3", email: "aya@example.com", display_name: "鈴木 彩", role: "editor" },
]

const SATO = 0
const TANAKA = 1
const SUZUKI = 2
const name = (member: number) => DEMO_MEMBERS[member].display_name

/** デモを操作している人。変更履歴の操作者として表示する。 */
export const DEMO_VIEWER = DEMO_MEMBERS[SATO]

const questions = (formId: string, titles: string[]): FormQuestion[] =>
  titles.map((title, index) => ({
    question_id: `${formId}-q${index + 1}`,
    title,
    question_type: "TEXT",
    options: [],
  }))

const HOUR = 60
const DAY = 24 * HOUR

const say = (minutesAgo: number, by: number | "respondent", text: string): DemoMessageSeed => ({
  minutesAgo,
  by,
  text,
})

const assign = (minutesAgo: number, by: number, to: number): DemoChangeSeed => ({
  minutesAgo,
  by,
  field: "assignee",
  from: null,
  to: name(to),
})

const status = (minutesAgo: number, by: number, from: string, to: string): DemoChangeSeed => ({
  minutesAgo,
  by,
  field: "status",
  from,
  to,
})

const priority = (
  minutesAgo: number,
  by: number,
  from: TicketPriority,
  to: TicketPriority,
): DemoChangeSeed => ({ minutesAgo, by, field: "priority", from, to })

export const DEMO_FORMS: DemoFormSeed[] = [
  {
    id: "demo-equipment",
    title: "備品の貸出申請",
    questions: questions("demo-equipment", ["備品", "お名前", "部署", "利用期間", "利用目的"]),
    titleQuestionId: null,
    tickets: [
      {
        email: "goto@example.com",
        answers: [
          "プロジェクター",
          "後藤 隼",
          "営業部",
          "11/20〜11/21",
          "取引先での提案に使います。",
        ],
        status: "open",
        priority: "medium",
        minutesAgo: 4 * HOUR,
      },
      {
        email: "ogawa@example.com",
        answers: [
          "移動式ホワイトボード",
          "小川 直人",
          "開発部",
          "11/14〜11/15",
          "設計レビューで使います。",
        ],
        status: "waiting",
        priority: "low",
        assignee: TANAKA,
        minutesAgo: DAY,
        changes: [
          assign(22 * HOUR, TANAKA, TANAKA),
          status(21 * HOUR, TANAKA, "未対応", "確認待ち"),
        ],
        messages: [
          say(
            21 * HOUR,
            TANAKA,
            "14日の午前中は別の部署の予約と重なっています。午後からでも大丈夫ですか。",
          ),
        ],
      },
      {
        email: "okamoto@example.com",
        answers: [
          "ノートパソコン（2台）",
          "岡本 七海",
          "人事部",
          "11/18〜11/30",
          "新入社員の研修で使います。",
        ],
        status: "progress",
        priority: "medium",
        assignee: SUZUKI,
        minutesAgo: DAY + 6 * HOUR,
        changes: [
          assign(DAY + 5 * HOUR, TANAKA, SUZUKI),
          status(DAY + 3 * HOUR, SUZUKI, "未対応", "対応中"),
        ],
        messages: [
          say(DAY + 3 * HOUR, SUZUKI, "2台ご用意できます。初期設定を済ませたもので大丈夫ですか。"),
          say(DAY + HOUR, "respondent", "はい、それでお願いします。"),
        ],
        memo: "PC-07 と PC-08 を貸し出す。17日の夕方に人事部へ届ける。",
      },
      {
        email: "nishida@example.com",
        answers: [
          "会議用マイクスピーカー",
          "西田 美穂",
          "総務部",
          "11/13",
          "全社会議の配信で使います。",
        ],
        status: "done",
        priority: "low",
        assignee: TANAKA,
        minutesAgo: 2 * DAY + 2 * HOUR,
        changes: [
          assign(2 * DAY, TANAKA, TANAKA),
          status(2 * DAY - HOUR, TANAKA, "未対応", "完了"),
        ],
        messages: [
          say(
            2 * DAY - HOUR,
            TANAKA,
            "受付に置いておきました。使い終わったら総務の棚に戻してください。",
          ),
        ],
      },
    ],
    incoming: [
      {
        email: "nishimura@example.com",
        answers: ["延長コード（10m）", "西村 楓", "営業部", "11/25", "展示会のブースで使います。"],
        status: "open",
        priority: "low",
        minutesAgo: 1,
      },
    ],
  },
  {
    id: "demo-business-cards",
    title: "名刺の発注依頼",
    questions: questions("demo-business-cards", ["お名前", "部署", "役職", "枚数", "備考"]),
    titleQuestionId: null,
    tickets: [
      {
        email: "murakami@example.com",
        answers: ["村上 翼", "企画部", "主任", "100枚", "裏面に英語表記も入れてください。"],
        status: "open",
        priority: "low",
        minutesAgo: 3 * HOUR,
      },
      {
        email: "maeda@example.com",
        answers: [
          "前田 裕子",
          "営業部",
          "課長",
          "200枚",
          "役職が変わったので、新しい役職で作り直したいです。",
        ],
        status: "waiting",
        priority: "medium",
        assignee: SUZUKI,
        minutesAgo: DAY + 2 * HOUR,
        changes: [
          assign(DAY + HOUR, SATO, SUZUKI),
          status(DAY, SUZUKI, "未対応", "対応中"),
          status(20 * HOUR, SUZUKI, "対応中", "確認待ち"),
        ],
        messages: [
          say(
            20 * HOUR,
            SUZUKI,
            "仕上がりの確認用データを社内チャットで送りました。役職と電話番号に誤りがないか見ていただけますか。",
          ),
        ],
      },
      {
        email: "hara@example.com",
        answers: [
          "原 悠希",
          "開発部",
          "",
          "100枚",
          "11/18入社です。入社日に間に合うとうれしいです。",
        ],
        status: "progress",
        priority: "high",
        assignee: SUZUKI,
        minutesAgo: 2 * DAY,
        changes: [
          assign(2 * DAY - HOUR, SATO, SUZUKI),
          priority(2 * DAY - HOUR, SATO, "medium", "high"),
          status(2 * DAY - 2 * HOUR, SUZUKI, "未対応", "対応中"),
        ],
        messages: [
          say(
            2 * DAY - 2 * HOUR,
            SUZUKI,
            "入社日に間に合うよう、急ぎで発注しました。11/17に届く予定です。",
          ),
        ],
        memo: "新入社員2名分（原さん・水野さん）をまとめて発注済み。注文番号 BC-1108。",
      },
      {
        email: "mizuno@example.com",
        answers: ["水野 颯太", "開発部", "", "100枚", ""],
        status: "progress",
        priority: "high",
        assignee: SUZUKI,
        minutesAgo: 2 * DAY + 10,
        changes: [
          assign(2 * DAY - HOUR, SATO, SUZUKI),
          priority(2 * DAY - HOUR, SATO, "medium", "high"),
          status(2 * DAY - 2 * HOUR, SUZUKI, "未対応", "対応中"),
        ],
        memo: "原さんの分とまとめて発注済み。注文番号 BC-1108。",
      },
      {
        email: "takahashi@example.com",
        answers: ["高橋 恵", "経理部", "", "100枚", "残りが少なくなったので追加をお願いします。"],
        status: "done",
        priority: "low",
        assignee: SUZUKI,
        minutesAgo: 6 * DAY,
        changes: [
          assign(6 * DAY - HOUR, SATO, SUZUKI),
          status(6 * DAY - HOUR, SUZUKI, "未対応", "対応中"),
          status(3 * DAY, SUZUKI, "対応中", "完了"),
        ],
        messages: [
          say(3 * DAY, SUZUKI, "名刺が届きました。総務の棚に置いてあります。"),
          say(3 * DAY - 2 * HOUR, "respondent", "受け取りました。ありがとうございます。"),
        ],
      },
    ],
    incoming: [
      {
        email: "ishii@example.com",
        answers: [
          "石井 健太",
          "営業部",
          "",
          "200枚",
          "来月の展示会で使うので、多めにお願いします。",
        ],
        status: "open",
        priority: "medium",
        minutesAgo: 1,
      },
    ],
  },
  {
    id: "demo-helpdesk",
    title: "社内ヘルプデスク",
    questions: questions("demo-helpdesk", ["件名", "お名前", "部署", "種類", "内容"]),
    titleQuestionId: "demo-helpdesk-q1",
    tickets: [
      {
        email: "maeda@example.com",
        answers: [
          "会議室Bのモニターに映らない",
          "前田 裕子",
          "営業部",
          "設備",
          "14時からの来客との打ち合わせで使いたいのですが、HDMIでつないでも画面が映りません。",
        ],
        status: "open",
        priority: "high",
        minutesAgo: 15,
      },
      {
        email: "murakami@example.com",
        answers: [
          "共有フォルダの閲覧権限がほしい",
          "村上 翼",
          "企画部",
          "アカウント・権限",
          "「2026年度_販促資料」のフォルダを見られるようにしてください。",
        ],
        status: "open",
        priority: "low",
        minutesAgo: 2 * HOUR,
      },
      {
        email: "ogawa@example.com",
        answers: [
          "VPNにつながらない",
          "小川 直人",
          "開発部",
          "ネットワーク",
          "在宅勤務中ですが、朝からVPNに接続できません。「認証に失敗しました」と表示されます。",
        ],
        status: "progress",
        priority: "high",
        assignee: SUZUKI,
        minutesAgo: 5 * HOUR,
        changes: [
          assign(4 * HOUR + 50, SATO, SUZUKI),
          priority(4 * HOUR + 50, SATO, "medium", "high"),
          status(4 * HOUR + 40, SUZUKI, "未対応", "対応中"),
        ],
        messages: [
          say(
            4 * HOUR + 40,
            SUZUKI,
            "証明書の期限が切れている可能性があります。VPNの設定画面で「証明書」の有効期限を確認していただけますか。",
          ),
          say(4 * HOUR, "respondent", "確認したところ、11/10で切れていました。"),
          say(
            3 * HOUR + 30,
            SUZUKI,
            "新しい証明書をメールでお送りしました。読み込んでから、もう一度接続してみてください。",
          ),
        ],
        memo: "証明書の期限切れは今月5件目。全社に更新の案内を出すか佐藤さんと相談する。",
      },
      {
        email: "okamoto@example.com",
        answers: [
          "新入社員のアカウント作成",
          "岡本 七海",
          "人事部",
          "アカウント・権限",
          "11/18入社の2名分、メールと社内チャットのアカウントを作成してください。",
        ],
        status: "progress",
        priority: "medium",
        assignee: TANAKA,
        minutesAgo: DAY + 3 * HOUR,
        changes: [
          assign(DAY + 2 * HOUR, SATO, TANAKA),
          status(DAY + HOUR, TANAKA, "未対応", "対応中"),
        ],
        messages: [
          say(
            DAY + HOUR,
            TANAKA,
            "承知しました。お二人の氏名のローマ字表記を教えていただけますか。",
          ),
          say(DAY - HOUR, "respondent", "Yuki Hara さんと Sota Mizuno さんです。"),
        ],
        memo: "11/17までに作成して、PCと一緒に渡す。初期パスワードは紙で手渡し。",
      },
      {
        email: "takahashi@example.com",
        answers: [
          "3階の複合機で紙詰まりが続く",
          "高橋 恵",
          "経理部",
          "設備",
          "3階の複合機で、印刷するたびに紙が詰まります。",
        ],
        status: "waiting",
        priority: "low",
        assignee: TANAKA,
        minutesAgo: DAY + 6 * HOUR,
        changes: [
          assign(DAY + 5 * HOUR, TANAKA, TANAKA),
          status(DAY + 4 * HOUR, TANAKA, "未対応", "確認待ち"),
        ],
        messages: [
          say(
            DAY + 4 * HOUR,
            TANAKA,
            "保守業者に点検を依頼しました。11/14の午前中に来てもらいます。それまでは2階の複合機を使ってください。",
          ),
        ],
        memo: "保守契約番号 MF-3021。点検のあとに請求書が届く。",
      },
      {
        email: "morita@example.com",
        answers: [
          "画像編集ソフトを使いたい",
          "森田 美里",
          "デザイン部",
          "ソフトウェア",
          "画像編集ソフトのライセンスを1つ追加してほしいです。",
        ],
        status: "waiting",
        priority: "medium",
        assignee: SATO,
        minutesAgo: 2 * DAY,
        changes: [
          assign(2 * DAY - HOUR, SATO, SATO),
          status(2 * DAY - 2 * HOUR, SATO, "未対応", "確認待ち"),
        ],
        messages: [
          say(
            2 * DAY - 2 * HOUR,
            SATO,
            "年間契約で月2,500円かかります。部長の承認をいただいたら、このチャットで教えてください。",
          ),
        ],
        memo: "承認が出たら管理画面からライセンスを追加する。",
      },
      {
        email: "fujii@example.com",
        answers: [
          "勤怠システムにログインできない",
          "藤井 大輝",
          "営業部",
          "アカウント・権限",
          "パスワードを何度か間違えてしまい、ロックされました。",
        ],
        status: "done",
        priority: "medium",
        assignee: SUZUKI,
        minutesAgo: 3 * DAY,
        changes: [
          assign(3 * DAY - 20, SATO, SUZUKI),
          status(3 * DAY - 30, SUZUKI, "未対応", "対応中"),
          status(3 * DAY - 2 * HOUR, SUZUKI, "対応中", "完了"),
        ],
        messages: [
          say(
            3 * DAY - 30,
            SUZUKI,
            "ロックを解除して、パスワードを初期化しました。仮のパスワードを社内チャットで送ったので、ログインしたら変更してください。",
          ),
          say(3 * DAY - 2 * HOUR, "respondent", "ログインできました。ありがとうございます。"),
        ],
      },
      {
        email: "ishii@example.com",
        answers: [
          "社用スマホの画面が割れた",
          "石井 健太",
          "営業部",
          "端末",
          "外出中に落として画面が割れてしまいました。タッチ操作はできます。",
        ],
        status: "done",
        priority: "medium",
        assignee: TANAKA,
        minutesAgo: 5 * DAY,
        changes: [
          assign(5 * DAY - HOUR, SATO, TANAKA),
          status(5 * DAY - HOUR, TANAKA, "未対応", "対応中"),
          status(4 * DAY, TANAKA, "対応中", "完了"),
        ],
        messages: [
          say(
            5 * DAY - HOUR,
            TANAKA,
            "端末の保険で修理できます。代わりの端末を用意したので、総務の窓口で受け取ってください。",
          ),
          say(4 * DAY + 2 * HOUR, "respondent", "代わりの端末を受け取りました。"),
        ],
        memo: "修理は3営業日。保険の申請は済み。",
      },
    ],
    incoming: [
      {
        email: "hara@example.com",
        answers: [
          "5階のWi-Fiが何度も切れる",
          "原 悠希",
          "開発部",
          "ネットワーク",
          "5階の窓側の席で、Wi-Fiが10分おきくらいに切れます。",
        ],
        status: "open",
        priority: "medium",
        minutesAgo: 1,
      },
    ],
  },
  {
    id: "demo-inquiry",
    title: "お問い合わせフォーム",
    questions: questions("demo-inquiry", [
      "件名",
      "お名前",
      "会社名",
      "お問い合わせ種別",
      "お問い合わせ内容",
    ]),
    titleQuestionId: "demo-inquiry-q1",
    tickets: [
      {
        email: "yamada@example.com",
        answers: [
          "請求書の再発行について",
          "山田 太郎",
          "株式会社ヤマダ商事",
          "お支払い",
          "10月分の請求書を紛失してしまいました。PDFで再発行していただけますか。宛名は「株式会社ヤマダ商事 経理部」でお願いします。",
        ],
        status: "open",
        priority: "high",
        minutesAgo: 25,
      },
      {
        email: "kobayashi@example.net",
        answers: [
          "パスワードの再設定メールが届きません",
          "小林 花子",
          "",
          "アカウント",
          "ログインできなくなったため再設定を試しましたが、メールが届きません。迷惑メールフォルダも確認しました。",
        ],
        status: "open",
        priority: "medium",
        minutesAgo: 3 * HOUR,
        memo: "今週2件目の同じお問い合わせ。メールの配信で止まっていないか、田中さんが配信ログを確認中。",
      },
      {
        email: "ito@example.com",
        answers: [
          "チームでの導入を検討しています",
          "伊藤 大輔",
          "株式会社イトウ設計",
          "導入のご相談",
          "設計事務所で10名ほどの利用を考えています。料金と、導入までの流れを教えてください。",
        ],
        status: "progress",
        priority: "high",
        assignee: SATO,
        minutesAgo: DAY + 2 * HOUR,
        changes: [
          assign(DAY + HOUR, SATO, SATO),
          status(DAY + HOUR, SATO, "未対応", "対応中"),
          priority(DAY + HOUR, SATO, "medium", "high"),
        ],
        messages: [
          say(
            DAY + 50,
            SATO,
            "お問い合わせありがとうございます。10名でのご利用でしたらチームプランが合うかと思います。資料をお送りしましたので、オンラインで30分ほどご説明のお時間をいただけますか。",
          ),
          say(
            5 * HOUR,
            "respondent",
            "資料ありがとうございます。来週の火曜か水曜の午後でお願いできますか。",
          ),
        ],
        memo: "チームプランで提案。11/18(火) 14時からで日程を調整中。\n見積もりは鈴木さんに依頼済み（11/14まで）。",
      },
      {
        email: "watanabe@example.org",
        answers: [
          "領収書の宛名を変更したい",
          "渡辺 由美",
          "株式会社ワタナベ",
          "お支払い",
          "先日発行していただいた領収書の宛名を、個人名から会社名に変更してほしいです。",
        ],
        status: "progress",
        priority: "low",
        assignee: TANAKA,
        minutesAgo: DAY + 5 * HOUR,
        changes: [assign(DAY + 3 * HOUR, SATO, TANAKA), status(DAY, TANAKA, "未対応", "対応中")],
        messages: [
          say(
            DAY,
            TANAKA,
            "承知しました。正式な会社名（「株式会社」の位置も含めて）を教えていただけますか。",
          ),
          say(20 * HOUR, "respondent", "「株式会社ワタナベ」でお願いします。"),
        ],
        memo: "経理に再発行を依頼済み。明日中にPDFで送る。",
      },
      {
        email: "nakamura@example.com",
        answers: [
          "回答をCSVで書き出したい",
          "中村 翔",
          "中村会計事務所",
          "機能について",
          "集めた回答をCSVで書き出して、会計ソフトに取り込みたいです。方法はありますか。",
        ],
        status: "waiting",
        priority: "medium",
        assignee: SUZUKI,
        minutesAgo: 2 * DAY + 3 * HOUR,
        changes: [
          assign(2 * DAY + HOUR, SATO, SUZUKI),
          status(2 * DAY, SUZUKI, "未対応", "対応中"),
          status(DAY + 22 * HOUR, SUZUKI, "対応中", "確認待ち"),
        ],
        messages: [
          say(
            DAY + 22 * HOUR,
            SUZUKI,
            "いまは画面からCSVを書き出す機能がないため、こちらで書き出してお送りします。必要な期間を教えていただけますか。",
          ),
        ],
        memo: "CSV出力の要望は今月3件目。開発チームに要望として共有済み。",
      },
      {
        email: null,
        answers: [
          "料金ページの表示が崩れています",
          "",
          "",
          "不具合の報告",
          "スマートフォンで料金ページを開くと、表の右側が切れて見えません。iPhoneのSafariです。",
        ],
        status: "open",
        priority: "medium",
        minutesAgo: 2 * DAY + 6 * HOUR,
        memo: "メールアドレスがないため返信はできない。\niPhone SE で再現した。開発チームに修正を依頼済み。",
      },
      {
        email: "kato@example.net",
        answers: [
          "プランを変更したい",
          "加藤 涼",
          "株式会社カトウ",
          "お支払い",
          "来月からスタンダードプランに変更したいです。月の途中で変えた場合は日割りになりますか。",
        ],
        status: "done",
        priority: "low",
        assignee: TANAKA,
        minutesAgo: 4 * DAY,
        changes: [
          assign(4 * DAY - HOUR, SATO, TANAKA),
          status(3 * DAY + 22 * HOUR, TANAKA, "未対応", "対応中"),
          status(3 * DAY + 2 * HOUR, TANAKA, "対応中", "完了"),
        ],
        messages: [
          say(
            3 * DAY + 22 * HOUR,
            TANAKA,
            "プランは管理画面の「契約」から変更できます。月の途中で変更された場合は、日割りで計算いたします。",
          ),
          say(3 * DAY + 3 * HOUR, "respondent", "変更できました。ありがとうございました。"),
        ],
      },
      {
        email: "yoshida@example.com",
        answers: [
          "登録の確認メールが届きません",
          "吉田 優",
          "吉田デザイン事務所",
          "アカウント",
          "新規登録したのですが、確認メールが届きません。",
        ],
        status: "done",
        priority: "high",
        assignee: SATO,
        minutesAgo: 6 * DAY,
        changes: [
          assign(6 * DAY - 30, SATO, SATO),
          priority(6 * DAY - 30, SATO, "medium", "high"),
          status(6 * DAY - 30, SATO, "未対応", "対応中"),
          status(5 * DAY, SATO, "対応中", "完了"),
        ],
        messages: [
          say(
            6 * DAY - 40,
            SATO,
            "ご不便をおかけしています。メールの受信設定で「example.com」からのメールを受け取れるようにしていただけますか。",
          ),
          say(5 * DAY + 2 * HOUR, "respondent", "受信設定を変えたら届きました。助かりました。"),
        ],
        memo: "会社のメールサーバーで止められていたケース。よくある質問に追記する（佐藤）。",
      },
    ],
    incoming: [
      {
        email: "matsumoto@example.com",
        answers: [
          "事務の担当者を追加したい",
          "松本 葵",
          "松本工務店",
          "機能について",
          "事務の担当者を1名追加したいのですが、どこから招待できますか。",
        ],
        status: "open",
        priority: "medium",
        minutesAgo: 1,
      },
    ],
  },
  {
    id: "demo-event",
    title: "【12/5開催】オンラインセミナー 参加申込",
    questions: questions("demo-event", [
      "お名前",
      "会社名",
      "参加方法",
      "参加人数",
      "ご質問・ご要望",
    ]),
    titleQuestionId: null,
    tickets: [
      {
        email: "inoue@example.com",
        answers: [
          "井上 誠",
          "株式会社サンプル",
          "会場",
          "1名",
          "仕事の都合で15分ほど遅れそうです。途中から参加しても大丈夫ですか。",
        ],
        status: "open",
        priority: "low",
        minutesAgo: 40,
      },
      {
        email: "kimura@example.net",
        answers: ["木村 さくら", "木村デザイン", "オンライン", "1名", "録画は後日見られますか。"],
        status: "waiting",
        priority: "medium",
        assignee: SUZUKI,
        minutesAgo: 9 * HOUR,
        changes: [assign(8 * HOUR, SATO, SUZUKI), status(6 * HOUR, SUZUKI, "未対応", "確認待ち")],
        messages: [
          say(
            6 * HOUR,
            SUZUKI,
            "お申し込みありがとうございます。録画はお申し込みの方に限り、開催後1週間ご覧いただけます。視聴用のURLは前日にお送りしますが、届け先はこのメールアドレスでよろしいですか。",
          ),
        ],
      },
      {
        email: "hayashi@example.org",
        answers: [
          "林 拓海",
          "株式会社ハヤシ物流",
          "会場",
          "3名",
          "同じ部署の3名で参加します。席を並びにしていただけますか。",
        ],
        status: "progress",
        priority: "medium",
        assignee: TANAKA,
        minutesAgo: DAY + 4 * HOUR,
        changes: [assign(DAY + 2 * HOUR, SATO, TANAKA), status(DAY, TANAKA, "未対応", "対応中")],
        messages: [
          say(
            DAY,
            TANAKA,
            "3名分のお席を並びでご用意できます。受付で名札をお渡しするため、参加される方のお名前を教えていただけますか。",
          ),
          say(20 * HOUR, "respondent", "林、山本、佐々木の3名です。"),
        ],
        memo: "前方の3席を確保。名札は3名分用意する。\n来年の導入を検討中とのこと。セミナー後に営業からご連絡する。",
      },
      {
        email: "shimizu@example.com",
        answers: [
          "清水 千尋",
          "株式会社テスト",
          "会場",
          "1名",
          "車椅子で参加します。会場にエレベーターはありますか。",
        ],
        status: "progress",
        priority: "high",
        assignee: SATO,
        minutesAgo: 2 * DAY,
        changes: [
          assign(2 * DAY - HOUR, SATO, SATO),
          priority(2 * DAY - HOUR, SATO, "medium", "high"),
          status(2 * DAY - HOUR, SATO, "未対応", "対応中"),
        ],
        messages: [
          say(
            DAY + 20 * HOUR,
            SATO,
            "会場の3階まではエレベーターで上がれます。当日は1階の入口までスタッフがお迎えにあがります。",
          ),
          say(DAY + 6 * HOUR, "respondent", "ありがとうございます。助かります。"),
        ],
        memo: "前列に車椅子のスペースを確保済み。当日の案内は鈴木さん。",
      },
      {
        email: "yamaguchi@example.com",
        answers: ["山口 陽菜", "株式会社サンプル", "オンライン", "1名", ""],
        status: "done",
        priority: "low",
        assignee: SUZUKI,
        minutesAgo: 3 * DAY,
        changes: [
          assign(3 * DAY - HOUR, SATO, SUZUKI),
          status(2 * DAY + 20 * HOUR, SUZUKI, "未対応", "完了"),
        ],
        messages: [
          say(
            2 * DAY + 20 * HOUR,
            SUZUKI,
            "お申し込みを受け付けました。視聴用のURLは前日にお送りします。当日はよろしくお願いいたします。",
          ),
        ],
      },
    ],
    incoming: [
      {
        email: "mori@example.net",
        answers: [
          "森 蓮",
          "森工務店",
          "オンライン",
          "1名",
          "初めて参加します。事前に準備しておくものはありますか。",
        ],
        status: "open",
        priority: "medium",
        minutesAgo: 1,
      },
    ],
  },
]
