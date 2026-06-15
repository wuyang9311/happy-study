"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Button } from "../../../components/ui/button";
import { Card, CardContent } from "../../../components/ui/card";
import { Badge } from "../../../components/ui/badge";
import { getUserCourses, type UserCourse } from "../../../lib/api";
import {
  BookOpen, Loader2, Calendar, Layers, ArrowRight, GraduationCap,
  Target, ChevronRight, BookMarked
} from "lucide-react";

export default function ProfileCoursesPage() {
  const router = useRouter();
  const [courses, setCourses] = useState<UserCourse[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getUserCourses()
      .then(d => setCourses(d.courses))
      .catch(e => console.error(e))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return (
    <div className="flex items-center justify-center min-h-[60vh]">
      <Loader2 className="w-8 h-8 animate-spin text-indigo-500" />
    </div>
  );

  return (
    <div className="flex flex-col gap-5 pb-8">
      {/* Header */}
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-sm shrink-0">
          <BookMarked className="w-4 h-4 text-white" />
        </div>
        <div>
          <h1 className="text-sm font-semibold text-foreground">我的课程</h1>
          <p className="text-xs text-muted-foreground">共 {courses.length} 门课程</p>
        </div>
      </div>

      {courses.length === 0 ? (
        /* Empty state */
        <div className="flex flex-col items-center justify-center min-h-[40vh] gap-4">
          <div className="w-16 h-16 rounded-2xl bg-secondary flex items-center justify-center">
            <BookOpen className="w-7 h-7 text-muted-foreground/50" />
          </div>
          <div className="text-center">
            <p className="text-sm font-medium text-foreground mb-1">还没有课程</p>
            <p className="text-xs text-muted-foreground">完成一次诊断学习后，课程就会出现在这里</p>
          </div>
          <Button
            onClick={() => router.push("/")}
            className="bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white text-xs rounded-xl"
          >
            开始学习
          </Button>
        </div>
      ) : (
        /* Course list */
        <div className="space-y-3">
          {courses.map((course) => (
            <Link key={course.session_id} href={`/profile/courses/${course.session_id}`}>
              <Card className="border border-border/60 bg-white shadow-sm hover:shadow-md hover:border-indigo-200/60 transition-all duration-200 cursor-pointer group">
                <CardContent className="p-4 md:p-5">
                  <div className="flex items-start gap-3">
                    {/* Icon */}
                    <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-50 to-purple-50 flex items-center justify-center shrink-0 shadow-sm border border-indigo-100/50">
                      <GraduationCap className="w-5 h-5 text-indigo-600" />
                    </div>

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <h3 className="text-sm font-semibold text-foreground">{course.title}</h3>
                        {course.has_report && course.overall_score !== undefined && (
                          <Badge variant="outline" className={`text-xs font-medium border ${
                            course.overall_score >= 70
                              ? "bg-emerald-50 text-emerald-700 border-emerald-200/60"
                              : course.overall_score >= 40
                              ? "bg-amber-50 text-amber-700 border-amber-200/60"
                              : "bg-red-50 text-red-700 border-red-200/60"
                          }`}>
                            {Math.round(course.overall_score)} 分
                          </Badge>
                        )}
                      </div>
                      <p className="text-xs text-muted-foreground mt-1 line-clamp-1">{course.topic} · {course.goal}</p>
                      <div className="flex items-center gap-3 mt-2 text-xs text-muted-foreground/70">
                        <span className="flex items-center gap-1"><Calendar className="w-3 h-3" />{course.created_at?.slice(0, 10)}</span>
                        <span className="w-1 h-1 rounded-full bg-border" />
                        <span className="flex items-center gap-1"><Layers className="w-3 h-3" />{course.total_weeks} 周 · {course.chapter_count} 章</span>
                      </div>
                    </div>

                    <ChevronRight className="w-4 h-4 text-muted-foreground/30 group-hover:text-indigo-400 transition-colors shrink-0 mt-1" />
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}

      {/* Back to home */}
      <div className="flex justify-center pt-2">
        <Button
          onClick={() => router.push("/")}
          variant="outline"
          className="text-muted-foreground border-border/80 hover:text-foreground hover:bg-secondary rounded-xl text-xs"
        >
          <ArrowRight className="w-4 h-4 mr-1.5 rotate-180" />
          返回首页
        </Button>
      </div>
    </div>
  );
}
