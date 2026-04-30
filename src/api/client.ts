const STORAGE_KEY = 'questionnaire_api_base';

// Default: use Vite proxy in dev, empty in production
const DEFAULT_BASE = import.meta.env.DEV ? '/api' : '';

export function getApiBase(): string {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored) return stored;
  return DEFAULT_BASE;
}

export function setApiBase(base: string): void {
  localStorage.setItem(STORAGE_KEY, base);
}

export function hasApiBase(): boolean {
  return !!getApiBase();
}

interface SubmissionPayload {
  answers: string[];
  scores: Record<string, number>;
  result: string;
  source?: string;
}

function apiUrl(path: string): string {
  const base = getApiBase();
  if (!base) throw new Error('API base not configured');
  const normalized = base.replace(/\/$/, '');
  // Auto-append /api if base doesn't already end with it
  const apiBase = normalized.endsWith('/api') ? normalized : `${normalized}/api`;
  return `${apiBase}${path}`;
}

export async function submitSubmission(payload: SubmissionPayload): Promise<{ success: boolean; id: string }> {
  const res = await fetch(apiUrl('/submit'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
  if (!res.ok) throw new Error(`Submit failed: ${res.status}`);
  return res.json();
}

export async function checkVersion(): Promise<string> {
  const res = await fetch(apiUrl('/version'));
  if (!res.ok) throw new Error(`Version check failed: ${res.status}`);
  const data = await res.json();
  return data.version;
}
