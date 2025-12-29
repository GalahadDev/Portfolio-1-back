package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	supabaseURL string
	serviceKey  string
	bucketName  string
}

func NewService() *Service {
	return &Service{
		supabaseURL: os.Getenv("SUPABASE_URL"),
		serviceKey:  os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		bucketName:  "proofs",
	}
}

// UploadFile sube el archivo y devuelve el PATH interno (no la URL)
func (s *Service) UploadFile(file io.Reader, filename string, contentType string) (string, error) {
	// 1. Generar Path único: proofs/uuid_nombre.jpg
	// Limpiamos el nombre de espacios
	safeFilename := strings.ReplaceAll(filename, " ", "_")
	storagePath := fmt.Sprintf("%s_%s", uuid.New().String(), safeFilename)

	// 2. URL de subida
	apiURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.supabaseURL, s.bucketName, storagePath)

	req, err := http.NewRequest("POST", apiURL, file)
	if err != nil {
		return "", err
	}

	// Headers
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error de red supabase: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error supabase (%d): %s", resp.StatusCode, string(body))
	}

	// Retornamos el path relativo para guardarlo en BD
	return storagePath, nil
}

// GetSignedURL convierte un path interno en una URL pública temporal
func (s *Service) GetSignedURL(storagePath string) (string, error) {
	if storagePath == "" {
		return "", nil
	}

	// 1. Codificar el path
	parts := strings.Split(storagePath, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	encodedPath := strings.Join(parts, "/")

	// 2. Endpoint de firma
	baseURL := strings.TrimRight(s.supabaseURL, "/")
	signURL := fmt.Sprintf("%s/storage/v1/object/sign/%s/%s", baseURL, s.bucketName, encodedPath)

	// 3. Body
	reqBody := bytes.NewBuffer([]byte(`{"expiresIn": 3600}`))

	req, _ := http.NewRequest("POST", signURL, reqBody)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("falló la firma de url")
	}

	// 4. Parsear respuesta
	var result struct {
		SignedURL string `json:"signedURL"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// 5. Contruccion URL final
	finalPath := result.SignedURL

	if strings.HasPrefix(finalPath, "/object/sign") {
		finalPath = "/storage/v1" + finalPath
	}

	// Si no empieza con http, le pegamos el dominio base
	if !strings.HasPrefix(finalPath, "http") {
		finalPath = fmt.Sprintf("%s%s", baseURL, finalPath)
	}

	return finalPath, nil
}
