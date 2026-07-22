export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("token");
}

export function setToken(token: string): void {
  if (typeof window === "undefined") return;
  localStorage.setItem("token", token);
}

export function requireAuth(): string {
  const token = getToken();
  if (!token) {
    window.location.href = "/login";
    throw new Error("Not authenticated");
  }
  return token;
}

export function logout(): void {
  localStorage.removeItem("token");
  window.location.href = "/login";
}