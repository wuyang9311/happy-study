"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { getReport, generateCurriculum } from "@/lib/api";
import type { Curriculum, DiagnosisReport } from "@/lib/api";
import { BookOpen, Calendar, Loader2, Target, Sparkles, ChevronRight, Clock, Layers } from "lucide-react";

const diffIco: Record<string, string> = { beginner: "🌱", intermediate: "🌿", advanced: "🌳" };

export default function CurriculumPage() {
  const params = useParams();
  const sessionId = params.sessionId as string;
  const [curriculum, setCurriculum] = useState<Curriculum | null>(null);
  const [report, setReport] = useState<DiagnosisReport | null>(null);
  const [loading, setLoading] = useState(true);

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
    <div className="flex items-center justify-center min-h-[80vh]">
      <Loader2 className="w-10 h-10 animate-spin text-sky-600" />
    </div>
  );
  if (!curriculum) return <div className="text-center mt-20 text-slate-500">课程加载失败</div>;

  return (
    <div className="flex flex-col gap-6 py-4">
      <div className="flex items-center gap-3">
        <div className="p-2 rounded-xl bg-gradient-to-br from-sky-500 to-blue-600 shadow-sm">
          <BookOpen className="w-5 h-5 text-white" />
        </div>
        <div className="flex-1">
          <h1 className="text-lg font-semibold text-slate-800">{curriculum.title}</h1>
          <div className="flex items-center gap-3 mt-1">
            <Badge variant="secondary" className="text-xs flex items-center gap-1">
              <Calendar className="w-3 h-3" />{curriculum.total_weeks} 周
            </Badge>
            <Badge variant="secondary" className="text-xs flex items-center gap-1">
              <Layers className="w-3 h-3" />{curriculum.chapters.length} 章
            </Badge>
          </div>
        </div>
      </div>

      <Card className="border-0 shadow-sm bg-gradient-to-r from-sky-50 to-blue-50">
        <CardContent className="p-4 flex gap-3">
          <Target className="w-5 h-5 text-sky-600 shrink-0 mt-0.5" />
          <p className="text-sm text-slate-600 leading-relaxed">{curriculum.goal}</p>
        </CardContent>
      </Card>

      <Separator />

      {report && (
        <div className="grid grid-cols-2 gap-3">
          <Card className="border-0 shadow-sm bg-green-50/50">
            <CardContent className="p-3 text-center">
              <p className="text-xs text-slate-500">当前水平</p>
              <p className="text-2xl font-bold text-green-600">{Math.round(report.overall_score)}</p>
              <p className="text-xs text-slate-400">{report.target_level}</p>
            </CardContent>
          </Card>
          <Card className="border-0 shadow-sm bg-blue-50/50">
            <CardContent className="p-3 text-center">
              <p className="text-xs text-slate-500">目标可达</p>
              <p className="text-2xl font-bold text-blue-600">{Math.min(100, Math.round(report.overall_score + 30))}</p>
              <p className="text-xs text-slate-400">约 {report.estimated_weeks} 周</p>
            </CardContent>
          </Card>
        </div>
      )}

      <div className="space-y-3">
        <h2 className="text-sm font-medium text-slate-700 flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-sky-600" />
          课程大纲
        </h2>
        {curriculum.chapters.map((ch, i) => (
          <Card key={i} className="border-0 shadow-sm hover:shadow-md transition-shadow">
            <CardContent className="p-4">
              <div className="flex items-start gap-3">
                <div className="w-8 h-8 rounded-full bg-sky-100 flex items-center justify-center shrink-0 text-sm">
                  {diffIco[ch.difficulty] || "📖"}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <h3 className="text-sm font-medium text-slate-800">{i+1}. {ch.title}</h3>
                    <Badge variant="outline" className="text-xs">{ch.difficulty === "beginner" ? "入门" : ch.difficulty === "intermediate" ? "进阶" : "高级"}</Badge>
                  </div>
                  <p className="text-xs text-slate-500 mt-1 line-clamp-2">{ch.description}</p>
                  <Badge variant="secondary" className="text-xs mt-2 flex items-center gap-1 w-fit">
                    <Clock className="w-3 h-3" />{ch.duration}
                  </Badge>
                  <div className="flex flex-wrap gap-1.5 mt-2">
                    {ch.topics.map((t, j) => <span key={j} className="px-2 py-0.5 text-xs rounded-full bg-slate-100 text-slate-600">{t}</span>)}
                  </div>
                </div>
                <ChevronRight className="w-4 h-4 text-slate-300 shrink-0 mt-2" />
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Separator />
      <div className="text-center">
        <Button onClick={() => window.location.href = "/"} variant="outline" className="text-slate-600">重新开始新的学习</Button>
      </div>
    </div>
  );
}
