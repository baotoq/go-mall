package main

import (
	"context"
	"fmt"

	"catalog/ent"
)

type seedProduct struct {
	name        string
	slug        string
	description string
	imageURL    string
	price       float64
	stock       int64
}

func seedIfEmpty(ctx context.Context, db *ent.Client) error {
	count, err := db.Product.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("count products: %w", err)
	}
	if count > 0 {
		return nil
	}

	categories := []struct {
		name        string
		slug        string
		description string
		products    []seedProduct
	}{
		{
			name:        "Mac",
			slug:        "mac",
			description: "Power your work and play with Mac.",
			products: []seedProduct{
				{"MacBook Pro 14\"", "macbook-pro-14", "Supercharged by M3 Pro and M3 Max chips.", "https://www.apple.com/v/macbook-pro/bz/images/overview/hero/hero_intro_endframe__e6khcva4hkeq_large.jpg", 199900, 25},
				{"MacBook Air 15\"", "macbook-air-15", "Big screen energy. Mighty M3 chip.", "https://www.apple.com/v/macbook-air/n/images/overview/hero/hero_intro__bxoaxa9b1iq6_large.jpg", 129900, 40},
				{"iMac", "imac", "Strikingly thin design. Strikingly fun colors.", "https://www.apple.com/v/imac/p/images/overview/hero/hero_intro_endframe__c2g7vtw8d1ye_large.jpg", 129900, 15},
			},
		},
		{
			name:        "iPhone",
			slug:        "iphone",
			description: "The latest iPhone lineup.",
			products: []seedProduct{
				{"iPhone 15 Pro", "iphone-15-pro", "Titanium. So strong. So light. So Pro.", "https://www.apple.com/v/iphone-15-pro/c/images/overview/hero/hero_intro_endframe__b1m9pdpfk2g6_large.jpg", 99900, 100},
				{"iPhone 15", "iphone-15", "New camera. New design. Newphoria.", "https://www.apple.com/v/iphone-15/d/images/overview/hero/hero_intro_endframe__c2g7vtw8d1ye_large.jpg", 79900, 150},
			},
		},
		{
			name:        "iPad",
			slug:        "ipad",
			description: "Lovable. Drawable. Magical.",
			products: []seedProduct{
				{"iPad Pro", "ipad-pro", "The thinnest Apple product ever.", "https://www.apple.com/v/ipad-pro/al/images/overview/hero/hero_intro_endframe__e6khcva4hkeq_large.jpg", 99900, 30},
				{"iPad Air", "ipad-air", "Two sizes. Faster chip. Still Air.", "https://www.apple.com/v/ipad-air/h/images/overview/hero/hero_intro_endframe__c2g7vtw8d1ye_large.jpg", 59900, 50},
			},
		},
		{
			name:        "Watch",
			slug:        "watch",
			description: "Smarter. Brighter. Mightier.",
			products: []seedProduct{
				{"Apple Watch Series 10", "apple-watch-series-10", "Thinstant classic.", "https://www.apple.com/v/apple-watch-series-10/c/images/overview/hero/hero_intro_endframe__b1m9pdpfk2g6_large.jpg", 39900, 80},
				{"Apple Watch Ultra 2", "apple-watch-ultra-2", "The most rugged and capable Apple Watch.", "https://www.apple.com/v/apple-watch-ultra-2/h/images/overview/hero/hero_intro_endframe__e6khcva4hkeq_large.jpg", 79900, 35},
			},
		},
		{
			name:        "Vision",
			slug:        "vision",
			description: "Welcome to the era of spatial computing.",
			products: []seedProduct{
				{"Apple Vision Pro", "apple-vision-pro", "A revolutionary spatial computer.", "https://www.apple.com/v/apple-vision-pro/d/images/overview/hero/hero_intro_endframe__c2g7vtw8d1ye_large.jpg", 349900, 10},
			},
		},
	}

	for _, cat := range categories {
		c, err := db.Category.Create().
			SetName(cat.name).
			SetSlug(cat.slug).
			SetDescription(cat.description).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("seed category %s: %w", cat.slug, err)
		}
		for _, p := range cat.products {
			if _, err := db.Product.Create().
				SetName(p.name).
				SetSlug(p.slug).
				SetDescription(p.description).
				SetImageURL(p.imageURL).
				SetPrice(p.price).
				SetTotalStock(p.stock).
				SetRemainingStock(p.stock).
				SetCategoryID(c.ID).
				Save(ctx); err != nil {
				return fmt.Errorf("seed product %s: %w", p.slug, err)
			}
		}
	}
	return nil
}
