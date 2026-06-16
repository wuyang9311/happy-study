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

// ====== 认证相关 ======

export interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  email: string;
  avatar: string;
  role: string;
  status: number;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  user_info: UserInfo;
}

export interface UserProfileResponse {
  user: UserInfo;
}

// Token 存储
export function saveToken(token: string) {
  localStorage.setItem('happy_study_token', token);
}

export function getToken(): string | null {
  if (typeof window !== 'undefined') {
    return localStorage.getItem('happy_study_token');
  }
  return null;
}

export function removeToken() {
  localStorage.removeItem('happy_study_token');
}

function authHeaders(): Record<string, string> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  const token = getToken();
  if (token) headers['Authorization'] = `Bearer ${token}`;
  return headers;
}

// 注册
export async function register(username: string, password: string, nickname?: string, email?: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/auth/register`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password, nickname, email }),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || '注册失败');
  return data;
}

// 登录
export async function login(username: string, password: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST', headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || '登录失败');
  return data;
}

// 获取个人信息
export async function getProfile(): Promise<UserProfileResponse> {
  const res = await fetch(`${API_BASE}/auth/profile`, { headers: authHeaders() });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || '获取个人信息失败');
  return data;
}

// ====== 自适应诊断（新） ======

export interface AdaptiveQuestion {
  id: number;
  content: string;
  category: string;
  difficulty: 'easy' | 'medium' | 'hard';
}

export interface AdaptiveStartResponse {
  session_id: string;
  question?: AdaptiveQuestion; // start 只返回 session_id，第一题由 SSE 流式获取
}

export interface AdaptiveAnswerResponse {
  done: boolean;
  question?: AdaptiveQuestion;
  message?: string;
}

// 开始自适应诊断
export async function startDiagnosis(topic: string, goal?: string): Promise<AdaptiveStartResponse> {
  const res = await fetch(`${API_BASE}/diagnosis/start`, {
    method: 'POST', headers: authHeaders(),
    body: JSON.stringify({ topic, goal: goal || '面试 P6' }),
  });
  if (!res.ok) {
    const data = await res.json().catch(() => ({}));
    throw new Error(data.error || `诊断启动失败: ${res.status}`);
  }
  return res.json();
}

// 提交答案，获取下一题（或结束）
export async function submitAnswer(sessionId: string, answer: string): Promise<AdaptiveAnswerResponse> {
  const res = await fetch(`${API_BASE}/diagnosis/answer`, {
    method: 'POST', headers: authHeaders(),
    body: JSON.stringify({ session_id: sessionId, answer }),
  });
  if (!res.ok) throw new Error(`提交答案失败: ${res.status}`);
  return res.json();
}

// 停止诊断，直接生成报告
export async function stopDiagnosis(sessionId: string): Promise<{ session_id: string; message: string }> {
  const res = await fetch(`${API_BASE}/diagnosis/stop`, {
    method: 'POST', headers: authHeaders(),
    body: JSON.stringify({ session_id: sessionId }),
  });
  if (!res.ok) throw new Error(`停止诊断失败: ${res.status}`);
  return res.json();
}

export async function getReport(sessionId: string): Promise<{ report: DiagnosisReport }> {
  const res = await fetch(`${API_BASE}/diagnosis/report/${sessionId}`, { headers: authHeaders() });
  if (!res.ok) throw new Error(`获取报告失败: ${res.status}`);
  return res.json();
}

export async function generateCurriculum(sessionId: string): Promise<{ curriculum: Curriculum }> {
  const res = await fetch(`${API_BASE}/curriculum/generate`, {
    method: 'POST', headers: authHeaders(),
    body: JSON.stringify({ session_id: sessionId }),
  });
  if (!res.ok) throw new Error(`生成课程失败: ${res.status}`);
  return res.json();
}

export async function generateLesson(sessionId: string, chapterIndex: number) {
  const res = await fetch(`${API_BASE}/curriculum/lesson`, {
    method: 'POST', headers: authHeaders(),
    body: JSON.stringify({ session_id: sessionId, chapter_index: chapterIndex }),
  });
  if (!res.ok) throw new Error(`生成教案失败: ${res.status}`);
  return res.json();
}

// ====== 课程列表 ======

export interface UserCourse {
  session_id: string;
  topic: string;
  goal: string;
  title: string;
  total_weeks: number;
  chapter_count: number;
  created_at: string;
  has_report: boolean;
  overall_score?: number;
}

export interface UserCoursesResponse {
  courses: UserCourse[];
}

export async function getUserCourses(): Promise<UserCoursesResponse> {
  const res = await fetch(`${API_BASE}/user/courses`, { headers: authHeaders() });
  if (!res.ok) throw new Error('获取课程列表失败');
  return res.json();
}

export async function getUserCourseDetail(sessionId: string): Promise<any> {
  const res = await fetch(`${API_BASE}/user/courses/${sessionId}`, { headers: authHeaders() });
  if (!res.ok) throw new Error('获取课程详情失败');
  return res.json();
}

// ====== 用户设置 ======

export async function getUserSettings(): Promise<{ preferred_model: string }> {
  const res = await fetch(`${API_BASE}/user/settings`, { headers: authHeaders() });
  if (!res.ok) throw new Error('获取设置失败');
  return res.json();
}

export async function updateUserModel(model: string): Promise<{ preferred_model: string }> {
  const res = await fetch(`${API_BASE}/user/settings/model`, {
    method: 'PUT',
    headers: authHeaders(),
    body: JSON.stringify({ model }),
  });
  if (!res.ok) throw new Error('更新模型失败');
  return res.json();
}
