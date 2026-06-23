import { useState } from 'react';
import { Check, Library, LogOut, Pencil, Plus, Trash2, X } from 'lucide-react';
import type { KnowledgeBase } from '../types/knowledgeBase';
import type { User } from '../types/user';
import { initials, labelOf } from '../lib/utils';
import { IconButton } from './ui';

export function Sidebar({
  user,
  kbs,
  selectedId,
  newKbName,
  onNewName,
  onCreate,
  onSignOut,
  onToggle,
  onEdit,
  onSave,
  onDelete,
}: {
  user: User;
  kbs: KnowledgeBase[];
  selectedId: number | null;
  newKbName: string;
  onNewName: (v: string) => void;
  onCreate: () => void;
  onSignOut: () => void;
  onToggle: (kb: KnowledgeBase) => void;
  onEdit: (patch: Partial<KnowledgeBase>) => void;
  onSave: () => void;
  onDelete: () => void;
}) {
  const [editingId, setEditingId] = useState<number | null>(null);
  const [draftName, setDraftName] = useState('');
  const publicKbs = kbs.filter((kb) => kb.visibility === 'public');
  const privateKbs = kbs.filter((kb) => kb.visibility !== 'public');

  function startEdit(kb: KnowledgeBase) {
    setEditingId(kb.id);
    setDraftName(kb.name);
  }

  function commitEdit() {
    const kb = kbs.find((k) => k.id === editingId);
    if (kb && draftName.trim()) {
      onEdit({ name: draftName.trim() });
      onSave();
    }
    setEditingId(null);
  }

  function cancelEdit() {
    setEditingId(null);
  }

  function renderKbItem(kb: KnowledgeBase) {
    const open = kb.id === selectedId;
    const editing = kb.id === editingId;
    const ownerName = kb.owner_user_display_name || `用户 #${kb.owner_user_id}`;
    return (
      <div key={kb.id} className={open ? 'kb-item active' : 'kb-item'}>
        <div className="kb-item-row">
          {editing ? (
            <div className="kb-item-name editing">
              <input
                autoFocus
                value={draftName}
                onChange={(e) => setDraftName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') commitEdit();
                  if (e.key === 'Escape') cancelEdit();
                }}
              />
              <IconButton icon={<Check size={15} />} title="保存" onClick={commitEdit} />
              <IconButton icon={<X size={15} />} title="取消" onClick={cancelEdit} />
            </div>
          ) : (
            <>
              <button className="kb-item-name" onClick={() => onToggle(kb)}>
                <strong>{kb.name}</strong>
                {kb.can_manage && (
                  <span
                    className="kb-edit-trigger"
                    title="重命名"
                    onClick={(e) => {
                      e.stopPropagation();
                      startEdit(kb);
                    }}
                  >
                    <Pencil size={13} />
                  </span>
                )}
              </button>
              {open && kb.can_manage && (
                <span
                  className="kb-delete-trigger"
                  title="删除"
                  onClick={(e) => {
                    e.stopPropagation();
                    onDelete();
                  }}
                >
                  <Trash2 size={14} />
                </span>
              )}
            </>
          )}
        </div>
        <button className="kb-item-sub" onClick={() => onToggle(kb)} disabled={editing}>
          <span>
            {ownerName} · {kb.document_count} 篇文档
          </span>
        </button>
      </div>
    );
  }

  return (
    <aside className="sidebar">
      <div className="brand">
        <span className="brand-mark">
          <Library size={18} />
        </span>
        <span>RagServer</span>
      </div>

      <div className="identity">
        <span className="avatar">{initials(user.display_name)}</span>
        <div className="meta">
          <strong>{user.display_name}</strong>
          <span>
            {user.email} · {labelOf(user.role)}
          </span>
        </div>
        <IconButton icon={<LogOut size={16} />} title="退出登录" onClick={onSignOut} />
      </div>

      <div className="kb-group">
        <div className="sidebar-section-label">知识库</div>
        <nav className="kb-list">
          {kbs.length === 0 && (
            <div className="muted" style={{ padding: '4px' }}>
              暂无知识库
            </div>
          )}
          {kbs.length > 0 && (
            <>
              <div className="kb-list-section public-section">
                <div className="kb-list-heading">公开</div>
                <div className="kb-list-section-body">
                  {publicKbs.length > 0 ? publicKbs.map(renderKbItem) : <div className="kb-list-empty">暂无公开知识库</div>}
                </div>
              </div>
              <div className="kb-list-section private-section">
                <div className="kb-list-heading">私有</div>
                <div className="kb-list-section-body">
                  {privateKbs.length > 0 ? privateKbs.map(renderKbItem) : <div className="kb-list-empty">暂无私有知识库</div>}
                </div>
              </div>
            </>
          )}
        </nav>
      </div>

      <div className="kb-create">
        <input
          value={newKbName}
          onChange={(e) => onNewName(e.target.value)}
          placeholder="新建知识库"
          onKeyDown={(e) => {
            if (e.key === 'Enter') onCreate();
          }}
        />
        <IconButton
          icon={<Plus size={18} />}
          title="新建知识库"
          onClick={onCreate}
          disabled={!newKbName.trim()}
        />
      </div>
    </aside>
  );
}
