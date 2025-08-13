package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptimizeImageURL(t *testing.T) {
	// Set up test CDN base URL
	DefaultImageConfig.CDNBaseURL = "https://cdn.example.com"

	tests := []struct {
		name        string
		originalURL string
		width       int
		height      int
		quality     int
		format      string
		expected    string
	}{
		{
			name:        "Basic optimization",
			originalURL: "https://example.com/image.jpg",
			width:       800,
			height:      600,
			quality:     85,
			format:      "webp",
			expected:    "https://cdn.example.com/image.jpg?auto=compress%2Cformat&f=webp&h=600&q=85&w=800",
		},
		{
			name:        "Width only",
			originalURL: "https://example.com/image.jpg",
			width:       400,
			height:      0,
			quality:     0,
			format:      "",
			expected:    "https://cdn.example.com/image.jpg?auto=compress%2Cformat&w=400",
		},
		{
			name:        "Empty URL",
			originalURL: "",
			width:       800,
			height:      600,
			quality:     85,
			format:      "webp",
			expected:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := OptimizeImageURL(test.originalURL, test.width, test.height, test.quality, test.format)
			assert.Equal(t, test.expected, result)
		})
	}

	// Test with no CDN configured
	DefaultImageConfig.CDNBaseURL = ""
	result := OptimizeImageURL("https://example.com/image.jpg", 800, 600, 85, "webp")
	assert.Equal(t, "https://example.com/image.jpg", result)
}

func TestGenerateResponsiveImageURLs(t *testing.T) {
	DefaultImageConfig.CDNBaseURL = "https://cdn.example.com"

	originalURL := "https://example.com/image.jpg"
	result := GenerateResponsiveImageURLs(originalURL)

	assert.NotNil(t, result)
	assert.Contains(t, result, "thumbnail")
	assert.Contains(t, result, "small")
	assert.Contains(t, result, "medium")
	assert.Contains(t, result, "large")
	assert.Contains(t, result, "xlarge")

	// Check that thumbnail URL contains expected parameters
	thumbnailURL := result["thumbnail"]
	assert.Contains(t, thumbnailURL, "w=150")
	assert.Contains(t, thumbnailURL, "h=150")
	assert.Contains(t, thumbnailURL, "q=80")

	// Test with empty URL
	emptyResult := GenerateResponsiveImageURLs("")
	assert.Nil(t, emptyResult)
}

func TestValidateImageURL(t *testing.T) {
	tests := []struct {
		name     string
		imageURL string
		expected bool
	}{
		{
			name:     "Valid HTTPS JPG",
			imageURL: "https://example.com/image.jpg",
			expected: true,
		},
		{
			name:     "Valid HTTP PNG",
			imageURL: "http://example.com/image.png",
			expected: true,
		},
		{
			name:     "Valid WebP",
			imageURL: "https://example.com/image.webp",
			expected: true,
		},
		{
			name:     "Valid SVG",
			imageURL: "https://example.com/icon.svg",
			expected: true,
		},
		{
			name:     "Invalid scheme",
			imageURL: "ftp://example.com/image.jpg",
			expected: false,
		},
		{
			name:     "Invalid extension",
			imageURL: "https://example.com/document.pdf",
			expected: false,
		},
		{
			name:     "Empty URL",
			imageURL: "",
			expected: false,
		},
		{
			name:     "Invalid URL",
			imageURL: "not-a-url",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ValidateImageURL(test.imageURL)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestSanitizeImageFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "Normal filename",
			filename: "image.jpg",
			expected: "image.jpg",
		},
		{
			name:     "Filename with path separators",
			filename: "../../../etc/passwd",
			expected: "etcpasswd",
		},
		{
			name:     "Filename with dangerous characters",
			filename: "image<script>.jpg",
			expected: "imagescript.jpg",
		},
		{
			name:     "Filename with spaces",
			filename: "  my image.jpg  ",
			expected: "my image.jpg",
		},
		{
			name:     "Complex malicious filename",
			filename: "..\\..\\<script>alert('xss')</script>image.jpg",
			expected: "scriptalert('xss')scriptimage.jpg",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SanitizeImageFilename(test.filename)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProcessImageURLs(t *testing.T) {
	DefaultImageConfig.CDNBaseURL = "https://cdn.example.com"

	urls := []string{
		"https://example.com/image1.jpg",
		"https://example.com/image2.png",
	}

	result := ProcessImageURLs(urls)

	assert.Len(t, result, 2)

	// Check first image response
	firstImage := result[0]
	assert.Equal(t, "https://example.com/image1.jpg", firstImage.Original)
	assert.Contains(t, firstImage.Optimized, "cdn.example.com")
	assert.NotNil(t, firstImage.Responsive)
	assert.Contains(t, firstImage.Responsive, "thumbnail")

	// Test with empty slice
	emptyResult := ProcessImageURLs([]string{})
	assert.Nil(t, emptyResult)
}
