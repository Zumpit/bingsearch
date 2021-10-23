package bingsearch

import (
	"context"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
	"strings"
)

type Result struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type SearchOptions struct {
	Limit     int
	Start     int
	UserAgent string
	OverLimit bool
	ProxyAddr string
}

const BaseUrl = "https://www.bing.com/search?q=site:\"linkedin.com\"+"

func Search(ctx context.Context, searchTerm string, opts ...SearchOptions) ([]Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := RateLimit.Wait(ctx); err != nil {
		return nil, err
	}
	c := colly.NewCollector(colly.MaxDepth(1))
	if len(opts) == 0 {
		opts = append(opts, SearchOptions{})
	}

	if opts[0].UserAgent == "" {
		c.UserAgent = "Mozilla/5.0 (Windows NT 10.0;Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36"
	} else {

		c.UserAgent = opts[0].UserAgent
	}

	results := []Result{}
	var rErr error
	rank := 1

	c.OnRequest(func(r *colly.Request) {
		if err := ctx.Err(); err != nil {
			r.Abrot()
			rErr = err
			return
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		rErr = err
	})

	c.OnHTML("li.b_algo", func(e *colly.HTMLElement) {
		sel := e.Dom

		linkHref, _ = sel.Find("a").Attr("href")
		linkText := strings.TrimSpace(linkHref)
		titleText := strings.TrimSpace(sel.Find("h2 > a"))

		if linkText != "" && linkText != "#" && titleText != "" {
			result := Result{
				URL:   linkText,
				Title: titleText,
			}
			results = append(results, result)
			rank += 1
		}
	})

	limit := opts[0].Limit
	if opts[0].OverLimit {
		limit = int(float64(opts[0].Limit) * 1.5)
	}

	url := url(searchTerm, limit, opts[0].Start)
	if opts[0].ProxyAddr != "" {
		rp, err := proxy.RoundRobinProxySwitcher(opts[0].ProxyAddr)

		if err != nil {
			return nil, err
		}

		c.SetProxyFunc(rp)
	}

	c.Visit(url)

	if rErr != nil {
		if strings.Contains(rErr.Error(), "Too many requests") {
			return nil, ErrBlocked
		}
		return nil, rErr
	}

	if opts[0].Limit != 0 && len(results) > opts[0].Limit {
		return results[:opts[0].Limit], nil
	}

	return results, nil
}

func base(url string) string {
	if strings.HasPrefix(url, "http") {
		return url
	} else {
		return BaseUrl + url
	}
}

func url(searchTerm string, limit int, start int) string {
	searchTerm = strings.Trim(searchTerm, " ")
	searchTerm = strings.Replace(searchTerm, " ", "+", -1)

	var url string

	if limit != 0 {
		url = fmt.Sprintf("%s&num=%d", url, limit)
	}
	return url
}
