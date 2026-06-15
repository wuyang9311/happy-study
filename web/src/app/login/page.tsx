"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "../../components/ui/card";
import { login, saveToken } from "../../lib/api";
import { LogIn, Loader2, ArrowLeft, Eye, EyeOff } from "lucide-react";
import Link from "next/link";

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [showPassword, setShowPassword] = useState(false);

  const handleLogin = async () => {
    if (!username.trim() || !password.trim()) return;
    setLoading(true);
    setError("");
    try {
      const data = await login(username.trim(), password);
      saveToken(data.token);
      router.push("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "登录失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[70vh] gap-6">
      {/* Back */}
      <div className="w-full max-w-sm">
        <button
          onClick={() => router.push("/")}
          className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors"
        >
          <ArrowLeft className="w-3.5 h-3.5" /> 返回首页
        </button>
      </div>

      <Card className="w-full max-w-sm border border-border/60 bg-white shadow-sm">
        <CardHeader className="text-center pb-4">
          <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-sky-500 to-indigo-600 flex items-center justify-center mx-auto mb-3 shadow-sm">
            <LogIn className="w-5.5 h-5.5 text-white" />
          </div>
          <CardTitle className="text-lg font-semibold">欢迎回来</CardTitle>
          <CardDescription className="text-xs text-muted-foreground">
            登录你的 Happy Study 账号
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-foreground/80">用户名</label>
            <Input
              placeholder="输入用户名"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleLogin()}
              className="h-11 text-sm bg-background border-border/80 focus-visible:ring-indigo-400/40"
              disabled={loading}
              autoFocus
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-medium text-foreground/80">密码</label>
            <div className="relative">
              <Input
                type={showPassword ? "text" : "password"}
                placeholder="输入密码"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleLogin()}
                className="h-11 text-sm bg-background border-border/80 focus-visible:ring-indigo-400/40 pr-10"
                disabled={loading}
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground/60 hover:text-muted-foreground"
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            </div>
          </div>

          {error && (
            <div className="flex items-center gap-2 p-2.5 rounded-lg bg-red-50 border border-red-200/60">
              <svg className="w-3.5 h-3.5 text-red-500 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="12" cy="12" r="10" /><line x1="15" y1="9" x2="9" y2="15" /><line x1="9" y1="9" x2="15" y2="15" />
              </svg>
              <p className="text-xs text-red-600">{error}</p>
            </div>
          )}

          <Button
            onClick={handleLogin}
            disabled={!username.trim() || !password.trim() || loading}
            className="w-full h-11 bg-gradient-to-r from-sky-500 to-indigo-600 hover:from-sky-600 hover:to-indigo-700 text-white shadow-sm disabled:opacity-50"
          >
            {loading ? (
              <span className="flex items-center gap-2"><Loader2 className="w-4 h-4 animate-spin" />登录中...</span>
            ) : "登录"}
          </Button>

          <p className="text-center text-xs text-muted-foreground">
            还没有账号？
            <Link href="/register" className="text-indigo-600 hover:text-indigo-700 font-medium ml-1">
              立即注册
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
