package main

import (
	"encoding/json"
	"net/http"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/samcday/hosted-javadocsets/docset"
	"github.com/samcday/hosted-javadocsets/mavencentral"
)

func main() {
	m := martini.Classic()
	m.Use(render.Renderer())

	m.Get("/", func(r render.Render) {
		r.HTML(200, "home", map[string]string{}, render.HTMLOptions{
			Layout: "layout",
		})
	})

	m.Get("/search", func(res http.ResponseWriter, req *http.Request) {
		results, err := mavencentral.Search(req.URL.Query().Get("q"), 10)
		if err != nil {
			panic(err)
		}
		if err := json.NewEncoder(res).Encode(results); err != nil {
			panic(err)
		}
	})

	m.Get("/feed/:groupId/:artifactId/Hosted_Javadocset:_Tested.xml", func(r render.Render) {
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
