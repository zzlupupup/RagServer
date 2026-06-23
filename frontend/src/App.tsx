import { useEffect, useMemo, useState } from 'react';
import { authApi } from './api/auth';
import { clearAuthToken, getAuthToken, setAuthToken } from './api/client';
import { knowledgeBasesApi } from './api/knowledgeBases';
import { documentsApi } from './api/documents';
import { apiKeysApi } from './api/apiKeys';
import { usersApi } from './api/users';
import type { ApiKey } from './types/apiKey';
import type { DocumentItem } from './types/document';
import type { KnowledgeBase } from './types/knowledgeBase';
import type { PaginationState } from './types/pagination';
import type { User, UserRole } from './types/user';
import { messageOf, readStoredUser } from './lib/utils';
import { Alert, ConfirmDialog } from './components/ui';
import { AuthPage } from './components/AuthPage';
import { Sidebar } from './components/Sidebar';
import { DocumentsTable } from './components/DocumentsTable';
import { ApiKeysTable } from './components/ApiKeysTable';

type AuthMode = 'login' | 'register';
const pageSize = 10;

export function App() {
  const [token, setToken] = useState(getAuthToken());
  const [user, setUser] = useState<User | null>(() => readStoredUser());
  const [authMode, setAuthMode] = useState<AuthMode>('login');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [displayName, setDisplayName] = useState('');
  const [role, setRole] = useState<UserRole>('student');

  const [kbs, setKbs] = useState<KnowledgeBase[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [documents, setDocuments] = useState<DocumentItem[]>([]);
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [documentPagination, setDocumentPagination] = useState<PaginationState>({ page: 1, pageSize, total: 0 });
  const [apiKeyPagination, setApiKeyPagination] = useState<PaginationState>({ page: 1, pageSize, total: 0 });
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');

  const [newKbName, setNewKbName] = useState('');
  const [confirm, setConfirm] = useState<{ kind: 'kb' | 'doc' | 'key'; doc?: DocumentItem; key?: ApiKey } | null>(null);
  const [confirming, setConfirming] = useState(false);

  const selected = useMemo(() => kbs.find((kb) => kb.id === selectedId) || null, [kbs, selectedId]);
  const isTeacher = user?.role === 'teacher';
  const hasPending = useMemo(
    () => documents.some((d) => d.index_status === 'pending' || d.index_status === 'indexing'),
    [documents],
  );

  useEffect(() => {
    const onUnauthorized = () => signOut();
    window.addEventListener('ragserver:unauthorized', onUnauthorized);
    return () => window.removeEventListener('ragserver:unauthorized', onUnauthorized);
  }, []);

  useEffect(() => {
    if (token && user) void loadAll();
  }, [token, user?.id]);

  // Auto-refresh documents while any are in a non-terminal index status.
  useEffect(() => {
    if (!token || !user || !selectedId || !hasPending) return;
    const timer = setInterval(async () => {
      try {
        await loadDocuments(selectedId, documentPagination.page);
        const kbList = await knowledgeBasesApi.list();
        setKbs(kbList);
      } catch {
        /* ignore polling errors */
      }
    }, 3000);
    return () => clearInterval(timer);
  }, [token, user?.id, selectedId, hasPending, documentPagination.page]);

  async function loadDocuments(kbId: number | null, page = documentPagination.page) {
    if (!kbId) {
      setDocuments([]);
      setDocumentPagination({ page: 1, pageSize, total: 0 });
      return;
    }
    const resp = await documentsApi.list(kbId, page, pageSize);
    setDocuments(resp.items);
    setDocumentPagination({ page: resp.page, pageSize: resp.page_size, total: resp.total });
  }

  async function loadApiKeys(page = apiKeyPagination.page) {
    const resp = await apiKeysApi.list(page, pageSize);
    setApiKeys(resp.items);
    setApiKeyPagination({ page: resp.page, pageSize: resp.page_size, total: resp.total });
  }

  async function loadAll(nextSelectedId = selectedId) {
    if (!token || !user) return;
    try {
      setError('');
      const kbList = await knowledgeBasesApi.list();
      setKbs(kbList);
      const nextId = nextSelectedId && kbList.some((kb) => kb.id === nextSelectedId) ? nextSelectedId : kbList[0]?.id || null;
      setSelectedId(nextId);
      await loadDocuments(nextId, nextId === selectedId ? documentPagination.page : 1);

      if (user.role === 'teacher') {
        const [_, userList] = await Promise.all([loadApiKeys(), usersApi.list()]);
        setUsers(userList);
      } else {
        setApiKeys([]);
        setUsers([]);
        setApiKeyPagination({ page: 1, pageSize, total: 0 });
      }
    } catch (err) {
      setError(messageOf(err, '加载失败'));
    }
  }

  async function submitAuth() {
    try {
      setError('');
      setNotice('');
      if (authMode === 'register') {
        await authApi.register({ email, password, display_name: displayName, role });
        setNotice('注册成功,请登录。');
        setAuthMode('login');
        return;
      }
      const resp = await authApi.login({ email, password });
      setAuthToken(resp.token);
      localStorage.setItem('ragserver_user', JSON.stringify(resp.user));
      setToken(resp.token);
      setUser(resp.user);
    } catch (err) {
      setError(messageOf(err, '认证失败'));
    }
  }

  function signOut() {
    clearAuthToken();
    localStorage.removeItem('ragserver_user');
    setToken('');
    setUser(null);
    setKbs([]);
    setDocuments([]);
    setApiKeys([]);
    setUsers([]);
    setSelectedId(null);
  }

  async function createKb() {
    if (!newKbName.trim()) return;
    try {
      setError('');
      const kb = await knowledgeBasesApi.create({ name: newKbName });
      setNewKbName('');
      await loadAll(kb.id);
    } catch (err) {
      setError(messageOf(err, '创建知识库失败'));
    }
  }

  function editSelected(patch: Partial<KnowledgeBase>) {
    if (!selected) return;
    setKbs(kbs.map((kb) => (kb.id === selected.id ? { ...kb, ...patch } : kb)));
  }

  async function saveKb() {
    if (!selected) return;
    try {
      setError('');
      await knowledgeBasesApi.update(selected.id, { name: selected.name });
      await loadAll(selected.id);
    } catch (err) {
      setError(messageOf(err, '保存失败'));
    }
  }

  async function deleteKb() {
    if (!selected) return;
    setConfirm({ kind: 'kb' });
  }

  async function confirmDelete() {
    if (!confirm) return;
    setConfirming(true);
    try {
      setError('');
      if (confirm.kind === 'kb') {
        if (!selected) return;
        await knowledgeBasesApi.remove(selected.id);
        await loadAll(null);
      } else if (confirm.kind === 'doc' && confirm.doc) {
        await documentsApi.remove(confirm.doc.id);
        if (selected) await loadDocuments(selected.id, documentPagination.page);
      } else if (confirm.kind === 'key' && confirm.key) {
        await apiKeysApi.remove(confirm.key.id);
        await loadApiKeys(apiKeyPagination.page);
      }
      setConfirm(null);
    } catch (err) {
      setError(messageOf(err, '删除失败'));
    } finally {
      setConfirming(false);
    }
  }

  async function upload(file?: File) {
    if (!selected || !file) return;
    try {
      setError('');
      await documentsApi.upload(selected.id, file);
      setDocumentPagination((prev) => ({ ...prev, page: 1 }));
      await Promise.all([loadDocuments(selected.id, 1), knowledgeBasesApi.list().then(setKbs)]);
    } catch (err) {
      setError(messageOf(err, '上传失败'));
    }
  }

  async function deleteDocument(doc: DocumentItem) {
    setConfirm({ kind: 'doc', doc });
  }

  async function toggleKb(kb: KnowledgeBase) {
    if (selectedId === kb.id) {
      setSelectedId(null);
      setDocuments([]);
      setDocumentPagination({ page: 1, pageSize, total: 0 });
      return;
    }
    setSelectedId(kb.id);
    try {
      await loadDocuments(kb.id, 1);
    } catch (err) {
      setError(messageOf(err, '加载文档失败'));
    }
  }

  async function createKey(studentId: number) {
    try {
      setError('');
      await apiKeysApi.create({ bound_user_id: studentId });
      await loadApiKeys(1);
    } catch (err) {
      setError(messageOf(err, '创建密钥失败'));
    }
  }

  async function disableKey(key: ApiKey) {
    try {
      setError('');
      await apiKeysApi.disable(key.id);
      await loadApiKeys(apiKeyPagination.page);
    } catch (err) {
      setError(messageOf(err, '禁用密钥失败'));
    }
  }

  async function enableKey(key: ApiKey) {
    try {
      setError('');
      await apiKeysApi.enable(key.id);
      await loadApiKeys(apiKeyPagination.page);
    } catch (err) {
      setError(messageOf(err, '启用密钥失败'));
    }
  }

  function deleteKey(key: ApiKey) {
    setConfirm({ kind: 'key', key });
  }

  if (!token || !user) {
    return (
      <AuthPage
        mode={authMode}
        email={email}
        password={password}
        displayName={displayName}
        role={role}
        error={error}
        notice={notice}
        onMode={setAuthMode}
        onEmail={setEmail}
        onPassword={setPassword}
        onDisplayName={setDisplayName}
        onRole={setRole}
        onSubmit={submitAuth}
      />
    );
  }

  return (
    <main className="shell">
      <Sidebar
        user={user}
        kbs={kbs}
        selectedId={selectedId}
        newKbName={newKbName}
        onNewName={setNewKbName}
        onCreate={createKb}
        onSignOut={signOut}
        onToggle={(kb) => void toggleKb(kb)}
        onEdit={editSelected}
        onSave={saveKb}
        onDelete={deleteKb}
      />

      <section className={`content ${isTeacher ? 'teacher-content' : 'student-content'}`}>
        {error && <Alert tone="error">{error}</Alert>}

        <DocumentsTable
          documents={documents}
          canUpload={!!selected}
          onUpload={(file) => void upload(file)}
          onDelete={(doc) => void deleteDocument(doc)}
          pagination={documentPagination}
          onPage={(page) => selectedId && void loadDocuments(selectedId, page)}
        />

        {isTeacher && (
          <ApiKeysTable
            apiKeys={apiKeys}
            users={users}
            onCreate={(studentId) => void createKey(studentId)}
            onDisable={(key) => void disableKey(key)}
            onEnable={(key) => void enableKey(key)}
            onDelete={(key) => deleteKey(key)}
            pagination={apiKeyPagination}
            onPage={(page) => void loadApiKeys(page)}
          />
        )}
      </section>

      <ConfirmDialog
        open={!!confirm}
        loading={confirming}
        title={confirm?.kind === 'kb' ? '删除知识库' : confirm?.kind === 'key' ? '删除密钥' : '删除文档'}
        confirmText="删除"
        message={
          confirm?.kind === 'kb' && selected
            ? `确定删除知识库「${selected.name}」吗?该操作不可撤销,且会一并删除其中的所有文档。`
            : confirm?.kind === 'doc' && confirm.doc
              ? `确定删除文档「${confirm.doc.original_filename}」吗?此操作不可撤销。`
              : confirm?.kind === 'key' && confirm.key
                ? `确定删除为「${confirm.key.bound_user_display_name || `用户 #${confirm.key.bound_user_id}`}」签发的密钥吗?此操作不可撤销。`
                : null
        }
        onConfirm={() => void confirmDelete()}
        onCancel={() => setConfirm(null)}
      />
    </main>
  );
}
