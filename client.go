package retraced

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	apiVersion = 2
)

// Client represents a client that can send events into the retraced service.
type Client struct {
	projectID string
	token     string
	Endpoint  string
}

// NewClient creates a new retraced api client that can be used to send events in
func NewClient(projectID string, apiToken string) (*Client, error) {
	return &Client{
		projectID: projectID,
		token:     apiToken,
		Endpoint:  "https://api.retraced.io",
	}, nil
}

// NewEventRecord is returned from the Retraced API when an event is created
type NewEventRecord struct {
	ID   string `json:"id"`
	Hash string `json:"hash"`
}

// ReportEvent is the method to call to send a new event.
func (c *Client) ReportEvent(event *Event) (*NewEventRecord, error) {
	event.apiVersion = apiVersion

	encoded, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/project/%s/event", c.Endpoint, c.projectID), bytes.NewBuffer(encoded))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token token=%s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Unexpected response from retraced api: %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var reqResp NewEventRecord
	if err := json.Unmarshal(bodyBytes, &reqResp); err != nil {
		return nil, err
	}

	return &reqResp, nil
}

// GetViewerToken will return a one-time use token that can be used to view a group's audit log.
func (c *Client) GetViewerToken(groupID string, isAdmin bool, targetID string) (*ViewerToken, error) {
	params := url.Values{}
	params.Add("group_id", groupID)
	params.Add("is_admin", strconv.FormatBool(isAdmin))

	if targetID != "" {
		params.Add("target_id", targetID)
	}

	u, err := url.Parse(fmt.Sprintf("%s/v1/project/%s/viewertoken", c.Endpoint, c.projectID))
	if err != nil {
		return nil, err
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token token=%s", c.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK { // There's a pending PR in the retraced API to match this.
		return nil, fmt.Errorf("Unexpected response from retraced api: %d", resp.StatusCode)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	viewerToken := ViewerToken{}
	if err := json.Unmarshal(contents, &viewerToken); err != nil {
		return nil, err
	}

	return &viewerToken, nil
}
