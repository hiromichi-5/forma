import React from "react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
  useLocation,
} from "react-router";
import FormsListPage from "@/pages/forms-list-page";
import FormManagementPage from "@/pages/form-management-page";
import LoginPage from "@/pages/login-page";
import SignupPage from "@/pages/signup-page";
import SettingsPage from "@/pages/settings-page";
import VerifyEmailPage from "@/pages/verify-email-page";
import PasswordResetPage from "@/pages/password-reset-page";
import PasswordResetConfirmPage from "@/pages/password-reset-confirm-page";
import AcceptInvitePage from "@/pages/accept-invite-page";
import { useRequireAuth } from "@/hooks/use-auth";

type RequireAuthProps = {
  children: React.ReactNode;
};

function RequireAuth({ children }: RequireAuthProps) {
  const { isLoading, shouldRedirect } = useRequireAuth();
  const location = useLocation();

  if (isLoading) {
    return <div className="p-6 text-sm text-muted-foreground">読み込み中...</div>;
  }

  if (shouldRedirect) {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  }

  return children;
}

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/signup" element={<SignupPage />} />
        <Route path="/verify-email" element={<VerifyEmailPage />} />
        <Route path="/password-reset" element={<PasswordResetPage />} />
        <Route path="/password-reset/confirm" element={<PasswordResetConfirmPage />} />
        <Route
          path="/invites/:inviteId/accept"
          element={
            <RequireAuth>
              <AcceptInvitePage />
            </RequireAuth>
          }
        />
        <Route
          path="/"
          element={
            <RequireAuth>
              <FormsListPage />
            </RequireAuth>
          }
        />
        <Route
          path="/forms/:id"
          element={
            <RequireAuth>
              <FormManagementPage />
            </RequireAuth>
          }
        />
        <Route
          path="/settings"
          element={
            <RequireAuth>
              <SettingsPage />
            </RequireAuth>
          }
        />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Router>
  );
}

export default App;
