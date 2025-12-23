import React from "react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import { ThemeProvider } from "./components/theme-provider";
import FormsListPage from "./pages/FormsListPage";
import FormManagementPage from "./pages/FormManagementPage";
import LoginPage from "./pages/LoginPage";
import SignupPage from "./pages/SignupPage";
import { useRequireAuth } from "./hooks/useAuth";

type RequireAuthProps = {
  children: React.ReactNode;
};

function RequireAuth({ children }: RequireAuthProps) {
  const { isLoading, shouldRedirect } = useRequireAuth();

  if (isLoading) {
    return <div className="p-6 text-sm text-muted-foreground">読み込み中...</div>;
  }

  if (shouldRedirect) {
    return <Navigate to="/login" replace />;
  }

  return children;
}

function App() {
  return (
    <ThemeProvider defaultTheme="light" storageKey="ui-theme">
      <Router>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/forms-list" element={<Navigate to="/" replace />} />
          <Route path="/login-new" element={<Navigate to="/login" replace />} />
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
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </ThemeProvider>
  );
}

export default App;
