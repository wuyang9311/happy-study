"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "../../components/ui/card";
import { register, saveToken } from "../../lib/api";
import { UserPlus, Loader2, ArrowLeft, Eye, EyeOff, CheckCircle2, XCircle } from "lucide-react";
import Link from "next/link";

export default function RegisterPage() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [nickname, setNickname] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [showPassword, setShowPassword] = useState(false);

  const passChecks = [
    { label: "至少 6 位", check: (p: string) => p.length >= 6 },
    { label: "包含字母", check: (p: string) => /[a-zA-Z]/.test(p) },
    { label: "包含数字", check: (p: string) => /[0-9]/.test(p) },
  ];

  const handleRegister = async () => {
    if (!username.trim() || !password.trim()) return;
    setLoading(true);
    setError("");
    try {
      const data = await register(username.trim(), password, nickname.trim() || undefined);
      saveToken(data.token);
      router.push("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "注册失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[70vh] gap-6">
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
          <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center mx-auto mb-3 shadow-sm">
            <UserPlus className="w-5.5 h-5.5 text-white" />
          </div>
          <CardTitle className="text-lg font-semibold">创建账号</CardTitle>
          <CardDescription className="text-xs text-muted-foreground">
            开启 AI 驱动的个性化学习之旅
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 用户名 */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-foreground/80">用户名</label>
            <Input
              placeholder="3-32位，字母或中文开头"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleRegister()}
              className="h-11 text-sm bg-background border-border/80 focus-visible:ring-indigo-400/40"
              disabled={loading}
              autoFocus
            />
          </div>

          {/* 昵称 */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-foreground/80">昵称（可选）</label>
            <Input
              placeholder="给自己取个名字"
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              className="h-11 text-sm bg-background border-border/80 focus-visible:ring-indigo-400/40"
              disabled={loading}
            />
          </div>

          {/* 密码 */}
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-foreground/80">密码</label>
            <div className="relative">
              <Input
                type={showPassword ? "text" : "password"}
                placeholder="至少包含字母和数字"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleRegister()}
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

            {/* 密码强度指示 */}
            <div className="flex gap-2 mt-1.5">
              {passChecks.map((c, i) => {
                const passed = c.check(password);
                return (
                  <span
                    key={i}
                    className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] border transition-colors ${
                      password === ""
                        ? "text-muted-foreground/50 border-border/50"
                        : passed
                        ? "text-emerald-600 border-emerald-200/60 bg-emerald-50"
                        : "text-muted-foreground border-border/60"
                    }`}
                  >
                    {password === "" ? (
                      <span className="w-3" />
                    ) : passed ? (
                      <CheckCircle2 className="w-3 h-3" />
                    ) : (
                      <XCircle className="w-3 h-3" />
                    )}
                    {c.label}
                  </span>
                );
              })}
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
            onClick={handleRegister}
            disabled={!username.trim() || !password.trim() || loading}
            className="w-full h-11 bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white shadow-sm disabled:opacity-50"
          >
            {loading ? (
              <span className="flex items-center gap-2"><Loader2 className="w-4 h-4 animate-spin" />注册中...</span>
            ) : "注册"}
          </Button>

          <p className="text-center text-xs text-muted-foreground">
            已有账号？
            <Link href="/login" className="text-indigo-600 hover:text-indigo-700 font-medium ml-1">
              立即登录
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
