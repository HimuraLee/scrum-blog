package model

import (
	"blog/db"
)

// Cate 分类
type Cate struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
}

func CateByID(sql *db.SqlClient, id int) (*Cate, error) {
	c := new(Cate)
	if err := sql.Table(db.CateTable).Where("id = ?", id).First(c).Error; err != nil {
		return nil, err
	} else {
		return c, nil
	}
}

func CateByName(sql *db.SqlClient, name string) (*Cate, error) {
	c := new(Cate)
	if err := sql.Table(db.CateTable).Where("name = ?", name).First(c).Error; err != nil {
		return nil, err
	} else {
		return c, nil
	}
}

func CateAll(sql *db.SqlClient) ([]*Cate, error) {
	c := make([]*Cate, 0)
	if err := sql.Table(db.CateTable).Find(&c).Error; err != nil {
		return nil, err
	} else {
		return c, nil
	}
}

func CateAdd(sql *db.SqlClient, c *Cate) error {
	return sql.Table(db.CateTable).Create(c).Error
}

func CateDrop(sql *db.SqlClient, id int) error {
	return sql.Table(db.CateTable).Where("id = ?", id).Delete(&Cate{}).Error
}

func CateEdit(sql *db.SqlClient, c *Cate) error {
	return sql.Table(db.CateTable).Where("id = ?", c.ID).Updates(c).Error
}

