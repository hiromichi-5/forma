"use client"

import { useState } from "react"
import type { FormResponse } from "@/types/form-response"
import { mockFormResponses } from "@/lib/mock-data"

export function useFormResponses() {
  const [responses, setResponses] = useState<FormResponse[]>(mockFormResponses)
  const [loading, setLoading] = useState(false)

  const updateResponseStatus = (id: string, status: FormResponse["status"]) => {
    setResponses((prev) => prev.map((r) => (r.id === id ? { ...r, status } : r)))
  }

  const assignResponse = (id: string, userId: string | null) => {
    setResponses((prev) => prev.map((r) => (r.id === id ? { ...r, assignedTo: userId } : r)))
  }

  const updatePriority = (id: string, priority: FormResponse["priority"]) => {
    setResponses((prev) => prev.map((r) => (r.id === id ? { ...r, priority } : r)))
  }

  return {
    responses,
    loading,
    updateResponseStatus,
    assignResponse,
    updatePriority,
  }
}
