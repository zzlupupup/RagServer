package rag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	einoredisindexer "github.com/cloudwego/eino-ext/components/indexer/redis"
	einoredisretriever "github.com/cloudwego/eino-ext/components/retriever/redis"
	einoembedding "github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
	"ragserver/backend/internal/dto"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/rag/parser"
	appstorage "ragserver/backend/internal/storage"
)

type Pipeline struct {
	redis     *redis.Client
	embedder  einoembedding.Embedder
	parser    *parser.Parser
	splitter  Splitter
	indexName string
	keyPrefix string
}

func NewPipeline(redisClient *redis.Client, embedder einoembedding.Embedder) *Pipeline {
	return &Pipeline{
		redis:     redisClient,
		embedder:  embedder,
		parser:    parser.New(),
		splitter:  NewSplitter(800, 120),
		indexName: appstorage.RedisIndexName,
		keyPrefix: appstorage.RedisKeyPrefix,
	}
}

func (p *Pipeline) BuildChunks(ctx context.Context, doc model.Document, data []byte) ([]*schema.Document, []model.DocumentChunk, error) {
	parsed, err := p.parser.Parse(ctx, doc.OriginalFilename, data)
	if err != nil {
		return nil, nil, err
	}
	chunks := p.splitter.Split(parsed)
	einoDocs := make([]*schema.Document, 0, len(chunks))
	dbChunks := make([]model.DocumentChunk, 0, len(chunks))
	for idx, chunk := range chunks {
		redisKey := fmt.Sprintf("%s%d:%d:%d", p.keyPrefix, doc.KBID, doc.ID, idx)
		contentHash := hashString(chunk.Content)
		meta := map[string]any{
			"kb_id":       doc.KBID,
			"document_id": doc.ID,
			"chunk_index": idx,
			"filename":    doc.OriginalFilename,
			"file_ext":    doc.FileExt,
		}
		metaJSON, _ := json.Marshal(meta)
		einoDoc := &schema.Document{
			ID:      redisKey[len(p.keyPrefix):],
			Content: chunk.Content,
			MetaData: map[string]any{
				"kb_id":       strconv.FormatUint(doc.KBID, 10),
				"document_id": strconv.FormatUint(doc.ID, 10),
				"chunk_id":    fmt.Sprintf("%d", idx),
				"chunk_index": idx,
				"filename":    doc.OriginalFilename,
			},
		}
		einoDocs = append(einoDocs, einoDoc)
		dbChunks = append(dbChunks, model.DocumentChunk{
			KBID:         doc.KBID,
			DocumentID:   doc.ID,
			ChunkIndex:   idx,
			Content:      chunk.Content,
			ContentHash:  contentHash,
			TokenCount:   len([]rune(chunk.Content)),
			RedisKey:     redisKey,
			MetadataJSON: datatypes.JSON(metaJSON),
		})
	}
	return einoDocs, dbChunks, nil
}

func (p *Pipeline) Index(ctx context.Context, docs []*schema.Document) error {
	indexer, err := einoredisindexer.NewIndexer(ctx, &einoredisindexer.IndexerConfig{
		Client:           p.redis,
		KeyPrefix:        p.keyPrefix,
		Embedding:        p.embedder,
		BatchSize:        10,
		DocumentToHashes: p.documentToHashes,
	})
	if err != nil {
		return err
	}
	_, err = indexer.Store(ctx, docs)
	return err
}

func (p *Pipeline) Search(ctx context.Context, kbID uint64, query string, topK int) ([]dto.SearchItem, error) {
	if topK <= 0 {
		topK = 5
	}
	if topK > 20 {
		topK = 20
	}
	r, err := einoredisretriever.NewRetriever(ctx, &einoredisretriever.RetrieverConfig{
		Client:       p.redis,
		Index:        p.indexName,
		VectorField:  "vector_content",
		ReturnFields: []string{"content", "filename", "document_id", "chunk_id", "chunk_index", einoredisretriever.SortByDistanceAttributeName},
		TopK:         topK,
		Embedding:    p.embedder,
		DocumentConverter: func(ctx context.Context, doc redis.Document) (*schema.Document, error) {
			fields := doc.Fields
			converted := &schema.Document{
				ID:      doc.ID,
				Content: fields["content"],
				MetaData: map[string]any{
					"filename":    fields["filename"],
					"document_id": fields["document_id"],
					"chunk_id":    fields["chunk_id"],
					"chunk_index": fields["chunk_index"],
				},
			}
			return converted.WithScore(scoreFromFields(fields, doc.Score)), nil
		},
	})
	if err != nil {
		return nil, err
	}
	docs, err := r.Retrieve(ctx, query, retriever.WithTopK(topK), einoredisretriever.WithFilterQuery(fmt.Sprintf("@kb_id:{%d}", kbID)))
	if err != nil {
		return nil, err
	}
	items := make([]dto.SearchItem, 0, len(docs))
	for _, doc := range docs {
		documentID, _ := strconv.ParseUint(fmt.Sprint(doc.MetaData["document_id"]), 10, 64)
		chunkID, _ := strconv.ParseUint(fmt.Sprint(doc.MetaData["chunk_id"]), 10, 64)
		items = append(items, dto.SearchItem{
			DocumentID: documentID,
			ChunkID:    chunkID,
			Filename:   fmt.Sprint(doc.MetaData["filename"]),
			Content:    doc.Content,
			Score:      doc.Score(),
			Metadata: map[string]any{
				"chunk_index": doc.MetaData["chunk_index"],
			},
		})
	}
	return items, nil
}

func (p *Pipeline) DeleteDocumentVectors(ctx context.Context, kbID, documentID uint64) error {
	pattern := fmt.Sprintf("%s%d:%d:*", p.keyPrefix, kbID, documentID)
	var cursor uint64
	for {
		keys, next, err := p.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := p.redis.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if next == 0 {
			return nil
		}
		cursor = next
	}
}

func (p *Pipeline) documentToHashes(ctx context.Context, doc *schema.Document) (*einoredisindexer.Hashes, error) {
	key := doc.ID
	key = filepath.ToSlash(key)
	return &einoredisindexer.Hashes{
		Key: key,
		Field2Value: map[string]einoredisindexer.FieldValue{
			"content": {
				Value:    doc.Content,
				EmbedKey: "vector_content",
			},
			"kb_id": {
				Value: doc.MetaData["kb_id"],
			},
			"document_id": {
				Value: doc.MetaData["document_id"],
			},
			"chunk_id": {
				Value: doc.MetaData["chunk_id"],
			},
			"chunk_index": {
				Value: doc.MetaData["chunk_index"],
			},
			"filename": {
				Value: doc.MetaData["filename"],
			},
		},
	}, nil
}

func scoreFromRedis(score *float64) float64 {
	if score == nil {
		return 0
	}
	return *score
}

func scoreFromFields(fields map[string]string, fallback *float64) float64 {
	if raw := fields[einoredisretriever.SortByDistanceAttributeName]; raw != "" {
		score, err := strconv.ParseFloat(raw, 64)
		if err == nil {
			return score
		}
	}
	return scoreFromRedis(fallback)
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
