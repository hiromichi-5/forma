import React from "react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
  useLocation,
} from "react-router-dom";
import FormsListPage from "./pages/FormsListPage";
import FormManagementPage from "./pages/FormManagementPage";
import LoginPage from "./pages/LoginPage";
import SignupPage from "./pages/SignupPage";
import SettingsPage from "./pages/SettingsPage";
import VerifyEmailPage from "./pages/VerifyEmailPage";
import PasswordResetPage from "./pages/PasswordResetPage";
import PasswordResetConfirmPage from "./pages/PasswordResetConfirmPage";
import AcceptInvitePage from "./pages/AcceptInvitePage";
import { useRequireAuth } from "./hooks/useAuth";

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
