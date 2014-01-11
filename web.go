package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/samcday/hosted-javadocsets/docset"
	"github.com/samcday/hosted-javadocsets/jobs"
	"github.com/samcday/hosted-javadocsets/mavencentral"
)

func main() {
	m := martini.Classic()
	m.Use(render.Renderer())

	var feedRoute martini.Route

	m.Get("/", func(r render.Render) {
		r.HTML(200, "home", map[string]string{}, render.HTMLOptions{
			Layout: "layout",
		})
	})

	m.Get("/search.json", func(res http.ResponseWriter, req *http.Request) {
		results, err := mavencentral.Search(req.URL.Query().Get("q"), 10)
		if err != nil {
			panic(err)
		}

		payload := make([]map[string]string, 0)
		for _, item := range results {
			payload = append(payload, map[string]string{
				"g":     item.GroupId,
				"a":     item.ArtifactId,
				"vc":    strconv.Itoa(item.VersionCount),
				"l":     item.LatestVersion,
				"value": item.Id,
			})
		}

		if err := json.NewEncoder(res).Encode(payload); err != nil {
			panic(err)
		}
	})

	m.Get("/artifact/:groupId/:artifactId", func(req *http.Request, params martini.Params, r render.Render) {
		feedUrl := feedRoute.URLWith([]string{params["groupId"], params["artifactId"], params["artifactId"]})
		absoluteFeedUrl := "http://localhost:5000" + feedUrl
		dashUrl := "dash-feed://" + url.QueryEscape(absoluteFeedUrl)

		view := map[string]interface{}{
			"Id":         "com.google.guava:guava",
			"ArtifactId": params["artifactId"],
			"URL":        feedUrl,
			"DashURL":    template.URL(dashUrl),
		}
		r.HTML(200, "artifact", view, render.HTMLOptions{
			Layout: "layout",
		})
	})

	feedRoute = m.Get("/feed/:groupId/Hosted_Javadocset_-_:artifactId.xml", func(r render.Render, params martini.Params, logger *log.Logger) {
		if err := jobs.QueueDocsetJob(params["groupId"], params["artifactId"], ""); err != nil {
			logger.Printf("Failed to queue docset job:", err)
		}
		r.HTML(200, "docset-feed", map[string]string{
			"Version":   "4.3.2.1",
			"DocsetUrl": "http://localhost:3000/Test.tgz",
		})
	})

	m.Get("/docset/:groupId/:artifactId/:version", func(res http.ResponseWriter, params martini.Params) {
		docset.Create(params["groupId"], params["artifactId"], params["version"], res)
	})

	m.Run()
}
