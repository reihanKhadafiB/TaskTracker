import { useState, useEffect } from "react";
import { getTasks, getContexts, updateTaskStatus, updateTask, deleteTask, type Task, type Context } from "../lib/api";
import { getToken, logout } from "../lib/auth";
import TaskForm from "./TaskForm";
import ContextManager from "./ContextManager";
import SubtaskList from "./SubtaskList";
import { ConfirmModal } from "./ConfirmModal";

export default function TaskApp() {
  const [token, setToken] = useState<string | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [contexts, setContexts] = useState<Context[]>([]);
  const [activeContext, setActiveContext] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editingTaskId, setEditingTaskId] = useState<number | null>(null);
  const [confirmAction, setConfirmAction] = useState<{title: string, message: string, isDanger?: boolean, onConfirm: () => void} | null>(null);

  useEffect(() => {
    const t = getToken();
    setToken(t);
    if (t) {
      loadContexts(t);
      loadTasks(t, activeContext);
    }

    const handleTaskUpdate = () => {
      if (t) loadTasks(t, activeContext);
    };
    const handleContextUpdate = () => {
      if (t) {
        loadContexts(t);
        loadTasks(t, activeContext);
      }
    };

    window.addEventListener("task-created", handleTaskUpdate);
    window.addEventListener("context-created", handleContextUpdate);

    return () => {
      window.removeEventListener("task-created", handleTaskUpdate);
      window.removeEventListener("context-created", handleContextUpdate);
    };
  }, [activeContext]);

  async function loadContexts(t: string) {
    try {
      const data = await getContexts(t);
      setContexts(data);
    } catch (err) {
      console.error(err);
    }
  }

  async function loadTasks(t: string, ctxId: number | null) {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      if (ctxId) params.append("context_id", ctxId.toString());
      
      const { data } = await getTasks(t, params);
      setTasks(data || []);
      setError(null);
    } catch (err) {
      setError("Gagal memuat task. Backend mungkin sedang sleep.");
    } finally {
      setLoading(false);
    }
  }

  async function handleToggleStatus(taskId: number, currentStatus: string) {
    if (!token) return;
    setConfirmAction({
      title: "Ubah Status Task",
      message: currentStatus === "done" ? "Tandai task ini sebagai belum selesai?" : "Tandai task ini sebagai selesai?",
      onConfirm: async () => {
        const newStatus = currentStatus === "done" ? "todo" : "done";
        await updateTaskStatus(token, taskId, newStatus);
        loadTasks(token, activeContext);
      }
    });
  }

  async function handleDeleteTask(taskId: number) {
    if (!token) return;
    setConfirmAction({
      title: "Hapus Task",
      message: "Apakah Anda yakin ingin menghapus task ini beserta semua subtasknya?",
      isDanger: true,
      onConfirm: async () => {
        await deleteTask(token, taskId);
        loadTasks(token, activeContext);
      }
    });
  }

  async function handleEditTask(e: React.FormEvent, taskId: number) {
    e.preventDefault();
    if (!token) return;
    const target = e.target as typeof e.target & {
      title: { value: string };
      dueDate: { value: string };
      contextId: { value: string };
    };

    const newTitle = target.title.value;
    const newDueDate = target.dueDate.value || null;
    const ctxId = target.contextId.value ? Number(target.contextId.value) : null;

    setConfirmAction({
      title: "Simpan Perubahan Task",
      message: "Apakah Anda yakin ingin menyimpan perubahan pada task ini?",
      onConfirm: async () => {
        await updateTask(token, taskId, {
          title: newTitle,
          due_date: newDueDate,
          context_id: ctxId
        });
        setEditingTaskId(null);
        loadTasks(token, activeContext);
      }
    });
  }

  function extractDateOnly(dueDate: string | null): string {
    if (!dueDate) return "";
    return dueDate.slice(0, 10);
  }

  const today = new Date().toISOString().split("T")[0];
  
  const groups: Record<string, Task[]> = {
    overdue: [],
    today: [],
    upcoming: [],
    noDeadline: [],
  };

  for (const task of tasks) {
    if (!task.due_date) {
      groups.noDeadline.push(task);
      continue;
    }
    const dateOnly = extractDateOnly(task.due_date);
    if (dateOnly < today) groups.overdue.push(task);
    else if (dateOnly === today) groups.today.push(task);
    else groups.upcoming.push(task);
  }

  if (!token) return <p>Silakan login terlebih dahulu.</p>;

  const renderGroup = (title: string, groupTasks: Task[]) => {
    if (groupTasks.length === 0) return null;
    return (
      <section style={{ marginBottom: "2rem" }}>
        <h3 style={{ marginBottom: "1rem", color: "var(--text-dark)", borderBottom: "1px solid var(--surface-border)", paddingBottom: "0.5rem" }}>{title}</h3>
        <ul>
          {groupTasks.map((task) => {
            const isDone = task.status === "done";
            return (
              <li key={task.id} className={`task-item ${isDone ? "task-done" : ""}`}>
                {editingTaskId === task.id ? (
                  <form onSubmit={(e) => handleEditTask(e, task.id)} className="responsive-form" style={{ gap: "8px", marginBottom: "8px" }}>
                    <input type="text" name="title" defaultValue={task.title} required style={{ flex: 1, minWidth: "150px" }} />
                    <input type="date" name="dueDate" defaultValue={extractDateOnly(task.due_date)} />
                    <select name="contextId" defaultValue={task.context_id?.toString() || ""}>
                      <option value="">Tanpa context</option>
                      {contexts.map(ctx => <option key={ctx.id} value={ctx.id}>{ctx.name}</option>)}
                    </select>
                    <button type="submit">Save</button>
                    <button type="button" onClick={() => setEditingTaskId(null)}>Cancel</button>
                  </form>
                ) : (
                  <>
                    <div className="task-header" style={{ alignItems: "flex-start", gap: "12px" }}>
                      <input
                        type="checkbox"
                        checked={isDone}
                        onChange={() => handleToggleStatus(task.id, task.status)}
                        style={{ marginTop: "4px", width: "18px", height: "18px" }}
                      />
                      <span className="task-title" style={{ fontSize: "1.05rem", fontWeight: 500, color: "var(--text-dark)", flex: 1, lineHeight: "1.4" }}>
                        {task.title}
                      </span>
                    </div>
                    
                    <div className="task-meta-row" style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", marginTop: "12px", flexWrap: "wrap", gap: "12px" }}>
                      <div style={{ display: "flex", flexDirection: "column", gap: "8px", alignItems: "flex-start" }}>
                        {task.due_date && (
                          <span className="task-date" style={{ background: "#f1f5f9", padding: "4px 8px", borderRadius: "4px", fontSize: "0.8rem", color: "var(--text-muted)", display: "inline-flex", alignItems: "center" }}>
                            Tenggat: {extractDateOnly(task.due_date)}
                          </span>
                        )}
                        
                        {task.context_id && (
                          <span className="context-pill-small" style={{
                            display: "inline-flex",
                            alignItems: "center",
                            border: "1px solid",
                            padding: "3px 8px",
                            borderRadius: "12px",
                            fontSize: "0.8rem",
                            borderColor: contexts.find(c => c.id === task.context_id)?.color,
                            color: contexts.find(c => c.id === task.context_id)?.color
                          }}>
                            <span style={{ display: "inline-block", width: "6px", height: "6px", borderRadius: "50%", background: contexts.find(c => c.id === task.context_id)?.color, marginRight: "6px" }}></span>
                            {contexts.find(c => c.id === task.context_id)?.name}
                          </span>
                        )}
                      </div>

                      <div style={{ display: "flex", gap: "6px" }}>
                        <button onClick={() => setEditingTaskId(task.id)} className="icon-btn secondary" style={{ padding: "4px 10px", fontSize: "0.8rem" }}>Edit</button>
                        <button onClick={() => handleDeleteTask(task.id)} className="icon-btn danger" style={{ padding: "4px 10px", fontSize: "0.8rem" }}>Del</button>
                      </div>
                    </div>
                  </>
                )}
                <div style={{ marginTop: "16px", marginLeft: "0", borderTop: "1px solid #f1f5f9", paddingTop: "8px" }}>
                  <SubtaskList taskId={task.id} />
                </div>
              </li>
            );
          })}
        </ul>
      </section>
    );
  };

  return (
    <div className="container">
      <div className="app-header" style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <h1 style={{ margin: 0 }}>Task Tracker</h1>
        <button
          onClick={() => setConfirmAction({
            title: "Logout",
            message: "Apakah Anda yakin ingin keluar?",
            onConfirm: () => logout()
          })}
          className="btn danger"
          style={{ padding: "6px 12px", fontSize: "0.9rem" }}
        >
          Logout
        </button>
      </div>

      <div className="context-filters">
        <button
          className={`context-pill ${activeContext === null ? "active" : ""}`}
          onClick={() => setActiveContext(null)}
          style={{ background: activeContext === null ? "var(--primary)" : "" }}
        >
          Semua Context
        </button>
        {contexts.map((ctx) => (
          <button
            key={ctx.id}
            className={`context-pill ${activeContext === ctx.id ? "active" : ""}`}
            onClick={() => setActiveContext(ctx.id)}
            style={{ 
              borderColor: ctx.color, 
              background: activeContext === ctx.id ? ctx.color : "transparent",
              color: activeContext === ctx.id ? "#fff" : "var(--text)"
            }}
          >
            {ctx.name}
          </button>
        ))}
      </div>

      <div className="panel">
        <h3>Manajemen Context</h3>
        <ContextManager />
      </div>

      <div className="panel">
        <h3>Tambah Task Baru</h3>
        <TaskForm />
      </div>

      {loading ? (
        <p>Loading tasks...</p>
      ) : error ? (
        <p style={{ color: "red" }}>{error}</p>
      ) : (
        <div className="panel" id="task-groups" style={{ padding: "2rem", marginBottom: "3rem" }}>
          {renderGroup("Overdue", groups.overdue)}
          {renderGroup("Hari Ini", groups.today)}
          {renderGroup("Akan Datang", groups.upcoming)}
          {renderGroup("Tanpa Deadline", groups.noDeadline)}
          {tasks.length === 0 && <p style={{ color: "var(--text-muted)", textAlign: "center", margin: "1rem 0" }}>Tidak ada task yang aktif.</p>}
        </div>
      )}

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
