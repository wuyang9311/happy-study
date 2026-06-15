"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "../../../components/ui/button";
import { Card, CardContent } from "../../../components/ui/card";
import { Badge } from "../../../components/ui/badge";
import { getReport, generateCurriculum } from "../../../lib/api";
import type { DiagnosisReport } from "../../../lib/api";
import {
  ClipboardList, TrendingUp, AlertTriangle, Target, ArrowRight,
  Loader2, CheckCircle2, ChevronLeft, Zap, BarChart3
} from "lucide-react";

const levelCfg: Record<string, { bar: string; label: string; color: string }> = {
  mastered: { bar: "bg-emerald-500", label: "掌握", color: "text-emerald-700 bg-emerald-50 border-emerald-200/60" },
  familiar: { bar: "bg-blue-500", label: "熟悉", color: "text-blue-700 bg-blue-50 border-blue-200/60" },
  weak: { bar: "bg-amber-500", label: "薄弱", color: "text-amber-700 bg-amber-50 border-amber-200/60" },
  unknown: { bar: "bg-rose-500", label: "未知", color: "text-rose-700 bg-rose-50 border-rose-200/60" },
};

function ScoreGauge({ score }: { score: number }) {
  const radius = 54;
  const circumference = 2 * Math.PI * radius;
  const progress = circumference * (1 - score / 100);
  const color = score >= 80 ? "#10b981" : score >= 60 ? "#3b82f6" : score >= 40 ? "#f59e0b" : "#f43f5e";

  return (
    <svg width="140" height="140" className="shrink-0">
      <circle cx="70" cy="70" r={radius} fill="none" stroke="#f0efec" strokeWidth="10" />
      <circle
        cx="70" cy="70" r={radius}
        fill="none" stroke={color} strokeWidth="10"
        strokeLinecap="round"
        strokeDasharray={circumference}
        strokeDashoffset={progress}
        transform="rotate(-90, 70, 70)"
        className="transition-all duration-1000 ease-out"
      />
      <text x="70" y="60" textAnchor="middle" className="text-2xl font-bold" fill="#1a1a2e">
        {Math.round(score)}
      </text>
      <text x="70" y="82" textAnchor="middle" className="text-[10px]" fill="#787884">
        综合掌握度
      </text>
    </svg>
  );
}

export default function ReportPage() {
  const params = useParams();
  const router = useRouter();
  const sessionId = params.sessionId as string;
  const [report, setReport] = useState<DiagnosisReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);

  useEffect(() => {
    getReport(sessionId).then(d => setReport(d.report)).catch(console.error).finally(() => setLoading(false));
  }, [sessionId]);

  const handleGenerate = async () => {
    setGenerating(true);
    try {
      await generateCurriculum(sessionId);
      router.push(`/curriculum/${sessionId}`);
    } catch (err) {
      console.error(err);
    }
  };

  if (loading) return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <Loader2 className="w-8 h-8 animate-spin text-indigo-500" />
    </div>
  );
  if (!report) return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] gap-3">
      <AlertTriangle className="w-10 h-10 text-muted-foreground/40" />
      <p className="text-sm text-muted-foreground">报告加载失败</p>
      <Button variant="outline" size="sm" onClick={() => router.push("/")} className="text-xs">返回首页</Button>
    </div>
  );

  return (
    <div className="flex flex-col gap-5 pb-8">
      {/* Header */}
      <div className="flex items-center gap-3">
        <button
          onClick={() => router.back()}
          className="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center hover:bg-border transition-colors"
        >
          <ChevronLeft className="w-4 h-4 text-muted-foreground" />
        </button>
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center shadow-sm">
            <ClipboardList className="w-4 h-4 text-white" />
          </div>
          <div>
            <h1 className="text-sm font-semibold text-foreground">诊断报告</h1>
            <p className="text-xs text-muted-foreground">{report.topic}</p>
          </div>
        </div>
      </div>

      {/* Score gauge + target */}
      <Card className="border border-border/60 bg-white shadow-sm">
        <CardContent className="p-5 md:p-6">
          <div className="flex flex-col sm:flex-row items-center gap-5">
            <ScoreGauge score={report.overall_score} />
            <div className="flex-1 space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div className="p-3 rounded-xl bg-secondary/60 border border-border/40">
                  <p className="text-xs text-muted-foreground mb-0.5">目标水平</p>
                  <p className="text-sm font-semibold text-foreground">{report.target_level}</p>
                </div>
                <div className="p-3 rounded-xl bg-secondary/60 border border-border/40">
                  <p className="text-xs text-muted-foreground mb-0.5">预估周期</p>
                  <p className="text-sm font-semibold text-foreground">{report.estimated_weeks} 周</p>
                </div>
              </div>
              <div className="h-2 bg-secondary rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-emerald-400 via-blue-500 to-indigo-600 rounded-full transition-all duration-1000"
                  style={{ width: `${report.overall_score}%` }}
                />
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Knowledge breakdown */}
      <Card className="border border-border/60 bg-white shadow-sm">
        <CardContent className="p-5 md:p-6">
          <h3 className="text-sm font-semibold flex items-center gap-2 text-foreground mb-4">
            <BarChart3 className="w-4 h-4 text-indigo-500" />
            知识掌握情况
          </h3>
          <div className="space-y-3.5">
            {report.scores.map((s, i) => {
              const cfg = levelCfg[s.level] || levelCfg.unknown;
              return (
                <div key={i} className="space-y-1.5">
                  <div className="flex items-center justify-between">
                    <span className="text-sm text-foreground font-medium">{s.category}</span>
                    <div className="flex items-center gap-2.5">
                      <Badge variant="outline" className={`text-xs font-medium border ${cfg.color}`}>
                        {cfg.label}
                      </Badge>
                      <span className="text-xs text-muted-foreground tabular-nums w-8 text-right">{Math.round(s.score)}</span>
                    </div>
                  </div>
                  <div className="h-2 bg-secondary rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full transition-all duration-1000 ${cfg.bar}`}
                      style={{ width: `${s.score}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* Strengths & Weaknesses */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <Card className="border border-emerald-200/50 bg-white shadow-sm">
          <div className="p-4 md:p-5">
            <h3 className="text-sm font-semibold flex items-center gap-2 text-emerald-700 mb-3">
              <CheckCircle2 className="w-4 h-4" /> 优势
            </h3>
            <ul className="space-y-2">
              {report.strengths.map((s, i) => (
                <li key={i} className="flex gap-2.5 text-sm text-muted-foreground leading-relaxed">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 shrink-0 mt-1.5" />
                  {s}
                </li>
              ))}
            </ul>
          </div>
        </Card>
        <Card className="border border-amber-200/50 bg-white shadow-sm">
          <div className="p-4 md:p-5">
            <h3 className="text-sm font-semibold flex items-center gap-2 text-amber-700 mb-3">
              <AlertTriangle className="w-4 h-4" /> 待加强
            </h3>
            <ul className="space-y-2">
              {report.weaknesses.map((w, i) => (
                <li key={i} className="flex gap-2.5 text-sm text-muted-foreground leading-relaxed">
                  <span className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0 mt-1.5" />
                  {w}
                </li>
              ))}
            </ul>
          </div>
        </Card>
      </div>

      {/* Summary */}
      <Card className="border border-border/60 bg-white shadow-sm">
        <div className="p-4 md:p-5 flex gap-3">
          <Target className="w-5 h-5 text-indigo-500 shrink-0 mt-0.5" />
          <p className="text-sm text-muted-foreground leading-relaxed">{report.summary}</p>
        </div>
      </Card>

      {/* Generate CTA */}
      <div className="flex justify-center pt-2">
        <Button
          onClick={handleGenerate}
          disabled={generating}
          className="h-12 px-8 bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white shadow-lg shadow-indigo-200/50 rounded-xl"
        >
          {generating ? (
            <span className="flex items-center gap-2">
              <Loader2 className="w-4 h-4 animate-spin" />
              生成课程方案...
            </span>
          ) : (
            <span className="flex items-center gap-2">
              <Zap className="w-4 h-4" />
              生成个性化课程
              <ArrowRight className="w-4 h-4" />
            </span>
          )}
        </Button>
      </div>
    </div>
  );
}
