package mavencentral

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	log "github.com/cihub/seelog"
)

const searchEndpoint = "http://search.maven.org/solrsearch/select"
const downloadEndpoint = "http://search.maven.org/remotecontent"

const templateStr = `"g:"{{.GroupId}}" AND a:"{{.ArtifactId}}" AND v:"{{.Version}}" AND l:"javadoc" AND p:"jar"`

var queryTemplate *template.Template = nil

type responseHeader struct {
	Status int `json:"status"`
	QTime  int
	Params map[string]string `json:"params"`
}

type artifact struct {
	Id         string   `json:"id"`
	GroupId    string   `json:"g"`
	ArtifactId string   `json:"a"`
	Packaging  string   `json:"p"`
	Extensions []string `json:"ec"`
	Timestamp  int      `json:"timestamp"`
}

type lookupResponseDoc struct {
	artifact
	Version string   `json:"v"`
	Tags    []string `json:"tags"`
}

type lookupResponseBody struct {
	NumFound int                 `json:"numFound"`
	Start    int                 `json:"start"`
	Docs     []lookupResponseDoc `json:"docs"`
}

type lookupResponse struct {
	ResponseHeader responseHeader     `json:"responseHeader"`
	Response       lookupResponseBody `json:"response"`
}

type searchResponse struct {
	ResponseHeader responseHeader     `json:"responseHeader"`
	Response       searchResponseBody `json:"response"`
}

type searchResponseBody struct {
	NumFound int            `json:"numFound"`
	Start    int            `json:"start"`
	Docs     []SearchResult `json:"docs"`
}

type SearchResult struct {
	artifact
	LatestVersion string   `json:"latestVersion"`
	RepositoryId  string   `json:"repositoryId"`
	VersionCount  int      `json:"versionCount"`
	Text          []string `json:"text"`
}

// Search executes a keyword query and returns the results.
func Search(s string, numResults int) ([]SearchResult, error) {
	log.Infof("Searching maven central for term %s", s)

	q := url.Values{}
	q.Add("q", s)
	q.Add("rows", strconv.Itoa(numResults))
	q.Add("wt", "json")

	searchUrl := searchEndpoint + "?" + q.Encode()
	log.Debugf("Querying %s", searchUrl)
	httpResp, err := http.Get(searchUrl)
	if err != nil {
		return nil, err
	}

	defer httpResp.Body.Close()
	decoder := json.NewDecoder(httpResp.Body)

	var resp searchResponse
	if err := decoder.Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Response.Docs, nil
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
	var resp lookupResponse
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
