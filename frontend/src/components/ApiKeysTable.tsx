import { useEffect, useMemo, useState } from 'react';
import { Check, Copy, KeyRound, Plus, Power, PowerOff, Trash2, X } from 'lucide-react';
import { apiKeysApi } from '../api/apiKeys';
import type { ApiKey } from '../types/apiKey';
import type { PaginationState } from '../types/pagination';
import type { User } from '../types/user';
import { formatDate, initials, labelOf, statusTone } from '../lib/utils';
import { Badge, Button, EmptyState, IconButton, Panel } from './ui';

export function ApiKeysTable({
  apiKeys,
  users,
  onCreate,
  onDisable,
  onEnable,
  onDelete,
  pagination,
  onPage,
}: {
  apiKeys: ApiKey[];
  users: User[];
  onCreate: (studentId: number) => void;
  onDisable: (key: ApiKey) => void;
  onEnable: (key: ApiKey) => void;
  onDelete: (key: ApiKey) => void;
  pagination: PaginationState;
  onPage: (page: number) => void;
}) {
  const [copiedKeyId, setCopiedKeyId] = useState<number | null>(null);
  const [picking, setPicking] = useState(false);
  const [revealedKeys, setRevealedKeys] = useState<Record<number, string>>({});
  const [failedKeyIds, setFailedKeyIds] = useState<Set<number>>(() => new Set());

  const students = useMemo(() => users.filter((u) => u.role === 'student'), [users]);
  const latestKeyByStudent = useMemo(() => {
    const map = new Map<number, ApiKey>();
    for (const key of apiKeys) {
      if (!map.has(key.bound_user_id)) map.set(key.bound_user_id, key);
    }
    return map;
  }, [apiKeys]);
  const studentCards = useMemo(
    () =>
      students
        .map((student) => ({ student, key: latestKeyByStudent.get(student.id) }))
        .filter((item): item is { student: User; key: ApiKey } => !!item.key),
    [latestKeyByStudent, students],
  );
  const studentsWithoutKeys = useMemo(
    () => students.filter((student) => !latestKeyByStudent.has(student.id)),
    [latestKeyByStudent, students],
  );

  useEffect(() => {
    const visibleIds = new Set(studentCards.map(({ key }) => key.id));

    setRevealedKeys((prev) => {
      const next = Object.fromEntries(Object.entries(prev).filter(([id]) => visibleIds.has(Number(id))));
      return Object.keys(next).length === Object.keys(prev).length ? prev : next;
    });
    setFailedKeyIds((prev) => {
      const next = new Set([...prev].filter((id) => visibleIds.has(id)));
      return next.size === prev.size ? prev : next;
    });
  }, [studentCards]);

  useEffect(() => {
    let cancelled = false;
    const missingKeys = studentCards.filter(({ key }) => !revealedKeys[key.id] && !failedKeyIds.has(key.id));
    if (missingKeys.length === 0) return;

    void Promise.all(
      missingKeys.map(async ({ key }) => {
        try {
          const resp = await apiKeysApi.reveal(key.id);
          return { id: key.id, value: resp.api_key };
        } catch {
          return { id: key.id, value: '' };
        }
      }),
    ).then((results) => {
      if (cancelled) return;
      setRevealedKeys((prev) => {
        const next = { ...prev };
        for (const result of results) {
          if (result.value) next[result.id] = result.value;
        }
        return next;
      });
      setFailedKeyIds((prev) => {
        const next = new Set(prev);
        for (const result of results) {
          if (!result.value) next.add(result.id);
        }
        return next;
      });
    });

    return () => {
      cancelled = true;
    };
  }, [failedKeyIds, revealedKeys, studentCards]);

  async function copyKey(keyId: number) {
    const keyValue = revealedKeys[keyId];
    if (!keyValue) return;
    try {
      await navigator.clipboard.writeText(keyValue);
      setCopiedKeyId(keyId);
      setTimeout(() => setCopiedKeyId(null), 1500);
    } catch {
      /* ignore clipboard errors */
    }
  }

  function pick(studentId: number) {
    setPicking(false);
    onCreate(studentId);
  }

  const pageCount = Math.max(1, Math.ceil(pagination.total / pagination.pageSize));

  return (
    <Panel
      className="api-keys-panel"
      title="API 密钥"
      subtitle="为每个学生签发 MCP 密钥,密钥以其绑定学生身份运行。"
      actions={
        <Button
          icon={<Plus size={15} />}
          onClick={() => setPicking(true)}
          disabled={students.length === 0}
          title={students.length === 0 ? '暂无学生可签发' : '签发密钥'}
        >
          签发
        </Button>
      }
    >
      {studentCards.length === 0 ? (
        <div style={{ marginTop: 14 }}>
          <EmptyState row>暂无已签发的密钥,点击右上角「签发」为学生签发。</EmptyState>
        </div>
      ) : (
        <div className="key-cards-scroll">
          <div className="key-cards">
            {studentCards.map(({ student, key }) => {
              const disabled = key.status === 'disabled';
              const revealedKey = revealedKeys[key.id];
              const secretText = revealedKey || (failedKeyIds.has(key.id) ? '密钥加载失败' : '密钥加载中...');
              return (
                <div key={student.id} className="student-card">
                  <div className="student-card-head">
                    <div className="student-card-title">
                      <span className="key-card-avatar">{initials(student.display_name)}</span>
                      <div className="key-card-name">
                        <div className="student-name-row">
                          <strong>{student.display_name}</strong>
                          <Badge tone={statusTone(key.status)} dot>
                            {labelOf(key.status)}
                          </Badge>
                        </div>
                        <span>{student.email}</span>
                      </div>
                    </div>
                    <div className="student-key-actions">
                      <IconButton
                        className={disabled ? 'student-enable-action' : undefined}
                        icon={disabled ? <Power size={15} /> : <PowerOff size={15} />}
                        title={disabled ? '启用密钥' : '禁用密钥'}
                        onClick={() => (disabled ? onEnable(key) : onDisable(key))}
                      />
                      <IconButton danger icon={<Trash2 size={15} />} title="删除密钥" onClick={() => onDelete(key)} />
                    </div>
                  </div>

                  <div className={`student-key${disabled ? ' disabled' : ''}`}>
                    <span className="student-key-meta">最近使用:{formatDate(key.last_used_at)}</span>
                    <div className="student-secret-row">
                      <KeyRound size={15} />
                      <code className="student-secret">{secretText}</code>
                      <IconButton
                        icon={copiedKeyId === key.id ? <Check size={15} /> : <Copy size={15} />}
                        title="复制密钥"
                        onClick={() => void copyKey(key.id)}
                        disabled={!revealedKey}
                      />
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}
      <div className="pagination">
        <span>
          共 {pagination.total} 条 · 第 {pagination.page} / {pageCount} 页
        </span>
        <div className="pagination-actions">
          <Button disabled={pagination.page <= 1} onClick={() => onPage(pagination.page - 1)}>
            上一页
          </Button>
          <Button disabled={pagination.page >= pageCount} onClick={() => onPage(pagination.page + 1)}>
            下一页
          </Button>
        </div>
      </div>

      {picking && (
        <div className="dialog-overlay" onClick={() => setPicking(false)}>
          <div className="dialog picker-dialog" onClick={(e) => e.stopPropagation()}>
            <button className="dialog-close" onClick={() => setPicking(false)} title="关闭">
              <X size={18} />
            </button>
            <h3 className="dialog-title">选择学生签发密钥</h3>
            <div className="picker-list">
              {students.map((s) => {
                const hasKey = latestKeyByStudent.has(s.id);
                return (
                  <button key={s.id} className="picker-item" onClick={() => pick(s.id)} disabled={hasKey}>
                    <span className="key-card-avatar">{initials(s.display_name)}</span>
                    <span className="picker-item-name">{s.display_name}</span>
                    <span className={hasKey ? 'picker-item-status signed' : 'picker-item-status'}>{hasKey ? '已签发' : '可签发'}</span>
                  </button>
                );
              })}
            </div>
          </div>
        </div>
      )}
    </Panel>
  );
}
