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

	doc.Find("div.clus").Each(func(_ int, clus *goquery.Selection) {
		mainItem := clus.Find("div.item").First()
		if mainItem.Length() == 0 {
			return
		}

		mainAnchor := mainItem.Find("strong[class] a.ourh").First()
		if mainAnchor.Length() == 0 {
			return
		}

		title := strings.TrimSpace(mainAnchor.Text())
		url, _ := mainAnchor.Attr("href")
		if url == "" || title == "" || strings.HasPrefix(url, "/r2/") {
			return
		}

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

		// The hidden expanded block (div#Np1) contains multiple div.dbpt sections:
		//   First dbpt  → "More:" news articles → Discussion (limit 3)
		//   Later dbpts → LinkedIn/Bluesky/Threads/Forums → Commentary (limit 3)
		mainItem.Find("div.dbpt").Each(func(i int, dbpt *goquery.Selection) {
			drhed := strings.TrimSpace(strings.TrimSuffix(dbpt.Find("div.drhed, span.drhed").First().Text(), ":"))
			isArticles := i == 0 || drhed == "More"

			dbpt.Find("div.di").EachWithBreak(func(_ int, di *goquery.Selection) bool {
				anchors := di.Find("a")
				if anchors.Length() == 0 {
					return true
				}
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

				if isArticles {
					if len(h.Discussion) >= 3 {
						return false
					}
					h.Discussion = append(h.Discussion, Discussion{
						Title:  dTitle,
						URL:    dURL,
						Source: dSource,
					})
				} else {
					author := dSource
					h.Commentary = append(h.Commentary, Commentary{
						Author: author,
						Text:   dTitle,
						URL:    dURL,
						Source: drhed, // "LinkedIn", "Bluesky", etc.
					})
				}
				return true
			})
		})

		feed.Headlines = append(feed.Headlines, h)
	})

	return feed, nil
}
