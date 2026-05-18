package main

import (
	"fmt"
	"io"
	"listener/internal/app"
	"listener/internal/naming"
	"listener/internal/paperless"
	"net/http"
	"os"
	"strings"
)

const configFilePath = "config.json"

var (
	loadConfigFn      = app.LoadConfig
	processDocumentFn = processDocument
)

func reloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	resolvedConfigPath, err := resolveConfigPath()
	if err != nil {
		http.Error(w, fmt.Sprintf("Fehler beim Finden der Config: %v", err), http.StatusInternalServerError)
		return
	}

	if err := loadConfigFn(resolvedConfigPath); err != nil {
		http.Error(w, fmt.Sprintf("Fehler beim Reload der Config: %v", err), http.StatusInternalServerError)
		return
	}

	writeOKText(w, "Config reloaded successfully")
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	logRequestHeaders(r)
	bodyBytes, err := readRequestBody(r)
	if err != nil {
		http.Error(w, "Fehler beim Lesen des Bodys", http.StatusBadRequest)
		fmt.Printf("Fehler beim Lesen des Bodys: %v\n", err)
		return
	}

	docID, err := extractDocumentIDFromBody(bodyBytes)
	if err != nil {
		http.Error(w, "Ungueltiger Webhook-Body", http.StatusBadRequest)
		fmt.Printf("Fehler beim Extrahieren der Dokument-ID: %v\n", err)
		return
	}

	fmt.Printf("Received document id: %s\n", docID)
	writeOKText(w, "Webhook received successfully")

	if err := processDocumentFn(docID); err != nil {
		fmt.Printf("Fehler bei der Dokumentverarbeitung: %v\n", err)
	}
}

func processDocument(docID string) error {
	pc := paperless.NewPaperlessClient(app.PaperlessHost, app.PaperlessToken)

	docummentContent, err := pc.GetDocument(docID)
	if err != nil {
		return fmt.Errorf("fehler beim Abrufen des Dokumentinhalts: %w", err)
	}

	parseResult := naming.CheckTitles(
		docummentContent,
		app.Title1,
		app.Title2,
		app.Title3,
		app.Title4,
		app.YearKeywords,
		app.AmountKeywords,
	)
	fmt.Printf("V_Title1: %s, V_Title2: %s, V_Title3: %s, V_Title4: %s DateString: %s\n", parseResult.Title1, parseResult.Title2, parseResult.Title3, parseResult.Title4, parseResult.DateString)

	finalTitle := naming.BuildFinalTitle(parseResult.Title1, parseResult.Title2, parseResult.Title3, parseResult.Title4, parseResult.DateString, parseResult.Year)
	fmt.Printf("Final Title: %s\n", finalTitle)

	fmt.Printf("Betrag: %v", parseResult.Amount)

	upderr := pc.UpdateTitle(docID, finalTitle, parseResult.Amount)
	if upderr != nil {
		return fmt.Errorf("fehler beim Aktualisieren des Titels: %w", upderr)
	}

	return nil
}

func main() {
	resolvedConfigPath, err := resolveConfigPath()
	if err != nil {
		fmt.Println("Error resolving configuration path:", err)
		return
	}

	err = app.LoadConfig(resolvedConfigPath)
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}

	registerRoutes()
	startHTTPServer()
}

func registerRoutes() {
	http.HandleFunc("/webhook/", webhookHandler)
	http.HandleFunc("/reload", reloadHandler)
}

func startHTTPServer() {
	fmt.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("HTTP-Server Fehler: %v\n", err)
	}
}

func logRequestHeaders(r *http.Request) {
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Printf("Header: %s = %s\n", name, value)
		}
	}
}

func readRequestBody(r *http.Request) ([]byte, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("BODY: %s\n", string(bodyBytes))
	return bodyBytes, nil
}

func extractDocumentIDFromBody(body []byte) (string, error) {
	return naming.ExtractDocumentID(string(body))
}

func writeOKText(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(message))
}

func resolveConfigPath() (string, error) {
	if envPath := os.Getenv("PAPERLESS_CONFIG_PATH"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
		return "", fmt.Errorf("PAPERLESS_CONFIG_PATH nicht gefunden: %s", envPath)
	}

	candidates := []string{configFilePath, "../../config.json", "/app/config.json"}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("keine Config gefunden (geprueft: %s)", strings.Join(candidates, ", "))
}

