import { createContext, useContext, useState, useEffect } from "react";
import type { ReactNode } from "react";
import { apiClient } from "@/lib/api";
import type {
  User,
  UserProfile,
  LoginRequest,
  SignupRequest,
  UpdateUserProfileRequest,
  ChangePasswordRequest,
} from "@/types";

type AuthContextType = {
  user: User | null;
  profile: UserProfile | null;
  isLoading: boolean;
  isProfileLoading: boolean;
  isAuthenticated: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  signup: (credentials: SignupRequest) => Promise<void>;
  logout: () => void;
  updateProfile: (request: UpdateUserProfileRequest) => Promise<UserProfile>;
  refreshProfile: () => Promise<void>;
  changePassword: (request: ChangePasswordRequest) => Promise<void>;
  deleteAccount: () => Promise<void>;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isProfileLoading, setIsProfileLoading] = useState(false);

  const isAuthenticated = !!user;

  const refreshProfile = async () => {
    if (!user) return;
    setIsProfileLoading(true);
    try {
      const profileData = await apiClient.getProfile();
      setProfile(profileData);
      // userのemailも更新
      setUser((prev) => (prev ? { ...prev, email: profileData.email } : null));
    } catch (error) {
      console.error("Failed to fetch profile:", error);
    } finally {
      setIsProfileLoading(false);
    }
  };

  useEffect(() => {
    const initializeAuth = async () => {
      const token = localStorage.getItem("forma_token");
      if (token) {
        try {
          const response = await apiClient.whoami();
          const newUser: User = {
            id: response.user_id,
          };
          setUser(newUser);

          try {
            const profileData = await apiClient.getProfile();
            setProfile(profileData);
            newUser.email = profileData.email;
            setUser(newUser);
          } catch (profileError) {
            console.error("Failed to fetch profile:", profileError);
          }
        } catch {
          localStorage.removeItem("forma_token");
          apiClient.clearToken();
        }
      }
      setIsLoading(false);
    };

    initializeAuth();
  }, []);

  const login = async (credentials: LoginRequest) => {
    const response = await apiClient.login(credentials);
    apiClient.setToken(response.token);

    const whoamiResponse = await apiClient.whoami();
    const newUser: User = {
      id: whoamiResponse.user_id,
    };

    setUser(newUser);
    await refreshProfile();
  };

  const signup = async (credentials: SignupRequest) => {
    const response = await apiClient.signup(credentials);
    apiClient.setToken(response.token);

    const whoamiResponse = await apiClient.whoami();
    const newUser: User = {
      id: whoamiResponse.user_id,
    };

    setUser(newUser);
    await refreshProfile();
  };

  const updateProfile = async (
    request: UpdateUserProfileRequest
  ): Promise<UserProfile> => {
    const updatedProfile = await apiClient.updateProfile(request);
    setProfile(updatedProfile);
    // userのemailも更新（メールが変更可能になった場合のため）
    setUser((prev) => (prev ? { ...prev, email: updatedProfile.email } : null));
    return updatedProfile;
  };

  const logout = () => {
    apiClient.clearToken();
    setUser(null);
    setProfile(null);
  };

  const changePassword = async (request: ChangePasswordRequest) => {
    await apiClient.changePassword(request);
  };

  const deleteAccount = async () => {
    try {
      await apiClient.deleteProfile();
      logout();
    } catch (error) {
      throw error;
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        profile,
        isLoading,
        isProfileLoading,
        isAuthenticated,
        login,
        signup,
        logout,
        updateProfile,
        refreshProfile,
        changePassword,
        deleteAccount,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

/* eslint-disable react-refresh/only-export-components */
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
