package provider

import (
	"github.com/KnutZuidema/godarr/pkg/model"
)

type Provider interface {
	ListBySearch(search string) ([]*model.Item, error)
	GetByID(id string) (*model.Item, error)
}
