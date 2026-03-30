package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gbsto/daisy/util"

	"github.com/gofiber/fiber/v2"
)

func UploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("uploadfile")
	if err != nil {
		return err
	}
	executable, err := os.Executable()
	if err != nil {
		log.Println(err)
	}
	serverDir := filepath.Dir(executable)
	fileLocation := filepath.Join(serverDir, "web", "public", "images", file.Filename)
	c.SaveFile(file, fileLocation)

	util.InitialPhotoUploadResizer(file.Filename)
	go util.ShrinkPhotos() //process entire images directory and create thumbnails of photos
	return c.Status(fiber.StatusOK).SendString("success")
}
