package vector

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/cblokkeel/limoncello/db"
	"github.com/sashabaranov/go-openai"
)

type Embedder struct {
	store  *db.Limoncello
	client *openai.Client
}

type SearchResults struct {
	ID         string  `json:"id"`
	Similarity float64 `json:"similarity"`
}

func NewEmbedder() (*Embedder, error) {
	store, err := db.NewLimoncello()
	if err != nil {
		return nil, err
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	return &Embedder{
		store,
		client,
	}, nil
}

func (e *Embedder) EmbeddDocument(ctx context.Context, collName string, k string, input string) error {
	res, err := e.embedd(ctx, input)
	if err != nil {
		return err
	}
	e.store.UpsertInCollection(collName, k, string(encode(getEmbeddingFromData(res.Data))))

	return nil
}

func (e *Embedder) NearestDocuments(ctx context.Context, colls []string, q string, n int) ([]*SearchResults, error) {
	res := []*SearchResults{}
	embeddedQuery, err := e.embedd(ctx, q)
	if err != nil {
		return nil, err
	}
	pairs, err := e.store.ReadCollections(colls)
	if err != nil {
		return nil, err
	}
	for _, pair := range pairs {
		sim, err := cosine(getEmbeddingFromData(embeddedQuery.Data), decode(pair.V))
		if err != nil {
			return nil, err
		}
		res = append(res, &SearchResults{
			ID:         string(pair.K),
			Similarity: sim,
		})
	}
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].Similarity > res[j].Similarity
	})
	return res[:n], nil
}

func (e *Embedder) CreateCollection(collName string) error {
	return e.store.CreateCollection(collName)
}

func (e *Embedder) embedd(ctx context.Context, q string) (*openai.EmbeddingResponse, error) {
	res, err := e.client.CreateEmbeddings(
		ctx,
		openai.EmbeddingRequestStrings{
			Input: []string{q},
			Model: openai.AdaEmbeddingV2,
			User:  "vecdb",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Error while embedding: %s", err)
	}
	return &res, nil
}

func getEmbeddingFromData(data []openai.Embedding) []float64 {
	if len(data) == 0 {
		return nil
	}
	e := data[0]
	res := make([]float64, len(e.Embedding))
	for i := range e.Embedding {
		res[i] = float64(e.Embedding[i])
	}
	return res
}

func encode(embeddings []float64) []byte {
	buf := make([]byte, len(embeddings)*8)
	for i, f := range embeddings {
		u := math.Float64bits(f)
		binary.LittleEndian.PutUint64(buf[i*8:], u)
	}
	return buf
}

func decode(buf []byte) []float64 {
	embeddings := make([]float64, len(buf)/8)
	for i := 0; i < len(buf); i += 8 {
		u := binary.LittleEndian.Uint64(buf[i : i+8])
		embeddings[i/8] = math.Float64frombits(u)
	}
	return embeddings
}
