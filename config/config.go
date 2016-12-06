package config

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/jinzhu/configor"
	"github.com/microcosm-cc/bluemonday"
	"github.com/qor/render"
)

type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Site     string
}

var Config = struct {
	Port uint   `default:"80" env:"PORT"`
	Host string `default:"172.31.19.190" env:"HOST"`
	DB   struct {
		Name     string `default:"qor_example"`
		Adapter  string `default:"mysql"`
		User     string
		Password string
	}
	SMTP SMTPConfig
}{}

var (
	Root   = os.Getenv("GOPATH") + "/src/github.com/sunwukonga/paypal-qor-admin"
	SGT, _ = time.LoadLocation("Asia/Singapore")
	View   *render.Render
)

func init() {
	if err := configor.Load(&Config, "config/database.yml", "config/smtp.yml"); err != nil {
		panic(err)
	}

	fmt.Println(Root)
	View = render.New()

	htmlSanitizer := bluemonday.UGCPolicy()
	View.RegisterFuncMap("raw", func(str string) template.HTML {
		return template.HTML(htmlSanitizer.Sanitize(str))
	})
}

func (s SMTPConfig) HostWithPort() string {
	return s.Host + ":" + s.Port
}
