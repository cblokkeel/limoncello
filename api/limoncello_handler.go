package api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cblokkeel/limoncello/vector"
	"github.com/gofiber/fiber/v2"
)

type LimoncelloHandler struct {
	embedder *vector.Embedder
}

func NewLimoncelloHandler() (*LimoncelloHandler, error) {
	e, err := vector.NewEmbedder()
	if err != nil {
		return nil, err
	}
	return &LimoncelloHandler{
		embedder: e,
	}, nil
}

type EmbeddBody struct {
	Key   string `json:"key"`
	Input string `json:"input"`
}

func (l *LimoncelloHandler) HandleEmbedd(c *fiber.Ctx) error {
	b := new(EmbeddBody)
	if err := c.BodyParser(b); err != nil {
		return err
	}
	coll := c.Params("coll")
	if b.Input == "" || b.Key == "" {
		return c.Status(400).JSON(map[string]string{"error": "bad parameters"})
	}
	if err := l.embedder.EmbeddDocument(c.Context(), coll, b.Key, b.Input); err != nil {
		return c.Status(400).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(map[string]string{"ok": fmt.Sprintf("Document %s embedded", b.Key)})
}

func (l *LimoncelloHandler) HandleCreateCollection(c *fiber.Ctx) error {
	coll := c.Params("coll")
	if err := l.embedder.CreateCollection(coll); err != nil {
		return err
	}
	return c.JSON(map[string]string{"ok": fmt.Sprintf("Collection %s created", coll)})
}

func (l *LimoncelloHandler) HandleSearch(c *fiber.Ctx) error {
	qColls := c.Query("colls")
	colls := strings.Split(qColls, ",")
	n, err := strconv.Atoi(c.Query("n"))
	if err != nil {
		return c.Status(400).JSON(map[string]string{"error": "bad parameters"})
	}
	q := c.Query("q")

	res, err := l.embedder.NearestDocuments(c.Context(), colls, q, n)
	if err != nil {
		return c.Status(400).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(res)
}
