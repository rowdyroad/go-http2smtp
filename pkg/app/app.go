package app

import (
	"context"
	"log"
	"net/http"
	"net/smtp"
	"net/textproto"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jordan-wright/email"
)


type authConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Config struct {
	Listen string `yaml:"listen"`
	Auth *authConfig `yaml:"auth"`
}

type request struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`

	FromEmail string `json:"fromEmail"`
	FromName  string `json:"fromName"`

	To          []string `json:"to"`
	Cc          []string `json:"cc"`
	Bcc         []string `json:"bcc"`
	ReadReceipt []string `json:"readReceipt"`

	Subject string `json:"subject"`

	Html string `json:"html"`
	Text string `json:"text"`
}

func formatFrom(req request) string {
	from := ""
	if req.FromName != "" {
		from = req.FromName
	}

	if from == "" {
		from = req.FromEmail
	} else {
		from = from + "<" + req.FromEmail + ">"
	}
	return from
}

type App struct {
	config Config
	server *http.Server
}

func NewApp(config Config) *App {
	app :=&App{
		config: config,
	}

	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/", func(c *gin.Context){
		c.Status(http.StatusOK);
	})

	router.POST("/", func(c *gin.Context) {
		if config.Auth != nil {
			log.Println("Need to auth");
			if username, password, ok := c.Request.BasicAuth(); !ok || username != config.Auth.Username || password != config.Auth.Password {
				log.Println("Forbidden")
				c.AbortWithStatus(http.StatusForbidden);
				return
			}
		}
		var req request
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithStatus(http.StatusBadRequest);
			return
		}

		e := &email.Email{
			To:          req.To,
			Bcc:         req.Bcc,
			Cc:          req.Cc,
			ReadReceipt: req.ReadReceipt,
			From:        formatFrom(req),
			Subject:     req.Subject,
			Text:        []byte(req.Text),
			HTML:        []byte(req.Html),
			Headers:     textproto.MIMEHeader{},
		}

		auth := smtp.PlainAuth("", req.Username, req.Password, req.Host)

		log.Println("Sending email");
		if err := e.Send(req.Host+":"+req.Port, auth); err != nil {
			log.Println("Error:", err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	})

	app.server = &http.Server{
		Addr:              config.Listen,
		Handler:           router,
	}

	return app
}


func (app *App) Run() {
	app.server.ListenAndServe()
}

func (app *App) Close() {
	app.server.Shutdown(context.Background());
}