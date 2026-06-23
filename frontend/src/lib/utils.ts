import type { User } from '../types/user';

export type Tone = 'green' | 'amber' | 'red' | 'gray';

const LABELS: Record<string, string> = {
  public: '公开',
  private: '私有',
  indexed: '已索引',
  indexing: '索引中',
  failed: '失败',
  pending: '待处理',
  active: '启用',
  disabled: '已禁用',
  teacher: '教师',
  student: '学生',
};

/** Map backend enum values to Chinese labels for display only. */
export function labelOf(value?: string) {
  if (!value) return value as string;
  return LABELS[value.toLowerCase()] ?? value;
}

/** Map an index/visibility/key status string to a badge tone. */
export function statusTone(status: string): Tone {
  const s = (status || '').toLowerCase();
  const green = ['indexed', 'active', 'public', 'ready'];
  const red = ['failed', 'disabled'];
  const amber = ['indexing', 'pending', 'private', 'processing'];
  if (green.includes(s)) return 'green';
  if (red.includes(s)) return 'red';
  if (amber.includes(s)) return 'amber';
  return 'gray';
}

export function messageOf(err: unknown, fallback: string) {
  return err instanceof Error ? err.message : fallback;
}

export function readStoredUser(): User | null {
  try {
    const raw = localStorage.getItem('ragserver_user');
    return raw ? (JSON.parse(raw) as User) : null;
  } catch {
    return null;
  }
}

export function formatSize(size: number) {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${Math.ceil(size / 1024)} KB`;
  return `${(size / 1024 / 1024).toFixed(1)} MB`;
}

export function formatDate(value?: string) {
  if (!value) return '-';
  return new Date(value).toLocaleString();
}

export function formatDateOnly(value?: string) {
  if (!value) return '-';
  return new Intl.DateTimeFormat('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
    .format(new Date(value))
    .replaceAll('/', '-');
}

export function initials(name?: string) {
  const base = (name || '').trim();
  if (!base) return '?';
  const parts = base.split(/\s+/);
  return (parts[0][0] + (parts[1]?.[0] ?? '')).toUpperCase();
}
