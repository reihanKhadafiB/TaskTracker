import { useState, useEffect } from "react";
import { createTask } from "../lib/api";
import { getToken } from "../lib/auth";

interface Props {
  onTaskCreated: () => void;
}

export default function TaskForm(){
  const [token, setToken] = useState<string | null>(null);
  const [title, setTitle] = useState("");
  const [dueDate, setDueDate] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setToken(getToken());
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;

    setSubmitting(true);
    setError(null);

    try {
      await createTask(token, {
        title,
        due_date: dueDate || null,
        status: "todo",
      });
      setTitle("");
      setDueDate("");
      window.dispatchEvent(new Event("task-created"));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create task");
    } finally {
      setSubmitting(false);
    }
  }

  if (!token) return null;
  
  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="Task title"
        required
      />
      <input
        type="date"
        value={dueDate}
        onChange={(e) => setDueDate(e.target.value)}
      />
      <button type="submit" disabled={submitting}>
        {submitting ? "Adding..." : "Add Task"}
      </button>
      {error && <p style={{ color: "red" }}>{error}</p>}
    </form>
  );
}