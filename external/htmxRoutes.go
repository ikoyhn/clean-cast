package external

import (
	"bytes"
	"html/template"
	"ikoyhn/podcast-sponsorblock/internal/database"
	"ikoyhn/podcast-sponsorblock/internal/models"
	"ikoyhn/podcast-sponsorblock/internal/services/apple"
	"net/http"

	"github.com/labstack/echo/v4"
)

func registerHtmxRoutes(e *echo.Echo) {
	e.Static("/", "../../public")
	e.GET("/", func(c echo.Context) error {
		return c.File("../../public/index.html")
	})

	e.GET("/dashboard", func(c echo.Context) error {
		appleResponse, err := apple.GetTopPodcasts(10)
		if err != nil {
			return c.HTML(http.StatusOK, "<h1>No podcasts found</h1>")
		}

		tmpl, err := template.ParseFiles("../../public/pages/dashboard/dashboard.html")
		if err != nil {
			return err
		}

		data := struct {
			AppleResponse []models.TopPodcast
		}{
			AppleResponse: appleResponse,
		}

		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, data)
		if err != nil {
			return err
		}

		return c.HTML(http.StatusOK, buf.String())
	})

	e.GET("/modal", func(c echo.Context) error {
		appleResponse, err := apple.GetTopPodcasts(10)
		if err != nil {
			return c.HTML(http.StatusOK, "<h1>No podcasts found</h1>")
		}

		tmpl, err := template.ParseFiles("../../public/pages/dashboard/modal.html")
		if err != nil {
			return err
		}

		data := struct {
			AppleResponse []models.TopPodcast
		}{
			AppleResponse: appleResponse,
		}

		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, data)
		if err != nil {
			return err
		}

		return c.HTML(http.StatusOK, buf.String())
	})

	e.GET("/podcasts", func(c echo.Context) error {
		podcasts, err := apple.GetAllPodcasts()
		if err != nil {
			return c.HTML(http.StatusOK, "<h1>No podcasts found</h1>")
		}

		tmpl, err := template.ParseFiles("../../public/pages/podcasts/podcasts.html")
		if err != nil {
			return err
		}

		data := struct {
			Podcasts []models.Podcast
		}{
			Podcasts: podcasts,
		}

		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, data)
		if err != nil {
			return err
		}

		return c.HTML(http.StatusOK, buf.String())
	})

	e.GET("/episodes/:id", func(c echo.Context) error {
		id := c.Param("id")
		episodes, err := database.GetPodcastEpisodesByPodcastId(id)
		if err != nil {
			return c.HTML(http.StatusOK, "<h1>No episodes found</h1>")
		}

		tmpl, err := template.ParseFiles("../../public/pages/podcasts/episodes.html")
		if err != nil {
			return err
		}

		data := struct {
			Episodes []models.PodcastEpisode
		}{
			Episodes: episodes,
		}

		buf := new(bytes.Buffer)
		err = tmpl.Execute(buf, data)
		if err != nil {
			return err
		}

		return c.HTML(http.StatusOK, buf.String())
	})

	e.GET("/downloads", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "<h1>Downloads</h1>")
	})
}
