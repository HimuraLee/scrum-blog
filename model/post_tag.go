package model

import (
	"blog/db"
	"fmt"
)

//PostTag 文章标签
type PostTag struct {
	ID     int `json:"id"`
	PostID int `json:"post_id" gorm:"column:post_id"`
	TagID  int `json:"tag_id" gorm:"column:tag_id"`
}

func TagsByPostID(sql *db.SqlClient, pid int) ([]int, error) {
	pt := make([]*PostTag, 0)
	err := sql.Table(db.PostTagTable).Where("post_id = ?", pid).Find(&pt).Error
	if err != nil {
		return nil, err
	}
	tid := make([]int, len(pt))
	for i := range pt {
		tid[i] = pt[i].TagID
	}
	return tid, nil
}

func PostTagAdd(sql *db.SqlClient, pid int, tid []int) error {
	if len(tid) == 0 {
		return nil
	}
	raw := fmt.Sprintf("INSERT INTO %s(post_id, tag_id) VALUES", db.PostTagTable)
	for i := range tid {
		raw = fmt.Sprintf("%s (%d, %d)", raw, pid, tid[i])
		if i < len(tid) - 1 {
			raw += fmt.Sprintf("%s,\n", raw)
		} else {
			raw += ";"
		}
	}
	return sql.Exec(raw).Error
}

func PostTagDrop(sql *db.SqlClient, pid int, tid []int) error {
	return sql.Table(db.PostTagTable).Where("post_id = ? AND tag_id IN (?)", pid, tid).Delete(&PostTag{}).Error
}

func PostTagsDrop(sql *db.SqlClient, pid int) error {
	return sql.Table(db.PostTagTable).Where("post_id = ?", pid).Delete(&PostTag{}).Error
}

func PostsTagDrop(sql *db.SqlClient, tid int) error {
	return sql.Table(db.PostTagTable).Where("tag_id = ?", tid).Delete(&PostTag{}).Error
}