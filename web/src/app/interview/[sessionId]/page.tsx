"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "../../../components/ui/button";
import { Card, CardContent } from "../../../components/ui/card";
import { ScrollArea } from "../../../components/ui/scroll-area";
import { Badge } from "../../../components/ui/badge";
import { submitAnswer, stopDiagnosis } from "../../../lib/api";
import type { AdaptiveQuestion } from "../../../lib/api";
import { Send, Loader2, Brain, User, StopCircle, ChevronLeft, AlertCircle } from "lucide-react";

const diffBadge: Record<string, string> = {
  easy: "bg-green-50 text-green-700 border-green-200/60",
  medium: "bg-amber-50 text-amber-700 border-amber-200/60",
  hard: "bg-red-50 text-red-700 border-red-200/60",
};
const diffDot: Record<string, string> = { easy: "bg-green-500", medium: "bg-amber-500", hard: "bg-red-500" };
const diffLabel: Record<string, string> = { easy: "简单", medium: "中等", hard: "困难" };

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

interface Message {
  type: "question" | "answer" | "system";
  content: string;
  streaming?: boolean;
  category?: string;
  difficulty?: string;
}

function getToken(): string | null {
  if (typeof window !== 'undefined') return localStorage.getItem('happy_study_token');
  return null;
}

export default function InterviewPage() {
  const params = useParams();
  const router = useRouter();
  const sessionId = params.sessionId as string;

  const [messages, setMessages] = useState<Message[]>([]);
  const [currentQuestion, setCurrentQuestion] = useState<AdaptiveQuestion | null>(null);
  const [currentAnswer, setCurrentAnswer] = useState("");
  const [loading, setLoading] = useState(false);
  const [stopping, setStopping] = useState(false);
  const [diagnosisDone, setDiagnosisDone] = useState(false);
  const [streamError, setStreamError] = useState<string | null>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    const token = getToken();
    if (!token) { router.push("/login"); return; }
    fetchFirstQuestion(token);
    return () => {
      // 组件卸载时中止请求
      if (abortRef.current) abortRef.current.abort();
    };
  }, [sessionId]);

  const fetchFirstQuestion = async (token: string) => {
    setStreamError(null);

    // 先显示 loading 状态
    const streamMsg: Message = { type: "question", content: "", streaming: true };
    setMessages([streamMsg]);

    const abortController = new AbortController();
    abortRef.current = abortController;

    // 30 秒超时保护
    const timeoutId = setTimeout(() => {
      abortController.abort();
      setStreamError("连接超时，请刷新页面重试");
      setMessages(prev => {
        const next = [...prev];
        const last = next[next.length - 1];
        if (last?.streaming) next[next.length - 1] = { ...last, streaming: false, content: last.content || "（加载超时）" };
        return next;
      });
    }, 30000);

    try {
      const url = API_BASE + "/diagnosis/question/" + sessionId;
      const res = await fetch(url, {
        headers: { 'Authorization': 'Bearer ' + token },
        signal: abortController.signal,
      });

      if (!res.ok) {
        clearTimeout(timeoutId);
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || `HTTP ${res.status}`);
      }

      const contentType = res.headers.get("content-type") || "";
      // 如果返回 JSON（已有缓存题目），直接解析
      if (contentType.includes("application/json") || contentType.includes("text/json")) {
        clearTimeout(timeoutId);
        const data = await res.json();
        const q = data?.question;
        if (q) {
          const parsed: AdaptiveQuestion = {
            id: q.id || q.ID || 1,
            content: q.content || q.Content || "",
            category: q.category || q.Category || "",
            difficulty: q.difficulty || q.Difficulty || "medium",
          };
          setCurrentQuestion(parsed);
          setMessages([{
            type: "question",
            content: parsed.content,
            category: parsed.category,
            difficulty: parsed.difficulty,
          }]);
        }
        return;
      }

      // SSE 流式解析
      const reader = res.body?.getReader();
      if (!reader) throw new Error("无法读取响应流");

      const decoder = new TextDecoder();
      let buffer = "";
      let fullContent = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() || "";

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            const data = line.slice(6);
            // 检查是否是完成事件 JSON
            if (data.startsWith("{") && (data.includes('"action"') || data.includes('"question"'))) {
              try {
                const parsed = JSON.parse(data);
                if (parsed.action === "done" && parsed.question) {
                  const q = parsed.question;
                  const aq: AdaptiveQuestion = {
                    id: q.id || q.ID || 1,
                    content: q.content || q.Content || "",
                    category: q.category || q.Category || "",
                    difficulty: q.difficulty || q.Difficulty || "medium",
                  };
                  setCurrentQuestion(aq);
                  setMessages(prev => {
                    const next = [...prev];
                    const last = next[next.length - 1];
                    if (last?.streaming) {
                      next[next.length - 1] = {
                        type: "question",
                        content: aq.content,
                        category: aq.category,
                        difficulty: aq.difficulty,
                      };
                    }
                    return next;
                  });
                  continue;
                }
              } catch { /* not a complete JSON yet */ }
            }
            // 普通 token：追加到流式消息
            fullContent += data;
            setMessages(prev => {
              const next = [...prev];
              const last = next[next.length - 1];
              if (last?.streaming) next[next.length - 1] = { ...last, content: fullContent };
              return next;
            });
          }
        }
      }
    } catch (err: any) {
      clearTimeout(timeoutId);
      if (err?.name === "AbortError") return; // 超时已处理或主动取消
      console.error("Stream error:", err);
      setStreamError(err?.message || "加载题目失败，请刷新页面重试");
      setMessages(prev => [...prev, { type: "system", content: "加载题目失败，请刷新页面重试" }]);
    } finally {
      clearTimeout(timeoutId);
    }
  };

  useEffect(() => {
    if (!loading && inputRef.current) inputRef.current.focus();
  }, [currentQuestion, loading]);

  useEffect(() => {
    if (scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
  }, [messages]);

  const handleSubmit = async () => {
    if (!currentAnswer.trim() || !currentQuestion) return;
    const userAnswer = currentAnswer.trim();
    setCurrentAnswer("");

    setMessages(prev => [...prev, { type: "answer", content: userAnswer }]);
    setCurrentQuestion(null);
    setLoading(true);

    try {
      const res = await submitAnswer(sessionId, userAnswer);
      if (res.done) {
        setDiagnosisDone(true);
        setMessages(prev => [...prev, { type: "system", content: res.message || "诊断完成！" }]);
        setTimeout(() => router.push("/report/" + sessionId), 1500);
      } else if (res.question) {
        const q = res.question;
        setCurrentQuestion(q);
        setMessages(prev => [...prev, { type: "question", content: q.content, category: q.category, difficulty: q.difficulty }]);
      }
    } catch (err) {
      console.error(err);
      setMessages(prev => [...prev, { type: "system", content: "提交失败，请重试" }]);
      if (currentQuestion) setCurrentQuestion(currentQuestion);
    } finally {
      setLoading(false);
    }
  };

  const handleStop = async () => {
    if (stopping) return;
    setStopping(true);
    try {
      const res = await stopDiagnosis(sessionId);
      setDiagnosisDone(true);
      setMessages(prev => [...prev, { type: "system", content: res.message || "已停止诊断" }]);
      setTimeout(() => router.push("/report/" + sessionId), 1500);
    } catch (err) {
      console.error(err);
      setStopping(false);
    }
  };

  const badgeCls = (d: string) => {
    return "inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium border " + (diffBadge[d] || "bg-slate-50 text-slate-600 border-slate-200/60");
  };
  const dotCls = (d: string) => {
    return "w-1.5 h-1.5 rounded-full " + (diffDot[d] || "bg-slate-400");
  };

  return (
    <div className="flex flex-col gap-4 pb-8">
      <div className="flex items-center gap-3">
        <button onClick={() => router.push("/")}
          className="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center hover:bg-border transition-colors">
          <ChevronLeft className="w-4 h-4 text-muted-foreground" />
        </button>
        <div className="flex items-center gap-2.5 flex-1 min-w-0">
          <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-sky-500 to-indigo-600 flex items-center justify-center shadow-sm shrink-0">
            <Brain className="w-4 h-4 text-white" />
          </div>
          <div className="min-w-0">
            <h1 className="text-sm font-semibold text-foreground">自适应诊断</h1>
            <p className="text-xs text-muted-foreground">AI 根据你的回答动态出题</p>
          </div>
        </div>
        <Button onClick={handleStop} disabled={stopping || diagnosisDone}
          variant="outline"
          className="text-xs border-red-200/60 text-red-600 hover:bg-red-50 hover:border-red-300 rounded-xl shrink-0 gap-1.5">
          {stopping ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : <StopCircle className="w-3.5 h-3.5" />}
          停止诊断
        </Button>
      </div>

      <Card className="border border-border/60 bg-white shadow-sm overflow-hidden">
        <ScrollArea ref={scrollRef} className="h-[calc(100dvh-240px)] p-4 md:p-6">
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-sky-100 to-indigo-100 flex items-center justify-center shrink-0 mt-0.5 shadow-sm">
                <Brain className="w-4 h-4 text-indigo-600" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="bg-gradient-to-r from-indigo-50/80 to-purple-50/80 rounded-xl rounded-tl-sm p-4 border border-indigo-100/60">
                  <p className="text-sm text-foreground leading-relaxed">欢迎！我来帮你诊断技术水平。请尽量详细回答每个问题。</p>
                  <p className="text-xs text-muted-foreground mt-2">不会的可以直接说"不太了解"，我会调整问题方向</p>
                </div>
              </div>
            </div>

            {messages.map((msg, i) => (
              <div key={i}>
                {msg.type === "question" && (
                  <div className="flex items-start gap-3">
                    <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-sky-100 to-indigo-100 flex items-center justify-center shrink-0 mt-0.5 shadow-sm">
                      <Brain className="w-4 h-4 text-indigo-600" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className={`bg-background rounded-xl rounded-tl-sm p-4 border ${msg.streaming ? 'border-indigo-300/60' : 'border-border/60'}`}>
                        {msg.streaming && !msg.content ? (
                          <div className="flex items-center gap-2">
                            <span className="inline-flex h-3 w-3">
                              <span className="animate-ping absolute h-3 w-3 rounded-full bg-indigo-400 opacity-75" />
                              <span className="relative rounded-full h-3 w-3 bg-indigo-500" />
                            </span>
                            <span className="text-xs text-muted-foreground">AI 正在思考题目...</span>
                          </div>
                        ) : (
                          <p className="text-sm text-foreground leading-relaxed whitespace-pre-wrap">
                            {msg.content}
                            {msg.streaming && <span className="inline-block w-2 h-4 bg-indigo-500 animate-pulse ml-0.5" />}
                          </p>
                        )}
                        {!msg.streaming && msg.difficulty && (
                          <div className="flex gap-2 mt-3">
                            <span className={badgeCls(msg.difficulty)}>
                              <span className={dotCls(msg.difficulty)} />
                              {diffLabel[msg.difficulty] || msg.difficulty}
                            </span>
                            {msg.category && (
                              <span className="px-2 py-0.5 rounded-full text-xs text-muted-foreground bg-secondary border border-border/60">
                                {msg.category}
                              </span>
                            )}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                )}
                {msg.type === "answer" && (
                  <div className="flex items-start gap-3 ml-10">
                    <div className="w-7 h-7 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shrink-0 mt-0.5 shadow-sm">
                      <User className="w-3.5 h-3.5 text-white" />
                    </div>
                    <div className="flex-1">
                      <div className="bg-indigo-50/80 rounded-xl rounded-tr-sm p-3.5 border border-indigo-100/60">
                        <p className="text-sm text-indigo-900/80 leading-relaxed whitespace-pre-wrap">{msg.content}</p>
                      </div>
                    </div>
                  </div>
                )}
                {msg.type === "system" && (
                  <div className="flex justify-center">
                    <div className="px-4 py-2 rounded-full bg-secondary border border-border/60">
                      <p className="text-xs text-muted-foreground">{msg.content}</p>
                    </div>
                  </div>
                )}
              </div>
            ))}

            {loading && (
              <div className="flex items-center gap-2 text-xs text-muted-foreground ml-11">
                <Loader2 className="w-3.5 h-3.5 animate-spin text-indigo-500" />
                <span>AI 正在分析你的回答...</span>
              </div>
            )}

            {!loading && currentQuestion && !diagnosisDone && !streamError && (
              <div className="flex gap-2 pt-2">
                <textarea
                  ref={inputRef}
                  value={currentAnswer}
                  onChange={e => setCurrentAnswer(e.target.value)}
                  onKeyDown={e => { if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); handleSubmit(); }}}
                  placeholder="输入你的回答... (Enter 发送，Shift+Enter 换行)"
                  className="flex-1 min-h-[56px] max-h-[120px] rounded-xl border border-border/80 bg-background p-3 text-sm text-foreground placeholder:text-muted-foreground/50 focus:outline-none focus:ring-2 focus:ring-indigo-400/30 focus:border-indigo-300 resize-none transition-all"
                />
                <Button onClick={handleSubmit}
                  disabled={!currentAnswer.trim() || loading}
                  className="h-[56px] w-[56px] shrink-0 bg-gradient-to-br from-sky-500 to-indigo-600 hover:from-sky-600 hover:to-indigo-700 text-white shadow-sm disabled:opacity-40 rounded-xl">
                  {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
                </Button>
              </div>
            )}

            {!currentQuestion && !diagnosisDone && !streamError && (
              <div className="flex items-center justify-center gap-2 py-4 text-xs text-muted-foreground">
                <Loader2 className="w-3.5 h-3.5 animate-spin text-indigo-400" />
                <span>正在生成第一道题...</span>
              </div>
            )}

            {streamError && (
              <div className="flex flex-col items-center gap-3 py-6 px-4 rounded-xl bg-red-50 border border-red-200/60">
                <div className="flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 text-red-500" />
                  <p className="text-xs text-red-600">{streamError}</p>
                </div>
                <Button onClick={() => fetchFirstQuestion(getToken() || "")}
                  variant="outline"
                  className="text-xs border-red-200/60 text-red-600 hover:bg-red-50 rounded-xl">
                  重新加载
                </Button>
              </div>
            )}
          </div>
        </ScrollArea>
      </Card>

      {diagnosisDone && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-white rounded-2xl p-8 text-center shadow-xl border border-border/60 max-w-sm mx-4">
            <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-sky-100 to-indigo-100 flex items-center justify-center mx-auto mb-5 shadow-sm">
              <Loader2 className="w-8 h-8 animate-spin text-indigo-600" />
            </div>
            <p className="text-foreground font-semibold mb-1">诊断已完成</p>
            <p className="text-xs text-muted-foreground">正在生成学习报告... 即将跳转</p>
          </div>
        </div>
      )}
    </div>
  );
}
