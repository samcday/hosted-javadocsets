package mavencentral

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	log "github.com/cihub/seelog"
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

// GetArtifact will download an artifact with given GAV and classifier from
// Maven Central. It returns an io.ReadCloser that represents an active http
// stream for the given artifact binary data. Caller must close this stream.
func GetArtifact(groupId, artifactId, version, classifier string) (io.ReadCloser, error) {
	log.Infof("Fetching %s %s:%s:%s from maven central", classifier, groupId, artifactId, version)

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

	searchUrl := searchEndpoint + "?" + q.Encode()
	log.Debug("Querying ", searchUrl)
	httpResp, err := http.Get(searchUrl)
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
	filePath := strings.Replace(artifact.GroupId, ".", "/", -1) + "/" + artifact.ArtifactId + "/" + artifact.Version + "/" + artifact.ArtifactId + "-" + artifact.Version + "-javadoc.jar"

	dlUrl := downloadEndpoint + "?filepath=" + url.QueryEscape(filePath)
	log.Debug("Downloading ", dlUrl)
	httpResp, err = http.Get(dlUrl)
	return httpResp.Body, err
}
