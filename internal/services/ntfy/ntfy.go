package ntfy

import (
	"bytes"
	"fmt"
	"ikoyhn/podcast-sponsorblock/internal/config"
	"net/http"
)

func SendNotification(message, title string) error {
	if config.Ntfy.Server == "" || config.Ntfy.Topic == "" {
		return nil
	}
	url := fmt.Sprintf("%s/%s", config.Ntfy.Server, config.Ntfy.Topic)
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(message))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")
	if title != "" {
		req.Header.Set("Title", title)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("ntfy returned status: %s", resp.Status)
	}
	return nil
}
