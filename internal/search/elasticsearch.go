package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"ecommerce-website/internal/models"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

const ProductIndex = "products"

type ElasticsearchService struct {
	client *elasticsearch.Client
}

type SearchFilters struct {
	CategoryID *string
	MinPrice   *float64
	MaxPrice   *float64
	InStock    *bool
	Search     *string
}

type SearchSort struct {
	Field string // name, price, created_at, popularity
	Order string // asc, desc
}

type SearchResponse struct {
	Products    []models.Product `json:"products"`
	Total       int64            `json:"total"`
	Page        int              `json:"page"`
	PageSize    int              `json:"pageSize"`
	TotalPages  int              `json:"totalPages"`
	Suggestions []string         `json:"suggestions,omitempty"`
	Facets      *SearchFacets    `json:"facets,omitempty"`
}

type SearchFacets struct {
	Categories  []CategoryFacet   `json:"categories"`
	PriceRanges []PriceRangeFacet `json:"priceRanges"`
}

type CategoryFacet struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type PriceRangeFacet struct {
	Range string  `json:"range"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Count int64   `json:"count"`
}

func NewElasticsearchService() (*ElasticsearchService, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200",
		},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	service := &ElasticsearchService{client: client}

	// Initialize index
	if err := service.initializeIndex(); err != nil {
		log.Printf("Warning: Failed to initialize Elasticsearch index: %v", err)
	}

	return service, nil
}

func (es *ElasticsearchService) initializeIndex() error {
	// Check if index exists
	req := esapi.IndicesExistsRequest{
		Index: []string{ProductIndex},
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	// If index exists, return
	if res.StatusCode == 200 {
		return nil
	}

	// Create index with mapping
	mapping := `{
		"mappings": {
			"properties": {
				"id": {"type": "keyword"},
				"name": {
					"type": "text",
					"analyzer": "standard",
					"fields": {
						"keyword": {"type": "keyword"},
						"suggest": {
							"type": "completion",
							"analyzer": "simple"
						}
					}
				},
				"description": {
					"type": "text",
					"analyzer": "standard"
				},
				"price": {"type": "float"},
				"compareAtPrice": {"type": "float"},
				"sku": {"type": "keyword"},
				"inventory": {"type": "integer"},
				"isActive": {"type": "boolean"},
				"categoryId": {"type": "keyword"},
				"categoryName": {
					"type": "text",
					"fields": {
						"keyword": {"type": "keyword"}
					}
				},
				"images": {"type": "keyword"},
				"specifications": {"type": "object"},
				"seoTitle": {"type": "text"},
				"seoDescription": {"type": "text"},
				"createdAt": {"type": "date"},
				"updatedAt": {"type": "date"},
				"popularity": {"type": "float"}
			}
		},
		"settings": {
			"analysis": {
				"analyzer": {
					"product_analyzer": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["lowercase", "stop", "snowball"]
					}
				}
			}
		}
	}`

	createReq := esapi.IndicesCreateRequest{
		Index: ProductIndex,
		Body:  strings.NewReader(mapping),
	}

	createRes, err := createReq.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("failed to create index: %s", createRes.String())
	}

	return nil
}

func (es *ElasticsearchService) IndexProduct(product *models.Product) error {
	// Prepare document for indexing
	doc := map[string]interface{}{
		"id":             product.ID,
		"name":           product.Name,
		"description":    product.Description,
		"price":          product.Price,
		"compareAtPrice": product.CompareAtPrice,
		"sku":            product.SKU,
		"inventory":      product.Inventory,
		"isActive":       product.IsActive,
		"categoryId":     product.CategoryID,
		"images":         product.Images,
		"specifications": product.Specifications,
		"seoTitle":       product.SEOTitle,
		"seoDescription": product.SEODescription,
		"createdAt":      product.CreatedAt,
		"updatedAt":      product.UpdatedAt,
		"popularity":     0.0, // Default popularity score
	}

	// Add category name if available
	if product.Category != nil {
		doc["categoryName"] = product.Category.Name
	}

	docBytes, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      ProductIndex,
		DocumentID: product.ID,
		Body:       bytes.NewReader(docBytes),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index document: %s", res.String())
	}

	return nil
}

func (es *ElasticsearchService) DeleteProduct(productID string) error {
	req := esapi.DeleteRequest{
		Index:      ProductIndex,
		DocumentID: productID,
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("failed to delete document: %s", res.String())
	}

	return nil
}

func (es *ElasticsearchService) SearchProducts(filters SearchFilters, sort SearchSort, page, pageSize int, includeFacets bool) (*SearchResponse, error) {
	query := es.buildSearchQuery(filters)

	// Build aggregations for facets
	var aggs map[string]interface{}
	if includeFacets {
		aggs = es.buildAggregations()
	}

	// Build sort
	sortClause := es.buildSort(sort)

	// Calculate pagination
	from := (page - 1) * pageSize

	searchBody := map[string]interface{}{
		"query": query,
		"sort":  sortClause,
		"from":  from,
		"size":  pageSize,
	}

	if aggs != nil {
		searchBody["aggs"] = aggs
	}

	searchBytes, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{ProductIndex},
		Body:  bytes.NewReader(searchBytes),
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	return es.parseSearchResponse(searchResult, page, pageSize)
}

func (es *ElasticsearchService) GetSuggestions(query string, size int) ([]string, error) {
	if size <= 0 {
		size = 5
	}

	searchBody := map[string]interface{}{
		"suggest": map[string]interface{}{
			"product_suggest": map[string]interface{}{
				"prefix": query,
				"completion": map[string]interface{}{
					"field": "name.suggest",
					"size":  size,
				},
			},
		},
	}

	searchBytes, err := json.Marshal(searchBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal suggest query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{ProductIndex},
		Body:  bytes.NewReader(searchBytes),
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute suggest: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("suggest error: %s", res.String())
	}

	var suggestResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&suggestResult); err != nil {
		return nil, fmt.Errorf("failed to decode suggest response: %w", err)
	}

	return es.parseSuggestResponse(suggestResult)
}

func (es *ElasticsearchService) buildSearchQuery(filters SearchFilters) map[string]interface{} {
	must := []map[string]interface{}{
		{"term": map[string]interface{}{"isActive": true}},
	}

	// Text search
	if filters.Search != nil && *filters.Search != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     *filters.Search,
				"fields":    []string{"name^3", "description^2", "categoryName", "sku"},
				"type":      "best_fields",
				"fuzziness": "AUTO",
			},
		})
	}

	// Category filter
	if filters.CategoryID != nil {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{"categoryId": *filters.CategoryID},
		})
	}

	// Price range filter
	if filters.MinPrice != nil || filters.MaxPrice != nil {
		priceRange := make(map[string]interface{})
		if filters.MinPrice != nil {
			priceRange["gte"] = *filters.MinPrice
		}
		if filters.MaxPrice != nil {
			priceRange["lte"] = *filters.MaxPrice
		}
		must = append(must, map[string]interface{}{
			"range": map[string]interface{}{"price": priceRange},
		})
	}

	// In stock filter
	if filters.InStock != nil && *filters.InStock {
		must = append(must, map[string]interface{}{
			"range": map[string]interface{}{"inventory": map[string]interface{}{"gt": 0}},
		})
	}

	return map[string]interface{}{
		"bool": map[string]interface{}{
			"must": must,
		},
	}
}

func (es *ElasticsearchService) buildSort(sort SearchSort) []map[string]interface{} {
	if sort.Field == "" {
		sort.Field = "created_at"
	}
	if sort.Order == "" {
		sort.Order = "desc"
	}

	// Map sort fields
	var sortField string
	switch sort.Field {
	case "name":
		sortField = "name.keyword"
	case "price":
		sortField = "price"
	case "created_at":
		sortField = "createdAt"
	case "popularity":
		sortField = "popularity"
	default:
		sortField = "createdAt"
	}

	return []map[string]interface{}{
		{sortField: map[string]interface{}{"order": sort.Order}},
		{"_score": map[string]interface{}{"order": "desc"}}, // Secondary sort by relevance
	}
}

func (es *ElasticsearchService) buildAggregations() map[string]interface{} {
	return map[string]interface{}{
		"categories": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "categoryId",
				"size":  20,
			},
		},
		"price_ranges": map[string]interface{}{
			"range": map[string]interface{}{
				"field": "price",
				"ranges": []map[string]interface{}{
					{"key": "0-25", "to": 25},
					{"key": "25-50", "from": 25, "to": 50},
					{"key": "50-100", "from": 50, "to": 100},
					{"key": "100-250", "from": 100, "to": 250},
					{"key": "250+", "from": 250},
				},
			},
		},
	}
}

func (es *ElasticsearchService) parseSearchResponse(result map[string]interface{}, page, pageSize int) (*SearchResponse, error) {
	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid search response format")
	}

	total, ok := hits["total"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid total format")
	}

	totalValue, ok := total["value"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid total value format")
	}

	hitsList, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid hits format")
	}

	var products []models.Product
	for _, hit := range hitsList {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		product, err := es.mapSourceToProduct(source)
		if err != nil {
			log.Printf("Error mapping product: %v", err)
			continue
		}

		products = append(products, *product)
	}

	totalPages := int((int64(totalValue) + int64(pageSize) - 1) / int64(pageSize))

	response := &SearchResponse{
		Products:   products,
		Total:      int64(totalValue),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	// Parse facets if available
	if aggs, ok := result["aggregations"].(map[string]interface{}); ok {
		response.Facets = es.parseFacets(aggs)
	}

	return response, nil
}

func (es *ElasticsearchService) parseSuggestResponse(result map[string]interface{}) ([]string, error) {
	suggest, ok := result["suggest"].(map[string]interface{})
	if !ok {
		return []string{}, nil
	}

	productSuggest, ok := suggest["product_suggest"].([]interface{})
	if !ok || len(productSuggest) == 0 {
		return []string{}, nil
	}

	firstSuggest, ok := productSuggest[0].(map[string]interface{})
	if !ok {
		return []string{}, nil
	}

	options, ok := firstSuggest["options"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	var suggestions []string
	for _, option := range options {
		optionMap, ok := option.(map[string]interface{})
		if !ok {
			continue
		}

		text, ok := optionMap["text"].(string)
		if ok {
			suggestions = append(suggestions, text)
		}
	}

	return suggestions, nil
}

func (es *ElasticsearchService) parseFacets(aggs map[string]interface{}) *SearchFacets {
	facets := &SearchFacets{}

	// Parse category facets
	if catAgg, ok := aggs["categories"].(map[string]interface{}); ok {
		if buckets, ok := catAgg["buckets"].([]interface{}); ok {
			for _, bucket := range buckets {
				if bucketMap, ok := bucket.(map[string]interface{}); ok {
					if key, ok := bucketMap["key"].(string); ok {
						if count, ok := bucketMap["doc_count"].(float64); ok {
							facets.Categories = append(facets.Categories, CategoryFacet{
								ID:    key,
								Name:  key, // Would need to resolve category name
								Count: int64(count),
							})
						}
					}
				}
			}
		}
	}

	// Parse price range facets
	if priceAgg, ok := aggs["price_ranges"].(map[string]interface{}); ok {
		if buckets, ok := priceAgg["buckets"].([]interface{}); ok {
			for _, bucket := range buckets {
				if bucketMap, ok := bucket.(map[string]interface{}); ok {
					if key, ok := bucketMap["key"].(string); ok {
						if count, ok := bucketMap["doc_count"].(float64); ok {
							var min, max float64
							if from, ok := bucketMap["from"].(float64); ok {
								min = from
							}
							if to, ok := bucketMap["to"].(float64); ok {
								max = to
							} else {
								max = 999999 // For open-ended ranges
							}

							facets.PriceRanges = append(facets.PriceRanges, PriceRangeFacet{
								Range: key,
								Min:   min,
								Max:   max,
								Count: int64(count),
							})
						}
					}
				}
			}
		}
	}

	return facets
}

func (es *ElasticsearchService) mapSourceToProduct(source map[string]interface{}) (*models.Product, error) {
	product := &models.Product{}

	if id, ok := source["id"].(string); ok {
		product.ID = id
	}

	if name, ok := source["name"].(string); ok {
		product.Name = name
	}

	if description, ok := source["description"].(string); ok {
		product.Description = description
	}

	if price, ok := source["price"].(float64); ok {
		product.Price = price
	}

	if compareAtPrice, ok := source["compareAtPrice"].(float64); ok {
		product.CompareAtPrice = &compareAtPrice
	}

	if sku, ok := source["sku"].(string); ok {
		product.SKU = sku
	}

	if inventory, ok := source["inventory"].(float64); ok {
		product.Inventory = int(inventory)
	}

	if isActive, ok := source["isActive"].(bool); ok {
		product.IsActive = isActive
	}

	if categoryId, ok := source["categoryId"].(string); ok {
		product.CategoryID = categoryId
	}

	if images, ok := source["images"].([]interface{}); ok {
		var imageStrings []string
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				imageStrings = append(imageStrings, imgStr)
			}
		}
		product.Images = models.StringArray(imageStrings)
	}

	return product, nil
}
