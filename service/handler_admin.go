package service

import (
	"blog/db"
	"blog/internal/crypto"
	"blog/internal/jwt"
	"blog/internal/md5view"
	"blog/internal/other"
	"blog/model"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/*----------------------------------------------------------------*/
// 获取操作系统信息
/*----------------------------------------------------------------*/
func (svc *BackendService) Sys(ctx echo.Context) error {
	state := struct {
		ARCH    string `json:"arch"`
		OS      string `json:"os"`
		Version string `json:"version"`
		NumCPU  int    `json:"num_cpu"`
	}{
		ARCH:    runtime.GOARCH,
		OS:      runtime.GOOS,
		Version: runtime.Version(),
		NumCPU:  runtime.NumCPU(),
	}
	return ctx.JSON(Suc("系统信息", state))
}

/*----------------------------------------------------------------*/
// 获取博客统计信息
/*----------------------------------------------------------------*/
func (svc *BackendService) Collect(ctx echo.Context) error {
	s, err := model.Collect(svc.sql)
	if err == nil {
		return ctx.JSON(Suc("统计信息", s))
	}
	return ctx.JSON(BadRequest("查询统计信息失败", err))
}

/*----------------------------------------------------------------*/
// markdown上传图片
/*----------------------------------------------------------------*/
func (svc *BackendService) UploadImg(ctx echo.Context) error {
	file, err := ctx.FormFile("img")
	if err != nil {
		return ctx.JSON(BadRequest("未发现文件,请重试", err))
	}
	src, err := file.Open()
	if err != nil {
		return ctx.JSON(BadRequest("文件打开失败,请重试", err))
	}
	defer src.Close()
	os.MkdirAll("image", 0777)
	basePath := filepath.Join("image", time.Now().Format(md5view.ShortSeqTime))
	mask := syscall.Umask(0)
	defer syscall.Umask(mask)
	err = os.MkdirAll(basePath, 0777)
	if err != nil {
		return ctx.JSON(BadRequest("创建目录失败,请重试", err))
	}
	fileExt := filepath.Ext(file.Filename)
	fileBaseName := strings.Split(filepath.Base(file.Filename), ".")[0]
	fileName := fileBaseName + fileExt
	var filePathName string
	k := 0
	for {
		if k > 0 {
			fileName = fmt.Sprintf("%s(%d)%s", fileBaseName, k, fileExt)
		}
		filePathName = filepath.Join(basePath, fileName)
		exist, err := other.PathExist(filePathName)
		if err != nil {
			return ctx.JSON(BadRequest("检查文件状态失败", err))
		} else if !exist {
			tmp := func() error {
				dst, err := os.Create(filePathName)
				if err != nil {
					return ctx.JSON(BadRequest("目标文件创建失败,请重试", err))
				}
				defer dst.Close()
				if _, err = io.Copy(dst, src); err != nil {
					return ctx.JSON(BadRequest("文件写入失败,请重试", err))
				}
				return nil
			}
			if err = tmp(); err != nil {
				return err
			}
			break
		}
		k ++
	}
	return ctx.JSON(Suc("文件上传成功", "/" + filePathName))
}

/*----------------------------------------------------------------*/
// 修改用户信息
/*----------------------------------------------------------------*/
func (svc *BackendService) UserEdit(ctx echo.Context) error {
	user := new(model.User)
	err := ctx.Bind(user)
	if err != nil {
		return ctx.JSON(BadRequest("数据输入错误,请重试", err))
	}
	tx := svc.sql.Begin()
	defer func() {
		if v := recover(); v != nil || err != nil {
			tx.Rollback()
		}
	}()
	sql := &db.SqlClient{DB: tx}
	err = model.UserEdit(sql, user)
	if err != nil {
		return ctx.JSON(BadRequest("用户信息修改失败", err))
	}
	if user.AuthorName != "" {
		err = md5view.EditConfigJS(filepath.Join(svc.cfg.Storage.VuePressBlogPath, "docs/.vuepress/config/theme/theme.js"), "author", user.AuthorName)
		if err != nil {
			return ctx.JSON(BadRequest("js信息修改失败", err))
		}
	}
	tx.Commit()
	return ctx.JSON(Suc("用户信息修改成功"))
}

/*----------------------------------------------------------------*/
// 修改用户密码
/*----------------------------------------------------------------*/
func (svc *BackendService) UserPass(ctx echo.Context) error {
	ipt := struct {
		Opass string `json:"opass" form:"opass"`
		Npass string `json:"npass" form:"npass"`
	}{}
	err := ctx.Bind(&ipt)
	if err != nil {
		return ctx.JSON(BadRequest("输入数据有误", err))
	}
	uid := ctx.Get("uid").(int)
	user, err := model.UserByID(svc.sql, uid)
	if err != nil {
		return ctx.JSON(BadRequest("未找到用户", err))
	}
	_op := crypto.CheckPassWord(ipt.Opass, user.Passwd)
	if  _op != user.Passwd {
		return ctx.JSON(BadRequest("原始密码输入错误,请重试"))
	}
	err = model.UserEdit(svc.sql, &model.User{ID: uid, Passwd: crypto.CheckPassWord(ipt.Npass, user.Passwd)})
	if err != nil {
		return ctx.JSON(BadRequest("密码修改失败", err))
	}
	return ctx.JSON(Suc("密码修改成功"))
}

/*----------------------------------------------------------------*/
// 删除分类
/*----------------------------------------------------------------*/
func (svc *BackendService) CateDrop(ctx echo.Context) error {
	cid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(BadRequest("获取分类id失败", err))
	}
	tx := svc.sql.Begin()
	defer func() {
		if v := recover(); v != nil || err != nil {
			tx.Rollback()
		}
	}()
	sql := &db.SqlClient{DB: tx}
	c, err := model.CateByID(sql, cid)
	if err != nil {
		return ctx.JSON(BadRequest("获取分类信息失败,请重试", err))
	}
	err = model.CateDrop(sql, cid)
	if err != nil {
		return ctx.JSON(BadRequest("分类删除失败,请重试", err))
	}
	p, err := model.PostByCateID(sql, cid, nil)
	if err != nil {
		return ctx.JSON(BadRequest("获取分类文章失败,请重试", err))
	}
	err = model.PostCateDrop(sql, cid)
	if err != nil {
		return ctx.JSON(BadRequest("更新对应文章分类失败,请重试", err))
	}
	if len(p) > 0 {
		for i := 0; i < len(p); i ++ {
			err = md5view.EditVuePressDoc(filepath.Join(svc.cfg.Storage.VuePressBlogPath, "docs/views", p[i].CreatedAt.Format(md5view.LongSeqTime) + ".md"), "categories", "other", c.Name)
			if err != nil {
				return ctx.JSON(BadRequest("转移分类下文章失败,请重试", err))
			}
		}
		go md5view.YarnBuild(svc.cfg.Script.VisitorBuildScript, svc.cfg.Log.VuePressLogPath)
	}
	tx.Commit()
	return ctx.JSON(Suc("分类删除成功"))
}

/*----------------------------------------------------------------*/
// 添加分类
/*----------------------------------------------------------------*/
func (svc *BackendService) CateAdd(ctx echo.Context) error {
	c := new(model.Cate)
	err := ctx.Bind(c)
	if err != nil {
		return ctx.JSON(BadRequest("数据输入错误,请重试", err))
	}
	tx := svc.sql.Begin()
	defer func() {
		if v := recover(); v != nil || err != nil {
			tx.Rollback()
		}
	}()
	sql := &db.SqlClient{DB: tx}
	_, err = model.CateByName(sql, c.Name)
	if err != nil && err != gorm.ErrRecordNotFound {
		return ctx.JSON(BadRequest("查詢分类信息失败,请重试", err))
	} else if err == nil {
		return ctx.JSON(Suc("添加分类成功"))
	}
	err = model.CateAdd(sql, c)
	if err != nil {
		return ctx.JSON(BadRequest("添加分类失败,请重试", err))
	}
	tx.Commit()
	return ctx.JSON(Suc("添加分类成功"))
}

/*----------------------------------------------------------------*/
// 编辑分类
/*----------------------------------------------------------------*/
func (svc *BackendService) CateEdit(ctx echo.Context) error {
	c := new(model.Cate)
	err := ctx.Bind(c)
	if err != nil {
		return ctx.JSON(BadRequest("数据输入错误,请重试", err))
	}
	tx := svc.sql.Begin()
	defer func() {
		if v := recover(); v != nil || err != nil {
			tx.Rollback()
		}
	}()
	sql := &db.SqlClient{DB: tx}
	oc, err := model.CateByID(sql, c.ID)
	if err != nil {
		return ctx.JSON(BadRequest("获取分类信息失败,请重试", err))
	}
	err = model.CateEdit(sql, c)
	if err != nil {
		return ctx.JSON(BadRequest("修改分类信息失败,请重试", err))
	}

	p, err := model.PostByCateID(sql, c.ID, nil)
	if err != nil {
		return ctx.JSON(BadRequest("获取分类文章失败,请重试", err))
	}
	if len(p) > 0 {
		for i := 0; i < len(p); i ++ {
			err = md5view.EditVuePressDoc(filepath.Join(svc.cfg.Storage.VuePressBlogPath, "docs/views", p[i].CreatedAt.Format(md5view.LongSeqTime) + ".md"), "categories", c.Name, oc.Name)
			if err != nil {
				return ctx.JSON(BadRequest("修改文章信息失败,请重试", err))
			}
		}
		go md5view.YarnBuild(svc.cfg.Script.VisitorBuildScript, svc.cfg.Log.VuePressLogPath)
	}
	tx.Commit()
	return ctx.JSON(Suc("分类修改成功"))
}

/*----------------------------------------------------------------*/
// 删除文章
/*----------------------------------------------------------------*/
func (svc *BackendService) PostDrop(ctx echo.Context) error {
	pid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(BadRequest("获取文章id失败", err))
	}
	tx := svc.sql.Begin()
	defer func() {
		if v := recover(); v != nil || err != nil {
			tx.Rollback()
		}
	}()
	sql := &db.SqlClient{DB: tx}
	p, err := model.PostByID(sql, []int{pid})
	if err != nil {
		return ctx.JSON(BadRequest("获取文章信息失败,请重试", err))
	}
	if len(p) < 1 {
		return ctx.JSON(NotFound("未查询到文章信息", err))
	}
	err = model.PostDrop(sql, pid)
	if err != nil {
		return ctx.JSON(BadRequest("删除文章失败,请重试", err))
	}
	err = model.PostTagsDrop(sql, pid)
	if err != nil {
		return ctx.JSON(BadRequest("删除文章标签关系失败,请重试", err))
	}
	os.Remove(filepath.Join(svc.cfg.Storage.VuePressBlogPath, "docs/views", p[0].CreatedAt.Format(md5view.LongSeqTime) + ".md"))
	go md5view.YarnBuild(svc.cfg.Script.VisitorBuildScript, svc.cfg.Log.VuePressLogPath)
	tx.Commit()
	return ctx.JSON(Suc("删除成功"))
}

/*----------------------------------------------------------------*/
// 添加标签
/*----------------------------------------------------------------*/
func (svc *BackendService) TagAdd(ctx echo.Context) error {
	t := new(model.Tag)
	err := ctx.Bind(t)
	if err != nil {
		return ctx.JSON(BadRequest("数据输入错误,请重试", err))
	}
	_, err = model.TagByName(svc.sql, t.Name)
	if err != nil && err != gorm.ErrRecordNotFound {
		return ctx.JSON(BadRequest("查詢标签信息失败,请重试", err))
	} else if err == nil {
		return ctx.JSON(Suc("添加标签成功"))
	}
	err = model.TagAdd(svc.sql, t)
	if err != nil {
		return ctx.JSON(BadRequest("添加标签失败,请重试", err))
	}
	return ctx.JSON(Suc("添加标签成功"))
}

/*----------------------------------------------------------------*/
// 编辑标签
/*----------------------------------------------------------------*/
func (svc *BackendService) TagEdit(ctx echo.Context) error {
	t := new(model.Tag)
	err := ctx.Bind(t)
	if err != nil {
		return ctx.JSON(BadRequest("数据输入错误,请重试", err))
	}
	tx := svc.sql.Begin()
	defer func() {
		if v := recover(); v != nil || err != nil {
			tx.Rollback()
		}
	}()
	sql := &db.SqlClient{DB: tx}
	ot, err := model.TagByID(sql, []int{t.ID})
	if err != nil {
		return ctx.JSON(BadRequest("获取标签信息失败", err))
	}
	if len(ot) < 1 {
		return ctx.JSON(NotFound("未发现标签信息"))
	}
	err = model.TagEdit(sql, t)
	if err != nil {
		return ctx.JSON(BadRequest("标签修改失败", err))
	}
	p, err := model.PostByTagID(sql, t.ID)
	if err != nil {
		return ctx.JSON(BadRequest("查询标签关联文章失败", err))
	}
	if len(p) > 0 {
		for i := 0; i < len(p); i ++ {
			err = md5view.EditVuePressDoc(filepath.Join(svc.cfg.Storage.VuePressBlogPath, "docs/views", p[i].CreatedAt.Format(md5view.LongSeqTime) + ".md"), "tags", t.Name, ot[0].Name)
			if err != nil {
				return ctx.JSON(BadRequest("修改文章信息失败", err))
			}
		}
		go md5view.YarnBuild(svc.cfg.Script.VisitorBuildScript, svc.cfg.Log.VuePressLogPath)
	}
	tx.Commit()
	return ctx.JSON(Suc("标签修改成功"))
}

/*----------------------------------------------------------------*/
// 删除标签
/*----------------------------------------------------------------*/
func (svc *BackendService) TagDrop(ctx echo.Context) error {
	tid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(BadRequest("获取标签id失败", err))
	}
	tx := svc.sql.Begin()
	defer func() {
		if v := recover(); v != nil || err != nil {
			tx.Rollback()
		}
	}()
	sql := &db.SqlClient{DB: tx}
	ot, err := model.TagByID(sql, []int{tid})
	if err != nil {
		return ctx.JSON(BadRequest("获取标签失败,请重试", err))
	}
	if len(ot) < 1 {
		return ctx.JSON(NotFound("未获取到该标签"))
	}
	err = model.TagDrop(sql, tid)
	if err != nil {
		return ctx.JSON(BadRequest("标签删除失败,请重试", err))
	}
	p, err := model.PostByTagID(sql, tid)
	if err != nil {
		return ctx.JSON(BadRequest("查询标签关联文章失败", err))
	}
	err = model.PostsTagDrop(sql, tid)
	if err != nil {
		return ctx.JSON(BadRequest("文章标签关系删除失败,请重试", err))
	}
	if len(p) > 0 {
		for i := 0; i < len(p); i ++ {
			err = md5view.EditVuePressDoc(filepath.Join(svc.cfg.Storage.VuePressBlogPath, "docs/views", p[i].CreatedAt.Format(md5view.LongSeqTime) + ".md"), "tags", "", ot[0].Name)
			if err != nil {
				return ctx.JSON(BadRequest("修改文章信息失败", err))
			}
		}
		go md5view.YarnBuild(svc.cfg.Script.VisitorBuildScript, svc.cfg.Log.VuePressLogPath)
	}
	tx.Commit()
	return ctx.JSON(Suc("标签删除成功"))
}

/*----------------------------------------------------------------*/
// 获取所有分类文章，分页
/*----------------------------------------------------------------*/
func (svc *BackendService) PostAll(ctx echo.Context) error {
	ipt := &model.Page{}
	err := ctx.Bind(ipt)
	if err != nil {
		return ctx.JSON(BadRequest("数据输入错误,请重试", err))
	}
	p, err := model.PostAll(svc.sql, ipt)
	if err != nil {
		return ctx.JSON(BadRequest("查询页面失败", err))
	}
	return ctx.JSON(Suc("页面信息", p))
}

/*----------------------------------------------------------------*/
// 获取单篇文章
/*----------------------------------------------------------------*/
func (svc *BackendService) PostGet(ctx echo.Context) error {
	pid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(BadRequest("获取文章id失败", err))
	}
	p, err := model.PostByID(svc.sql, []int{pid})
	if err != nil {
		return ctx.JSON(BadRequest("查询页面失败", err))
	}
	if len(p) < 1 {
		return ctx.JSON(BadRequest("未查询到页面信息"))
	}
	return ctx.JSON(Suc("页面信息", p[0]))
}

/*----------------------------------------------------------------*/
// 根据关键词获取文章
/*----------------------------------------------------------------*/
func (svc *BackendService) PostGetFuzzy(ctx echo.Context) error {
	type param struct {
		FzTitle string `json:"fz_title"`
		Page *model.Page `json:"page"`
	}
	ipt := &param{}
	err := ctx.Bind(ipt)
	if err != nil {
		return ctx.JSON(BadRequest("数据输入错误,请重试", err))
	}
	cnt, err := model.PostFuzzyCount(svc.sql, ipt.FzTitle)
	if err != nil {
		return ctx.JSON(BadRequest("获取关键词文章总数失败,请重试", err))
	}
	p, err := model.PostGetFuzzy(svc.sql, ipt.FzTitle, ipt.Page)
	if err != nil {
		return ctx.JSON(BadRequest("获取关键词文章信息失败,请重试", err))
	}
	return ctx.JSON(Suc("文章信息", struct {
		Count int         `json:"count"`
		Items interface{} `json:"items"`
	}{cnt, p}))
}

/*----------------------------------------------------------------*/
// 根据分类获取文章
/*----------------------------------------------------------------*/
func (svc *BackendService) PostByCateID(ctx echo.Context) error {
	cid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(BadRequest("获取分类id失败", err))
	}
	ipt := &model.Page{}
	err = ctx.Bind(ipt)
	if err != nil {
		return ctx.JSON(BadRequest("数据输入错误,请重试", err))
	}
	if cid == 0 {
		p, err := model.PostAll(svc.sql, ipt)
		if err != nil {
			return ctx.JSON(BadRequest("获取文章信息失败,请重试", err))
		}
		cnt, err := model.PostAllCount(svc.sql)
		if err != nil {
			return ctx.JSON(BadRequest("获取文章总数信息失败,请重试", err))
		}
		return ctx.JSON(Suc("文章信息", struct {
			Count int         `json:"count"`
			Items interface{} `json:"items"`
		}{cnt, p}))
	}
	cnt, err := model.PostByCateIDCount(svc.sql, cid)
	if err != nil {
		return ctx.JSON(BadRequest("获取分类文章总数失败,请重试", err))
	}
	p, err := model.PostByCateID(svc.sql, cid, ipt)
	if err != nil {
		return ctx.JSON(BadRequest("获取分类文章信息失败,请重试", err))
	}
	return ctx.JSON(Suc("文章信息", struct {
		Count int         `json:"count"`
		Items interface{} `json:"items"`
	}{cnt, p}))
}

/*----------------------------------------------------------------*/
// 获取全部分类
/*----------------------------------------------------------------*/
func (svc *BackendService) CateAll(ctx echo.Context) error {
	c, err := model.CateAll(svc.sql)
	if err != nil {
		return ctx.JSON(BadRequest("查询页面失败", err))
	}
	return ctx.JSON(Suc("分类信息", c))
}

/*----------------------------------------------------------------*/
// 获取全部标签
/*----------------------------------------------------------------*/
func (svc *BackendService) TagAll(ctx echo.Context) error {
	t, err := model.TagAll(svc.sql)
	if err != nil {
		return ctx.JSON(BadRequest("查询标签失败", err))
	}
	return ctx.JSON(Suc("标签信息", t))
}

/*----------------------------------------------------------------*/
// 获取特定文章的所有标签
/*----------------------------------------------------------------*/
func (svc *BackendService) TagByPostID(ctx echo.Context) error {
	pid, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return ctx.JSON(BadRequest("获取文章id失败", err))
	}
	t, err := model.TagsByPostID(svc.sql, pid)
	if err != nil {
		return ctx.JSON(BadRequest("查询标签失败", err))
	}
	return ctx.JSON(Suc("标签信息", t))
}

/*----------------------------------------------------------------*/
// 编辑博客信息
/*----------------------------------------------------------------*/
func (svc *BackendService) InfoEdit(ctx echo.Context) error {
	var _update bool
	p := struct {
		Title string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
	}{}
	ctx.Bind(&p)
	if p.Title != "" {
		svc.cfg.Website.Title = p.Title
		_update = true
		err := md5view.EditConfigJS(filepath.Join(svc.cfg.Storage.VuePressBlogPath, "docs/.vuepress/config.js"), "title", p.Title)
		if err != nil {
			return ctx.JSON(BadRequest("js信息修改失败", err))
		}
	}
	if p.Description != "" {
		svc.cfg.Website.Description = p.Description
		_update = true
		err := md5view.EditConfigJS(filepath.Join(svc.cfg.Storage.VuePressBlogPath, "docs/.vuepress/config.js"), "description", p.Description)
		if err != nil {
			return ctx.JSON(BadRequest("js信息修改失败", err))
		}
	}
	if _update {
		err := viper.WriteConfig()
		if err != nil {
			return ctx.JSON(BadRequest("博客信息修改失败", err))
		}
		go md5view.YarnBuild(svc.cfg.Script.VisitorBuildScript, svc.cfg.Log.VuePressLogPath)
	}
	return ctx.JSON(Suc("博客信息修改成功"))
}

/*----------------------------------------------------------------*/
// 获取博客信息
/*----------------------------------------------------------------*/
func (svc *BackendService) InfoBase(ctx echo.Context) error {
	return ctx.JSON(Suc("博客信息", svc.cfg.Website))
}

/*----------------------------------------------------------------*/
// 新建/编辑文章
/*----------------------------------------------------------------*/
func (svc *BackendService) PostAddOrEdit(ctx echo.Context) error {
	ipt := &struct {
		Post model.Post `json:"post" form:"post"` // 文章信息
		Tags []int         `json:"tags" form:"tags"` // 标签
		Edit bool          `json:"edit" form:"edit"` // 是否编辑
	}{}
	err := ctx.Bind(ipt)
	if err != nil {
		return ctx.JSON(BadRequest("数据输入错误,请重试", err))
	}
	tx := svc.sql.Begin()
	defer func() {
		if v := recover(); v != nil || err != nil {
			tx.Rollback()
		}
	}()
	sql := &db.SqlClient{DB: tx}
	if ipt.Edit {
		p, err := model.PostByID(sql, []int{ipt.Post.ID})
		if err != nil {
			return ctx.JSON(BadRequest("获取文章信息失败,请重试", err))
		}
		if len(p) < 1 {
			err = ctx.JSON(BadRequest("未获取到文章信息"))
			return err
		}
		if ipt.Post.CateID != 0 {
			c, err := model.CateByID(sql, ipt.Post.CateID)
			if err != nil {
				return ctx.JSON(BadRequest("获取文章分类信息失败,请重试", err))
			}
			ipt.Post.CateName = c.Name
		}
		err = model.PostEdit(sql, &ipt.Post)
		if err != nil {
			return ctx.JSON(BadRequest("文章修改失败,请重试", err))
		}
		_old, err := model.TagsByPostID(sql, ipt.Post.ID)
		if err != nil {
			return ctx.JSON(BadRequest("获取文章标签关系失败", err))
		}
		_new := ipt.Tags
		add := make([]int, 0)
		del := make([]int, 0)
		newM := make(map[int]bool, len(_new))
		oldM := make(map[int]bool, len(_old))
		for _, v := range _new {
			newM[v] = true
		}
		for _, v := range _old {
			oldM[v] = true
		}
		for _, itm := range _old {
			if !newM[itm] {
				del = append(del, itm)
			}
		}
		for _, itm := range _new {
			if !oldM[itm] {
				add = append(add, itm)
			}
		}
		// 删除标签
		err = model.PostTagDrop(sql, ipt.Post.ID, del)
		if err != nil {
			return ctx.JSON(BadRequest("删除文章标签关系失败", err))
		}
		// 添加标签
		err = model.PostTagAdd(sql, ipt.Post.ID, add)
		if err != nil {
			return ctx.JSON(BadRequest("插入文章标签关系失败", err))
		}
		tags, err := model.TagByID(sql, ipt.Tags)
		if err != nil {
			return ctx.JSON(BadRequest("获取标签信息失败", err))
		}
		tagsName := make([]string, len(tags))
		for i := range tags {
			tagsName[i] = tags[i].Name
		}
		vpd := &md5view.VuePressDoc{
			FrontMatter: &md5view.FrontMatter {
				Title:      ipt.Post.Title,
				Tags:       tagsName,
				Publish:    ipt.Post.Status == 1,
				Date:       ipt.Post.CreatedAt.Format(md5view.LongSplitTime),
			},
			Doc:         ipt.Post.MarkdownContent,
		}
		if ipt.Post.CateID != 0 {
			vpd.FrontMatter.Categories = []string{ipt.Post.CateName}
		}
		if ipt.Post.Passwd != "" {
			h := md5.New()
			h.Write([]byte(ipt.Post.Passwd))
			vpd.FrontMatter.Passwd = []string{hex.EncodeToString(h.Sum(nil))}
		}
		path := filepath.Join(svc.cfg.Storage.VuePressBlogPath, "docs/views", ipt.Post.Filename + ".md")
		err = ioutil.WriteFile(path, []byte(vpd.String()), 0777)
		if err != nil {
			return ctx.JSON(BadRequest("写入Markdown文件失败", err))
		}
		go md5view.YarnBuild(svc.cfg.Script.VisitorBuildScript, svc.cfg.Log.VuePressLogPath)
		tx.Commit()
		return ctx.JSON(Suc("文章修改成功"))
	}
	c, err := model.CateByID(sql, ipt.Post.CateID)
	if err != nil {
		return ctx.JSON(BadRequest("获取文章分类信息失败,请重试", err))
	}
	ipt.Post.CateName = c.Name
	err = model.PostAdd(sql, &ipt.Post)
	if err != nil {
		return ctx.JSON(BadRequest("文章添加失败,请重试", err))
	}
	ipt.Post.Filename = ipt.Post.CreatedAt.Format(md5view.LongSeqTime)
	err = model.PostEdit(sql, &ipt.Post)
	if err != nil {
		return ctx.JSON(BadRequest("文章更新失败,请重试", err))
	}
	err = model.PostTagAdd(sql, ipt.Post.ID, ipt.Tags)
	if err != nil {
		return ctx.JSON(BadRequest("插入文章标签关系失败", err))
	}
	tags, err := model.TagByID(sql, ipt.Tags)
	if err != nil {
		return ctx.JSON(BadRequest("获取标签信息失败", err))
	}
	tagsName := make([]string, len(tags))
	for i := range tags {
		tagsName[i] = tags[i].Name
	}
	vpd := &md5view.VuePressDoc{
		FrontMatter: &md5view.FrontMatter{
			Title:      ipt.Post.Title,
			Tags:       tagsName,
			Categories: []string{ipt.Post.CateName},
			Publish:    ipt.Post.Status == 1,
			Date:       ipt.Post.CreatedAt.Format(md5view.LongSplitTime),
		},
		Doc:         ipt.Post.MarkdownContent,
	}
	if ipt.Post.Passwd != "" {
		h := md5.New()
		h.Write([]byte(ipt.Post.Passwd))
		vpd.FrontMatter.Passwd = []string{hex.EncodeToString(h.Sum(nil))}
	}
	path := filepath.Join(svc.cfg.Storage.VuePressBlogPath, "docs/views", ipt.Post.CreatedAt.Format(md5view.LongSeqTime) + ".md")
	err = ioutil.WriteFile(path, []byte(vpd.String()), 0777)
	if err != nil {
		return ctx.JSON(BadRequest("写入Markdown文件失败", err))
	}
	go md5view.YarnBuild(svc.cfg.Script.VisitorBuildScript, svc.cfg.Log.VuePressLogPath)
	tx.Commit()
	return ctx.JSON(Suc("文章添加成功"))
}

/*----------------------------------------------------------------*/
// 用户登录
/*----------------------------------------------------------------*/
func (svc *BackendService) UserLogin(ctx echo.Context) error {
	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := other.LoginLimiter.Wait(ct)
	if err != nil {
		return ctx.JSON(BadRequest("当前登录人数过多,请等待", err))
	}
	form := struct {
		Name   string `json:"name" form:"name"`
		Passwd string `json:"passwd" form:"passwd"`
	}{}
	err = ctx.Bind(&form)
	if err != nil {
		return ctx.JSON(BadRequest("请输入用户名和密码", err))
	}
	if form.Name == "" && len(form.Name) > 18 {
		return ctx.JSON(BadRequest("请输入正确的账号"))
	}
	user, err := model.UserByName(svc.sql, form.Name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ctx.JSON(BadRequest("该用户不存在"))
		}
		return ctx.JSON(BadRequest(err))
	}
	if crypto.CheckPassWord(form.Passwd, user.Passwd) != user.Passwd {
		return ctx.JSON(BadRequest("密码错误"))
	}
	auth := jwt.JwtAuth{
		ID:    user.ID,
		ExpAt: time.Now().Add(time.Hour * 6).Unix(),
	}
	return ctx.JSON(Suc("登录成功", auth.Encode(svc.cfg.Auth.JwtKey)))
}

/*----------------------------------------------------------------*/
// 用户认证
/*----------------------------------------------------------------*/
func (svc *BackendService) UserAuth(ctx echo.Context) error {
	user, err := model.UserByID(svc.sql, ctx.Get("uid").(int))
	if err != nil {
		return ctx.JSON(NotFound("没有找到该用户信息", err))
	}
	return ctx.JSON(Suc("登录用户信息", user))
}