"use client";

import { useState, useEffect, useRef } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "../../../components/ui/button";
import { Card, CardContent } from "../../../components/ui/card";
import { Progress } from "../../../components/ui/progress";
import { ScrollArea } from "../../../components/ui/scroll-area";
import { Badge } from "../../../components/ui/badge";
import { submitAnswer, getReport } from "../../../lib/api";
import type { Question } from "../../../lib/api";
import { Send, Loader2, Brain } from "lucide-react";

const diffBg: Record<string, string> = {
  easy: "bg-green-100 text-green-700",
  medium: "bg-yellow-100 text-yellow-700",
  hard: "bg-red-100 text-red-700",
};

const roundLabels = ["广度扫描", "深度追问", "综合题"];

export default function InterviewPage() {
  const params = useParams();
  const router = useRouter();
  const sessionId = params.sessionId as string;

  const [questions, setQuestions] = useState<Question[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [currentAnswer, setCurrentAnswer] = useState("");
  const [loading, setLoading] = useState(false);
  const [answered, setAnswered] = useState<{ qId: number; answer: string }[]>([]);
  const [generating, setGenerating] = useState(false);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    const stored = sessionStorage.getItem(`questions_${sessionId}`);
    if (stored) setQuestions(JSON.parse(stored));
  }, [sessionId]);

  useEffect(() => {
    if (questions.length > 0) {
      sessionStorage.setItem(`questions_${sessionId}`, JSON.stringify(questions));
    }
  }, [questions, sessionId]);

  useEffect(() => {
    if (!loading && inputRef.current) inputRef.current.focus();
  }, [currentIndex, loading]);

  const currentQ = questions[currentIndex];
  const progress = questions.length ? (currentIndex / questions.length) * 100 : 0;
  const roundIdx = currentIndex < 5 ? 0 : currentIndex < 8 ? 1 : 2;

  const handleSubmit = async () => {
    if (!currentAnswer.trim() || !currentQ) return;
    setAnswered(prev => [...prev, { qId: currentQ.id, answer: currentAnswer }]);
    setLoading(true);
    try {
      const res = await submitAnswer(sessionId, currentQ.id, currentAnswer);
      setCurrentAnswer("");
      if (res.done) {
        setGenerating(true);
        await getReport(sessionId);
        router.push(`/report/${sessionId}`);
      } else {
        setCurrentIndex(i => i + 1);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex flex-col gap-4 py-4">
      <div className="flex items-center gap-3">
        <div className="p-2 rounded-xl bg-gradient-to-br from-sky-500 to-blue-600 shadow-sm">
          <Brain className="w-5 h-5 text-white" />
        </div>
        <div className="flex-1">
          <h1 className="text-lg font-semibold text-slate-800">面试诊断</h1>
          <p className="text-xs text-slate-500">{currentIndex + 1}/{questions.length} 题 · {roundLabels[roundIdx] || ""}</p>
        </div>
        <Badge variant="secondary" className="text-xs">{currentIndex + 1}/{questions.length}</Badge>
      </div>
      <Progress value={progress} className="h-1.5" />

      <ScrollArea className="h-[calc(100vh-320px)]">
        <div className="space-y-4">
          {answered.map((item, i) => (
            <div key={i} className="space-y-2">
              <div className="flex items-start gap-3">
                <div className="w-8 h-8 rounded-full bg-sky-100 flex items-center justify-center shrink-0 mt-0.5">
                  <Brain className="w-4 h-4 text-sky-600" />
                </div>
                <div className="flex-1 bg-white rounded-2xl rounded-tl-sm p-4 shadow-sm border border-slate-100">
                  <p className="text-sm text-slate-800">{questions[i]?.content}</p>
                  <div className="flex gap-2 mt-2">
                    <Badge className={`text-xs ${diffBg[questions[i]?.difficulty] || ""}`}>{questions[i]?.difficulty}</Badge>
                    <Badge variant="outline" className="text-xs text-slate-500">{questions[i]?.category}</Badge>
                  </div>
                </div>
              </div>
              <div className="flex items-start gap-3 ml-11">
                <div className="w-7 h-7 rounded-full bg-blue-600 flex items-center justify-center shrink-0 mt-0.5">
                  <span className="text-xs text-white font-medium">你</span>
                </div>
                <div className="flex-1 bg-blue-50 rounded-2xl rounded-tl-sm p-3">
                  <p className="text-sm text-slate-700">{item.answer}</p>
                </div>
              </div>
            </div>
          ))}

          {currentQ && (
            <div className="flex items-start gap-3">
              <div className="w-8 h-8 rounded-full bg-sky-100 flex items-center justify-center shrink-0 mt-0.5">
                <Brain className="w-4 h-4 text-sky-600" />
              </div>
              <div className="flex-1">
                <div className="bg-white rounded-2xl rounded-tl-sm p-4 shadow-sm border border-slate-100">
                  <p className="text-sm text-slate-800">{currentQ.content}</p>
                  <div className="flex gap-2 mt-2">
                    <Badge className={`text-xs ${diffBg[currentQ.difficulty] || ""}`}>{currentQ.difficulty}</Badge>
                    <Badge variant="outline" className="text-xs text-slate-500">{currentQ.category}</Badge>
                  </div>
                </div>
                <div className="mt-3 flex gap-2">
                  <textarea
                    ref={inputRef}
                    value={currentAnswer}
                    onChange={(e) => setCurrentAnswer(e.target.value)}
                    onKeyDown={(e) => { if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); handleSubmit(); }}}
                    placeholder="输入你的回答..."
                    className="flex-1 min-h-[60px] max-h-[120px] rounded-xl border border-slate-200 bg-white p-3 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500 resize-none"
                    disabled={loading}
                  />
                  <Button onClick={handleSubmit} disabled={!currentAnswer.trim() || loading} className="self-end h-[60px] bg-gradient-to-r from-sky-500 to-blue-600">
                    {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : <Send className="w-4 h-4" />}
                  </Button>
                </div>
              </div>
            </div>
          )}
        </div>
      </ScrollArea>

      {generating && (
        <div className="fixed inset-0 bg-white/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-white rounded-2xl p-8 text-center shadow-xl">
            <Loader2 className="w-10 h-10 animate-spin text-sky-600 mx-auto mb-4" />
            <p className="text-slate-700 font-medium">正在生成诊断报告...</p>
            <p className="text-xs text-slate-400 mt-2">AI 正在分析你的回答，请稍候</p>
          </div>
        </div>
      )}
    </div>
  );
}
