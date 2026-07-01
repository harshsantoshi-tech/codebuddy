package vectorstore

import (
	"bytes"
	"codebuddy/embeddings"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ChromaClient struct {
	BaseURL  string
	TenantID string 
	Database string
	client   *http.Client
}

func NewChromaClient(baseURL string) *ChromaClient {
	return &ChromaClient{
		BaseURL:  baseURL,
		TenantID: "default_tenant",  
		Database: "default_database", 
		client:   &http.Client{},
	}
}

func (c *ChromaClient) basePath() string {
	return fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s",
		c.BaseURL, c.TenantID, c.Database)
}

func (c *ChromaClient) CreateCollection(name string) (string, error) {
	name = sanitizeName(name)

	body := map[string]interface{}{
		"name": name,
		"metadata": map[string]string{
			"hnsw:space": "cosine",
		},
	}

	url := c.basePath() + "/collections"
	resp, statusCode, err := c.post(url, body)
	if err != nil {
		return "", err
	}

	if statusCode == http.StatusConflict {
		fmt.Println("Collection already exists, fetching ID...")
		return c.GetCollectionID(name)
	}

	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return "", fmt.Errorf("create collection failed with status %d: %v", statusCode, resp)
	}

	id, _ := resp["id"].(string)
	fmt.Println("✅ Created ChromaDB collection:", name, "| id:", id)
	return id, nil
}

func (c *ChromaClient) GetCollectionID(name string) (string, error) {
	name = sanitizeName(name)
	url := fmt.Sprintf("%s/collections/%s", c.basePath(), name)

	resp, err := c.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("get collection request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	id, ok := result["id"].(string)
	if !ok {
		return "", fmt.Errorf("collection '%s' not found: %s", name, string(body))
	}
	return id, nil
}


func (c *ChromaClient) UpsertChunks(collectionID string, chunks []embeddings.EmbeddedChunk) error {
	ids := make([]string, len(chunks))
	vectors := make([][]float32, len(chunks))
	documents := make([]string, len(chunks))
	metadatas := make([]map[string]interface{}, len(chunks))

	for i, chunk := range chunks {
		ids[i] = chunk.ID
		vectors[i] = chunk.Vector
		documents[i] = chunk.Content
		metadatas[i] = map[string]interface{}{
			"file_path":  chunk.FilePath,
			"language":   chunk.Language,
			"repo_url":   chunk.RepoURL,
			"start_line": chunk.StartLine,
			"end_line":   chunk.EndLine,
		}
	}

	url := fmt.Sprintf("%s/collections/%s/upsert", c.basePath(), collectionID)
	_, statusCode, err := c.post(url, map[string]interface{}{
		"ids":        ids,
		"embeddings": vectors,
		"documents":  documents,
		"metadatas":  metadatas,
	})

	if err != nil {
		return err
	}
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return fmt.Errorf("upsert failed with status %d", statusCode)
	}

	fmt.Printf("✅ Upserted %d chunks into collection %s\n", len(chunks), collectionID)
	return nil
}

func (c *ChromaClient) QueryCollection(collectionID string, queryVector []float32, topK int) ([]QueryResult, error) {
	url := fmt.Sprintf("%s/collections/%s/query", c.basePath(), collectionID)

	body := map[string]interface{}{
		"query_embeddings": [][]float32{queryVector}, // batch of 1 query
		"n_results":        topK,
		"include":          []string{"documents", "metadatas", "distances"},
	}

	resp, statusCode, err := c.post(url, body)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed with status %d", statusCode)
	}

	return parseQueryResults(resp), nil
}

type QueryResult struct {
	Document string
	Metadata map[string]interface{}
	Distance float32 
}

// parseQueryResults extracts results from Chroma's nested response format
func parseQueryResults(resp map[string]interface{}) []QueryResult {

	var results []QueryResult

	docs, _ := resp["documents"].([]interface{})
	metas, _ := resp["metadatas"].([]interface{})
	dists, _ := resp["distances"].([]interface{})

	if len(docs) == 0 {
		return results
	}

	docList, _ := docs[0].([]interface{})
	metaList, _ := metas[0].([]interface{})
	distList, _ := dists[0].([]interface{})

	for i := range docList {
		doc, _ := docList[i].(string)
		meta, _ := metaList[i].(map[string]interface{})
		dist := float32(0)
		if i < len(distList) {
			if d, ok := distList[i].(float64); ok {
				dist = float32(d)
			}
		}
		results = append(results, QueryResult{
			Document: doc,
			Metadata: meta,
			Distance: dist,
		})
	}

	return results
}

func (c *ChromaClient) post(url string, body interface{}) (map[string]interface{}, int, error) {
	bodyBytes, _ := json.Marshal(body)

	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("POST %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	return result, resp.StatusCode, nil
}

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ".", "-")
	return strings.ToLower(name)
}