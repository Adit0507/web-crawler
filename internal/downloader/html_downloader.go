package downloader

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// handles http reqs to fetch web pages
type HTMLDownloader struct {
	client    *http.Client
	userAgent string
}

func NewHTMLDownloader() *HTMLDownloader {
	return &HTMLDownloader{
		client: &http.Client{
			Timeout: 30*time.Second,	
		},
		userAgent: "WebCrawler/1.0",
	}
}

func (d *HTMLDownloader) Dowload(url string)(io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", d.userAgent)
	
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return resp.Body, nil
}

