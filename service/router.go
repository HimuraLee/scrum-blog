package service

import (
	"github.com/labstack/echo/v4/middleware"
)

func (svc *BackendService) ApiAdminRouter() {
	svc.echo.Use(NewRecoverMiddleware(), NewAccessLogMiddleware(svc.cfg))
	svc.echo.Use(middleware.CORSWithConfig(crosConfig))
	svc.echo.Static("/.well-known", "/root/himura/")
	v1 := svc.echo.Group(`/api/v1`)

	v1.POST(`/login`, svc.UserLogin)            // 登陆

	if svc.cfg.Auth.Enable {
		v1.Use(NewAuthMiddleware(svc.cfg.Auth.JwtKey))
	}

	v1.GET(`/sys`, svc.Sys)                      // 服务器信息
	v1.GET(`/collect`, svc.Collect)              // 统计信息

	v1.POST(`/img`, svc.UploadImg)         // 图片上传

	v1.GET(`/auth`, svc.UserAuth)                // 获取当前登陆信息
	v1.PATCH(`/user`, svc.UserEdit) // 修改自身信息
	v1.PATCH(`/user/passwd`, svc.UserPass)          // 修改密码

	v1.DELETE(`/cate/:id`, svc.CateDrop)   // 删除分类
	v1.POST(`/cate`, svc.CateAdd)        // 添加分类
	v1.PATCH(`/cate`, svc.CateEdit)      // 编辑分类
	v1.GET(`/cates`, svc.CateAll)         // 分类列表

	v1.DELETE(`/post/:id`, svc.PostDrop)   // 删除文章/页面
	v1.POST(`/post`, svc.PostAddOrEdit)     // 文章/页面-编辑/添加
	v1.GET(`/posts`, svc.PostGetFuzzy)     // 文章/页面-编辑/添加
	v1.GET(`/posts`, svc.PostAll)         // 页面
	v1.GET(`/post/:id`, svc.PostGet)        // 文章
	v1.GET(`/cate/:id/posts`, svc.PostByCateID)     // 通过分类查询文章


	v1.DELETE(`/tag/:id`, svc.TagDrop)     // 删除标签
	v1.POST(`/tag`, svc.TagAdd)          // 添加标签
	v1.PATCH(`/tag`, svc.TagEdit)        // 编辑标签
	v1.GET(`/tags`,  svc.TagAll)           // 标签列表
	v1.GET(`/post/:id/tags`, svc.TagByPostID)

	v1.POST(`/info`, svc.InfoEdit)
	v1.GET(`/info`, svc.InfoBase)
}