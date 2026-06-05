const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

export interface Question {
  id: number;
  content: string;
  category: string;
  difficulty: 'easy' | 'medium' | 'hard';
}

export interface StartDiagnosisResponse {
  session_id: string;
  question: Question;
  total_questions: number;
  question_number: number;
}

export interface AnswerResponse {
  done: boolean;
  question?: Question;
  question_number?: number;
  message?: string;
}

export interface DiagnosisReport {
  topic: string;
  overall_score: number;
  scores: { category: string; score: number; level: string; feedback: string }[];
  weaknesses: string[];
  strengths: string[];
  summary: string;
  target_level: string;
  estimated_weeks: number;
}

export interface Curriculum {
  topic: string;
  title: string;
  goal: string;
  chapters: { title: string; description: string; duration: string; topics: string[]; difficulty: string }[];
  total_weeks: number;
}

export async function startDiagnosis(topic: string, goal?: string): Promise<StartDiagnosisResponse> {
  const res = await fetch(`${API_BASE}/diagnosis/start`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ topic, goal: goal || '面试 P6' }),
  });
  if (!res.ok) throw new Error(`诊断启动失败: ${res.status}`);
  return res.json();
}

export async function submitAnswer(sessionId: string, questionId: number, answer: string): Promise<AnswerResponse> {
  const res = await fetch(`${API_BASE}/diagnosis/answer`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ session_id: sessionId, question_id: questionId, answer }),
  });
  if (!res.ok) throw new Error(`提交答案失败: ${res.status}`);
  return res.json();
}

export async function getReport(sessionId: string): Promise<{ report: DiagnosisReport }> {
  const res = await fetch(`${API_BASE}/diagnosis/report/${sessionId}`);
  if (!res.ok) throw new Error(`获取报告失败: ${res.status}`);
  return res.json();
}

export async function generateCurriculum(sessionId: string): Promise<{ curriculum: Curriculum }> {
  const res = await fetch(`${API_BASE}/curriculum/generate`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ session_id: sessionId }),
  });
  if (!res.ok) throw new Error(`生成课程失败: ${res.status}`);
  return res.json();
}

export async function generateLesson(sessionId: string, chapterIndex: number) {
  const res = await fetch(`${API_BASE}/curriculum/lesson`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ session_id: sessionId, chapter_index: chapterIndex }),
  });
  if (!res.ok) throw new Error(`生成教案失败: ${res.status}`);
  return res.json();
}
