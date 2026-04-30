package data

import "gomall/app/catalog/internal/data/ent"

func NewTestData(client *ent.Client) *Data {
	return &Data{db: client}
}
