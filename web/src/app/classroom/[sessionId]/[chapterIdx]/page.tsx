"use client";

import { useState, useEffect, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "../../../../components/ui/button";
import {
  getUserCourseDetail,
  generateSections,
  generateSectionContent,
  type Section,
  type SectionContent,
} from "../../../../lib/api";
import {
  BookOpen, Loader2, ChevronLeft, Clock, CheckCircle,
  Code, Lightbulb, Play, ArrowRight, ListTree, Sparkles,
  FileText, GraduationCap, BookMarked
} from "lucide-react";

export default function ClassroomPage() {
  const params = useParams();
  const router = useRouter();
  const sessionId = params.sessionId as string;
  const chapterIdx = parseInt(params.chapterIdx as string, 10);

  const [sections, setSections] = useState<Section[]>([]);
  const [activeSection, setActiveSection] = useState<number>(0);
  const [sectionContent, setSectionContent] = useState<SectionContent | null>(null);
  const [loadingSections, setLoadingSections] = useState(true);
  const [loadingContent, setLoadingContent] = useState(false);
  const [courseTitle, setCourseTitle] = useState("");
  const [chapterTitle, setChapterTitle] = useState("");

  // 页面进入时加载课程信息和小节目录
  useEffect(() => {
    (async () => {
      try {
        const data = await getUserCourseDetail(sessionId);
        if (data?.curriculum) {
          setCourseTitle(data.curriculum.title);
          const ch = data.curriculum.chapters[chapterIdx];
          if (ch) {
            setChapterTitle(ch.title);
          }
        }

        // 生成小节目录
        const res = await generateSections(sessionId, chapterIdx);
        if (res.sections && res.sections.length > 0) {
          setSections(res.sections);
          // 默认选中第一个小节并加载内容
          setActiveSection(0);
        }
      } catch (e) {
        console.error(e);
      } finally {
        setLoadingSections(false);
      }
    })();
  }, [sessionId, chapterIdx]);

  // activeSection 变化时自动加载内容
  useEffect(() => {
    if (sections.length === 0) return;
    (async () => {
      setLoadingContent(true);
      setSectionContent(null);
      try {
        const res = await generateSectionContent(sessionId, chapterIdx, activeSection);
        if (res.content) {
          setSectionContent(res.content);
        }
      } catch (e) {
        console.error(e);
      } finally {
        setLoadingContent(false);
      }
    })();
  }, [sessionId, chapterIdx, activeSection, sections.length]);

  // 回到课程大纲
  const goBack = () => {
    router.push(`/profile/courses/${sessionId}`);
  };

  // 渲染 Markdown 样式的文本（简单处理）
  const renderContent = (text: string) => {
    // 将 Markdown 标题/代码块等简单转为 HTML
    const html = text
      .replace(/### (.+)/g, '<h3 class="text-sm font-semibold text-foreground mt-4 mb-2">$1</h3>')
      .replace(/## (.+)/g, '<h2 class="text-base font-semibold text-foreground mt-5 mb-2">$1</h2>')
      .replace(/# (.+)/g, '<h1 class="text-lg font-bold text-foreground mt-6 mb-3">$1</h1>')
      .replace(/```(\w*)\n?([\s\S]*?)```/g, '<pre class="p-3 rounded-lg bg-slate-900 text-slate-100 text-xs leading-relaxed overflow-x-auto my-3"><code>$2</code></pre>')
      .replace(/`([^`]+)`/g, '<code class="px-1.5 py-0.5 rounded bg-slate-100 text-indigo-600 text-xs font-mono">$1</code>')
      .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
      .replace(/\n\n/g, '</p><p class="text-xs text-muted-foreground leading-relaxed mb-2">')
      .replace(/\n/g, '<br/>');

    return `<p class="text-xs text-muted-foreground leading-relaxed mb-2">${html}</p>`;
  };

  if (loadingSections) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="w-8 h-8 animate-spin text-indigo-500" />
          <p className="text-xs text-muted-foreground">AI 正在生成小节目录...</p>
        </div>
      </div>
    );
  }

  if (sections.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-3">
        <BookOpen className="w-10 h-10 text-muted-foreground/40" />
        <p className="text-sm text-muted-foreground">小节目录生成失败</p>
        <Button variant="outline" size="sm" onClick={goBack} className="text-xs">返回课程大纲</Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col min-h-[calc(100vh-3.5rem)]">
      {/* 顶部导航 */}
      <div className="sticky top-14 z-40 bg-background/80 backdrop-blur-lg border-b border-border">
        <div className="max-w-6xl mx-auto px-4 h-12 flex items-center gap-3">
          <button
            onClick={goBack}
            className="w-7 h-7 rounded-lg bg-secondary flex items-center justify-center hover:bg-border transition-colors shrink-0"
          >
            <ChevronLeft className="w-3.5 h-3.5 text-muted-foreground" />
          </button>
          <div className="flex items-center gap-2 min-w-0">
            <GraduationCap className="w-4 h-4 text-indigo-500 shrink-0" />
            <span className="text-xs text-muted-foreground truncate">{courseTitle}</span>
            <span className="text-xs text-muted-foreground/40">/</span>
            <span className="text-xs font-medium text-foreground truncate">{chapterTitle}</span>
          </div>
          <div className="ml-auto flex items-center gap-2">
            <span className="text-[10px] text-muted-foreground/60 bg-secondary px-2 py-0.5 rounded-full flex items-center gap-1">
              <ListTree className="w-3 h-3" />{sections.length} 个小节
            </span>
          </div>
        </div>
      </div>

      {/* 主体：左右布局 */}
      <div className="flex flex-1 max-w-6xl mx-auto w-full">
        {/* 左侧菜单栏 */}
        <aside className="w-56 shrink-0 border-r border-border/40 hidden md:block">
          <div className="p-3">
            <div className="flex items-center gap-1.5 mb-2 px-2">
              <Sparkles className="w-3.5 h-3.5 text-indigo-500" />
              <span className="text-[11px] font-semibold text-foreground uppercase tracking-wider">小节目录</span>
            </div>
            <div className="space-y-0.5">
              {sections.map((sec, i) => (
                <button
                  key={i}
                  onClick={() => setActiveSection(i)}
                  className={`w-full text-left px-3 py-2 rounded-lg transition-all duration-150 ${
                    activeSection === i
                      ? "bg-gradient-to-r from-indigo-50 to-purple-50 text-indigo-700 border border-indigo-100/60"
                      : "text-muted-foreground hover:bg-secondary hover:text-foreground"
                  }`}
                >
                  <div className="flex items-start gap-2">
                    <span className={`text-[10px] font-bold mt-0.5 shrink-0 w-4 h-4 rounded flex items-center justify-center ${
                      activeSection === i
                        ? "bg-indigo-500 text-white"
                        : "bg-secondary text-muted-foreground"
                    }`}>
                      {i + 1}
                    </span>
                    <div className="min-w-0">
                      <p className={`text-xs font-medium truncate ${
                        activeSection === i ? "text-indigo-700" : "text-foreground"
                      }`}>
                        {sec.title}
                      </p>
                      <p className="text-[10px] text-muted-foreground/60 mt-0.5 flex items-center gap-1">
                        <Clock className="w-2.5 h-2.5" />
                        {sec.estimated_minutes} 分钟
                      </p>
                    </div>
                  </div>
                </button>
              ))}
            </div>
          </div>
        </aside>

        {/* 移动端小节目录选择器 */}
        <div className="md:hidden px-4 pt-3 pb-2">
          <div className="flex gap-1 overflow-x-auto pb-1 scrollbar-none">
            {sections.map((sec, i) => (
              <button
                key={i}
                onClick={() => setActiveSection(i)}
                className={`shrink-0 px-3 py-1.5 rounded-full text-xs font-medium transition-all ${
                  activeSection === i
                    ? "bg-indigo-500 text-white shadow-sm"
                    : "bg-secondary text-muted-foreground hover:bg-border"
                }`}
              >
                {i + 1}. {sec.title}
              </button>
            ))}
          </div>
        </div>

        {/* 右侧内容区 */}
        <main className="flex-1 min-w-0 px-4 md:px-6 py-4 md:py-6">
          {loadingContent ? (
            <div className="flex flex-col items-center justify-center py-20 gap-3">
              <div className="relative">
                <Loader2 className="w-8 h-8 animate-spin text-indigo-500" />
                <div className="absolute inset-0 flex items-center justify-center">
                  <BookOpen className="w-3.5 h-3.5 text-indigo-300" />
                </div>
              </div>
              <p className="text-xs text-muted-foreground">AI 正在生成学习内容...</p>
              <div className="flex gap-1.5 mt-1">
                <div className="w-1.5 h-1.5 rounded-full bg-indigo-300 animate-bounce" style={{ animationDelay: "0ms" }} />
                <div className="w-1.5 h-1.5 rounded-full bg-indigo-400 animate-bounce" style={{ animationDelay: "150ms" }} />
                <div className="w-1.5 h-1.5 rounded-full bg-indigo-500 animate-bounce" style={{ animationDelay: "300ms" }} />
              </div>
            </div>
          ) : sectionContent ? (
            <div className="max-w-3xl space-y-5">
              {/* 小节标题 */}
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <div className="w-6 h-6 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center shadow-sm shrink-0">
                    <span className="text-[10px] font-bold text-white">{activeSection + 1}</span>
                  </div>
                  <h1 className="text-base font-bold text-foreground">{sectionContent.section_title}</h1>
                </div>
                <p className="text-xs text-muted-foreground/60 ml-8">
                  第 {chapterIdx + 1} 章 · 小节 {activeSection + 1}/{sections.length}
                </p>
              </div>

              {/* 正文内容 */}
              <div
                className="p-4 md:p-5 rounded-xl bg-white border border-border/60 shadow-sm
                  [&_h1]:text-base [&_h1]:font-bold [&_h1]:text-foreground [&_h1]:mt-6 [&_h1]:mb-3
                  [&_h2]:text-sm [&_h2]:font-semibold [&_h2]:text-foreground [&_h2]:mt-5 [&_h2]:mb-2
                  [&_h3]:text-xs [&_h3]:font-semibold [&_h3]:text-foreground [&_h3]:mt-4 [&_h3]:mb-2
                  [&_p]:text-xs [&_p]:text-muted-foreground [&_p]:leading-relaxed [&_p]:mb-2
                  [&_pre]:p-3 [&_pre]:rounded-lg [&_pre]:bg-slate-900 [&_pre]:text-slate-100 [&_pre]:text-xs [&_pre]:leading-relaxed [&_pre]:overflow-x-auto [&_pre]:my-3
                  [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:bg-slate-100 [&_code]:text-indigo-600 [&_code]:text-xs [&_code]:font-mono
                  [&_pre_code]:bg-transparent [&_pre_code]:text-slate-100 [&_pre_code]:px-0 [&_pre_code]:py-0
                  [&_strong]:font-semibold [&_strong]:text-foreground
                  [&_ul]:text-xs [&_ul]:text-muted-foreground [&_ul]:space-y-1 [&_ul]:my-2 [&_ul]:pl-4
                  [&_li]:leading-relaxed
                "
                dangerouslySetInnerHTML={{ __html: renderContent(sectionContent.content) }}
              />

              {/* 关键要点 */}
              {sectionContent.key_points && sectionContent.key_points.length > 0 && (
                <div className="p-4 rounded-xl bg-amber-50/80 border border-amber-200/50">
                  <div className="flex items-center gap-2 mb-3">
                    <Lightbulb className="w-4 h-4 text-amber-500" />
                    <h3 className="text-xs font-semibold text-foreground">要点总结</h3>
                  </div>
                  <div className="space-y-2">
                    {sectionContent.key_points.map((pt, j) => (
                      <div key={j} className="flex items-start gap-2">
                        <div className="w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0 mt-1.5" />
                        <span className="text-xs text-muted-foreground leading-relaxed">{pt}</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* 代码示例 */}
              {sectionContent.code_examples && sectionContent.code_examples.length > 0 && (
                <div>
                  <div className="flex items-center gap-2 mb-2">
                    <Code className="w-4 h-4 text-sky-500" />
                    <h3 className="text-xs font-semibold text-foreground">代码示例</h3>
                  </div>
                  <div className="space-y-2">
                    {sectionContent.code_examples.map((code, j) => (
                      <pre key={j} className="p-3 rounded-lg bg-slate-900 text-slate-100 text-xs leading-relaxed overflow-x-auto">
                        <code>{code}</code>
                      </pre>
                    ))}
                  </div>
                </div>
              )}

              {/* 练习任务 */}
              {sectionContent.practice_task && (
                <div className="p-4 rounded-xl bg-emerald-50/80 border border-emerald-200/50">
                  <div className="flex items-center gap-2 mb-2">
                    <Play className="w-4 h-4 text-emerald-500" />
                    <h3 className="text-xs font-semibold text-foreground">练习任务</h3>
                  </div>
                  <p className="text-xs text-emerald-700 leading-relaxed">{sectionContent.practice_task}</p>
                </div>
              )}

              {/* 底部导航：上一节/下一节 */}
              <div className="flex items-center justify-between pt-4 border-t border-border/60">
                <div>
                  {activeSection > 0 && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setActiveSection(activeSection - 1)}
                      className="text-xs text-muted-foreground border-border/60"
                    >
                      <ChevronLeft className="w-3 h-3 mr-1" />
                      {sections[activeSection - 1].title}
                    </Button>
                  )}
                </div>
                <div>
                  {activeSection < sections.length - 1 && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setActiveSection(activeSection + 1)}
                      className="text-xs text-muted-foreground border-border/60"
                    >
                      {sections[activeSection + 1].title}
                      <ArrowRight className="w-3 h-3 ml-1" />
                    </Button>
                  )}
                  {activeSection === sections.length - 1 && (
                    <Button
                      size="sm"
                      className="text-xs bg-gradient-to-r from-emerald-500 to-teal-600 hover:from-emerald-600 hover:to-teal-700 text-white"
                      onClick={() => router.push(`/profile/courses/${sessionId}`)}
                    >
                      <CheckCircle className="w-3 h-3 mr-1" /> 完成本章学习
                    </Button>
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-20 gap-3">
              <BookOpen className="w-10 h-10 text-muted-foreground/40" />
              <p className="text-xs text-muted-foreground">内容加载失败</p>
              <Button variant="outline" size="sm" onClick={() => setActiveSection(activeSection)} className="text-xs">
                重新加载
              </Button>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
