package server

import (
	"html/template"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UIControllers struct {}

type HTTPError struct {
    StatusCode int
    Message string
}

func serveError(logger *logrus.Logger, c *gin.Context, httpError HTTPError) {
    c.Status(httpError.StatusCode)
    tmpl, err := template.ParseFiles("internal/server/templates/error.html")
    if err != nil {
        logger.Error(err)
        return
    }

    tmpl.Execute(c.Writer, httpError)
}

func (uc *UIControllers) Dashboard(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    tmpl, err := template.ParseFiles("internal/server/templates/index.html")
    if err != nil {
        logger.Error(err)
        serveError(logger, c, HTTPError{500, "Something went wrong!"})
        return
    }

    tmpl.Execute(c.Writer, nil)
}

func (uc *UIControllers) Dags(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    tmpl, err := template.ParseFiles("internal/server/templates/dags.html")
    if err != nil {
        logger.Error(err)
        serveError(logger, c, HTTPError{500, "Something went wrong!"})
        return
    }

    tmpl.Execute(c.Writer, nil)
}

func (uc *UIControllers) Dag(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    tmpl, err := template.ParseFiles("internal/server/templates/dag.html")
    if err != nil {
        logger.Error(err)
        serveError(logger, c, HTTPError{500, "Something went wrong!"})
        return
    }

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
        serveError(logger, c, HTTPError{404, "Dag not found!!"})
		return
	}

    var tmplData struct {
        ID int
    }
    tmplData.ID = id

    tmpl.Execute(c.Writer, tmplData)
}

func (uc *UIControllers) Executors(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    tmpl, err := template.ParseFiles("internal/server/templates/executors.html")
    if err != nil {
        logger.Error(err)
        serveError(logger, c, HTTPError{500, "Something went wrong!"})
        return
    }

    tmpl.Execute(c.Writer, nil)
}

