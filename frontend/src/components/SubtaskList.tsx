import { useState, useEffect } from "react";
import { createSubtask, updateSubtaskDone, getSubtasksByTask, updateSubtask, deleteSubtask, type Subtask } from "../lib/api";
import { getToken } from "../lib/auth";
import { ConfirmModal } from "./ConfirmModal";

interface Props {
  taskId: number;
}

export default function SubtaskList({ taskId }: Props) {
  const [token] = useState<string | null>(getToken());
  const [subtasks, setSubtasks] = useState<Subtask[]>([]);
  const [newTitle, setNewTitle] = useState("");
  const [loading, setLoading] = useState(true);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [confirmAction, setConfirmAction] = useState<{title: string, message: string, isDanger?: boolean, onConfirm: () => void} | null>(null);

  useEffect(() => {
    if (!token) return;
    loadSubtasks();
  }, []);

    async function loadSubtasks() {
        if (!token) return;
        try {
            const data = await getSubtasksByTask(token, taskId);
            setSubtasks(data);
        } finally {
            setLoading(false);
        }
    }

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault();
    if (!token || !newTitle.trim()) return;
    
    setConfirmAction({
      title: "Tambah Subtask",
      message: `Tambahkan subtask "${newTitle}"?`,
      onConfirm: async () => {
        await createSubtask(token, taskId, newTitle, subtasks.length);
        setNewTitle("");
        loadSubtasks();
      }
    });
  }

  async function handleToggle(subtaskId: number, isDone: boolean) {
    if (!token) return;
    
    setConfirmAction({
      title: "Ubah Status Subtask",
      message: isDone ? "Tandai subtask ini sebagai selesai?" : "Tandai subtask ini sebagai belum selesai?",
      onConfirm: async () => {
        await updateSubtaskDone(token, taskId, subtaskId, isDone);
        loadSubtasks();
        window.dispatchEvent(new Event("task-created"));
      }
    });
  }

  async function handleDelete(subtaskId: number) {
    if (!token) return;
    
    setConfirmAction({
      title: "Hapus Subtask",
      message: "Apakah Anda yakin ingin menghapus subtask ini?",
      isDanger: true,
      onConfirm: async () => {
        await deleteSubtask(token, taskId, subtaskId);
        loadSubtasks();
      }
    });
  }

  async function handleEditSubmit(subtaskId: number, title: string) {
    if (!token || !title.trim()) return;
    
    setConfirmAction({
      title: "Simpan Perubahan Subtask",
      message: "Simpan perubahan pada subtask ini?",
      onConfirm: async () => {
        await updateSubtask(token, taskId, subtaskId, title);
        setEditingId(null);
        loadSubtasks();
      }
    });
  }

  if (loading) return <p>Loading subtasks...</p>;

  return (
    <div style={{ fontSize: "0.95em" }}>
      <ul style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
        {subtasks.map((st) => (
          <li key={st.id} style={{ display: "flex", alignItems: "center", gap: "10px", padding: "6px 0", borderBottom: "1px solid #f1f5f9" }}>
            <input
              type="checkbox"
              checked={st.is_done}
              onChange={(e) => handleToggle(st.id, e.target.checked)}
              style={{ width: "18px", height: "18px", marginTop: "2px" }}
            />
            {editingId === st.id ? (
              <form className="inline-form" onSubmit={(e) => {
                  e.preventDefault();
                  const target = e.target as typeof e.target & {
                    title: { value: string };
                  };
                  handleEditSubmit(st.id, target.title.value);
                }}>
                <input type="text" name="title" defaultValue={st.title} autoFocus style={{ flex: 1, border: "none", borderBottom: "1px solid var(--primary)", borderRadius: 0, padding: "4px 0", background: "transparent", outline: "none", boxShadow: "none" }} />
                <button type="submit" className="icon-btn">Save</button>
                <button type="button" className="icon-btn secondary" onClick={() => setEditingId(null)}>Cancel</button>
              </form>
            ) : (
              <>
                <span style={{ flex: 1, textDecoration: st.is_done ? "line-through" : "none", color: st.is_done ? "var(--text-muted)" : "var(--text)", fontSize: "0.95rem" }}>{st.title}</span>
                <button onClick={() => setEditingId(st.id)} className="icon-btn secondary" style={{ padding: "4px 8px", fontSize: "0.8rem", background: "transparent", border: "none", boxShadow: "none" }}>Edit</button>
                <button onClick={() => handleDelete(st.id)} className="icon-btn danger" style={{ padding: "4px 8px", fontSize: "0.8rem", background: "transparent", border: "none", boxShadow: "none", color: "var(--danger)" }}>Del</button>
              </>
            )}
          </li>
        ))}
      </ul>
      <form onSubmit={handleAdd} style={{ display: "flex", gap: "12px", marginTop: "8px", alignItems: "center", width: "100%", overflow: "hidden" }}>
        <div style={{ width: "18px", height: "18px", border: "2px solid #cbd5e1", borderRadius: "4px", flexShrink: 0 }}></div>
        <input
          type="text"
          value={newTitle}
          onChange={(e) => setNewTitle(e.target.value)}
          placeholder="Tambahkan subtask..."
          style={{ flex: 1, minWidth: 0, padding: "8px 0", border: "none", borderBottom: "1px solid var(--surface-border)", borderRadius: 0, background: "transparent", boxShadow: "none", outline: "none", fontSize: "0.95rem" }}
        />
        <button type="submit" className="icon-btn primary" style={{ background: "transparent", color: "var(--primary)", border: "none", boxShadow: "none", padding: "4px", fontSize: "1.2rem", fontWeight: "bold", flexShrink: 0, cursor: "pointer" }}>+</button>
      </form>
      {confirmAction && (
        <ConfirmModal 
          isOpen={true}
          title={confirmAction.title}
          message={confirmAction.message}
          isDanger={confirmAction.isDanger}
          onConfirm={() => {
            confirmAction.onConfirm();
            setConfirmAction(null);
          }}
          onCancel={() => setConfirmAction(null)}
        />
      )}
    </div>
  );
}