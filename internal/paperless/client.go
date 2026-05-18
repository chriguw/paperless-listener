package paperless

import (
	"fmt"
	"log/slog"

	"github.com/go-resty/resty/v2"
)

type PaperlessClient struct {
	client *resty.Client
	logger *slog.Logger
}

func NewPaperlessClient(paperlessHost string, token string) PaperlessClient {
	return PaperlessClient{
		client: resty.New().
			SetBaseURL(paperlessHost).
			SetHeader("Authorization", fmt.Sprintf("Token %s", token)).
			SetHeader("Accept", "application/json"),
		logger: slog.Default(),
	}
}

func (pc PaperlessClient) GetDocument(docId string) (string, error) {
	url := fmt.Sprintf("/api/documents/?id=%s", docId)
	resp, err := pc.client.R().
		SetQueryParam("id", fmt.Sprintf("%s", docId)).
		SetResult(&Document{}).
		Get(url)
	if err != nil {
		pc.logger.Error("Fehler beim Senden der Anfrage", slog.Any("error", err))
		return "", err
	}

	if resp.IsError() {
		pc.logger.Error("Unerwarteter Statuscode", slog.Int("status", resp.StatusCode()))
		return "", fmt.Errorf("Unerwarteter Statuscode: %d", resp.StatusCode())
	}

	document := resp.Result().(*Document)
	if len(document.Results) > 0 {
		return document.Results[0].Content, nil
	}

	pc.logger.Warn("Kein Inhalt gefunden", slog.String("docId", docId))
	return "", fmt.Errorf("Kein Inhalt gefunden")
}

func (pc PaperlessClient) UpdateTitle(docId string, finalTitle string, amount float64, cfId int) error {
	url := fmt.Sprintf("/api/documents/%s/", docId)

	payload := DocumentPayload{
		Title: finalTitle,
	}

	if cfId > 0 {
		payload.CustomFields = []CustomField{
			{
				Value: amount,
				Field: cfId,
			},
		}
	}

	resp, err := pc.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Patch(url)
	if err != nil {
		pc.logger.Error("Fehler beim Senden der PATCH-Anfrage", slog.Any("error", err))
		return err
	}

	if resp.IsError() {
		pc.logger.Error("Unerwarteter Statuscode beim PATCH", slog.Int("status", resp.StatusCode()))
		return fmt.Errorf("Unerwarteter Statuscode: %d", resp.StatusCode())
	}

	return nil
}
