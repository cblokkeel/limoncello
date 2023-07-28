package api

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cblokkeel/limoncello/vector"
	"github.com/gofiber/fiber/v2"
)

type DatabseHandler struct {
	embedder *vector.Embedder
}

func NewDatabaseHandler() (*DatabseHandler, error) {
	e, err := vector.NewEmbedder()
	if err != nil {
		return nil, err
	}
	return &DatabseHandler{
		embedder: e,
	}, nil
}

type EmbeddBody struct {
	Key   string `json:"key"`
	Input string `json:"input"`
}

func (h *DatabseHandler) HandleEmbedd(c *fiber.Ctx) error {
	b := new(EmbeddBody)
	if err := c.BodyParser(b); err != nil {
		return err
	}
	coll := c.Params("coll")
	if b.Input == "" || b.Key == "" {
		return c.Status(400).JSON(map[string]string{"error": "bad parameters"})
	}
	if err := h.embedder.EmbeddDocument(c.Context(), coll, b.Key, b.Input); err != nil {
		return c.Status(400).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(map[string]string{"ok": fmt.Sprintf("Document %s embedded", b.Key)})
}

func (h *DatabseHandler) HandleCreateCollection(c *fiber.Ctx) error {
	coll := c.Params("coll")
	if err := h.embedder.CreateCollection(coll); err != nil {
		return err
	}
	return c.JSON(map[string]string{"ok": fmt.Sprintf("Collection %s created", coll)})
}

func (h *DatabseHandler) HandleSearch(c *fiber.Ctx) error {
	qColls := c.Query("colls")
	colls := strings.Split(qColls, ",")
	n, err := strconv.Atoi(c.Query("n"))
	if err != nil {
		return c.Status(400).JSON(map[string]string{"error": "bad parameters"})
	}
	q := c.Query("q")

	res, err := h.embedder.NearestDocuments(c.Context(), colls, q, n)
	if err != nil {
		return c.Status(400).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(res)
}
