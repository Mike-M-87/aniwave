package notifications

import (
	"aniwave/models"
	"aniwave/structure"
	"aniwave/utils"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

func DisplayNotifications(c *fiber.Ctx) error {
	go FetchAllNotifications()
	var nots []models.Not
	err := utils.DB.Find(&nots).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not find notifications")
	}
	return c.JSON(nots)
}

func MarkasDone(c *fiber.Ctx) error {
	dne := new(structure.MarkBody)
	err := c.BodyParser(dne)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Could not parse request")
	}
	if dne.Key != os.Getenv("KEY") {
		return c.Status(fiber.StatusUnauthorized).SendString("Check your key parameter UwU. Retrying will not help")
	}
	var not models.Not
	err = utils.DB.Where("id = ?", dne.Id).First(&not).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Notification not found")
	}
	not.Done = dne.Done
	err = utils.DB.Save(&not).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not save notification")
	}
	return c.SendStatus(fiber.StatusOK)
}