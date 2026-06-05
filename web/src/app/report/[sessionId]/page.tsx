"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "../../../components/ui/button";
import { Card, CardContent } from "../../../components/ui/card";
import { Badge } from "../../../components/ui/badge";
import { Separator } from "../../../components/ui/separator";
import { getReport, generateCurriculum } from "../../../lib/api";
import type { DiagnosisReport } from "../../../lib/api";
import { ClipboardList, TrendingUp, AlertTriangle, Target, ArrowRight, Loader2, CheckCircle2 } from "lucide-react";

const levelCfg: Record<string, { bar: string; label: string }> = {
  mastered: { bar: "bg-green-500", label: "掌握" },
  familiar: { bar: "bg-blue-500", label: "熟悉" },
  weak: { bar: "bg-yellow-500", label: "薄弱" },
  unknown: { bar: "bg-red-500", label: "未知" },
};

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
    <div className="flex items-center justify-center min-h-[80vh]">
      <Loader2 className="w-10 h-10 animate-spin text-sky-600" />
    </div>
  );
  if (!report) return <div className="text-center mt-20 text-slate-500">报告加载失败</div>;

  return (
    <div className="flex flex-col gap-6 py-4">
      <div className="flex items-center gap-3">
        <div className="p-2 rounded-xl bg-gradient-to-br from-sky-500 to-blue-600 shadow-sm">
          <ClipboardList className="w-5 h-5 text-white" />
        </div>
        <div>
          <h1 className="text-lg font-semibold text-slate-800">诊断报告</h1>
          <p className="text-xs text-slate-500">{report.topic}</p>
        </div>
      </div>

      <Card className="border-0 shadow-sm bg-gradient-to-r from-sky-50 to-blue-50">
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-slate-500">综合掌握度</p>
              <p className="text-4xl font-bold text-slate-800">{Math.round(report.overall_score)}<span className="text-lg text-slate-400">/100</span></p>
            </div>
            <div className="text-right">
              <p className="text-sm text-slate-500">目标水平</p>
              <p className="text-lg font-semibold">{report.target_level}</p>
              <p className="text-xs text-slate-400">{report.estimated_weeks} 周</p>
            </div>
          </div>
          <div className="mt-4 h-2 bg-slate-200 rounded-full overflow-hidden">
            <div className="h-full bg-gradient-to-r from-sky-400 to-blue-600 rounded-full transition-all duration-1000" style={{ width: `${report.overall_score}%` }} />
          </div>
        </CardContent>
      </Card>

      <Card className="border-0 shadow-sm">
        <CardContent className="space-y-3 pt-6">
          <h3 className="text-sm font-medium flex items-center gap-2 text-slate-700">
            <TrendingUp className="w-4 h-4 text-sky-600" />
            知识掌握情况
          </h3>
          {report.scores.map((s, i) => {
            const cfg = levelCfg[s.level] || levelCfg.unknown;
            return (
              <div key={i} className="space-y-1">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-slate-700">{s.category}</span>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className={`text-xs ${cfg.bar.replace('bg-', 'text-').replace('-500', '-700')}`}>{cfg.label}</Badge>
                    <span className="text-slate-500 w-8 text-right">{Math.round(s.score)}</span>
                  </div>
                </div>
                <div className="h-1.5 bg-slate-100 rounded-full overflow-hidden">
                  <div className={`h-full rounded-full transition-all duration-1000 ${cfg.bar}`} style={{ width: `${s.score}%` }} />
                </div>
              </div>
            );
          })}
        </CardContent>
      </Card>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Card className="border-0 shadow-sm bg-green-50/50">
          <div className="p-4">
            <h3 className="text-sm font-medium flex items-center gap-2 text-green-700 mb-2">
              <CheckCircle2 className="w-4 h-4" /> 优势
            </h3>
            <ul className="space-y-1">
              {report.strengths.map((s, i) => <li key={i} className="text-sm text-slate-700 flex gap-2"><span className="text-green-500">•</span>{s}</li>)}
            </ul>
          </div>
        </Card>
        <Card className="border-0 shadow-sm bg-red-50/50">
          <div className="p-4">
            <h3 className="text-sm font-medium flex items-center gap-2 text-red-700 mb-2">
              <AlertTriangle className="w-4 h-4" /> 待加强
            </h3>
            <ul className="space-y-1">
              {report.weaknesses.map((w, i) => <li key={i} className="text-sm text-slate-700 flex gap-2"><span className="text-red-500">•</span>{w}</li>)}
            </ul>
          </div>
        </Card>
      </div>

      <Card className="border-0 shadow-sm bg-white">
        <div className="p-4 flex gap-3">
          <Target className="w-5 h-5 text-sky-600 shrink-0 mt-0.5" />
          <p className="text-sm text-slate-600 leading-relaxed">{report.summary}</p>
        </div>
      </Card>

      <Separator />
      <div className="text-center">
        <Button onClick={handleGenerate} disabled={generating} className="h-12 px-8 bg-gradient-to-r from-sky-500 to-blue-600 shadow-lg shadow-sky-200">
          {generating ? (
            <span className="flex items-center gap-2"><Loader2 className="w-4 h-4 animate-spin" />生成课程方案...</span>
          ) : (
            <span className="flex items-center gap-2">生成个性化课程 <ArrowRight className="w-4 h-4" /></span>
          )}
        </Button>
      </div>
    </div>
  );
}
