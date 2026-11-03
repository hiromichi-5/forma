import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { QueryClientProvider } from "@tanstack/react-query"
import { Toaster } from "sonner"
import "./index.css"
import App from "@/app"
import { AuthProvider } from "@/hooks/use-auth"
import { createQueryClient } from "@/queries/query-client"

const queryClient = createQueryClient()

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <App />
      </AuthProvider>
      <Toaster />
    </QueryClientProvider>
  </StrictMode>,
)
