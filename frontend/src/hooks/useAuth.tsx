import { createContext, useContext, useState, useEffect, useCallback } from "react";
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
  logout: () => Promise<void>;
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

  const hydrateUser = useCallback(async () => {
    try {
      const response = await apiClient.whoami();
      const baseUser: User = { id: response.user_id };
      setUser(baseUser);
      try {
        const profileData = await apiClient.getProfile();
        setProfile(profileData);
        setUser({ ...baseUser, email: profileData.email });
      } catch (profileError) {
        console.error("Failed to fetch profile:", profileError);
      }
      return true;
    } catch (error) {
      console.warn("Failed to hydrate user:", error);
      setUser(null);
      setProfile(null);
      return false;
    }
  }, []);

  const refreshProfile = async () => {
    if (!user) return;
    setIsProfileLoading(true);
    try {
      const profileData = await apiClient.getProfile();
      setProfile(profileData);
      setUser((prev) => (prev ? { ...prev, email: profileData.email } : null));
    } catch (error) {
      console.error("Failed to fetch profile:", error);
    } finally {
      setIsProfileLoading(false);
    }
  };

  useEffect(() => {
    const initializeAuth = async () => {
      await hydrateUser();
      setIsLoading(false);
    };

    initializeAuth();
  }, [hydrateUser]);

  const login = async (credentials: LoginRequest) => {
    await apiClient.login(credentials);
    await hydrateUser();
  };

  const signup = async (credentials: SignupRequest) => {
    await apiClient.signup(credentials);
    // サインアップ後はメール認証が必要なため、自動ログインしない
  };

  const updateProfile = async (
    request: UpdateUserProfileRequest
  ): Promise<UserProfile> => {
    const updatedProfile = await apiClient.updateProfile(request);
    setProfile(updatedProfile);
    setUser((prev) => (prev ? { ...prev, email: updatedProfile.email } : null));
    return updatedProfile;
  };

  const logout = async () => {
    try {
      await apiClient.logout();
    } catch (error) {
      console.error("Failed to logout:", error);
    } finally {
      setUser(null);
      setProfile(null);
    }
  };

  const changePassword = async (request: ChangePasswordRequest) => {
    await apiClient.changePassword(request);
  };

  const deleteAccount = async () => {
    await apiClient.deleteProfile();
    await logout();
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
