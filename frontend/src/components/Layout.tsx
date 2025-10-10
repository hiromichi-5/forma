import { useState, useEffect } from "react";
import { Link, useLocation, Outlet } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { Button } from "../components/ui/Button";
import { Icon } from "../components/ui/Icon";
import type { LucideIcon } from "lucide-react";
import {
  Home,
  LayoutDashboard,
  FileText,
  Users,
  Settings,
  Menu,
  X,
  User,
  LogOut,
  UserPlus,
} from "lucide-react";

interface NavItem {
  label: string;
  href: string;
  icon: LucideIcon;
  current?: boolean;
}

function Header() {
  const { profile, logout } = useAuth();
  const [userMenuOpen, setUserMenuOpen] = useState(false);

  const displayName = profile?.display_name || "ユーザー";

  return (
    <header className="bg-blue-600 text-white shadow-md">
      <div className="px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <h1 className="text-xl font-bold">Forma</h1>
            </div>
          </div>

          <div className="flex items-center space-x-4">
            <div className="relative">
              <Button
                variant="ghost"
                onClick={() => setUserMenuOpen(!userMenuOpen)}
                className="flex items-center space-x-2 text-white hover:bg-blue-500"
                aria-label="ユーザーメニューを開く"
              >
                <Icon icon={User} />
                <span className="text-sm">{displayName}</span>
              </Button>

              {userMenuOpen && (
                <div className="absolute right-0 mt-2 w-48 bg-white text-gray-700 rounded-md shadow-lg ring-1 ring-black ring-opacity-5 z-50">
                  <div className="py-1">
                    <button
                      onClick={() => {
                        logout();
                        setUserMenuOpen(false);
                      }}
                      className="flex items-center w-full px-4 py-2 text-sm hover:bg-gray-100"
                    >
                      <Icon icon={LogOut} size="sm" className="mr-2" />
                      ログアウト
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {userMenuOpen && (
        <div
          className="fixed inset-0 z-40"
          onClick={() => setUserMenuOpen(false)}
        />
      )}
    </header>
  );
}

function Sidebar() {
  const location = useLocation();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const navigation: NavItem[] = [
    {
      label: "ホーム",
      href: "/",
      icon: Home,
      current: location.pathname === "/",
    },
    {
      label: "カンバン",
      href: "/kanban",
      icon: LayoutDashboard,
      current: location.pathname === "/kanban",
    },
    {
      label: "フォーム",
      href: "/forms",
      icon: FileText,
      current: location.pathname.startsWith("/forms"),
    },
    {
      label: "メンバー",
      href: "/members",
      icon: Users,
      current: location.pathname.startsWith("/members"),
    },
    {
      label: "招待受理",
      href: "/invites/accept",
      icon: UserPlus,
      current: location.pathname.startsWith("/invites/accept"),
    },
    {
      label: "設定",
      href: "/settings",
      icon: Settings,
      current: location.pathname === "/settings",
    },
  ];

  const SidebarContent = () => (
    <div className="flex flex-col h-full">
      <nav
        className="flex-1 px-2 py-4 space-y-1"
        aria-label="メインナビゲーション"
      >
        {navigation.map((item) => (
          <Link
            key={item.label}
            to={item.href}
            className={`group flex items-center px-2 py-2 text-sm font-medium rounded-md transition-colors ${
              item.current
                ? "bg-blue-50 text-blue-700"
                : "text-gray-700 hover:bg-gray-50 hover:text-gray-900"
            }`}
            aria-current={item.current ? "page" : undefined}
            onClick={() => setSidebarOpen(false)}
          >
            <Icon
              icon={item.icon}
              className={`mr-3 flex-shrink-0 ${
                item.current
                  ? "text-blue-600"
                  : "text-gray-400 group-hover:text-gray-500"
              }`}
            />
            {item.label}
          </Link>
        ))}
      </nav>
    </div>
  );

  return (
    <>
      <div className="lg:hidden">
        <Button
          variant="ghost"
          onClick={() => setSidebarOpen(true)}
          className="fixed top-4 left-4 z-40"
          aria-label="サイドバーを開く"
        >
          <Icon icon={Menu} size="lg" />
        </Button>

        {sidebarOpen && (
          <>
            <div
              className="fixed inset-0 z-40 bg-gray-600 bg-opacity-75"
              onClick={() => setSidebarOpen(false)}
            />
            <div className="fixed inset-y-0 left-0 z-50 w-64 bg-white shadow-xl">
              <div className="flex items-center justify-between h-16 px-4 border-b border-gray-200">
                <h2 className="text-lg font-semibold text-gray-900">
                  ナビゲーション
                </h2>
                <Button
                  variant="ghost"
                  onClick={() => setSidebarOpen(false)}
                  aria-label="サイドバーを閉じる"
                >
                  <Icon icon={X} size="lg" />
                </Button>
              </div>
              <SidebarContent />
            </div>
          </>
        )}
      </div>

      <div className="hidden lg:flex lg:flex-shrink-0">
        <div className="flex flex-col w-64">
          <div className="flex flex-col h-full bg-white border-r border-gray-200 shadow-md">
            <SidebarContent />
          </div>
        </div>
      </div>
    </>
  );
}

export function Layout() {
  const { user, profile, refreshProfile, isProfileLoading } = useAuth();

  useEffect(() => {
    if (user && !profile && !isProfileLoading) {
      refreshProfile();
    }
  }, [user, profile, isProfileLoading, refreshProfile]);

  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <div className="flex">
        <Sidebar />
        <main className="flex-1 lg:pl-0 pl-16" role="main">
          <div className="py-6 px-4 sm:px-6 lg:px-8">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}
