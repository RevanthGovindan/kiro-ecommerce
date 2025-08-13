package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ImageOptimizationConfig holds configuration for image optimization
type ImageOptimizationConfig struct {
	CDNBaseURL    string
	DefaultWidth  int
	DefaultHeight int
	Quality       int
	Format        string
}

// DefaultImageConfig provides default image optimization settings
var DefaultImageConfig = ImageOptimizationConfig{
	CDNBaseURL:    "", // Will be set from environment
	DefaultWidth:  800,
	DefaultHeight: 600,
	Quality:       85,
	Format:        "webp",
}

// OptimizeImageURL generates an optimized image URL with CDN parameters
func OptimizeImageURL(originalURL string, width, height int, quality int, format string) string {
	if originalURL == "" {
		return ""
	}

	// If no CDN base URL is configured, return original URL
	if DefaultImageConfig.CDNBaseURL == "" {
		return originalURL
	}

	// Parse the original URL to extract the path
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return originalURL
	}

	// Build optimized URL with CDN parameters
	params := url.Values{}

	if width > 0 {
		params.Add("w", strconv.Itoa(width))
	}
	if height > 0 {
		params.Add("h", strconv.Itoa(height))
	}
	if quality > 0 && quality <= 100 {
		params.Add("q", strconv.Itoa(quality))
	}
	if format != "" {
		params.Add("f", format)
	}

	// Add auto optimization
	params.Add("auto", "compress,format")

	// Construct the optimized URL
	optimizedURL := fmt.Sprintf("%s%s", DefaultImageConfig.CDNBaseURL, parsedURL.Path)
	if len(params) > 0 {
		optimizedURL += "?" + params.Encode()
	}

	return optimizedURL
}

// GenerateResponsiveImageURLs creates multiple image URLs for different screen sizes
func GenerateResponsiveImageURLs(originalURL string) map[string]string {
	if originalURL == "" {
		return nil
	}

	sizes := map[string]struct {
		width   int
		height  int
		quality int
	}{
		"thumbnail": {150, 150, 80},
		"small":     {300, 300, 85},
		"medium":    {600, 600, 85},
		"large":     {1200, 1200, 90},
		"xlarge":    {1920, 1920, 90},
	}

	urls := make(map[string]string)

	for size, config := range sizes {
		urls[size] = OptimizeImageURL(originalURL, config.width, config.height, config.quality, DefaultImageConfig.Format)
	}

	return urls
}

// ValidateImageURL checks if an image URL is valid and safe
func ValidateImageURL(imageURL string) bool {
	if imageURL == "" {
		return false
	}

	// Parse URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return false
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	// Check for valid image extensions
	validExtensions := []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg"}
	path := strings.ToLower(parsedURL.Path)

	for _, ext := range validExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}

// GetImageDimensions extracts width and height from query parameters
func GetImageDimensions(queryParams url.Values) (width, height int) {
	if w := queryParams.Get("w"); w != "" {
		if parsed, err := strconv.Atoi(w); err == nil && parsed > 0 && parsed <= 4000 {
			width = parsed
		}
	}

	if h := queryParams.Get("h"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 && parsed <= 4000 {
			height = parsed
		}
	}

	return width, height
}

// SanitizeImageFilename removes potentially dangerous characters from filenames
func SanitizeImageFilename(filename string) string {
	// Remove path separators and other dangerous characters
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "<", "")
	filename = strings.ReplaceAll(filename, ">", "")
	filename = strings.ReplaceAll(filename, ":", "")
	filename = strings.ReplaceAll(filename, "\"", "")
	filename = strings.ReplaceAll(filename, "|", "")
	filename = strings.ReplaceAll(filename, "?", "")
	filename = strings.ReplaceAll(filename, "*", "")

	return strings.TrimSpace(filename)
}

// ImageResponse represents an optimized image response
type ImageResponse struct {
	Original   string            `json:"original"`
	Optimized  string            `json:"optimized"`
	Responsive map[string]string `json:"responsive,omitempty"`
}

// ProcessImageURLs processes a list of image URLs and returns optimized versions
func ProcessImageURLs(urls []string) []ImageResponse {
	if len(urls) == 0 {
		return nil
	}

	responses := make([]ImageResponse, len(urls))

	for i, url := range urls {
		responses[i] = ImageResponse{
			Original:   url,
			Optimized:  OptimizeImageURL(url, DefaultImageConfig.DefaultWidth, DefaultImageConfig.DefaultHeight, DefaultImageConfig.Quality, DefaultImageConfig.Format),
			Responsive: GenerateResponsiveImageURLs(url),
		}
	}

	return responses
}
