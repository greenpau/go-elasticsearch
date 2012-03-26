// Go interface to elasticsearch.
package elasticsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Reference to an ElasticSearch server.
type ElasticSearch struct {
	// Base URL to elasticsearch
	URL string
}

type response struct {
	OK     bool                   `json:"ok"`
	Index  string                 `json:"_index"`
	Type   string                 `json:"_type"`
	Id     string                 `json:"_id"`
	Found  bool                   `json:"found"`
	Source map[string]interface{} `json:"_source"`
}

func (es *ElasticSearch) mkURL(index, doctype, id string,
	params map[string]string) string {

	paramv := url.Values{}
	for k, v := range params {
		paramv.Set(k, v)
	}

	return fmt.Sprintf("%s/%s/%s/%s?%s", es.URL,
		url.QueryEscape(index),
		url.QueryEscape(doctype),
		url.QueryEscape(id),
		paramv.Encode())
}

func handleResponse(resp *http.Response) (*response, error) {
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	d := json.NewDecoder(resp.Body)
	iresp := &response{}
	err := d.Decode(iresp)
	if err != nil {
		return nil, err
	}
	if !iresp.OK {
		return nil, errors.New("Response wasn't OK")
	}

	return iresp, nil
}

// Store a document in the index.
//
// The ID is optional in which case the ID will be generated by the
// server.
//
// Returns the new ID on success, otherwise an error.
func (es *ElasticSearch) Index(index, doctype, id string,
	doc map[string]interface{},
	params map[string]string) (string, error) {

	data, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Post(es.mkURL(index, doctype, id, params),
		"application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	iresp, err := handleResponse(resp)
	if err != nil {
		return "", err
	}

	return iresp.Id, nil
}

// Delete an index entry.
func (es *ElasticSearch) Delete(index, doctype, id string,
	params map[string]string) (bool, error) {

	req, err := http.NewRequest("DELETE", es.mkURL(index, doctype, id, params),
		nil)
	if err != nil {
		return false, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	iresp, err := handleResponse(resp)
	if err != nil {
		return false, err
	}

	return iresp.Found, nil
}
