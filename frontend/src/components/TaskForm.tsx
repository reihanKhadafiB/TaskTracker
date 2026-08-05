import { useState, useEffect } from "react";
import { createTask, getContexts, type Context } from "../lib/api";
import { getToken } from "../lib/auth";
import { ConfirmModal } from "./ConfirmModal";

export default function TaskForm() {
  const [token, setToken] = useState<string | null>(null);
  const [contexts, setContexts] = useState<Context[]>([]);
  const [title, setTitle] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [contextId, setContextId] = useState<string>("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [confirmAction, setConfirmAction] = useState<{title: string, message: string, onConfirm: () => void} | null>(null);

  useEffect(() => {
    const t = getToken();
    setToken(t);
    if (t) loadContexts(t);

    window.addEventListener("context-created", () => {
      if (t) loadContexts(t);
    });
  }, []);

  async function loadContexts(t: string) {
    try {
      const data = await getContexts(t);
      setContexts(data);
    } catch {
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;

    setConfirmAction({
      title: "Tambah Task Baru",
      message: `Apakah Anda yakin ingin menambahkan task "${title}"?`,
      onConfirm: async () => {
        setSubmitting(true);
        setError(null);

        try {
          await createTask(token, {
            title,
            due_date: dueDate || null,
            context_id: contextId ? Number(contextId) : null,
            status: "todo",
          });
          setTitle("");
          setDueDate("");
          setContextId("");
          window.dispatchEvent(new Event("task-created"));
        } catch (err) {
          setError(err instanceof Error ? err.message : "Failed to create task");
        } finally {
          setSubmitting(false);
        }
      }
    });
  }

  if (!token) return null;

  return (
    <div className="panel google-form-card">
      <h2 style={{ fontSize: "1.3rem", borderBottom: "1px solid var(--surface-border)", paddingBottom: "12px", marginBottom: "8px" }}>Tambah Task Baru</h2>
      <form onSubmit={handleSubmit} style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
        
        <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
          <label style={{ fontSize: "0.9rem", fontWeight: 500, color: "var(--text-dark)" }}>Nama Task <span style={{color: "var(--danger)"}}>*</span></label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Jawaban Anda"
            required
            style={{ border: "none", borderBottom: "1px solid var(--surface-border)", borderRadius: "0", padding: "8px 0", background: "transparent", boxShadow: "none" }}
          />
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
          <label style={{ fontSize: "0.9rem", fontWeight: 500, color: "var(--text-dark)" }}>Tenggat Waktu</label>
          <input
            type="date"
            value={dueDate}
            onChange={(e) => setDueDate(e.target.value)}
            style={{ width: "100%", border: "1px solid var(--surface-border)", borderRadius: "var(--radius-sm)", padding: "8px 12px" }}
          />
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: "6px" }}>
          <label style={{ fontSize: "0.9rem", fontWeight: 500, color: "var(--text-dark)" }}>Pilih Context</label>
          <select
            value={contextId}
            onChange={(e) => setContextId(e.target.value)}
            style={{ width: "100%", border: "1px solid var(--surface-border)", borderRadius: "var(--radius-sm)", padding: "8px 12px" }}
          >
            <option value="">Tanpa context</option>
            {contexts.map((ctx) => (
              <option key={ctx.id} value={ctx.id}>{ctx.name}</option>
            ))}
          </select>
        </div>

        <div style={{ marginTop: "8px", display: "flex", justifyContent: "flex-start" }}>
          <button type="submit" disabled={submitting} className="btn primary" style={{ padding: "8px 24px" }}>
            {submitting ? "Menyimpan..." : "Kirim"}
          </button>
        </div>

      {error && <p style={{ color: "var(--danger)", width: "100%", margin: "0" }}>{error}</p>}
      
      {confirmAction && (
        <ConfirmModal 
          isOpen={true}
          title={confirmAction.title}
          message={confirmAction.message}
          onConfirm={() => {
            confirmAction.onConfirm();
            setConfirmAction(null);
          }}
          onCancel={() => setConfirmAction(null)}
        />
      )}
      </form>
    </div>
  );
}