package main

import (
	"log"

	"github.com/cblokkeel/limoncello/api"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	apiv1 := app.Group("/api/v1")
	dbHandler, err := api.NewDatabaseHandler()
	if err != nil {
		log.Fatal(err)
	}

	apiv1.Post("/:coll", dbHandler.HandleCreateCollection)
	apiv1.Post("/:coll/embedd", dbHandler.HandleEmbedd)
	apiv1.Get("/:coll/search", dbHandler.HandleSearch)

	app.Listen(":8000")

}
