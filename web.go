package main

import (
	"net/http"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/samcday/hosted-javadocsets/docset"
)

func main() {
	m := martini.Classic()
	m.Use(render.Renderer())

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
