package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func fetchFeed() (Feed, error) {
	req, err := http.NewRequest("GET", "https://www.techmeme.com/", nil)
	if err != nil {
		return Feed{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; techmeme-cli)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Feed{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Feed{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return Feed{}, err
	}

	var feed Feed

	doc.Find("div.item").Each(func(_ int, s *goquery.Selection) {
		anchor := s.Find("a.ourh").First()
		if anchor.Length() == 0 {
			return
		}

		title := strings.TrimSpace(anchor.Text())
		url, _ := anchor.Attr("href")
		source := strings.TrimSpace(s.Find("cite").First().Text())
		if source == "" {
			source = strings.TrimSpace(s.Find("span.src").First().Text())
		}
		timeStr := strings.TrimSpace(s.Find("span.time").First().Text())

		if url == "" || title == "" {
			return
		}

		h := Headline{
			Title:  title,
			URL:    url,
			Source: source,
			Time:   timeStr,
		}

		// Discussion links
		s.Find("div.ii a, a.ii").Each(func(_ int, a *goquery.Selection) {
			dTitle := strings.TrimSpace(a.Text())
			dURL, _ := a.Attr("href")
			dSource := strings.TrimSpace(a.Find("cite").Text())
			if dSource == "" {
				dSource = strings.TrimSpace(a.Find("span.src").Text())
			}
			if dTitle != "" && dURL != "" {
				h.Discussion = append(h.Discussion, Discussion{
					Title:  dTitle,
					URL:    dURL,
					Source: dSource,
				})
			}
		})

		// Commentary
		s.Find(".cmtt").Each(func(_ int, c *goquery.Selection) {
			author := strings.TrimSpace(c.Find(".cmttauthor").Text())
			text := strings.TrimSpace(c.Find(".cmttxt").Text())
			cURL, _ := c.Find("a").First().Attr("href")
			cSource := strings.TrimSpace(c.Find("cite").Text())
			if cSource == "" {
				cSource = strings.TrimSpace(c.Find("span.src").Text())
			}
			if author != "" || text != "" {
				h.Commentary = append(h.Commentary, Commentary{
					Author: author,
					Text:   text,
					URL:    cURL,
					Source: cSource,
				})
			}
		})

		feed.Headlines = append(feed.Headlines, h)
	})

	return feed, nil
}
