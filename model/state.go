package model

import (
	"blog/db"
)

// State 统计信息
type State struct {
	Post int64 `json:"post"`
	Cate int64 `json:"cate"`
	Tag  int64 `json:"tag"`
}

func Collect(sql *db.SqlClient) (*State, error) {
	var posts, cates, tags int64
	if err := sql.Table(db.PostTable).Count(&posts).Error; err != nil {
		return nil, err
	}
	if err := sql.Table(db.CateTable).Count(&cates).Error; err != nil {
		return nil, err
	}
	if err := sql.Table(db.TagTable).Count(&tags).Error; err != nil {
		return nil, err
	}
	return &State{
		Post: posts,
		Cate: cates,
		Tag:  tags,
	}, nil
}