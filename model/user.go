package model

import (
	"blog/db"
	"time"
)

// tips int(11)、tinyint(4)、smallint(6)、mediumint(9)、bigint(20)

// User 用户
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Passwd    string    `json:"passwd"`
	AuthorName string `json:"author_name"`
	CreatedAt time.Time `json:"created_at"`
}

//UserByName
func UserByName(sql *db.SqlClient, name string) (*User, error) {
	u := new(User)
	if err := sql.Table(db.UserTable).Where("name = ?", name).First(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

//UserByID
func UserByID(sql *db.SqlClient, id int) (*User, error) {
	u := new(User)
	if err := sql.Table(db.UserTable).Where("id = ?", id).First(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

//UserEdit 更新用户信息
func UserEdit(sql *db.SqlClient, u *User) error {
	return sql.Table(db.UserTable).Where("id = ?", u.ID).Updates(u).Error
}