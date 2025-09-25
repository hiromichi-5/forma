import { createContext, useContext, useState, useEffect } from "react";
import type { ReactNode } from "react";
import { apiClient } from "@/lib/api";
import type { User, LoginRequest, SignupRequest } from "@/types";

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  signup: (credentials: SignupRequest) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const isAuthenticated = !!user;

  useEffect(() => {
    const initializeAuth = async () => {
      const token = localStorage.getItem("forma_token");
      if (token) {
        try {
          const response = await apiClient.whoami();
          setUser({
            id: response.user_id,
          });
        } catch (error) {
          localStorage.removeItem("forma_token");
          apiClient.clearToken();
        }
      }
      setIsLoading(false);
    };

    initializeAuth();
  }, []);

  const login = async (credentials: LoginRequest) => {
    try {
      const response = await apiClient.login(credentials);
      apiClient.setToken(response.token);

      const whoamiResponse = await apiClient.whoami();
      const newUser: User = {
        id: whoamiResponse.user_id,
      };

      setUser(newUser);
    } catch (error) {
      throw error;
    }
  };

  const signup = async (credentials: SignupRequest) => {
    try {
      const response = await apiClient.signup(credentials);
      apiClient.setToken(response.token);

      const whoamiResponse = await apiClient.whoami();
      const newUser: User = {
        id: whoamiResponse.user_id,
      };

      setUser(newUser);
    } catch (error) {
      throw error;
    }
  };

  const logout = () => {
    apiClient.clearToken();
    setUser(null);
  };

  return (
    <AuthContext.Provider
      value={{ user, isLoading, isAuthenticated, login, signup, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}

export function useRequireAuth() {
  const { user, isLoading } = useAuth();

  if (isLoading) {
    return { isLoading: true, user: null };
  }

  if (!user) {
    return { isLoading: false, user: null, shouldRedirect: true };
  }

  return { isLoading: false, user, shouldRedirect: false };
}
