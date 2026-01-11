"use client";

import { useState, useEffect } from "react";
import type { TicketHistory } from "@/types";
import { apiClient } from "@/lib/api";

export function useTicketHistories(ticketId: string) {
  const [histories, setHistories] = useState<TicketHistory[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchHistories = async () => {
      try {
        setIsLoading(true);
        setError(null);
        const response = await apiClient.getTicketHistories(ticketId);
        setHistories(response.histories);
      } catch (err) {
        setError(err instanceof Error ? err : new Error("チケットヒストリーの取得に失敗しました"));
      } finally {
        setIsLoading(false);
      }
    };

    fetchHistories();
  }, [ticketId]);

  return {
    histories,
    isLoading,
    error,
  };
}
