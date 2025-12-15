package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/VersusControl/versus-incident/pkg/config"
	"github.com/VersusControl/versus-incident/pkg/services"

	"github.com/gofiber/fiber/v2"
)

func CreateIncident(c *fiber.Ctx) error {
	cfg := config.GetConfig()
	format := strings.ToLower(c.Query("format", "json"))

	if cfg.Alert.DebugBody {
		rawBody := c.Body()

		// Log the raw request body for debugging purposes
		fmt.Println("Raw Request Body:", string(rawBody))
	}

	switch format {
	case "json_stream":
		decoder := json.NewDecoder(bytes.NewReader(c.Body()))
		count := 0
		var merged map[string]interface{}
		var logs []string

		for decoder.More() {
			record := map[string]interface{}{}
			if err := decoder.Decode(&record); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid json_stream input"})
			}

			count++

			if merged == nil {
				merged = map[string]interface{}{}
				for k, v := range record {
					merged[k] = v
				}
			}

			if logVal, ok := record["Logs"].(string); ok && logVal != "" {
				logs = append(logs, logVal)
			} else if logVal, ok := record["log"].(string); ok && logVal != "" {
				logs = append(logs, logVal)
			} else {
				if raw, err := json.Marshal(record); err == nil {
					logs = append(logs, string(raw))
				}
			}
		}

		if count == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No incidents found in json_stream"})
		}

		if len(logs) > 0 {
			merged["Logs"] = strings.Join(logs, "\n")
			merged["LogsList"] = logs
		}

		var err error
		if len(c.Queries()) > 0 {
			overwriteVaule := c.Queries()
			err = services.CreateIncident("", &merged, &overwriteVaule)
		} else {
			err = services.CreateIncident("", &merged)
		}

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status": "Incident created",
			"count":  count,
		})
	case "json_array":
		var records []map[string]interface{}
		if err := json.Unmarshal(c.Body(), &records); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid json_array input"})
		}

		if len(records) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No incidents found in json_array"})
		}

		merged := map[string]interface{}{}
		for k, v := range records[0] {
			merged[k] = v
		}

		var logs []string
		for _, record := range records {
			if logVal, ok := record["Logs"].(string); ok && logVal != "" {
				logs = append(logs, logVal)
			} else if logVal, ok := record["log"].(string); ok && logVal != "" {
				logs = append(logs, logVal)
			} else {
				if raw, err := json.Marshal(record); err == nil {
					logs = append(logs, string(raw))
				}
			}
		}

		if len(logs) > 0 {
			merged["Logs"] = strings.Join(logs, "\n")
			merged["LogsList"] = logs
		}

		var err error
		if len(c.Queries()) > 0 {
			overwriteVaule := c.Queries()
			err = services.CreateIncident("", &merged, &overwriteVaule)
		} else {
			err = services.CreateIncident("", &merged)
		}

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"status": "Incident created",
			"count":  len(records),
		})
	default:
		body := &map[string]interface{}{}

		if err := c.BodyParser(body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
		}

		var err error

		// If query parameters exist, get the value to overwrite the default configuration
		if len(c.Queries()) > 0 {
			overwriteVaule := c.Queries()
			err = services.CreateIncident("", body, &overwriteVaule)
		} else {
			err = services.CreateIncident("", body)
		}

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "Incident created"})
	}
}
