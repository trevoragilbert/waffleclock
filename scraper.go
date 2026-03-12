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

	// Each story cluster lives in a div.clus. Inside it:
	//   First div.item      — the main headline row
	//   div.relitems        — optional related-item rows (same story, different angle)
	//   div.di (inside div#Np1, hidden) — full discussion links with titles
	doc.Find("div.clus").Each(func(_ int, clus *goquery.Selection) {
		// The very first div.item in the cluster is the main story.
		mainItem := clus.Find("div.item").First()
		if mainItem.Length() == 0 {
			return
		}

		// Main headline: STRONG with an L-class (L1–L4) containing A.ourh
		mainAnchor := mainItem.Find("strong[class] a.ourh").First()
		if mainAnchor.Length() == 0 {
			return
		}

		title := strings.TrimSpace(mainAnchor.Text())
		url, _ := mainAnchor.Attr("href")
		if url == "" || title == "" || strings.HasPrefix(url, "/r2/") {
			return // skip ads
		}

		// Source: last <a> in the <cite> of the headline's shrtbl
		citeEl := mainItem.Find("table.shrtbl cite").First()
		source := strings.TrimSpace(citeEl.Find("a").Last().Text())
		if source == "" {
			raw := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(citeEl.Text()), ":"))
			if idx := strings.LastIndex(raw, "/"); idx != -1 {
				source = strings.TrimSpace(raw[idx+1:])
			} else {
				source = raw
			}
		}

		h := Headline{Title: title, URL: url, Source: source}

		// Discussions: div.di items inside the main item's hidden expanded block.
		// Structure: div#Np1 > div.di > cite + a (title/url)
		// goquery can find hidden elements; display:none doesn't matter.
		mainItem.Find("div.di").EachWithBreak(func(_ int, di *goquery.Selection) bool {
			if len(h.Discussion) >= 3 {
				return false
			}
			anchors := di.Find("a")
			if anchors.Length() == 0 {
				return true
			}
			// Last <a> is the article link; cite's <a> is the publication.
			articleAnchor := anchors.Last()
			dTitle := strings.TrimSpace(articleAnchor.Text())
			dURL, _ := articleAnchor.Attr("href")
			if dTitle == "" || dURL == "" {
				return true
			}

			diCite := di.Find("cite")
			dSource := strings.TrimSpace(diCite.Find("a").Last().Text())
			if dSource == "" {
				raw := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(diCite.Text()), ":"))
				if idx := strings.LastIndex(raw, "/"); idx != -1 {
					dSource = strings.TrimSpace(raw[idx+1:])
				}
			}

			h.Discussion = append(h.Discussion, Discussion{
				Title:  dTitle,
				URL:    dURL,
				Source: dSource,
			})
			return true
		})

		feed.Headlines = append(feed.Headlines, h)
	})

	return feed, nil
}
