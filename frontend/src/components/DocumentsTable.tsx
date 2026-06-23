import { Trash2, Upload } from 'lucide-react';
import type { DocumentItem } from '../types/document';
import type { PaginationState } from '../types/pagination';
import { formatDateOnly, labelOf, statusTone } from '../lib/utils';
import { Badge, Button, EmptyState, Panel } from './ui';

export function DocumentsTable({
  documents,
  canUpload,
  onUpload,
  onDelete,
  pagination,
  onPage,
}: {
  documents: DocumentItem[];
  canUpload: boolean;
  onUpload: (file?: File) => void;
  onDelete: (doc: DocumentItem) => void;
  pagination: PaginationState;
  onPage: (page: number) => void;
}) {
  const pageCount = Math.max(1, Math.ceil(pagination.total / pagination.pageSize));

  return (
    <Panel
      className="documents-panel"
      title="文档"
      subtitle="当前用户可向可见的知识库上传文档。"
      actions={
        <label className={`upload${!canUpload ? ' disabled' : ''}`}>
          <Upload size={16} />
          上传
          <input
            type="file"
            accept=".pdf,.md,.markdown,.docx"
            disabled={!canUpload}
            onChange={(event) => {
              onUpload(event.target.files?.[0]);
              event.target.value = '';
            }}
          />
        </label>
      }
    >
      <div className="table-scroll">
        <table className="doc-table">
          <colgroup>
            <col />
            <col style={{ width: '88px' }} />
            <col style={{ width: '88px' }} />
            <col style={{ width: '88px' }} />
            <col style={{ width: '112px' }} />
            <col style={{ width: '52px' }} />
          </colgroup>
          <thead>
            <tr>
              <th>文件</th>
              <th className="th-center">状态</th>
              <th className="th-center">分块</th>
              <th className="th-center">上传者</th>
              <th className="th-center">上传时间</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {documents.map((doc) => (
              <tr key={doc.id}>
                <td className="cell-strong cell-file">{doc.original_filename}</td>
                <td className="cell-center">
                  <Badge tone={statusTone(doc.index_status)} dot>
                    {labelOf(doc.index_status)}
                  </Badge>
                </td>
                <td className="cell-center">{doc.chunk_count}</td>
                <td className="cell-center cell-ellipsis">{doc.uploaded_by_user_display_name || `用户 #${doc.uploaded_by_user_id}`}</td>
                <td className="cell-center">{formatDateOnly(doc.created_at)}</td>
                <td className="col-actions">
                  <button
                    className="icon-button danger"
                    title={doc.can_delete ? '删除文档' : '无删除权限'}
                    onClick={() => doc.can_delete && onDelete(doc)}
                    disabled={!doc.can_delete}
                  >
                    <Trash2 size={16} />
                  </button>
                </td>
              </tr>
            ))}
            {documents.length === 0 && (
              <tr>
                <td colSpan={6}>
                  <EmptyState row>暂无文档</EmptyState>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
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
    </Panel>
  );
}
