import type { ReactNode } from 'react';
import { AlertTriangle, Loader2, X } from 'lucide-react';
import type { Tone } from '../lib/utils';

type ButtonVariant = 'primary' | 'default' | 'danger';

export function Button({
  variant = 'default',
  icon,
  full,
  className,
  children,
  ...rest
}: {
  variant?: ButtonVariant;
  icon?: ReactNode;
  full?: boolean;
} & React.ButtonHTMLAttributes<HTMLButtonElement>) {
  const cls = ['button'];
  if (variant === 'primary') cls.push('primary');
  if (variant === 'danger') cls.push('danger');
  if (full) cls.push('full');
  if (className) cls.push(className);
  return (
    <button className={cls.join(' ')} {...rest}>
      {icon}
      {children}
    </button>
  );
}

export function IconButton({
  icon,
  danger,
  className,
  ...rest
}: {
  icon: ReactNode;
  danger?: boolean;
} & React.ButtonHTMLAttributes<HTMLButtonElement>) {
  const cls = ['icon-button'];
  if (danger) cls.push('danger');
  if (className) cls.push(className);
  return (
    <button className={cls.join(' ')} {...rest}>
      {icon}
    </button>
  );
}

export function Badge({ tone = 'gray', dot, children }: { tone?: Tone; dot?: boolean; children: ReactNode }) {
  return (
    <span className={`badge tone-${tone}`}>
      {dot && <span className="dot" />}
      {children}
    </span>
  );
}

export function Panel({
  title,
  subtitle,
  icon,
  actions,
  className,
  children,
}: {
  title: string;
  subtitle?: string;
  icon?: ReactNode;
  actions?: ReactNode;
  className?: string;
  children: ReactNode;
}) {
  return (
    <section className={className ? `panel ${className}` : 'panel'}>
      <div className="panel-title">
        <div className="heading">
          <h2>{title}</h2>
          {subtitle && <p>{subtitle}</p>}
        </div>
        {icon && !actions && <span className="panel-icon">{icon}</span>}
        {actions}
      </div>
      {children}
    </section>
  );
}

export function Alert({ tone, children }: { tone: 'error' | 'notice'; children: ReactNode }) {
  return <div className={tone}>{children}</div>;
}

export function EmptyState({
  icon,
  children,
  row,
}: {
  icon?: ReactNode;
  children: ReactNode;
  row?: boolean;
}) {
  return <div className={`empty${row ? ' empty-row' : ''}`}>{icon}
    <span>{children}</span>
  </div>;
}

export function Spinner({ size = 16 }: { size?: number }) {
  return <Loader2 size={size} className="spinner" />;
}

export function ConfirmDialog({
  open,
  title,
  message,
  confirmText = '确认',
  cancelText = '取消',
  loading,
  onConfirm,
  onCancel,
}: {
  open: boolean;
  title: string;
  message: ReactNode;
  confirmText?: string;
  cancelText?: string;
  loading?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}) {
  if (!open) return null;
  return (
    <div className="dialog-overlay" onClick={onCancel}>
      <div className="dialog" onClick={(e) => e.stopPropagation()}>
        <button className="dialog-close" onClick={onCancel} title="关闭">
          <X size={18} />
        </button>
        <div className="dialog-icon">
          <AlertTriangle size={22} />
        </div>
        <h3 className="dialog-title">{title}</h3>
        <div className="dialog-message">{message}</div>
        <div className="dialog-actions">
          <Button onClick={onCancel} disabled={loading}>
            {cancelText}
          </Button>
          <Button variant="danger" onClick={onConfirm} disabled={loading} icon={loading ? <Spinner size={15} /> : undefined}>
            {confirmText}
          </Button>
        </div>
      </div>
    </div>
  );
}
