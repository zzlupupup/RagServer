import { Library } from 'lucide-react';
import type { UserRole } from '../types/user';
import { Alert, Button } from './ui';

export function AuthPage({
  mode,
  email,
  password,
  displayName,
  role,
  error,
  notice,
  onMode,
  onEmail,
  onPassword,
  onDisplayName,
  onRole,
  onSubmit,
}: {
  mode: 'login' | 'register';
  email: string;
  password: string;
  displayName: string;
  role: UserRole;
  error: string;
  notice: string;
  onMode: (mode: 'login' | 'register') => void;
  onEmail: (v: string) => void;
  onPassword: (v: string) => void;
  onDisplayName: (v: string) => void;
  onRole: (v: UserRole) => void;
  onSubmit: () => void;
}) {
  return (
    <main className="auth-shell">
      <section className="auth-panel">
        <div className="auth-brand">
          <span className="brand-mark">
            <Library size={20} />
          </span>
          <span>RagServer</span>
        </div>

        <div className="auth-tabs">
          <button className={mode === 'login' ? 'active' : ''} onClick={() => onMode('login')}>
            登录
          </button>
          <button className={mode === 'register' ? 'active' : ''} onClick={() => onMode('register')}>
            注册
          </button>
        </div>

        {error && <Alert tone="error">{error}</Alert>}
        {notice && <Alert tone="notice">{notice}</Alert>}

        <label className="field">
          <span>邮箱</span>
          <input value={email} onChange={(e) => onEmail(e.target.value)} autoComplete="email" />
        </label>
        <label className="field">
          <span>密码</span>
          <input
            value={password}
            onChange={(e) => onPassword(e.target.value)}
            type="password"
            autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
          />
        </label>
        {mode === 'register' && (
          <>
            <label className="field">
              <span>显示名称</span>
              <input value={displayName} onChange={(e) => onDisplayName(e.target.value)} />
            </label>
            <label className="field">
              <span>角色</span>
              <select value={role} onChange={(e) => onRole(e.target.value as UserRole)}>
                <option value="student">学生</option>
                <option value="teacher">教师</option>
              </select>
            </label>
          </>
        )}

        <Button variant="primary" full onClick={onSubmit}>
          {mode === 'login' ? '登录' : '创建账号'}
        </Button>
      </section>
    </main>
  );
}
