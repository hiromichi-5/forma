import { lazy, Suspense } from "react"
import { createBrowserRouter, Navigate, Outlet, useLocation, useParams } from "react-router"
import { RouterProvider } from "react-router/dom"
import { AppLayout } from "@/components/app-layout"
import { PageLoading } from "@/components/page-loading"
import { useRequireAuth } from "@/hooks/use-auth"
import RouteErrorPage from "@/pages/route-error-page"

// ページは表示するときに読み込む。初回に読み込む JavaScript を減らし、ログイン画面を速く表示する。
const WelcomePage = lazy(() => import("@/pages/welcome-page"))
const LoginPage = lazy(() => import("@/pages/login-page"))
const SignupPage = lazy(() => import("@/pages/signup-page"))
const VerifyEmailPage = lazy(() => import("@/pages/verify-email-page"))
const PasswordResetPage = lazy(() => import("@/pages/password-reset-page"))
const PasswordResetConfirmPage = lazy(() => import("@/pages/password-reset-confirm-page"))
const AcceptInvitePage = lazy(() => import("@/pages/accept-invite-page"))
const FormsListPage = lazy(() => import("@/pages/forms-list-page"))
const FormManagementPage = lazy(() => import("@/pages/form-management-page"))
const SettingsPage = lazy(() => import("@/pages/settings-page"))

function RootLayout() {
  return (
    <Suspense fallback={<PageLoading />}>
      <Outlet />
    </Suspense>
  )
}

/**
 * ログインが必要な画面。未ログインならログイン画面へ送り、ログイン後に元の画面へ戻す。
 * トップを開いた未ログインの人には、ログイン画面ではなくサービスの紹介を見せる。
 */
function RequireAuth() {
  const { isLoading, shouldRedirect } = useRequireAuth()
  const location = useLocation()

  if (isLoading) {
    return <PageLoading />
  }

  if (shouldRedirect) {
    if (location.pathname === "/") {
      return <Navigate to="/welcome" replace />
    }
    return <Navigate to="/login" state={{ from: location.pathname + location.search }} replace />
  }

  return <Outlet />
}

// 別のフォームへ移ったときに、絞り込みや表示形式などを前のフォームから持ち越さない。
function FormManagementRoute() {
  const { id } = useParams()
  return <FormManagementPage key={id} />
}

const router = createBrowserRouter([
  {
    element: <RootLayout />,
    errorElement: <RouteErrorPage />,
    children: [
      { path: "/welcome", element: <WelcomePage /> },
      { path: "/login", element: <LoginPage /> },
      { path: "/signup", element: <SignupPage /> },
      { path: "/verify-email", element: <VerifyEmailPage /> },
      { path: "/password-reset", element: <PasswordResetPage /> },
      { path: "/password-reset/confirm", element: <PasswordResetConfirmPage /> },
      {
        element: <RequireAuth />,
        children: [
          { path: "/invites/:inviteId/accept", element: <AcceptInvitePage /> },
          {
            element: <AppLayout />,
            children: [
              { path: "/", element: <FormsListPage /> },
              { path: "/forms/:id", element: <FormManagementRoute /> },
              { path: "/settings", element: <SettingsPage /> },
            ],
          },
        ],
      },
      { path: "*", element: <Navigate to="/" replace /> },
    ],
  },
])

export default function App() {
  return <RouterProvider router={router} />
}
