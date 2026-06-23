CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  email VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  display_name VARCHAR(128) NOT NULL,
  role VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  last_login_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_users_email (email),
  INDEX idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS knowledge_bases (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  owner_user_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(128) NOT NULL,
  description TEXT NULL,
  visibility VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  created_at DATETIME(3) NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  deleted_at DATETIME(3) NULL,
  INDEX idx_knowledge_bases_owner_user_id (owner_user_id),
  INDEX idx_knowledge_bases_visibility (visibility),
  INDEX idx_knowledge_bases_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS documents (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  kb_id BIGINT UNSIGNED NOT NULL,
  uploaded_by_user_id BIGINT UNSIGNED NOT NULL,
  filename VARCHAR(255) NOT NULL,
  original_filename VARCHAR(255) NOT NULL,
  file_ext VARCHAR(32) NOT NULL,
  mime_type VARCHAR(128) NOT NULL,
  file_size BIGINT NOT NULL,
  file_hash VARCHAR(128) NOT NULL,
  storage_path TEXT NOT NULL,
  index_status VARCHAR(32) NOT NULL DEFAULT 'pending',
  index_error TEXT NULL,
  chunk_count INT NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  deleted_at DATETIME(3) NULL,
  INDEX idx_documents_kb_id (kb_id),
  INDEX idx_documents_uploaded_by_user_id (uploaded_by_user_id),
  INDEX idx_documents_deleted_at (deleted_at),
  INDEX idx_documents_file_hash (file_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS document_chunks (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  kb_id BIGINT UNSIGNED NOT NULL,
  document_id BIGINT UNSIGNED NOT NULL,
  chunk_index INT NOT NULL,
  content MEDIUMTEXT NOT NULL,
  content_hash VARCHAR(128) NOT NULL,
  token_count INT NOT NULL DEFAULT 0,
  redis_key VARCHAR(255) NOT NULL,
  metadata_json JSON NULL,
  created_at DATETIME(3) NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  INDEX idx_document_chunks_kb_id (kb_id),
  INDEX idx_document_chunks_document_id (document_id),
  UNIQUE KEY uk_document_chunks_redis_key (redis_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ingestion_jobs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  kb_id BIGINT UNSIGNED NOT NULL,
  document_id BIGINT UNSIGNED NOT NULL,
  created_by_user_id BIGINT UNSIGNED NOT NULL,
  job_type VARCHAR(32) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  error_message TEXT NULL,
  started_at DATETIME(3) NULL,
  finished_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  INDEX idx_ingestion_jobs_document_id (document_id),
  INDEX idx_ingestion_jobs_created_by_user_id (created_by_user_id),
  INDEX idx_ingestion_jobs_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS api_keys (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  created_by_user_id BIGINT UNSIGNED NOT NULL,
  bound_user_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(128) NOT NULL,
  key_hash VARCHAR(128) NOT NULL,
  encrypted_key TEXT NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  last_used_at DATETIME(3) NULL,
  created_at DATETIME(3) NOT NULL,
  updated_at DATETIME(3) NOT NULL,
  deleted_at DATETIME(3) NULL,
  UNIQUE KEY uk_api_keys_key_hash (key_hash),
  INDEX idx_api_keys_created_by_user_id (created_by_user_id),
  INDEX idx_api_keys_bound_user_id (bound_user_id),
  INDEX idx_api_keys_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
