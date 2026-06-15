"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "../components/ui/card";
import { startDiagnosis, getToken, removeToken, getProfile } from "../lib/api";
import { GraduationCap, Sparkles, Brain, BookOpen, ArrowRight, Loader2 } from "lucide-react";

export default function Home() {
  const [topic, setTopic] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [checkingAuth, setCheckingAuth] = useState(true);
  const router = useRouter();

  // 检查登录状态
  useEffect(() => {
    const token = getToken();
    if (!token) {
      setCheckingAuth(false);
      return;
    }
    // 验证 token 有效性
    getProfile().then(() => setCheckingAuth(false)).catch(() => {
      removeToken();
      setCheckingAuth(false);
    });
  }, []);

  const handleStart = async () => {
    if (!topic.trim()) return;
    setLoading(true);
    setError("");
    try {
      const data = await startDiagnosis(topic);
      // startDiagnosis 只返回 session_id，第一题由诊断页 SSE 流式获取
      router.push(`/interview/${data.session_id}`);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "启动失败";
      if (msg.includes("令牌") || msg.includes("401")) {
        router.push("/login");
        return;
      }
      setError(msg);
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col items-center gap-10 py-8">
      {checkingAuth ? (
        <div className="flex items-center justify-center min-h-[40vh]">
          <Loader2 className="w-6 h-6 animate-spin text-indigo-500" />
        </div>
      ) : (
        <>
      {/* Hero Section */}
      <div className="text-center max-w-xl">
        <div className="inline-flex items-center gap-2 px-3 py-1 mb-4 rounded-full bg-secondary text-xs font-medium text-muted-foreground/80 border border-border/60">
          <Sparkles className="w-3 h-3 text-indigo-500" />
          AI 驱动的自适应学习引擎
        </div>
        <h1 className="text-3xl sm:text-4xl font-bold tracking-tight text-foreground leading-[1.1]">
          先考后学，
          <br />
          <span className="bg-gradient-to-r from-sky-600 via-indigo-600 to-purple-600 bg-clip-text text-transparent">
            只学不会的
          </span>
        </h1>
        <p className="mt-4 text-sm text-muted-foreground leading-relaxed max-w-md mx-auto">
          AI 面试官三遍扫描你的知识盲区，生成个性化课程方案。
          掌握的直接跳过，不会的精准补齐。
        </p>
      </div>

      {/* Feature Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3 w-full max-w-2xl">
        {[
          { icon: Brain, title: "诊断面试", desc: "AI 面试官三遍扫描，精准摸清你的知识底子", color: "from-sky-500 to-blue-600" },
          { icon: Sparkles, title: "定制课程", desc: "只教你不会的，已掌握的直接跳过", color: "from-indigo-500 to-purple-600" },
          { icon: BookOpen, title: "面试对标", desc: "能力可视化，对标大厂职级要求", color: "from-purple-500 to-pink-600" },
        ].map((item, i) => (
          <Card key={i} className="border border-border/60 bg-white shadow-sm hover:shadow-md hover:-translate-y-0.5 transition-all duration-200">
            <CardHeader className="pb-2">
              <div className={`w-9 h-9 rounded-xl bg-gradient-to-br ${item.color} flex items-center justify-center shadow-sm mb-2`}>
                <item.icon className="w-4.5 h-4.5 text-white" />
              </div>
              <CardTitle className="text-sm font-semibold">{item.title}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-xs text-muted-foreground leading-relaxed">{item.desc}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Input Section */}
      <Card className="w-full max-w-xl border border-border/60 bg-white shadow-sm">
        <CardHeader>
          <CardTitle className="text-lg font-semibold">你想学什么？</CardTitle>
          <CardDescription className="text-xs text-muted-foreground">
            输入你想学习的技术方向，AI 将为你定制专属学习方案
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex gap-2">
            <Input
              placeholder="例如：Go 并发编程、系统设计、React 源码"
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleStart()}
              className="h-12 text-sm bg-background border-border/80 focus-visible:ring-indigo-400/40"
              disabled={loading}
            />
            <Button
              onClick={handleStart}
              disabled={!topic.trim() || loading}
              className="h-12 px-6 bg-gradient-to-r from-sky-500 to-indigo-600 hover:from-sky-600 hover:to-indigo-700 text-white shadow-sm shadow-indigo-200/50 disabled:opacity-50"
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  准备中...
                </span>
              ) : (
                <span className="flex items-center gap-2">
                  开始诊断 <ArrowRight className="w-4 h-4" />
                </span>
              )}
            </Button>
          </div>

          {error && (
            <div className="flex items-center gap-2 p-3 rounded-lg bg-red-50 border border-red-200/60">
              <svg className="w-4 h-4 text-red-500 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="12" cy="12" r="10" />
                <line x1="15" y1="9" x2="9" y2="15" />
                <line x1="9" y1="9" x2="15" y2="15" />
              </svg>
              <p className="text-xs text-red-600">{error}</p>
            </div>
          )}

          <div className="flex flex-wrap gap-2">
            <span className="text-xs text-muted-foreground self-center mr-1">热门话题</span>
            {["Go 并发编程", "系统设计", "React 源码", "Redis 原理", "Kubernetes"].map((t) => (
              <button
                key={t}
                onClick={() => setTopic(t)}
                className="px-3 py-1.5 text-xs rounded-full bg-secondary text-muted-foreground hover:bg-indigo-50 hover:text-indigo-600 hover:border-indigo-200 border border-transparent transition-all"
              >
                {t}
              </button>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Trust indicators */}
      <div className="flex items-center gap-6 text-xs text-muted-foreground/50">
        <span className="flex items-center gap-1.5">
          <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          </svg>
          隐私安全
        </span>
        <span className="flex items-center gap-1.5">
          <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="10" />
            <path d="M12 6v6l4 2" />
          </svg>
          实时反馈
        </span>
        <span className="flex items-center gap-1.5">
          <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M9 12l2 2 4-4" />
            <circle cx="12" cy="12" r="10" />
          </svg>
          自适应学习
        </span>
      </div>
        </>
      )}
    </div>
  );
}
