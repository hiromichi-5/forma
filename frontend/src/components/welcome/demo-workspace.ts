import { QueryClient, type InfiniteData } from "@tanstack/react-query"
import { queryKeys, type TicketListFilter } from "@/queries/keys"
import { compareTickets, matchesFilter } from "@/lib/ticket-cache"
import { respondentEmailLabel } from "@/lib/ticket-display"
import type { ResponseChatMessage } from "@/components/response-detail"
import {
  DEMO_FORMS,
  DEMO_MEMBERS,
  DEMO_VIEWER,
  demoStatusId,
  demoStatuses,
  type DemoFormSeed,
  type DemoTicketSeed,
} from "@/components/welcome/demo-data"
import type {
  FormStatus,
  Member,
  ListTicketCountsResponse,
  ListTicketHistoriesResponse,
  ListTicketsResponse,
  NotificationType,
  TicketDetail,
  TicketHistory,
  TicketPriority,
} from "@/types"

export type DemoTicketChange = {
  status_id?: string
  assignee_id?: string | null
  priority?: TicketPriority
}

export type DemoFormListItem = {
  id: string
  title: string
  responseCount: number
  latestSubmittedAt: Date | null
}

type DemoFormState = {
  seed: DemoFormSeed
  statuses: FormStatus[]
  titleQuestionId: string | null
  tickets: TicketDetail[]
  incoming: DemoTicketSeed[]
  histories: Map<string, TicketHistory[]>
  chats: Map<string, ResponseChatMessage[]>
  memos: Map<string, string>
}

export function createDemoQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { enabled: false, staleTime: Infinity, gcTime: Infinity, retry: false },
    },
  })
}

export type DemoWorkspace = ReturnType<typeof createDemoWorkspace>

export function createDemoWorkspace(queryClient: QueryClient, now: Date = new Date()) {
  let nextTicketNumber = 0
  let nextHistoryNumber = 0

  const forms = new Map<string, DemoFormState>(
    DEMO_FORMS.map((seed) => {
      const state: DemoFormState = {
        seed,
        statuses: demoStatuses(seed.id),
        titleQuestionId: seed.titleQuestionId,
        tickets: [],
        incoming: [...seed.incoming],
        histories: new Map(),
        chats: new Map(),
        memos: new Map(),
      }
      state.tickets = seed.tickets.map((ticket) => toTicket(state, ticket))
      return [seed.id, state]
    }),
  )

  function toTicket(state: DemoFormState, seed: DemoTicketSeed): TicketDetail {
    const submittedAt = new Date(now.getTime() - seed.minutesAgo * 60_000).toISOString()
    const status = state.statuses.find((s) => s.id === demoStatusId(state.seed.id, seed.status))!
    const assignee = seed.assignee === undefined ? null : DEMO_MEMBERS[seed.assignee]
    nextTicketNumber += 1
    const id = `${state.seed.id}-ticket-${nextTicketNumber}`
    const ago = (minutes: number) => new Date(now.getTime() - minutes * 60_000)
    const respondentName =
      seed.answers[state.seed.questions.findIndex((q) => q.title === "お名前")] ||
      respondentEmailLabel(seed.email)

    for (const [index, change] of (seed.changes ?? []).entries()) {
      recordHistory(state, id, change.field, change.from, change.to, {
        at: new Date(ago(change.minutesAgo).getTime() + index * 1000),
        by: DEMO_MEMBERS[change.by],
      })
    }
    state.chats.set(
      id,
      (seed.messages ?? []).map((message, index) => ({
        id: `${id}-message-${index + 1}`,
        from: message.by === "respondent" ? "respondent" : "staff",
        authorName:
          message.by === "respondent" ? respondentName : DEMO_MEMBERS[message.by].display_name,
        text: message.text,
        sentAt: ago(message.minutesAgo).toISOString(),
      })),
    )
    if (seed.memo) state.memos.set(id, seed.memo)

    return {
      id,
      form_id: state.seed.id,
      form_title: state.seed.title,
      response_id: `response-${nextTicketNumber}`,
      respondent_email: seed.email,
      status: { id: status.id, name: status.name, color: status.color },
      priority: seed.priority,
      title_question_id: state.titleQuestionId,
      title: titleOf(state, seed.answers),
      assignee: assignee
        ? { id: assignee.id, display_name: assignee.display_name, email: assignee.email }
        : null,
      submitted_at: submittedAt,
      created_at: submittedAt,
      answers: state.seed.questions.map((question, index) => ({
        question_id: question.question_id,
        question_title: question.title,
        question_type: question.question_type,
        values: [seed.answers[index] ?? ""],
        display_value: seed.answers[index] ?? "",
      })),
      notifications: [
        { notification_type: "status_change", last_sent_at: null },
        { notification_type: "assignee_assigned", last_sent_at: null },
      ],
    }
  }

  function titleOf(state: DemoFormState, answers: string[]): string {
    const questionId = state.titleQuestionId ?? state.seed.questions[0]?.question_id
    const index = state.seed.questions.findIndex((q) => q.question_id === questionId)
    return answers[index] ?? ""
  }

  function stateOf(formId: string): DemoFormState {
    const state = forms.get(formId)
    if (!state) throw new Error(`unknown demo form: ${formId}`)
    return state
  }

  function listPages(state: DemoFormState, filter: TicketListFilter) {
    const tickets = state.tickets.filter((t) => matchesFilter(t, filter)).sort(compareTickets)
    const data: InfiniteData<ListTicketsResponse, string | null> = {
      pages: [{ tickets, next_cursor: null }],
      pageParams: [null],
    }
    return data
  }

  function counts(state: DemoFormState, query: string): ListTicketCountsResponse {
    return {
      counts: state.statuses.map((status) => ({
        status_id: status.id,
        count: state.tickets.filter((t) => matchesFilter(t, { statusIds: [status.id], query }))
          .length,
      })),
    }
  }

  function writeTicket(state: DemoFormState, ticket: TicketDetail) {
    queryClient.setQueryData(queryKeys.ticket(ticket.id), ticket)
    queryClient.setQueryData<ListTicketHistoriesResponse>(queryKeys.ticketHistories(ticket.id), {
      histories: state.histories.get(ticket.id) ?? [],
    })
  }

  function refresh(state: DemoFormState) {
    for (const [key] of queryClient.getQueriesData({
      queryKey: queryKeys.ticketLists(state.seed.id),
    })) {
      const filter = key[key.length - 1] as TicketListFilter
      queryClient.setQueryData(key, listPages(state, filter))
    }
    for (const [key] of queryClient.getQueriesData({
      queryKey: queryKeys.ticketCountsAll(state.seed.id),
    })) {
      const { query } = key[key.length - 1] as { query: string }
      queryClient.setQueryData(key, counts(state, query))
    }
  }

  function recordHistory(
    state: DemoFormState,
    ticketId: string,
    fieldName: string,
    oldValue: string | null,
    newValue: string | null,
    { at = new Date(), by = DEMO_VIEWER }: { at?: Date; by?: Member } = {},
  ) {
    nextHistoryNumber += 1
    const history: TicketHistory = {
      id: `demo-history-${nextHistoryNumber}`,
      ticket_id: ticketId,
      changed_by: by.id,
      changed_by_name: by.display_name,
      field_name: fieldName,
      old_value: oldValue,
      new_value: newValue,
      created_at: at.toISOString(),
    }
    state.histories.set(ticketId, [history, ...(state.histories.get(ticketId) ?? [])])
  }

  function prepare(formId: string, filter: TicketListFilter) {
    const state = stateOf(formId)
    const filters = [
      filter,
      ...state.statuses.map((status) => ({ statusIds: [status.id], query: filter.query })),
    ]
    for (const f of filters) {
      queryClient.setQueryData(queryKeys.ticketList(formId, f), listPages(state, f))
    }
    queryClient.setQueryData(
      queryKeys.ticketCounts(formId, filter.query),
      counts(state, filter.query),
    )
  }

  forms.forEach((state, formId) => {
    state.tickets.forEach((ticket) => writeTicket(state, ticket))
    prepare(formId, { statusIds: [], query: "" })
  })

  return {
    formList(): DemoFormListItem[] {
      return [...forms.values()].map((state) => {
        const latest = [...state.tickets].sort(compareTickets)[0]
        return {
          id: state.seed.id,
          title: state.seed.title,
          responseCount: state.tickets.length,
          latestSubmittedAt: latest ? new Date(latest.submitted_at) : null,
        }
      })
    },

    form(formId: string) {
      const state = stateOf(formId)
      return {
        title: state.seed.title,
        questions: state.seed.questions,
        statuses: state.statuses,
        members: DEMO_MEMBERS,
        titleQuestionId: state.titleQuestionId,
      }
    },

    prepare,

    updateTicket(formId: string, ticketId: string, change: DemoTicketChange) {
      const state = stateOf(formId)
      const index = state.tickets.findIndex((t) => t.id === ticketId)
      if (index === -1) return
      const before = state.tickets[index]
      const after = { ...before }

      if (change.status_id !== undefined && change.status_id !== before.status.id) {
        const status = state.statuses.find((s) => s.id === change.status_id)
        if (!status) return
        after.status = { id: status.id, name: status.name, color: status.color }
        recordHistory(state, ticketId, "status", before.status.name, status.name)
      }
      if (
        change.assignee_id !== undefined &&
        change.assignee_id !== (before.assignee?.id ?? null)
      ) {
        const member = DEMO_MEMBERS.find((m) => m.id === change.assignee_id)
        after.assignee = member
          ? { id: member.id, display_name: member.display_name, email: member.email }
          : null
        recordHistory(
          state,
          ticketId,
          "assignee",
          before.assignee?.display_name ?? null,
          member?.display_name ?? null,
        )
      }
      if (change.priority !== undefined && change.priority !== before.priority) {
        after.priority = change.priority
        recordHistory(state, ticketId, "priority", before.priority, change.priority)
      }

      state.tickets[index] = after
      writeTicket(state, after)
      refresh(state)
    },

    setTitleQuestion(formId: string, questionId: string | null) {
      const state = stateOf(formId)
      state.titleQuestionId = questionId
      state.tickets = state.tickets.map((ticket) => ({
        ...ticket,
        title_question_id: questionId,
        title: titleOf(
          state,
          ticket.answers.map((answer) => answer.display_value),
        ),
      }))
      state.tickets.forEach((ticket) => writeTicket(state, ticket))
      refresh(state)
    },

    sync(formId: string): number {
      const state = stateOf(formId)
      const arrived = state.incoming.splice(0)
      now = new Date()
      for (const seed of arrived) {
        const ticket = toTicket(state, seed)
        state.tickets.push(ticket)
        writeTicket(state, ticket)
      }
      refresh(state)
      return arrived.length
    },

    chat(formId: string, ticketId: string): ResponseChatMessage[] {
      return stateOf(formId).chats.get(ticketId) ?? []
    },

    sendMessage(formId: string, ticketId: string, text: string): ResponseChatMessage[] {
      const state = stateOf(formId)
      const messages = state.chats.get(ticketId) ?? []
      const next: ResponseChatMessage[] = [
        ...messages,
        {
          id: `${ticketId}-message-${messages.length + 1}`,
          from: "staff",
          authorName: DEMO_VIEWER.display_name,
          text,
          sentAt: new Date().toISOString(),
        },
      ]
      state.chats.set(ticketId, next)
      return next
    },

    memo(formId: string, ticketId: string): string {
      return stateOf(formId).memos.get(ticketId) ?? ""
    },

    setMemo(formId: string, ticketId: string, memo: string) {
      stateOf(formId).memos.set(ticketId, memo)
    },

    sendNotification(formId: string, ticketId: string, type: NotificationType) {
      const state = stateOf(formId)
      const index = state.tickets.findIndex((t) => t.id === ticketId)
      if (index === -1) return
      const ticket = state.tickets[index]
      const sentAt = new Date().toISOString()
      state.tickets[index] = {
        ...ticket,
        notifications: ticket.notifications.map((n) =>
          n.notification_type === type ? { ...n, last_sent_at: sentAt } : n,
        ),
      }
      writeTicket(state, state.tickets[index])
    },
  }
}
