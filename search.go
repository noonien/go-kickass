package kickass

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Torrent struct {
	Name     string
	Category string
	Uploader string
	Verified bool
	Magnet   string
	Torrent  string

	Size    int
	Files   int
	Age     string
	Seeds   int
	Leeches int
}

type SearchResults struct {
	Torrents []Torrent

	Pages      int
	Categories map[string]int
}

type SearchOptions struct {
	Page     int
	Category string

	Sort      string
	Ascending bool
}

func (c *Client) Search(query string, opt *SearchOptions) (*SearchResults, error) {
	// Default search options
	if opt == nil {
		opt = &SearchOptions{}
	}

	// Add category to query
	if opt.Category != "" {
		query += " category:" + opt.Category
	}

	// If a sort field is defined, add it to the query
	var sort string
	if opt.Sort != "" {
		order := "desc"
		if opt.Ascending {
			order = "asc"
		}

		sort = fmt.Sprintf("?field=%s&order=%s", opt.Sort, order)
	}

	// Pages start at 1
	if opt.Page < 1 {
		opt.Page = 1
	}

	url := fmt.Sprintf("usearch/%s/%d/%s", url.QueryEscape(query), opt.Page, sort)

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Grab categories
	categories := map[string]int{}
	doc.Find("ul.tabNavigation a").Each(func(i int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		parts := strings.Split(href, ":")

		// Skip non-category
		if len(parts) != 2 {
			return
		}

		category := strings.TrimRight(parts[1], "/")

		var no int
		noStr := sel.Find(".menuValue").Text()

		// If number ends in a k, multiply it by 1000
		if strings.HasSuffix(noStr, "k") {
			noK, _ := strconv.ParseFloat(strings.TrimRight(noStr, "k"), 32)
			no = int(noK * 1000)
		} else {
			no, _ = strconv.Atoi(noStr)
		}

		categories[category] = no
	})

	// Grab torrents
	torrents := extractTorrents(doc.Find("table.data tr"))

	// Total number of pages
	pages, _ := strconv.Atoi(doc.Find("div.pages a").Last().Text())

	return &SearchResults{
		Torrents:   torrents,
		Pages:      pages,
		Categories: categories,
	}, nil
}

func extractTorrents(sel *goquery.Selection) []Torrent {
	var torrents []Torrent
	sel.Each(func(i int, sel *goquery.Selection) {
		// Skip table header
		if sel.HasClass("firstr") {
			return
		}

		tds := sel.Find("td")
		t := Torrent{}

		nameDiv := tds.Eq(0).Find(".torrentname")
		t.Name = nameDiv.Find(".cellMainLink").Text()
		t.Category = nameDiv.Find("span strong a").Last().Text()
		t.Uploader = nameDiv.Find("span a").Eq(0).Text()

		iconsDiv := tds.Eq(0).Find(".iaconbox")
		t.Verified = iconsDiv.Find(".ka-green").Length() == 1
		t.Magnet, _ = iconsDiv.Find("a.imagnet").Attr("href")
		t.Torrent, _ = iconsDiv.Find("a.idownload").Eq(1).Attr("href")

		size := strings.Split(tds.Eq(1).Text(), " ")
		if len(size) == 2 {
			var sizeUnits float64 = 1
			switch size[1] {
			case "KB":
				sizeUnits = 1024
			case "MB":
				sizeUnits = 1024 * 1024
			case "GB":
				sizeUnits = 1024 * 1024 * 1024
			case "TB":
				sizeUnits = 1024 * 1024 * 1024 * 1024
			}
			sizeBase, _ := strconv.ParseFloat(size[0], 32)
			t.Size = int(sizeBase * sizeUnits)
		}

		t.Files, _ = strconv.Atoi(tds.Eq(2).Text())
		t.Age = strings.Replace(tds.Eq(3).Text(), "\u00a0", " ", -1)
		t.Seeds, _ = strconv.Atoi(tds.Eq(4).Text())
		t.Leeches, _ = strconv.Atoi(tds.Eq(5).Text())

		torrents = append(torrents, t)
	})

	return torrents
}
