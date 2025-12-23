"use client";

import type React from "react";

import { useState, useRef, useEffect } from "react";
import type { FormResponse } from "@/types/form-response";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { X, Send, User, Bot } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { ja } from "date-fns/locale";
import { useChatMessages } from "@/hooks/use-chat-messages";
import { cn } from "@/lib/utils";

interface ChatInterfaceProps {
  response: FormResponse;
  onClose: () => void;
  currentUserId: string;
  currentUserName: string;
}

export function ChatInterface({
  response,
  onClose,
  currentUserId,
  currentUserName,
}: ChatInterfaceProps) {
  const { messages, sendMessage } = useChatMessages(response.id);
  const [inputValue, setInputValue] = useState("");
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = () => {
    if (inputValue.trim()) {
      sendMessage(
        inputValue.trim(),
        currentUserId,
        currentUserName,
        response.respondentEmail
      );
      setInputValue("");
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <Card className="w-full max-w-3xl h-[80vh] flex flex-col">
        <div className="flex items-center justify-between p-4 border-b">
          <div className="flex-1">
            <h3 className="text-lg font-semibold">
              {response.respondentName}さんとのチャット
            </h3>
            <p className="text-sm text-muted-foreground">
              {response.respondentEmail}
            </p>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-5 w-5" />
          </Button>
        </div>

        <div className="p-4 bg-muted/50 border-b">
          <div className="flex items-center gap-2 mb-2">
            <Badge variant="outline">元の回答</Badge>
            <span className="text-xs text-muted-foreground">
              {formatDistanceToNow(response.submittedAt, {
                addSuffix: true,
                locale: ja,
              })}
            </span>
          </div>
          <div className="space-y-2 text-sm">
            {Object.entries(response.responses).map(([key, value], index) => (
              <div key={key} className="bg-background/50 p-2 rounded">
                <p className="font-medium text-muted-foreground text-xs">
                  質問 {index + 1}
                </p>
                <p>{value}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-center">
              <Bot className="h-12 w-12 text-muted-foreground mb-3" />
              <p className="text-muted-foreground">
                まだメッセージがありません
              </p>
              <p className="text-sm text-muted-foreground">
                追加の質問や不備の確認を送信してください
              </p>
            </div>
          ) : (
            messages.map((message) => (
              <div
                key={message.id}
                className={cn(
                  "flex gap-3",
                  message.senderType === "staff"
                    ? "justify-end"
                    : "justify-start"
                )}
              >
                <div
                  className={cn(
                    "flex gap-3 max-w-[80%]",
                    message.senderType === "staff"
                      ? "flex-row-reverse"
                      : "flex-row"
                  )}
                >
                  <div
                    className={cn(
                      "flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center",
                      message.senderType === "staff"
                        ? "bg-primary text-primary-foreground"
                        : "bg-muted"
                    )}
                  >
                    {message.senderType === "staff" ? (
                      <User className="h-4 w-4" />
                    ) : (
                      <Bot className="h-4 w-4" />
                    )}
                  </div>

                  <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-medium">
                        {message.senderName}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {formatDistanceToNow(message.timestamp, {
                          addSuffix: true,
                          locale: ja,
                        })}
                      </span>
                    </div>

                    <div
                      className={cn(
                        "rounded-lg p-3",
                        message.senderType === "staff"
                          ? "bg-primary text-primary-foreground"
                          : "bg-muted"
                      )}
                    >
                      <p className="text-sm whitespace-pre-wrap">
                        {message.message}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            ))
          )}
          <div ref={messagesEndRef} />
        </div>

        <div className="p-4 border-t">
          <div className="flex gap-2">
            <Input
              placeholder="メッセージを入力..."
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyPress={handleKeyPress}
              className="flex-1"
            />
            <Button
              onClick={handleSend}
              disabled={!inputValue.trim()}
              className="gap-2"
            >
              <Send className="h-4 w-4" />
              送信
            </Button>
          </div>
          <p className="text-xs text-muted-foreground mt-2">
            Enter で送信 / Shift + Enter で改行 /
            回答者にメール通知が送信されます
          </p>
        </div>
      </Card>
    </div>
  );
}
