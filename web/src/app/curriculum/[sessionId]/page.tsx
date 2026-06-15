"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "../../../components/ui/button";
import { Card, CardContent } from "../../../components/ui/card";
import { Badge } from "../../../components/ui/badge";
import { getReport, generateCurriculum } from "../../../lib/api";
import type { Curriculum, DiagnosisReport } from "../../../lib/api";
import {
  BookOpen, Calendar, Loader2, Target, Sparkles, ChevronDown,
  Clock, Layers, ChevronLeft, GraduationCap, ArrowRight
} from "lucide-react";

const diffIco: Record<string, string> = { beginner: "🌱", intermediate: "🌿", advanced: "🌳" };
const diffLabel: Record<string, string> = { beginner: "入门", intermediate: "进阶", advanced: "高级" };
const diffColor: Record<string, string> = {
  beginner: "bg-emerald-50 text-emerald-700 border-emerald-200/60",
  intermediate: "bg-indigo-50 text-indigo-700 border-indigo-200/60",
  advanced: "bg-purple-50 text-purple-700 border-purple-200/60",
};

export default function CurriculumPage() {
  const params = useParams();
  const router = useRouter();
  const sessionId = params.sessionId as string;
  const [curriculum, setCurriculum] = useState<Curriculum | null>(null);
  const [report, setReport] = useState<DiagnosisReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [expandedChapter, setExpandedChapter] = useState<number | null>(null);

  useEffect(() => {
    (async () => {
      try {
        const rd = await getReport(sessionId);
        setReport(rd.report);
        const cd = await generateCurriculum(sessionId);
        setCurriculum(cd.curriculum);
      } catch (e) { console.error(e); }
      finally { setLoading(false); }
    })();
  }, [sessionId]);

  if (loading) return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <Loader2 className="w-8 h-8 animate-spin text-indigo-500" />
    </div>
  );
  if (!curriculum) return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] gap-3">
      <BookOpen className="w-10 h-10 text-muted-foreground/40" />
      <p className="text-sm text-muted-foreground">课程加载失败</p>
      <Button variant="outline" size="sm" onClick={() => router.push("/")} className="text-xs">返回首页</Button>
    </div>
  );

  return (
    <div className="flex flex-col gap-5 pb-8">
      {/* Header */}
      <div className="flex items-center gap-3">
        <button
          onClick={() => router.push(`/report/${sessionId}`)}
          className="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center hover:bg-border transition-colors"
        >
          <ChevronLeft className="w-4 h-4 text-muted-foreground" />
        </button>
        <div className="flex items-center gap-2.5 flex-1 min-w-0">
          <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-sm shrink-0">
            <BookOpen className="w-4 h-4 text-white" />
          </div>
          <div className="min-w-0">
            <h1 className="text-sm font-semibold text-foreground truncate">{curriculum.title}</h1>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <span className="flex items-center gap-1"><Calendar className="w-3 h-3" />{curriculum.total_weeks} 周</span>
              <span className="w-1 h-1 rounded-full bg-border" />
              <span className="flex items-center gap-1"><Layers className="w-3 h-3" />{curriculum.chapters.length} 章</span>
            </div>
          </div>
        </div>
        <Badge className="hidden sm:inline-flex text-xs bg-gradient-to-r from-indigo-50 to-purple-50 text-indigo-700 border-indigo-200/60">
          <GraduationCap className="w-3 h-3 mr-1" />学习方案
        </Badge>
      </div>

      {/* Goal banner */}
      <div className="p-4 rounded-xl bg-gradient-to-r from-indigo-50/80 to-purple-50/80 border border-indigo-100/60">
        <div className="flex gap-3">
          <Target className="w-5 h-5 text-indigo-600 shrink-0 mt-0.5" />
          <p className="text-sm text-muted-foreground leading-relaxed">{curriculum.goal}</p>
        </div>
      </div>

      {/* Stats cards */}
      {report && (
        <div className="grid grid-cols-2 gap-3">
          <Card className="border border-emerald-200/50 bg-white shadow-sm">
            <div className="p-4 text-center">
              <p className="text-xs text-muted-foreground mb-1">当前水平</p>
              <p className="text-2xl font-bold text-emerald-600">{Math.round(report.overall_score)}</p>
              <p className="text-xs text-muted-foreground/70 mt-0.5">{report.target_level}</p>
            </div>
          </Card>
          <Card className="border border-indigo-200/50 bg-white shadow-sm">
            <div className="p-4 text-center">
              <p className="text-xs text-muted-foreground mb-1">目标可达</p>
              <p className="text-2xl font-bold text-indigo-600">{Math.min(100, Math.round(report.overall_score + 30))}</p>
              <p className="text-xs text-muted-foreground/70 mt-0.5">约 {report.estimated_weeks} 周</p>
            </div>
          </Card>
        </div>
      )}

      {/* Chapter list */}
      <div className="space-y-3">
        <h2 className="text-sm font-semibold text-foreground flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-indigo-500" />
          课程大纲
        </h2>

        {curriculum.chapters.map((ch, i) => (
          <Card
            key={i}
            className="border border-border/60 bg-white shadow-sm hover:shadow-md transition-all duration-200 cursor-pointer"
            onClick={() => setExpandedChapter(expandedChapter === i ? null : i)}
          >
            <div className="p-4 md:p-5">
              <div className="flex items-start gap-3">
                {/* Chapter number */}
                <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-indigo-100 to-purple-100 flex items-center justify-center shrink-0 shadow-sm">
                  <span className="text-sm font-bold text-indigo-600">{i + 1}</span>
                </div>

                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <h3 className="text-sm font-semibold text-foreground">{ch.title}</h3>
                    <Badge
                      variant="outline"
                      className={`text-xs font-medium border ${diffColor[ch.difficulty] || "bg-slate-50 text-slate-600 border-slate-200/60"}`}
                    >
                      {diffIco[ch.difficulty] || "📖"} {diffLabel[ch.difficulty] || ch.difficulty}
                    </Badge>
                  </div>
                  <p className="text-xs text-muted-foreground mt-1.5 leading-relaxed">{ch.description}</p>

                  <div className="flex items-center gap-3 mt-2.5">
                    <span className="text-xs text-muted-foreground/70 flex items-center gap-1">
                      <Clock className="w-3 h-3" />{ch.duration}
                    </span>
                  </div>

                  {/* Expandable topics */}
                  <div className={`overflow-hidden transition-all duration-300 ${expandedChapter === i ? "max-h-96 mt-3" : "max-h-0 mt-0"}`}>
                    <div className="pt-3 border-t border-border/60">
                      <p className="text-[11px] font-medium text-muted-foreground mb-2 uppercase tracking-wider">涵盖知识点</p>
                      <div className="flex flex-wrap gap-1.5">
                        {ch.topics.map((t, j) => (
                          <span
                            key={j}
                            className="px-2.5 py-1 text-xs rounded-full bg-secondary text-muted-foreground hover:bg-indigo-50 hover:text-indigo-600 transition-colors"
                          >
                            {t}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>

                <ChevronDown
                  className={`w-4 h-4 text-muted-foreground/40 shrink-0 mt-2 transition-transform duration-200 ${expandedChapter === i ? "rotate-180" : ""}`}
                />
              </div>
            </div>
          </Card>
        ))}
      </div>

      {/* Restart CTA */}
      <div className="flex justify-center pt-2">
        <Button
          onClick={() => router.push("/")}
          variant="outline"
          className="text-muted-foreground border-border/80 hover:text-foreground hover:bg-secondary rounded-xl"
        >
          <ArrowRight className="w-4 h-4 mr-1.5 rotate-180" />
          重新开始新的学习
        </Button>
      </div>
    </div>
  );
}
