const API_BASE_URL = import.meta.env.PUBLIC_API_BASE_URL;

interface ApiOptions {
  method?: string;
  body?: unknown;
  token?: string;
}

async function apiFetch<T>(path: string, options: ApiOptions = {}): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (options.token) {
    headers["Authorization"] = `Bearer ${options.token}`;
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: options.method ?? "GET",
    headers,
    body: options.body ? JSON.stringify(options.body) : undefined,
  });

  if (!response.ok) {
    const errorBody = await response.json().catch(() => ({ error: "Unknown error" }));
    throw new Error(errorBody.error ?? `Request failed with status ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json();
}
export interface Task {
  id: number;
  context_id: number | null;
  title: string;
  description: string;
  status: string;
  due_date: string | null;
  created_at: string;
  completed_at: string | null;
}

export interface TaskListResponse {
  data: Task[];
  next_cursor: number | null;
}

export interface Subtask {
  id: number;
  task_id: number;
  title: string;
  is_done: boolean;
  sort_order: number;
  created_at: string;
}

export interface Context {
  id: number;
  name: string;
  color: string;
  created_at: string;
}
export function login(email: string, password: string): Promise<{ token: string }> {
  return apiFetch<{ token: string }>("/auth/login", {
    method: "POST",
    body: { email, password },
  });
}

export function register(email: string, password: string): Promise<{ message: string }> {
  return apiFetch<{ message: string }>("/auth/register", {
    method: "POST",
    body: { email, password },
  });
}
export function getTasks(token: string, params?: URLSearchParams): Promise<TaskListResponse> {
  const query = params ? `?${params.toString()}` : "";
  return apiFetch<TaskListResponse>(`/tasks${query}`, { token });
}

export function getTaskById(token: string, id: number): Promise<Task> {
  return apiFetch<Task>(`/tasks/${id}`, { token });
}

export function createTask(token: string, task: Partial<Task>): Promise<Task> {
  return apiFetch<Task>("/tasks", { method: "POST", body: task, token });
}

export function updateTask(token: string, id: number, task: Partial<Task>): Promise<void> {
  return apiFetch<void>(`/tasks/${id}`, { method: "PUT", body: task, token });
}

export function updateTaskStatus(token: string, id: number, status: string): Promise<void> {
  return apiFetch<void>(`/tasks/${id}/status`, { method: "PATCH", body: { status }, token });
}

export function deleteTask(token: string, id: number): Promise<void> {
  return apiFetch<void>(`/tasks/${id}`, { method: "DELETE", token });
}
export function createSubtask(token: string, taskId: number, title: string, sortOrder = 0): Promise<Subtask> {
  return apiFetch<Subtask>(`/tasks/${taskId}/subtasks`, {
    method: "POST",
    body: { title, sort_order: sortOrder },
    token,
  });
}

export function updateSubtaskDone(token: string, taskId: number, subtaskId: number, isDone: boolean): Promise<void> {
  return apiFetch<void>(`/tasks/${taskId}/subtasks/${subtaskId}`, {
    method: "PATCH",
    body: { is_done: isDone },
    token,
  });
}
export function getContexts(token: string): Promise<Context[]> {
  return apiFetch<Context[]>("/contexts", { token });
}

export function createContext(token: string, name: string, color: string): Promise<Context> {
  return apiFetch<Context>("/contexts", { method: "POST", body: { name, color }, token });
}

export function updateContext(token: string, id: number, name: string, color: string): Promise<Context> {
  return apiFetch<Context>(`/contexts/${id}`, { method: "PUT", body: { name, color }, token });
}

export function deleteContext(token: string, id: number): Promise<void> {
  return apiFetch<void>(`/contexts/${id}`, { method: "DELETE", token });
}

export async function downloadContextPDF(token: string, contextId: number): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/contexts/${contextId}/export`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!response.ok) throw new Error("Failed to download PDF");
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  window.open(url, '_blank');
  setTimeout(() => { window.URL.revokeObjectURL(url); }, 1000);
}

export function getSubtasksByTask(token: string, taskId: number): Promise<Subtask[]> {
  return apiFetch<Subtask[]>(`/tasks/${taskId}/subtasks`, { token });
}

export function updateSubtask(token: string, taskId: number, subtaskId: number, title: string): Promise<void> {
  return apiFetch<void>(`/tasks/${taskId}/subtasks/${subtaskId}`, {
    method: "PUT",
    body: { title },
    token,
  });
}

export function deleteSubtask(token: string, taskId: number, subtaskId: number): Promise<void> {
  return apiFetch<void>(`/tasks/${taskId}/subtasks/${subtaskId}`, {
    method: "DELETE",
    token,
  });
}