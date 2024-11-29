package server

import (
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)


func serveInternalServerError(c *gin.Context) {
    c.File("internal/server/templates/500.html")
}

func Dashboard(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    tmpl, err := template.ParseFiles("internal/server/templates/index.html")
    if err != nil {
        logger.Error(err)
        serveInternalServerError(c)
        return
    }

    tmpl.Execute(c.Writer, nil)
}

func Dags(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    tmpl, err := template.ParseFiles("internal/server/templates/dags.html")
    if err != nil {
        logger.Error(err)
        serveInternalServerError(c)
        return
    }

    tmpl.Execute(c.Writer, nil)
}

func Executors(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    tmpl, err := template.ParseFiles("internal/server/templates/executors.html")
    if err != nil {
        logger.Error(err)
        serveInternalServerError(c)
        return
    }

    tmpl.Execute(c.Writer, nil)
}

