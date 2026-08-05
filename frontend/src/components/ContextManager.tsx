import { useState, useEffect } from "react";
import { getContexts, createContext, updateContext, deleteContext, downloadContextPDF, type Context } from "../lib/api";
import { getToken } from "../lib/auth";
import { ConfirmModal } from "./ConfirmModal";

export default function ContextManager() {
  const [token, setToken] = useState<string | null>(null);
  const [contexts, setContexts] = useState<Context[]>([]);
  const [name, setName] = useState("");
  const [color, setColor] = useState("#3B82F6");
  const [error, setError] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [confirmAction, setConfirmAction] = useState<{title: string, message: string, isDanger?: boolean, onConfirm: () => void} | null>(null);

  useEffect(() => {
    const t = getToken();
    setToken(t);
    if (t) loadContexts(t);
  }, []);

  async function loadContexts(t: string) {
    try {
      const data = await getContexts(t);
      setContexts(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load contexts");
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;

    setConfirmAction({
      title: "Buat Context",
      message: "Apakah Anda yakin ingin membuat context baru ini?",
      onConfirm: async () => {
        try {
          await createContext(token, name, color);
          setName("");
          await loadContexts(token);
          window.dispatchEvent(new Event("context-created"));
        } catch (err) {
          setError(err instanceof Error ? err.message : "Failed to create context");
        }
      }
    });
  }

  async function handleEditSubmit(e: React.FormEvent, id: number) {
    e.preventDefault();
    if (!token) return;
    const target = e.target as typeof e.target & {
      name: { value: string };
      color: { value: string };
    };

    const newName = target.name.value;
    const newColor = target.color.value;

    setConfirmAction({
      title: "Simpan Perubahan Context",
      message: "Apakah Anda yakin ingin menyimpan perubahan pada context ini?",
      onConfirm: async () => {
        try {
          await updateContext(token, id, newName, newColor);
          setEditingId(null);
          await loadContexts(token);
          window.dispatchEvent(new Event("context-created"));
        } catch (err) {
          setError(err instanceof Error ? err.message : "Failed to update context");
        }
      }
    });
  }

  async function handleDelete(id: number) {
    if (!token) return;
    
    setConfirmAction({
      title: "Hapus Context",
      message: "Hapus context ini? Semua task dan subtask di dalamnya akan otomatis ikut terhapus.",
      isDanger: true,
      onConfirm: async () => {
        try {
          await deleteContext(token, id);
          await loadContexts(token);
          window.dispatchEvent(new Event("context-created"));
        } catch (err) {
          setError(err instanceof Error ? err.message : "Failed to delete context");
        }
      }
    });
  }

  async function handleDownloadPDF(id: number) {
    if (!token) return;
    setConfirmAction({
      title: "Preview PDF",
      message: "Tampilkan preview PDF untuk context ini?",
      onConfirm: async () => {
        try {
          await downloadContextPDF(token, id);
        } catch (err) {
          setError(err instanceof Error ? err.message : "Failed to download PDF");
        }
      }
    });
  }

  if (!token) return null;

  return (
    <div>
      <form onSubmit={handleSubmit} className="responsive-form" style={{ marginBottom: "16px" }}>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Nama context (mis. PT IGP, Kuliah)"
          required
          style={{ flex: 1 }}
        />
        <input
          type="color"
          value={color}
          onChange={(e) => setColor(e.target.value)}
          title="Pilih warna context"
        />
        <button type="submit">Tambah Context</button>
      </form>
      {error && <p style={{ color: "var(--danger)" }}>{error}</p>}
      <ul style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
        {contexts.map((ctx) => (
          <li key={ctx.id} style={{ display: "flex", alignItems: "center", gap: "12px", background: "var(--item-bg)", padding: "10px 14px", borderRadius: "var(--radius-sm)", border: "1px solid var(--surface-border)" }}>
            {editingId === ctx.id ? (
              <form 
                className="inline-form"
                onSubmit={(e) => handleEditSubmit(e, ctx.id)}
              >
                <input type="text" name="name" defaultValue={ctx.name} required style={{ flex: 1 }} />
                <input type="color" name="color" defaultValue={ctx.color} title="Pilih warna context" />
                <button type="submit" className="icon-btn">Save</button>
                <button type="button" className="icon-btn secondary" onClick={() => setEditingId(null)}>Cancel</button>
              </form>
            ) : (
              <>
                <div style={{ width: "12px", height: "12px", borderRadius: "50%", backgroundColor: ctx.color, boxShadow: `0 0 8px ${ctx.color}40` }}></div>
                <span style={{ fontWeight: "500", flex: 1, color: "var(--text-dark)" }}>{ctx.name}</span>
                <button onClick={() => handleDownloadPDF(ctx.id)} className="icon-btn secondary" title="Download PDF Report">PDF</button>
                <button onClick={() => setEditingId(ctx.id)} className="icon-btn secondary">Edit</button>
                <button onClick={() => handleDelete(ctx.id)} className="icon-btn danger">Del</button>
              </>
            )}
          </li>
        ))}
      </ul>

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