package db

import (
	"github.com/KnutZuidema/godarr/pkg/model"
)

type Database interface {
	GetItem(id string) (*model.Item, error)
	GetItemByExternalID(externalID string) (*model.Item, error)
	CreateItem(item *model.Item) (*model.Item, error)
	ListItems(offset, count int) ([]*model.Item, error)
}
