import { useEffect, useMemo, useState } from 'react';
import { KeyRound, Library, Plus, RefreshCw, Save, Trash2, Upload } from 'lucide-react';
import { setAdminToken, getAdminToken } from './api/client';
import { knowledgeBasesApi } from './api/knowledgeBases';
import { documentsApi } from './api/documents';
import { apiKeysApi } from './api/apiKeys';
import type { KnowledgeBase } from './types/knowledgeBase';
import type { DocumentItem } from './types/document';
import type { ApiKey } from './types/apiKey';

export function App() {
  const [token, setToken] = useState(getAdminToken());
  const [kbs, setKbs] = useState<KnowledgeBase[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [documents, setDocuments] = useState<DocumentItem[]>([]);
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [error, setError] = useState('');
  const [newKbName, setNewKbName] = useState('');
  const [newKbDescription, setNewKbDescription] = useState('');
  const [keyName, setKeyName] = useState('');
  const [revealedKey, setRevealedKey] = useState('');

  const selected = useMemo(() => kbs.find((kb) => kb.id === selectedId) || null, [kbs, selectedId]);

  async function loadAll() {
    if (!token) return;
    try {
      setError('');
      const [kbList, keyList] = await Promise.all([knowledgeBasesApi.list(), apiKeysApi.list()]);
      setKbs(kbList);
      setApiKeys(keyList);
      const nextSelected = selectedId || kbList[0]?.id || null;
      setSelectedId(nextSelected);
      if (nextSelected) setDocuments(await documentsApi.list(nextSelected));
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
    }
  }

  useEffect(() => {
    void loadAll();
  }, [token]);

  async function createKb() {
    if (!newKbName.trim()) return;
    const kb = await knowledgeBasesApi.create({ name: newKbName, description: newKbDescription });
    setNewKbName('');
    setNewKbDescription('');
    setSelectedId(kb.id);
    await loadAll();
  }

  async function saveKb() {
    if (!selected) return;
    await knowledgeBasesApi.update(selected.id, { name: selected.name, description: selected.description });
    await loadAll();
  }

  async function upload(file?: File) {
    if (!selected || !file) return;
    await documentsApi.upload(selected.id, file);
    setDocuments(await documentsApi.list(selected.id));
    await loadAll();
  }

  async function createKey() {
    if (!keyName.trim()) return;
    const created = await apiKeysApi.create(keyName);
    setKeyName('');
    setRevealedKey(created.api_key || '');
    await loadAll();
  }

  return (
    <main className="shell">
      <aside className="sidebar">
        <div className="brand">
          <Library size={22} />
          <span>RagServer</span>
        </div>
        <label className="field">
          <span>Admin Token</span>
          <input
            value={token}
            type="password"
            onChange={(event) => {
              setToken(event.target.value);
              setAdminToken(event.target.value);
            }}
            placeholder="输入 ADMIN_TOKEN"
          />
        </label>
        <button className="button" onClick={loadAll}>
          <RefreshCw size={16} />
          刷新
        </button>
        <nav className="kb-list">
          {kbs.map((kb) => (
            <button
              className={kb.id === selectedId ? 'kb-item active' : 'kb-item'}
              key={kb.id}
              onClick={async () => {
                setSelectedId(kb.id);
                setDocuments(await documentsApi.list(kb.id));
              }}
            >
              <strong>{kb.name}</strong>
              <span>{kb.document_count} 个文档</span>
            </button>
          ))}
        </nav>
      </aside>

      <section className="content">
        {error && <div className="error">{error}</div>}

        <section className="panel">
          <div className="panel-title">
            <h2>知识库</h2>
            <button className="icon-button" onClick={createKb} title="创建知识库">
              <Plus size={18} />
            </button>
          </div>
          <div className="form-grid">
            <input value={newKbName} onChange={(e) => setNewKbName(e.target.value)} placeholder="新知识库名称" />
            <input value={newKbDescription} onChange={(e) => setNewKbDescription(e.target.value)} placeholder="描述" />
          </div>

          {selected && (
            <div className="detail">
              <input
                value={selected.name}
                onChange={(e) => setKbs(kbs.map((kb) => (kb.id === selected.id ? { ...kb, name: e.target.value } : kb)))}
              />
              <textarea
                value={selected.description}
                onChange={(e) =>
                  setKbs(kbs.map((kb) => (kb.id === selected.id ? { ...kb, description: e.target.value } : kb)))
                }
              />
              <div className="actions">
                <button className="button" onClick={saveKb}>
                  <Save size={16} />
                  保存
                </button>
                <button
                  className="button danger"
                  onClick={async () => {
                    await knowledgeBasesApi.remove(selected.id);
                    setSelectedId(null);
                    await loadAll();
                  }}
                >
                  <Trash2 size={16} />
                  删除
                </button>
              </div>
            </div>
          )}
        </section>

        <section className="panel">
          <div className="panel-title">
            <h2>文档</h2>
            <label className="upload">
              <Upload size={16} />
              上传
              <input
                type="file"
                accept=".pdf,.md,.markdown,.docx"
                onChange={(event) => void upload(event.target.files?.[0])}
              />
            </label>
          </div>
          <table>
            <thead>
              <tr>
                <th>文件</th>
                <th>状态</th>
                <th>分块</th>
                <th>大小</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {documents.map((doc) => (
                <tr key={doc.id}>
                  <td>{doc.original_filename}</td>
                  <td>
                    <span className={`badge ${doc.index_status}`}>{doc.index_status}</span>
                    {doc.index_error && <small>{doc.index_error}</small>}
                  </td>
                  <td>{doc.chunk_count}</td>
                  <td>{Math.ceil(doc.file_size / 1024)} KB</td>
                  <td>
                    <button
                      className="icon-button danger"
                      title="删除文档"
                      onClick={async () => {
                        await documentsApi.remove(doc.id);
                        if (selected) setDocuments(await documentsApi.list(selected.id));
                      }}
                    >
                      <Trash2 size={16} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>

        <section className="panel">
          <div className="panel-title">
            <h2>API Key</h2>
            <KeyRound size={18} />
          </div>
          <div className="form-grid">
            <input value={keyName} onChange={(e) => setKeyName(e.target.value)} placeholder="Key 名称" />
            <button className="button" onClick={createKey}>
              <Plus size={16} />
              生成
            </button>
          </div>
          {revealedKey && <code className="secret">{revealedKey}</code>}
          <table>
            <thead>
              <tr>
                <th>名称</th>
                <th>状态</th>
                <th>最近使用</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {apiKeys.map((key) => (
                <tr key={key.id}>
                  <td>{key.name}</td>
                  <td><span className={`badge ${key.status}`}>{key.status}</span></td>
                  <td>{key.last_used_at || '-'}</td>
                  <td className="row-actions">
                    <button className="button" onClick={async () => setRevealedKey((await apiKeysApi.reveal(key.id)).api_key)}>
                      查看
                    </button>
                    <button className="button" onClick={async () => { await apiKeysApi.disable(key.id); await loadAll(); }}>
                      禁用
                    </button>
                    <button className="icon-button danger" onClick={async () => { await apiKeysApi.remove(key.id); await loadAll(); }}>
                      <Trash2 size={16} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      </section>
    </main>
  );
}

