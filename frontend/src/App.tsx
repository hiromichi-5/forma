import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import { AuthProvider, useRequireAuth } from "./hooks/useAuth";
import { Layout } from "./components/Layout";
import { LoginPage } from "./pages/LoginPage";
import { SignupPage } from "./pages/SignupPage";
import { HomePage } from "./pages/HomePage";
import { KanbanPage } from "./pages/KanbanPage";
import FormsPage from "./pages/FormsPage";
import MembersPage from "./pages/MembersPage";
import SettingsPage from "./pages/SettingsPage";
import InviteAcceptPage from "./pages/InviteAcceptPage";

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isLoading, shouldRedirect } = useRequireAuth();

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (shouldRedirect) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/signup" element={<SignupPage />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route index element={<HomePage />} />
        <Route path="kanban" element={<KanbanPage />} />
        <Route path="kanban/:form_id" element={<KanbanPage />} />
        <Route path="forms" element={<FormsPage />} />
        <Route path="members" element={<MembersPage />} />
        <Route path="members/:form_id" element={<MembersPage />} />
        <Route path="invites/accept" element={<InviteAcceptPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

function App() {
  return (
    <AuthProvider>
      <Router>
        <AppRoutes />
      </Router>
    </AuthProvider>
  );
}

export default App;
