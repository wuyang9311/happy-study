"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { getToken, getProfile } from "../../../lib/api";
import type { UserInfo } from "../../../lib/api";
import Link from "next/link";
import { ChevronLeft, Brain, Loader2, Settings } from "lucide-react";

const sidebarItems = [
  { href: "/profile/settings/model", label: "模型设置", icon: Brain },
];

export default function SettingsLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const [user, setUser] = useState<UserInfo | null>(null);
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    const token = getToken();
    if (!token) {
      router.push("/login");
      return;
    }
    getProfile().then(d => {
      setUser(d.user);
      setChecking(false);
    }).catch(() => {
      router.push("/login");
    });
  }, []);

  if (checking) {
    return (
      <div className="flex items-center justify-center min-h-[40vh]">
        <Loader2 className="w-5 h-5 animate-spin text-indigo-500" />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 pb-8">
      {/* 顶部返回 */}
      <div className="flex items-center gap-3">
        <button onClick={() => router.push("/")}
          className="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center hover:bg-border transition-colors">
          <ChevronLeft className="w-4 h-4 text-muted-foreground" />
        </button>
        <div>
          <h1 className="text-sm font-semibold text-foreground">设置</h1>
          <p className="text-xs text-muted-foreground">个性化你的学习体验</p>
        </div>
      </div>

      <div className="flex gap-6">
        {/* 左侧菜单 */}
        <aside className="w-44 shrink-0">
          <nav className="space-y-1">
            {sidebarItems.map((item) => {
              const isActive = pathname === item.href || pathname.startsWith(item.href + "/");
              const Icon = item.icon;
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`flex items-center gap-2.5 px-3 py-2 rounded-xl text-xs font-medium transition-all ${
                    isActive
                      ? "bg-indigo-50 text-indigo-700 shadow-sm border border-indigo-100/60"
                      : "text-muted-foreground hover:text-foreground hover:bg-secondary border border-transparent"
                  }`}
                >
                  <Icon className={`w-4 h-4 ${isActive ? "text-indigo-500" : "text-muted-foreground"}`} />
                  {item.label}
                </Link>
              );
            })}
          </nav>
        </aside>

        {/* 右侧内容 */}
        <main className="flex-1 min-w-0">
          {children}
        </main>
      </div>
    </div>
  );
}
