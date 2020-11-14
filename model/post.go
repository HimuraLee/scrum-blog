package model

import (
	"blog/db"
	"time"
)

type Post struct {
	ID              int      `json:"id"`
	CateID          int      `json:"cate_id"  gorm:"column:cate_id"`
	CateName string `json:"cate_name"`
	Status          int      `json:"status"`
	Title           string   `json:"title"`
	Passwd string `json:"passwd"`
	Filename string `json:"filename"`
	MarkdownContent string   `json:"markdown_content"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func PostByID(sql *db.SqlClient, id []int) ([]*Post, error) {
	p := make([]*Post, 0)
	if err := sql.Table(db.PostTable).Where("id IN (?)", id).Find(&p).Error; err != nil {
		return nil, err
	} else {
		return p, nil
	}
}

func PostAll(sql *db.SqlClient, pg *Page) ([]*Post, error) {
	p := make([]*Post, 0)
	if err := sql.Table(db.PostTable).Order("id desc").Offset(pg.Ps*(pg.Pi-1)).Limit(pg.Ps).Find(&p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func PostAllCount(sql *db.SqlClient) (int, error) {
	var cnt int64
	if err := sql.Table(db.PostTable).Find(&[]*Post{}).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return int(cnt), nil
}

func PostByCateID(sql *db.SqlClient, id int, pg *Page) ([]*Post, error) {
	var err error
	p := make([]*Post, 0)
	if pg == nil {
		err = sql.Table(db.PostTable).Where("cate_id = ?", id).Find(&p).Error
	} else {
		err = sql.Table(db.PostTable).Where("cate_id = ?", id).Order("id desc").Offset(pg.Ps*(pg.Pi-1)).Limit(pg.Ps).Find(&p).Error
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func PostByTagID(sql *db.SqlClient, id int) ([]*Post, error) {
	p := make([]*Post, 0)
	err := sql.Table(db.PostTable).Where("id IN (?)", sql.Table(db.PostTagTable).Where("tag_id = ?", id).Select("post_id").Find(&[]*PostTag{})).Find(&p).Error
	if err != nil {
		return nil, err
	}
	return p, nil
}

func PostByCateIDCount(sql *db.SqlClient, id int) (int, error) {
	var cnt int64
	err := sql.Table(db.PostTable).Where("cate_id = ?", id).Find(&[]*Post{}).Count(&cnt).Error
	if err != nil {
		return 0, err
	}
	return int(cnt), nil
}

func PostGetFuzzy(sql *db.SqlClient, fzTitle string, pg *Page) ([]*Post, error) {
	p := make([]*Post, 0)
	if err := sql.Table(db.PostTable).Where("title LIKE '%?%'", fzTitle).Order("id desc").Offset(pg.Ps*(pg.Pi-1)).Limit(pg.Ps).Find(&p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func PostFuzzyCount(sql *db.SqlClient, fzTitle string) (int, error) {
	var cnt int64
	if err := sql.Table(db.PostTable).Where("title LIKE '%?%'", fzTitle).Find(&[]*Post{}).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return int(cnt), nil
}

func PostAdd(sql *db.SqlClient, p *Post) error {
	return sql.Table(db.PostTable).Create(p).Error
}

func PostDrop(sql *db.SqlClient, id int) error {
	return sql.Table(db.PostTable).Where("id = ?", id).Delete(&Post{}).Error
}

func PostCateDrop(sql *db.SqlClient, id int) error {
	return sql.Table(db.PostTable).Where("cate_id = ?", id).Updates(map[string]interface{}{
		"cate_id": 1,
		"cate_name": "未分类",
	}).Error
}

func PostEdit(sql *db.SqlClient, p *Post) error {
	return  sql.Table(db.PostTable).Where("id = ?", p.ID).Updates(p).Error
}
