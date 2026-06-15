"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { getToken, removeToken, getProfile, type UserInfo } from "../lib/api";
import { LogIn, UserPlus, LogOut, BookOpen } from "lucide-react";

export default function Nav() {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [showMenu, setShowMenu] = useState(false);

  useEffect(() => {
    const token = getToken();
    if (token) {
      getProfile().then(d => setUser(d.user)).catch(() => removeToken());
    }
  }, []);

  const handleLogout = () => {
    removeToken();
    setUser(null);
    setShowMenu(false);
    window.location.href = "/";
  };

  return (
    <header className="sticky top-0 z-50 bg-background/80 backdrop-blur-lg border-b border-border">
      <div className="max-w-4xl mx-auto px-4 h-14 flex items-center justify-between">
        <Link href="/" className="flex items-center gap-2.5 group">
          <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-sky-500 to-indigo-600 flex items-center justify-center shadow-sm group-hover:shadow-md transition-shadow">
            <svg className="w-4.5 h-4.5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M22 10v6M2 10l10-5 10 5-10 5z" />
              <path d="M6 12v5c3 3 9 3 12 0v-5" />
            </svg>
          </div>
          <span className="font-semibold text-sm tracking-tight text-foreground">Happy Study</span>
        </Link>

        <nav className="flex items-center gap-1">
          {user ? (
            /* 已登录 */
            <div className="relative">
              <button
                onClick={() => setShowMenu(!showMenu)}
                className="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-secondary transition-colors"
              >
                <div className="w-6 h-6 rounded-lg bg-gradient-to-br from-indigo-100 to-purple-100 flex items-center justify-center">
                  <span className="text-[10px] font-bold text-indigo-600">
                    {user.nickname?.charAt(0) || user.username.charAt(0)}
                  </span>
                </div>
                <span className="text-xs font-medium text-foreground hidden sm:block">{user.nickname || user.username}</span>
              </button>

              {showMenu && (
                <>
                  <div className="fixed inset-0 z-40" onClick={() => setShowMenu(false)} />
                  <div className="absolute right-0 top-full mt-1 w-44 bg-white rounded-xl border border-border/60 shadow-lg z-50 py-1.5">
                    <div className="px-3 py-2 border-b border-border/60 mb-1">
                      <p className="text-xs font-medium text-foreground">{user.nickname || user.username}</p>
                      <p className="text-[10px] text-muted-foreground">{user.email || user.username}</p>
                    </div>
                    <Link
                      href="/profile/courses"
                      onClick={() => setShowMenu(false)}
                      className="flex items-center gap-2 px-3 py-1.5 text-xs text-muted-foreground hover:text-indigo-600 hover:bg-indigo-50 transition-colors"
                    >
                      <BookOpen className="w-3.5 h-3.5" /> 我的课程
                    </Link>
                    <button
                      onClick={handleLogout}
                      className="w-full flex items-center gap-2 px-3 py-1.5 text-xs text-muted-foreground hover:text-red-600 hover:bg-red-50 transition-colors"
                    >
                      <LogOut className="w-3.5 h-3.5" /> 退出登录
                    </button>
                  </div>
                </>
              )}
            </div>
          ) : (
            /* 未登录 */
            <>
              <Link
                href="/login"
                className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-muted-foreground hover:text-foreground rounded-lg hover:bg-secondary transition-colors"
              >
                <LogIn className="w-3.5 h-3.5" /> 登录
              </Link>
              <Link
                href="/register"
                className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 rounded-lg shadow-sm transition-all"
              >
                <UserPlus className="w-3.5 h-3.5" /> 注册
              </Link>
            </>
          )}
        </nav>
      </div>
    </header>
  );
}
