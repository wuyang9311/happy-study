"use client";

import { useState, useEffect } from "react";
import { getUserSettings, updateUserModel } from "../../../../lib/api";
import { Button } from "../../../../components/ui/button";
import { Input } from "../../../../components/ui/input";
import { Card, CardContent } from "../../../../components/ui/card";
import { Check, Loader2, Sparkles, AlertCircle } from "lucide-react";

const PRESET_MODELS = [
  { id: "deepseek-chat", name: "DeepSeek V3", desc: "高速高性价比，适合日常诊断", provider: "DeepSeek" },
  { id: "deepseek-reasoner", name: "DeepSeek R1", desc: "推理更强，适合深度分析", provider: "DeepSeek" },
  { id: "deepseek-chat-v3-0324", name: "DeepSeek V3-0324", desc: "最新版本，性能全面提升", provider: "DeepSeek" },
];

export default function ModelSettingsPage() {
  const [currentModel, setCurrentModel] = useState("");
  const [customModel, setCustomModel] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    getUserSettings().then(data => {
      setCurrentModel(data.preferred_model || "");
      setCustomModel(data.preferred_model || "");
      setLoading(false);
    }).catch(() => {
      setError("加载设置失败");
      setLoading(false);
    });
  }, []);

  const selectPreset = async (modelId: string) => {
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      await updateUserModel(modelId);
      setCurrentModel(modelId);
      setCustomModel(modelId);
      setSuccess("模型已切换为 " + (PRESET_MODELS.find(m => m.id === modelId)?.name || modelId));
    } catch (err: any) {
      setError(err?.message || "保存失败");
    } finally {
      setSaving(false);
    }
  };

  const saveCustomModel = async () => {
    if (!customModel.trim()) return;
    setSaving(true);
    setError("");
    setSuccess("");
    try {
      await updateUserModel(customModel.trim());
      setCurrentModel(customModel.trim());
      setSuccess("模型已设置为 " + customModel.trim());
    } catch (err: any) {
      setError(err?.message || "保存失败");
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[30vh]">
        <Loader2 className="w-5 h-5 animate-spin text-indigo-500" />
      </div>
    );
  }

  return (
    <div className="space-y-5">
      <div>
        <h2 className="text-base font-semibold text-foreground">模型设置</h2>
        <p className="text-xs text-muted-foreground mt-1">
          选择用于诊断和教学的 AI 模型。你需要拥有对应模型的 API 访问权限。
        </p>
      </div>

      {/* 预设模型 */}
      <div>
        <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-3">推荐模型</h3>
        <div className="space-y-2">
          {PRESET_MODELS.map((model) => {
            const isActive = currentModel === model.id;
            return (
              <button
                key={model.id}
                onClick={() => selectPreset(model.id)}
                disabled={saving}
                className={`w-full text-left flex items-center gap-3 p-3.5 rounded-xl border transition-all ${
                  isActive
                    ? "bg-indigo-50 border-indigo-300/60 shadow-sm"
                    : "bg-white border-border/60 hover:border-indigo-200 hover:shadow-sm"
                }`}
              >
                <div className={`w-8 h-8 rounded-xl flex items-center justify-center shrink-0 ${
                  isActive
                    ? "bg-gradient-to-br from-indigo-500 to-purple-600"
                    : "bg-secondary"
                }`}>
                  <Sparkles className={`w-4 h-4 ${isActive ? "text-white" : "text-muted-foreground"}`} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className={`text-sm font-medium ${isActive ? "text-indigo-700" : "text-foreground"}`}>
                      {model.name}
                    </span>
                    <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-secondary text-muted-foreground border border-border/60">
                      {model.provider}
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5">{model.desc}</p>
                </div>
                {isActive && (
                  <div className="w-6 h-6 rounded-full bg-indigo-500 flex items-center justify-center shrink-0">
                    <Check className="w-3.5 h-3.5 text-white" />
                  </div>
                )}
              </button>
            );
          })}
        </div>
      </div>

      {/* 自定义模型 */}
      <div>
        <h3 className="text-xs font-medium text-muted-foreground uppercase tracking-wider mb-3">自定义模型</h3>
        <Card className="border border-border/60 bg-white shadow-sm">
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground mb-3">
              如果你使用其他兼容 OpenAI API 的模型，可以在这里手动输入模型名称。
            </p>
            <div className="flex gap-2">
              <Input
                value={customModel}
                onChange={e => setCustomModel(e.target.value)}
                placeholder="输入模型名称，如 gpt-4o、claude-sonnet-4"
                className="flex-1 text-sm bg-background border-border/80 focus-visible:ring-indigo-400/40"
                disabled={saving}
                onKeyDown={e => { if (e.key === "Enter") saveCustomModel(); }}
              />
              <Button
                onClick={saveCustomModel}
                disabled={!customModel.trim() || saving}
                className="h-10 px-4 bg-gradient-to-r from-sky-500 to-indigo-600 hover:from-sky-600 hover:to-indigo-700 text-white shadow-sm text-xs rounded-xl"
              >
                {saving ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : "应用"}
              </Button>
            </div>
            {currentModel && !PRESET_MODELS.find(m => m.id === currentModel) && (
              <p className="text-xs text-indigo-600 mt-2 flex items-center gap-1">
                <Check className="w-3 h-3" /> 当前使用: {currentModel}
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 提示信息 */}
      {success && (
        <div className="flex items-center gap-2 p-3 rounded-xl bg-green-50 border border-green-200/60">
          <Check className="w-4 h-4 text-green-600 shrink-0" />
          <p className="text-xs text-green-700">{success}</p>
        </div>
      )}

      {error && (
        <div className="flex items-center gap-2 p-3 rounded-xl bg-red-50 border border-red-200/60">
          <AlertCircle className="w-4 h-4 text-red-500 shrink-0" />
          <p className="text-xs text-red-600">{error}</p>
        </div>
      )}

      {/* 使用说明 */}
      <div className="p-3.5 rounded-xl bg-amber-50 border border-amber-200/60">
        <p className="text-xs text-amber-800 leading-relaxed">
          <strong>💡 注意：</strong>切换模型后，新生成的诊断问题和课程内容将使用所选模型。
          如果模型不可用，系统会自动回退到默认模型（deepseek-chat）。
        </p>
      </div>
    </div>
  );
}
