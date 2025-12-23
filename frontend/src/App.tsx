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

function App() {
  return (
    <ThemeProvider defaultTheme="light" storageKey="ui-theme">
      <Router>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/" element={<FormsListPage />} />
          <Route path="/forms/:id" element={<FormManagementPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </ThemeProvider>
  );
}

export default App;
