package parser

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/SKharchenko87/foodix-parser/internal/config"
	"github.com/SKharchenko87/foodix-parser/internal/models"
)

type Calorizator struct {
	client http.Client
	cfg    config.SourceConfig
}

func NewCalorizator(cfg config.SourceConfig) Parser {
	res := new(Calorizator)
	res.client = http.Client{Timeout: time.Duration(cfg.Timeout) * time.Millisecond}
	res.cfg = cfg
	return res
}

func (c Calorizator) Parse() (res []models.Product, err error) {
	var chanProduct = make(chan models.Product, c.cfg.ProductPerPage)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for product := range chanProduct {
			res = append(res, product)
		}
	}()

	// Обрабатываем первую страницу
	doc, err := c.fetchPage(c.cfg.URL + "?page=0")
	if err != nil {
		return nil, fmt.Errorf("could not fetch page index=0: %w", err)
	}
	pageCount, err := parseCountPages(doc)
	if err != nil {
		return nil, fmt.Errorf("could not fetch pageCount: %w", err)
	}
	sizeProduct := c.cfg.ProductPerPage * pageCount
	res = make([]models.Product, 0, sizeProduct)

	err = parseProducts(doc, chanProduct)
	if err != nil {
		return nil, fmt.Errorf("could not parse products: %w", err)
	}

	// Обрабатываем оставшиеся страницы
	for i := 1; i < pageCount; i++ {
		time.Sleep(2 * time.Second) // Что бы не заблокировали
		doc, err = c.fetchPage(c.cfg.URL + "?page=" + strconv.Itoa(i))
		if err != nil {
			return nil, fmt.Errorf("could not fetch page index=%d: %w", i, err)
		}
		err = parseProducts(doc, chanProduct)
		if err != nil {
			return nil, fmt.Errorf("could not parse products: %w", err)
		}
	}
	close(chanProduct)

	wg.Wait()
	return res, nil
}

func (c Calorizator) fetchPage(url string) (doc *goquery.Document, err error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("response is nil")
	}
	body := resp.Body
	defer func(body io.ReadCloser) {
		err = body.Close()
		if err != nil {
			doc = nil
		}
	}(body)

	doc, err = goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}
	return doc, err
}

func parseCountPages(doc *goquery.Document) (int, error) {
	href, _ := doc.Find("li.pager-last a").Attr("href")
	u, _ := url.Parse(href)
	q := u.Query().Get("page")

	res, err := strconv.Atoi(q)
	if err != nil {
		return 0, err
	}
	return res + 1, nil
}

func parseProducts(doc *goquery.Document, ch chan models.Product) error {
	var err error
	doc.Find("#main-content tr.even, #main-content tr.odd").
		EachWithBreak(func(i int, s *goquery.Selection) bool {
			name, ok := parseName(s, &err)
			if !ok {
				return false
			}

			protein, ok := parseNutrient(s, i, "protein", &err)
			if !ok {
				return false
			}

			fat, ok := parseNutrient(s, i, "fat", &err)
			if !ok {
				return false
			}

			carbohydrate, ok := parseNutrient(s, i, "carbohydrate", &err)
			if !ok {
				return false
			}

			kcal, ok := parseKcal(s, i, "kcal", &err)
			if !ok {
				return false
			}

			product := models.Product{
				Name:         name,
				Protein:      protein,
				Fat:          fat,
				Carbohydrate: carbohydrate,
				Kcal:         kcal,
			}
			ch <- product
			return true
		})
	return err
}

func parseNutrient(s *goquery.Selection, i int, name string, err *error) (float64, bool) {
	selector := fmt.Sprintf(".views-field-field-%s-value", name)
	nutrientStr := strings.TrimSpace(s.Find(selector).Text())
	if nutrientStr == "" {
		return 0, true
	}
	nutrient, parseErr := strconv.ParseFloat(nutrientStr, 64)
	if parseErr != nil {
		*err = fmt.Errorf("invalid %s value at row %d: %v", name, i, parseErr)
		return 0, false
	}
	return nutrient, true
}

func parseKcal(s *goquery.Selection, i int, name string, err *error) (int, bool) {
	selector := fmt.Sprintf(".views-field-field-%s-value", name)
	kcalStr := strings.TrimSpace(s.Find(selector).Text())
	if kcalStr == "" {
		return 0, true
	}

	kcal, parseErr := strconv.Atoi(kcalStr)
	if parseErr != nil {
		*err = fmt.Errorf("invalid %s value at row %d: %v", name, i, parseErr)
		return 0, false
	}
	return kcal, true
}

func parseName(s *goquery.Selection, err *error) (name string, ok bool) {
	ok = true
	selector := ".views-field-title a"
	name = strings.TrimSpace(s.Find(selector).Text())
	if name == "" {
		*err = fmt.Errorf("product name is empty")
		ok = false
	}
	return name, ok
}
