package service

import (
	"blog/config"
	"blog/internal/jwt"
	"bytes"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	JwtToken    = "token"
)

func NewAuthMiddleware(key string) func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			tokenRaw := ctx.FormValue(JwtToken)
			if tokenRaw == "" {
				tokenRaw = ctx.Request().Header.Get(echo.HeaderAuthorization)
				if tokenRaw == "" {
					ctx.JSON(AuthFailed("Jwt token not found", "re-login please"))
					return nil
				}
				tokenRaw = tokenRaw[7:] // Bearer token len("Bearer ")==7
			}
			jwtAuth, err := jwt.Verify(tokenRaw, key)
			if err == nil {
				ctx.Set("auth", jwtAuth)
				ctx.Set("uid", jwtAuth.ID)
			} else {
				return ctx.JSON(AuthFailed("Jwt token verified failed","re-login please"))
			}
			return next(ctx)
		}
	}
}

func NewRecoverMiddleware() func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, 1<<10)
					length := runtime.Stack(stack, false)
					os.Stdout.Write(stack[:length])
					ctx.Error(err)
				}
			}()
			return next(ctx)
		}
	}
}

func NewAccessLogMiddleware(cfg *config.Config) func(echo.HandlerFunc) echo.HandlerFunc {
	if cfg.HTTP.AccessLog {
		acLog = &logrus.Logger{
			Formatter: new(logrus.JSONFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.InfoLevel,
		}
		os.MkdirAll(filepath.Dir(cfg.HTTP.AcLogPath), 0777)
		// 设置 rotatelogs
		logWriter, err := rotatelogs.New(
			// 分割后的文件名称
			cfg.HTTP.AcLogPath + ".%Y%m%d.log",
			// 生成软链，指向最新日志文件
			rotatelogs.WithLinkName(cfg.HTTP.AcLogPath),
			// 设置最大保存时间(7天)
			rotatelogs.WithMaxAge(7*24*time.Hour),
			// 设置日志切割时间间隔(1天)
			rotatelogs.WithRotationTime(24*time.Hour),
		)
		if err != nil {
			log.Errorf("failed to set rotatelogs, %s", err)
		}
		acLog.SetOutput(logWriter)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) (err error) {
			start := time.Now()
			if err = next(ctx); err != nil {
				ctx.Error(err)
			}
			stop := time.Now()
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			defer pool.Put(buf)
			buf.WriteString("\tip：" + ctx.RealIP())
			buf.WriteString("\tmethod：" + ctx.Request().Method)
			buf.WriteString("\tpath：" + ctx.Request().RequestURI)
			buf.WriteString("\tspan：" + stop.Sub(start).String())
			acLog.Infof(buf.String())
			return
		}
	}
}

var pool *sync.Pool
var acLog *logrus.Logger

func init() {
	pool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 512))
		},
	}
}

var crosConfig = middleware.CORSConfig{
	AllowOrigins: []string{"*"},
	AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
}