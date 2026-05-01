package service

import (
	"encoding/json"
	nethttp "net/http"

	"gomall/app/catalog/internal/biz"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/uuid"
)

type seedResponse struct {
	CategoriesSeeded int `json:"categories_seeded"`
	ProductsSeeded   int `json:"products_seeded"`
}

type cleanResponse struct {
	CategoriesDeleted int `json:"categories_deleted"`
	ProductsDeleted   int `json:"products_deleted"`
}

type seedCategory struct {
	name        string
	slug        string
	description string
}

type seedProduct struct {
	name         string
	slug         string
	description  string
	priceCents   int64
	currency     string
	imageURL     string
	theme        string
	stock        int
	categorySlug string
}

var seedCategories = []seedCategory{
	{"Electronics", "electronics", "Gadgets, devices, and tech accessories"},
	{"Clothing", "clothing", "Fashion and apparel for all occasions"},
	{"Books", "books", "Fiction, non-fiction, and educational titles"},
	{"Sports", "sports", "Equipment and gear for active lifestyles"},
	{"Home & Garden", "home-garden", "Furniture, decor, and outdoor essentials"},
}

var seedProducts = []seedProduct{
	{"Wireless Headphones", "wireless-headphones", "Premium noise-cancelling over-ear headphones", 9999, "USD", "", "dark", 50, "electronics"},
	{"Mechanical Keyboard", "mechanical-keyboard", "Compact TKL keyboard with RGB backlight", 12999, "USD", "", "dark", 30, "electronics"},
	{"USB-C Hub", "usb-c-hub", "7-in-1 hub with HDMI, USB 3.0, and SD card reader", 3999, "USD", "", "light", 80, "electronics"},
	{"Smartphone Stand", "smartphone-stand", "Adjustable aluminum desk stand for phones and tablets", 1999, "USD", "", "light", 120, "electronics"},
	{"Classic White Tee", "classic-white-tee", "100% organic cotton crew-neck t-shirt", 2999, "USD", "", "light", 200, "clothing"},
	{"Slim Fit Jeans", "slim-fit-jeans", "Stretch denim in modern slim cut", 5999, "USD", "", "light", 150, "clothing"},
	{"Puffer Jacket", "puffer-jacket", "Lightweight water-resistant insulated jacket", 8999, "USD", "", "dark", 60, "clothing"},
	{"Running Sneakers", "running-sneakers", "Breathable mesh trainers with cushioned sole", 7999, "USD", "", "light", 90, "clothing"},
	{"Go Programming Language", "go-programming-language", "The definitive guide to Go by Alan Donovan", 4999, "USD", "", "light", 40, "books"},
	{"Clean Code", "clean-code", "Handbook of agile software craftsmanship by Robert Martin", 3999, "USD", "", "light", 55, "books"},
	{"Designing Data-Intensive Applications", "designing-data-intensive-applications", "Martin Kleppmann's guide to scalable systems", 5499, "USD", "", "dark", 35, "books"},
	{"Yoga Mat", "yoga-mat", "Non-slip 6mm thick eco-friendly mat", 3499, "USD", "", "light", 70, "sports"},
	{"Resistance Bands Set", "resistance-bands-set", "Set of 5 bands with varying resistance levels", 2499, "USD", "", "light", 100, "sports"},
	{"Ceramic Plant Pot", "ceramic-plant-pot", "Modern matte-finish pot, 6-inch diameter", 1499, "USD", "", "light", 130, "home-garden"},
	{"LED Desk Lamp", "led-desk-lamp", "Touch-dimmer lamp with warm and cool modes", 4499, "USD", "", "dark", 65, "home-garden"},
}

func (s *CatalogService) HandleSeed(w nethttp.ResponseWriter, r *nethttp.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == nethttp.MethodOptions {
		w.WriteHeader(nethttp.StatusNoContent)
		return
	}
	if r.Method != nethttp.MethodPost {
		nethttp.Error(w, "method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	categoryIDs := map[string]string{}
	var categoriesSeeded int

	for _, sc := range seedCategories {
		cat, err := s.cuc.Create(ctx, &biz.Category{
			Name:        sc.name,
			Slug:        sc.slug,
			Description: sc.description,
		})
		if err != nil {
			if !errors.IsConflict(err) {
				nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
				return
			}
			// already exists — resolve its ID for product association
			res, listErr := s.cuc.List(ctx, biz.ListCategoriesFilter{Page: 1, PageSize: 100})
			if listErr != nil {
				nethttp.Error(w, listErr.Error(), nethttp.StatusInternalServerError)
				return
			}
			for _, c := range res.Categories {
				if c.Slug == sc.slug {
					categoryIDs[sc.slug] = c.ID.String()
					break
				}
			}
			continue
		}
		categoryIDs[cat.Slug] = cat.ID.String()
		categoriesSeeded++
	}

	var productsSeeded int
	for _, sp := range seedProducts {
		p := &biz.Product{
			Name:        sp.name,
			Slug:        sp.slug,
			Description: sp.description,
			PriceCents:  sp.priceCents,
			Currency:    sp.currency,
			ImageURL:    sp.imageURL,
			Theme:       sp.theme,
			Stock:       sp.stock,
		}
		if catIDStr, ok := categoryIDs[sp.categorySlug]; ok {
			parsed, err := uuid.Parse(catIDStr)
			if err == nil {
				p.CategoryID = &parsed
			}
		}
		_, err := s.puc.Create(ctx, p)
		if err != nil {
			if !errors.IsConflict(err) {
				nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
				return
			}
			continue
		}
		productsSeeded++
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusOK)
	_ = json.NewEncoder(w).Encode(seedResponse{
		CategoriesSeeded: categoriesSeeded,
		ProductsSeeded:   productsSeeded,
	})
}

func (s *CatalogService) HandleClean(w nethttp.ResponseWriter, r *nethttp.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	if r.Method == nethttp.MethodOptions {
		w.WriteHeader(nethttp.StatusNoContent)
		return
	}
	if r.Method != nethttp.MethodDelete {
		nethttp.Error(w, "method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	// products must be deleted before categories (FK constraint)
	productsDeleted, err := s.puc.DeleteAll(ctx)
	if err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
		return
	}
	categoriesDeleted, err := s.cuc.DeleteAll(ctx)
	if err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(nethttp.StatusOK)
	_ = json.NewEncoder(w).Encode(cleanResponse{
		CategoriesDeleted: categoriesDeleted,
		ProductsDeleted:   productsDeleted,
	})
}
