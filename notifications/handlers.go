package notifications

import (
	"aniwave/models"
	"aniwave/utils"
	"github.com/gofiber/fiber/v2"
)

func CheckNotifications(c *fiber.Ctx) error {
	go FetchAllNotifications()
	return c.SendStatus(fiber.StatusAccepted)
}
func DisplayNotifications(c *fiber.Ctx) error {
	var nots []models.Not
	err := utils.DB.Find(&nots).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not find notifications")
	}
	return c.JSON(nots)
}

func ChangeDone(c *fiber.Ctx) error {
	animid := c.Query("id")
	var not models.Not
	err := utils.DB.Where("id = ?", animid).First(&not).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Notification not found")
	}
	not.Done = !not.Done
	err = utils.DB.Save(&not).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not save notification")
	}
	return c.SendStatus(fiber.StatusOK)
}
