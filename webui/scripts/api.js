// Thin fetch wrappers. No game rules here: the Go engine decides everything.

export async function getHealth(signal) {
  const res = await fetch("/api/health", { cache: "no-store", signal });
  if (!res.ok) {
    throw new Error(`health check failed: HTTP ${res.status}`);
  }
  return res.json();
}
