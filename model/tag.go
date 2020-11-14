package model

import "blog/db"

// Tag 标签
type Tag struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
}

func TagByID(sql *db.SqlClient, id []int) ([]*Tag, error) {
	t := make([]*Tag, 0)
	if err := sql.Table(db.TagTable).Where("id IN (?)", id).Find(&t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

func TagByName(sql *db.SqlClient, name string) (*Tag, error) {
	t := new(Tag)
	if err := sql.Table(db.TagTable).Where("name = ?", name).First(t).Error; err != nil {
		return nil, err
	} else {
		return t, nil
	}
}


func TagAll(sql *db.SqlClient) ([]*Tag, error) {
	t := make([]*Tag, 0)
	if err := sql.Table(db.TagTable).Find(&t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

func TagAdd(sql *db.SqlClient, t *Tag) error {
	return sql.Table(db.TagTable).Create(t).Error
}

func TagDrop(sql *db.SqlClient, id int) error {
	return sql.Table(db.TagTable).Where("id = ?", id).Delete(&Tag{}).Error
}

func TagEdit(sql *db.SqlClient, t *Tag) error {
	return sql.Table(db.TagTable).Where("id = ?", t.ID).Updates(t).Error
}