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

	// Each story cluster lives inside a div.itc2. Within it:
	//   div.item        — the main headline row
	//   div.di          — individual discussion/related-article items
	doc.Find("div.itc2").Each(func(_ int, cluster *goquery.Selection) {
		item := cluster.Find("div.item").First()
		if item.Length() == 0 {
			return
		}

		anchor := item.Find("strong.L1 a.ourh, a.ourh").First()
		if anchor.Length() == 0 {
			return
		}

		title := strings.TrimSpace(anchor.Text())
		url, _ := anchor.Attr("href")
		if url == "" || title == "" {
			return
		}

		// Source: publication name from the <cite> anchor inside the headline row
		citeEl := item.Find("table.shrtbl cite, cite").First()
		source := strings.TrimSpace(citeEl.Find("a").First().Text())
		if source == "" {
			// Fall back to full cite text, strip author prefix (e.g. "Author / Pub:")
			raw := strings.TrimSpace(citeEl.Text())
			if idx := strings.LastIndex(raw, "/"); idx != -1 {
				source = strings.TrimSpace(raw[idx+1:])
			} else {
				source = strings.TrimSuffix(strings.TrimSpace(raw), ":")
			}
		}

		h := Headline{
			Title:  title,
			URL:    url,
			Source: source,
		}

		// Discussion items: each div.di in the cluster is a related article.
		// Structure: <div class="di"><cite>Author / <a>Pub</a>:</cite> <a href="...">Title</a></div>
		cluster.Find("div.di").Each(func(_ int, di *goquery.Selection) {
			// The last <a> is the article link; the cite's <a> is the publication.
			anchors := di.Find("a")
			if anchors.Length() == 0 {
				return
			}
			articleAnchor := anchors.Last()
			dTitle := strings.TrimSpace(articleAnchor.Text())
			dURL, _ := articleAnchor.Attr("href")

			diCite := di.Find("cite")
			dSource := strings.TrimSpace(diCite.Find("a").First().Text())
			if dSource == "" {
				raw := strings.TrimSpace(diCite.Text())
				if idx := strings.LastIndex(raw, "/"); idx != -1 {
					dSource = strings.TrimSpace(raw[idx+1:])
				}
			}

			if dTitle != "" && dURL != "" {
				h.Discussion = append(h.Discussion, Discussion{
					Title:  dTitle,
					URL:    dURL,
					Source: dSource,
				})
			}
		})

		feed.Headlines = append(feed.Headlines, h)
	})

	return feed, nil
}
