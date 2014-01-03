package mavencentral

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"text/template"
)

const searchEndpoint = "http://search.maven.org/solrsearch/select"
const downloadEndpoint = "http://search.maven.org/remotecontent"

const templateStr = `"g:"{{.GroupId}}" AND a:"{{.ArtifactId}}" AND v:"{{.Version}}" AND l:"javadoc" AND p:"jar"`

var queryTemplate *template.Template = nil

type centralResponseHeader struct {
	Status int `json:"status"`
	QTime  int
	Params map[string]string `json:"params"`
}

type centralResponseDoc struct {
	Id         string   `json:"id"`
	GroupId    string   `json:"g"`
	ArtifactId string   `json:"a"`
	Version    string   `json:"v"`
	Packaging  string   `json:"p"`
	Timestamp  int      `json:"timestamp"`
	Tags       []string `json:"tags"`
	Extensions []string `json:"ec"`
}

type centralResponseBody struct {
	NumFound int                  `json:"numFound"`
	Start    int                  `json:"start"`
	Docs     []centralResponseDoc `json:"docs"`
}

type centralResponse struct {
	ResponseHeader centralResponseHeader `json:"responseHeader"`
	Response       centralResponseBody   `json:"response"`
}

func GetArtifact(groupId, artifactId, version, classifier string) (io.ReadCloser, error) {
	if queryTemplate == nil {
		t, err := template.New("query").Parse(templateStr)
		if err != nil {
			panic(err)
		}
		queryTemplate = t
	}

	q := url.Values{}
	q.Add("q", "g:\""+groupId+"\" AND a:\""+artifactId+"\" AND v:\""+version+"\" AND l:\""+classifier+"\" AND p:\"jar\"")
	q.Add("rows", "1")
	q.Add("wt", "json")

	httpResp, err := http.Get(searchEndpoint + "?" + q.Encode())
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	decoder := json.NewDecoder(httpResp.Body)
	var resp centralResponse
	if err := decoder.Decode(&resp); err != nil {
		return nil, err
	}
	if resp.Response.NumFound < 1 {
		return nil, errors.New("no artifact found")
	}

	artifact := resp.Response.Docs[0]
	filePath := strings.Replace(artifact.GroupId, ".", "/", -1) + "/" + artifact.ArtifactId + "/" + artifact.Version + "/" + artifact.ArtifactId + "-javadoc.jar"

	dlUrl := downloadEndpoint + "?filepath=" + url.QueryEscape(filePath)
	fmt.Println(dlUrl)
	httpResp, err = http.Get(dlUrl)
	return httpResp.Body, err
}
