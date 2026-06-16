"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "../../../../components/ui/button";
import { Card, CardContent } from "../../../../components/ui/card";
import { Badge } from "../../../../components/ui/badge";
import { getUserCourseDetail, generateLesson } from "../../../../lib/api";
import {
  BookOpen, Loader2, Calendar, Target, Sparkles, ChevronDown,
  Clock, Layers, ChevronLeft, GraduationCap, ArrowRight, FileText,
  Code, CheckCircle, Lightbulb, BookMarked, Play
} from "lucide-react";

const diffIco: Record<string, string> = { beginner: "🌱", intermediate: "🌿", advanced: "🌳" };
const diffLabel: Record<string, string> = { beginner: "入门", intermediate: "进阶", advanced: "高级" };
const diffColor: Record<string, string> = {
  beginner: "bg-emerald-50 text-emerald-700 border-emerald-200/60",
  intermediate: "bg-indigo-50 text-indigo-700 border-indigo-200/60",
  advanced: "bg-purple-50 text-purple-700 border-purple-200/60",
};

interface LessonPlan {
  chapter_title: string;
  section_title: string;
  content: string;
  code_examples: string[];
  key_points: string[];
  practice_question: string;
}

export default function CourseDetailPage() {
  const params = useParams();
  const router = useRouter();
  const sessionId = params.sessionId as string;

  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [expandedChapter, setExpandedChapter] = useState<number | null>(null);
  const [lessonContent, setLessonContent] = useState<Record<number, LessonPlan>>({});
  const [generatingLesson, setGeneratingLesson] = useState<number | null>(null);

  useEffect(() => {
    getUserCourseDetail(sessionId)
      .then(d => {
        setData(d);
        // Load cached lesson plans
        if (d.lesson_plans) {
          setLessonContent(d.lesson_plans);
        }
      })
      .catch(e => console.error(e))
      .finally(() => setLoading(false));
  }, [sessionId]);

  const handleToggleChapter = async (index: number) => {
    if (expandedChapter === index) {
      setExpandedChapter(null);
      return;
    }
    setExpandedChapter(index);

    // Generate lesson if not cached
    if (!lessonContent[index]) {
      setGeneratingLesson(index);
      try {
        const res = await generateLesson(sessionId, index);
        if (res.lesson_plan) {
          setLessonContent(prev => ({ ...prev, [index]: res.lesson_plan }));
        }
      } catch (e) {
        console.error(e);
      } finally {
        setGeneratingLesson(null);
      }
    }
  };

  if (loading) return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <Loader2 className="w-8 h-8 animate-spin text-indigo-500" />
    </div>
  );
  if (!data || !data.curriculum) return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] gap-3">
      <BookOpen className="w-10 h-10 text-muted-foreground/40" />
      <p className="text-sm text-muted-foreground">课程加载失败</p>
      <Button variant="outline" size="sm" onClick={() => router.push("/profile/courses")} className="text-xs">返回课程列表</Button>
    </div>
  );

  const { curriculum, report, topic, goal, created_at } = data;

  return (
    <div className="flex flex-col gap-5 pb-8">
      {/* Header */}
      <div className="flex items-center gap-3">
        <button
          onClick={() => router.push("/profile/courses")}
          className="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center hover:bg-border transition-colors"
        >
          <ChevronLeft className="w-4 h-4 text-muted-foreground" />
        </button>
        <div className="flex items-center gap-2.5 flex-1 min-w-0">
          <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-sm shrink-0">
            <BookMarked className="w-4 h-4 text-white" />
          </div>
          <div className="min-w-0">
            <h1 className="text-sm font-semibold text-foreground truncate">{curriculum.title}</h1>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <span className="flex items-center gap-1"><Calendar className="w-3 h-3" />{created_at?.slice(0, 10)}</span>
              <span className="w-1 h-1 rounded-full bg-border" />
              <span className="flex items-center gap-1"><Layers className="w-3 h-3" />{curriculum.total_weeks} 周</span>
              <span className="w-1 h-1 rounded-full bg-border" />
              <span className="flex items-center gap-1"><BookOpen className="w-3 h-3" />{curriculum.chapters.length} 章</span>
            </div>
          </div>
        </div>
      </div>

      {/* Info banner */}
      <div className="p-4 rounded-xl bg-gradient-to-r from-indigo-50/80 to-purple-50/80 border border-indigo-100/60">
        <div className="flex gap-3">
          <Target className="w-5 h-5 text-indigo-600 shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-medium text-foreground">{topic}</p>
            <p className="text-xs text-muted-foreground mt-0.5">目标：{goal}</p>
          </div>
        </div>
      </div>

      {/* Stats cards */}
      {report && (
        <div className="grid grid-cols-2 gap-3">
          <Card className="border border-emerald-200/50 bg-white shadow-sm">
            <div className="p-4 text-center">
              <p className="text-xs text-muted-foreground mb-1">诊断得分</p>
              <p className="text-2xl font-bold text-emerald-600">{Math.round(report.overall_score)}</p>
              <p className="text-xs text-muted-foreground/70 mt-0.5">{report.target_level}</p>
            </div>
          </Card>
          <Card className="border border-indigo-200/50 bg-white shadow-sm">
            <div className="p-4 text-center">
              <p className="text-xs text-muted-foreground mb-1">预计达成</p>
              <p className="text-2xl font-bold text-indigo-600">{Math.min(100, Math.round((report.overall_score || 0) + 30))}</p>
              <p className="text-xs text-muted-foreground/70 mt-0.5">约 {report.estimated_weeks} 周</p>
            </div>
          </Card>
        </div>
      )}

      {/* Chapter list with expandable lessons */}
      <div className="space-y-3">
        <h2 className="text-sm font-semibold text-foreground flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-indigo-500" />
          课程大纲
        </h2>

        {curriculum.chapters.map((ch: any, i: number) => (
          <Card
            key={i}
            className={`border bg-white shadow-sm transition-all duration-200 ${
              expandedChapter === i ? "border-indigo-200/80 shadow-md" : "border-border/60 hover:shadow-md"
            }`}
          >
            <div
              className="p-4 md:p-5 cursor-pointer"
              onClick={() => handleToggleChapter(i)}
            >
              <div className="flex items-start gap-3">
                {/* Chapter number */}
                <div className={`w-9 h-9 rounded-xl flex items-center justify-center shrink-0 shadow-sm transition-colors ${
                  expandedChapter === i
                    ? "bg-gradient-to-br from-indigo-500 to-purple-600 text-white"
                    : "bg-gradient-to-br from-indigo-100 to-purple-100 text-indigo-600"
                }`}>
                  <span className="text-sm font-bold">{i + 1}</span>
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
                    {lessonContent[i] && (
                      <Badge variant="outline" className="text-xs font-medium bg-emerald-50 text-emerald-700 border-emerald-200/60">
                        <CheckCircle className="w-3 h-3 mr-0.5" /> 已学习
                      </Badge>
                    )}
                  </div>
                  <p className="text-xs text-muted-foreground mt-1.5 leading-relaxed">{ch.description}</p>

                  <div className="flex items-center gap-3 mt-2.5">
                    <span className="text-xs text-muted-foreground/70 flex items-center gap-1">
                      <Clock className="w-3 h-3" />{ch.duration}
                    </span>
                    {generatingLesson === i && (
                      <span className="text-xs text-indigo-500 flex items-center gap-1">
                        <Loader2 className="w-3 h-3 animate-spin" /> 生成中...
                      </span>
                    )}
                  </div>
                </div>

                <div className="flex items-center gap-2 shrink-0">
                  <Button
                    size="sm"
                    className="text-xs bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white rounded-lg shadow-sm"
                    onClick={(e) => {
                      e.stopPropagation();
                      router.push(`/classroom/${sessionId}/${i}`);
                    }}
                  >
                    <Play className="w-3 h-3 mr-1" /> 开始学习
                  </Button>
                  <ChevronDown
                    className={`w-4 h-4 text-muted-foreground/40 shrink-0 transition-transform duration-200 ${expandedChapter === i ? "rotate-180" : ""}`}
                  />
                </div>
              </div>
            </div>

            {/* Expanded Lesson Content */}
            {expandedChapter === i && lessonContent[i] && (
              <div className="px-4 md:px-5 pb-4 md:pb-5 border-t border-border/60">
                <div className="pt-4 space-y-4">
                  {/* Section title */}
                  <div className="flex items-center gap-2">
                    <FileText className="w-4 h-4 text-indigo-500" />
                    <h4 className="text-sm font-medium text-foreground">{lessonContent[i].section_title}</h4>
                  </div>

                  {/* Content */}
                  <div className="p-4 rounded-xl bg-slate-50 border border-slate-100">
                    <div className="text-xs text-muted-foreground leading-relaxed whitespace-pre-wrap">
                      {lessonContent[i].content}
                    </div>
                  </div>

                  {/* Key Points */}
                  {lessonContent[i].key_points && lessonContent[i].key_points.length > 0 && (
                    <div>
                      <div className="flex items-center gap-2 mb-2">
                        <Lightbulb className="w-4 h-4 text-amber-500" />
                        <h5 className="text-xs font-semibold text-foreground">关键要点</h5>
                      </div>
                      <div className="space-y-1.5">
                        {lessonContent[i].key_points.map((pt: string, j: number) => (
                          <div key={j} className="flex items-start gap-2">
                            <div className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0 mt-1.5" />
                            <span className="text-xs text-muted-foreground">{pt}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Code Examples */}
                  {lessonContent[i].code_examples && lessonContent[i].code_examples.length > 0 && (
                    <div>
                      <div className="flex items-center gap-2 mb-2">
                        <Code className="w-4 h-4 text-sky-500" />
                        <h5 className="text-xs font-semibold text-foreground">代码示例</h5>
                      </div>
                      <div className="space-y-2">
                        {lessonContent[i].code_examples.map((code: string, j: number) => (
                          <pre key={j} className="p-3 rounded-lg bg-slate-900 text-slate-100 text-xs leading-relaxed overflow-x-auto">
                            <code>{code}</code>
                          </pre>
                        ))}
                      </div>
                    </div>
                  )}

                  {/* Practice Question */}
                  {lessonContent[i].practice_question && (
                    <div>
                      <div className="flex items-center gap-2 mb-2">
                        <CheckCircle className="w-4 h-4 text-emerald-500" />
                        <h5 className="text-xs font-semibold text-foreground">练习题</h5>
                      </div>
                      <div className="p-3 rounded-xl bg-emerald-50 border border-emerald-100">
                        <p className="text-xs text-emerald-800 leading-relaxed">{lessonContent[i].practice_question}</p>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Loading state */}
            {expandedChapter === i && generatingLesson === i && (
              <div className="px-4 md:px-5 pb-4 md:pb-5 border-t border-border/60">
                <div className="pt-6 flex flex-col items-center gap-3">
                  <Loader2 className="w-6 h-6 animate-spin text-indigo-500" />
                  <p className="text-xs text-muted-foreground">AI 正在生成教案...</p>
                </div>
              </div>
            )}
          </Card>
        ))}
      </div>

      {/* Actions */}
      <div className="flex justify-center gap-3 pt-2">
        <Button
          onClick={() => router.push(`/report/${sessionId}`)}
          variant="outline"
          className="text-muted-foreground border-border/80 hover:text-foreground hover:bg-secondary rounded-xl text-xs"
        >
          查看诊断报告
        </Button>
        <Button
          onClick={() => router.push("/")}
          variant="outline"
          className="text-muted-foreground border-border/80 hover:text-foreground hover:bg-secondary rounded-xl text-xs"
        >
          <ArrowRight className="w-4 h-4 mr-1.5 rotate-180" />
          重新开始
        </Button>
      </div>
    </div>
  );
}
