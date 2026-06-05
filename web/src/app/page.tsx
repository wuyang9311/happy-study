"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { startDiagnosis } from "../lib/api";
import { GraduationCap, Sparkles, ArrowRight, Brain, BookOpen } from "lucide-react";

export default function Home() {
  const [topic, setTopic] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const router = useRouter();

  const handleStart = async () => {
    if (!topic.trim()) return;
    setLoading(true);
    setError("");
    try {
      const data = await startDiagnosis(topic);
      sessionStorage.setItem(`questions_${data.session_id}`, JSON.stringify([data.question]));
      router.push(`/interview/${data.session_id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "启动失败，请检查后端服务是否运行");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[80vh] gap-8">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-3 rounded-2xl bg-gradient-to-br from-sky-500 to-blue-600 shadow-lg">
          <GraduationCap className="w-8 h-8 text-white" />
        </div>
        <div>
          <h1 className="text-3xl font-bold bg-gradient-to-r from-sky-600 to-blue-700 bg-clip-text text-transparent">Happy Study</h1>
          <p className="text-slate-500 text-sm">AI 驱动 · 先考后学</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 w-full max-w-2xl mb-4">
        <Card className="border-0 shadow-sm bg-white/70 backdrop-blur">
          <CardHeader className="pb-2">
            <Brain className="w-5 h-5 text-sky-600 mb-1" />
            <CardTitle className="text-sm font-medium">诊断面试</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-slate-500">AI 面试官三遍扫描，精准摸清你的知识底子</p>
          </CardContent>
        </Card>
        <Card className="border-0 shadow-sm bg-white/70 backdrop-blur">
          <CardHeader className="pb-2">
            <Sparkles className="w-5 h-5 text-sky-600 mb-1" />
            <CardTitle className="text-sm font-medium">定制课程</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-slate-500">只教你不会的，已掌握的直接跳过</p>
          </CardContent>
        </Card>
        <Card className="border-0 shadow-sm bg-white/70 backdrop-blur">
          <CardHeader className="pb-2">
            <BookOpen className="w-5 h-5 text-sky-600 mb-1" />
            <CardTitle className="text-sm font-medium">面试对标</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-xs text-slate-500">能力可视化，对标大厂职级要求</p>
          </CardContent>
        </Card>
      </div>

      <Card className="w-full max-w-xl border-0 shadow-lg bg-white/90 backdrop-blur">
        <CardHeader>
          <CardTitle className="text-xl">你想学什么？</CardTitle>
          <CardDescription>
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
              className="h-12 text-base"
              disabled={loading}
            />
            <Button
              onClick={handleStart}
              disabled={!topic.trim() || loading}
              className="h-12 px-6 bg-gradient-to-r from-sky-500 to-blue-600 hover:from-sky-600 hover:to-blue-700"
            >
              {loading ? (
                <span className="flex items-center gap-2">
                  <span className="animate-spin w-4 h-4 border-2 border-white border-t-transparent rounded-full" />
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
            <p className="text-red-500 text-sm bg-red-50 p-3 rounded-lg">{error}</p>
          )}
          <div className="flex flex-wrap gap-2 pt-2">
            {["Go 并发编程", "系统设计", "React 源码", "Redis 原理", "Kubernetes"].map((t) => (
              <button
                key={t}
                onClick={() => setTopic(t)}
                className="px-3 py-1.5 text-xs rounded-full bg-slate-100 hover:bg-sky-100 hover:text-sky-700 text-slate-600 transition-colors"
              >
                {t}
              </button>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
