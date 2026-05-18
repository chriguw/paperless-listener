package paperless

import "time"

type Document struct {
	Count    int         `json:"count"`
	Next     interface{} `json:"next"`
	Previous interface{} `json:"previous"`
	All      []int       `json:"all"`
	Results  []struct {
		Id                  int           `json:"id"`
		Correspondent       interface{}   `json:"correspondent"`
		DocumentType        interface{}   `json:"document_type"`
		StoragePath         interface{}   `json:"storage_path"`
		Title               string        `json:"title"`
		Content             string        `json:"content"`
		Tags                []interface{} `json:"tags"`
		Created             string        `json:"created"`
		CreatedDate         string        `json:"created_date"`
		Modified            time.Time     `json:"modified"`
		Added               time.Time     `json:"added"`
		DeletedAt           interface{}   `json:"deleted_at"`
		ArchiveSerialNumber interface{}   `json:"archive_serial_number"`
		OriginalFileName    string        `json:"original_file_name"`
		ArchivedFileName    string        `json:"archived_file_name"`
		Owner               interface{}   `json:"owner"`
		UserCanChange       bool          `json:"user_can_change"`
		IsSharedByRequester bool          `json:"is_shared_by_requester"`
		Notes               []interface{} `json:"notes"`
		CustomFields        []interface{} `json:"custom_fields"`
		PageCount           int           `json:"page_count"`
		MimeType            string        `json:"mime_type"`
	} `json:"results"`
}

type CustomField struct {
	Value any `json:"value"`
	Field int `json:"field"`
}

type DocumentPayload struct {
	Title        string        `json:"title"`
	CustomFields []CustomField `json:"custom_fields,omitempty"`
}
