package database

import (
	"io"

	"github.com/jmoiron/sqlx"

	"github.com/KnutZuidema/godarr/pkg/model"
)

type Database interface {
	io.Closer
	GetItem(id string) (*model.Item, error)
	GetItemByExternalID(externalID string) (*model.Item, error)
	CreateItem(item *model.Item) (*model.Item, error)
	ListItems(offset, count int) ([]*model.Item, error)
	SetItemStatus(id string, status model.ItemStatus) error
	GetItemStatus(id string) (model.ItemStatus, error)
}

const (
	getItem = `
		select * from item where id = $1
	`

	getItemByExternalID = `
		select * from item where external_id = $1
	`

	createItem = `
		insert into item values (
			:id,
			:external_id,
			:kind,
			:title,
			:description,
			:image_path,
			:rating
		) on conflict (id) do update set
			external_id=:external_id,
			kind=:kind,
			title=:title,
			description=:description,
			image_path=:image_path,
			rating=:rating
		returning *
	`

	listItems = `
		select * from item
		order by id
		offset $1 limit $2
	`

	setItemStatus = `
		insert into item_status values (
			$1, now(), $2
		)
	`

	getItemStatus = `
		select status from item_status
		where item_id = $1
		order by received_at desc
		limit 1
	`
)

type database struct {
	getItem             *sqlx.Stmt
	getItemByExternalID *sqlx.Stmt
	createItem          *sqlx.NamedStmt
	listItems           *sqlx.Stmt
	setItemStatus       *sqlx.Stmt
	getItemStatus       *sqlx.Stmt
}

func New(db *sqlx.DB) (Database, error) {
	getItem, err := db.Preparex(getItem)
	if err != nil {
		return nil, err
	}
	getItemByExternalID, err := db.Preparex(getItemByExternalID)
	if err != nil {
		return nil, err
	}
	createItem, err := db.PrepareNamed(createItem)
	if err != nil {
		return nil, err
	}
	listItems, err := db.Preparex(listItems)
	if err != nil {
		return nil, err
	}
	setItemStatus, err := db.Preparex(setItemStatus)
	if err != nil {
		return nil, err
	}
	getItemStatus, err := db.Preparex(getItemStatus)
	if err != nil {
		return nil, err
	}
	return &database{
		getItem:             getItem,
		getItemByExternalID: getItemByExternalID,
		createItem:          createItem,
		listItems:           listItems,
		setItemStatus:       setItemStatus,
		getItemStatus:       getItemStatus,
	}, nil
}

func (d *database) Close() error {
	for _, stmt := range []io.Closer{d.getItem, d.getItemByExternalID, d.createItem, d.listItems, d.getItemStatus, d.setItemStatus} {
		if err := stmt.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (d *database) GetItem(id string) (*model.Item, error) {
	var item model.Item
	if err := d.getItem.Get(&item, id); err != nil {
		return nil, err
	}
	return &item, nil
}

func (d *database) GetItemByExternalID(externalID string) (*model.Item, error) {
	var item model.Item
	if err := d.getItemByExternalID.Get(&item, externalID); err != nil {
		return nil, err
	}
	return &item, nil
}

func (d *database) CreateItem(item *model.Item) (*model.Item, error) {
	var res model.Item
	if err := d.createItem.Get(&res, item); err != nil {
		return nil, err
	}
	return &res, nil
}

func (d *database) ListItems(offset, count int) ([]*model.Item, error) {
	items := []*model.Item{}
	if err := d.listItems.Select(&items, offset, count); err != nil {
		return nil, err
	}
	return items, nil
}

func (d *database) SetItemStatus(id string, status model.ItemStatus) error {
	if _, err := d.setItemStatus.Exec(id, status); err != nil {
		return err
	}
	return nil
}

func (d *database) GetItemStatus(id string) (model.ItemStatus, error) {
	var res string
	if err := d.getItemStatus.Get(&res, id); err != nil {
		return "", err
	}
	return model.ItemStatus(res), nil
}
