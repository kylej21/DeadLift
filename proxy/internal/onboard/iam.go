package onboard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"proxy/internal/models"
)

const (
	googleTokenURL    = "https://oauth2.googleapis.com/token"
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

type iamPolicy struct {
	Bindings []iamBinding `json:"bindings"`
	Etag     string       `json:"etag"`
	Version  int          `json:"version"`
}

type iamBinding struct {
	Role    string   `json:"role"`
	Members []string `json:"members"`
}

func exchangeCode(clientID, clientSecret, redirectURI, code string) (string, *models.UserInfo, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}
	resp, err := http.PostForm(googleTokenURL, data)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var tr struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", nil, fmt.Errorf("parse token response: %w", err)
	}
	if tr.Error != "" {
		return "", nil, fmt.Errorf("%s: %s", tr.Error, tr.ErrorDesc)
	}
	info, err := getUserInfo(tr.AccessToken)
	if err != nil {
		return "", nil, fmt.Errorf("get user info: %w", err)
	}
	return tr.AccessToken, info, nil
}

func getUserInfo(token string) (*models.UserInfo, error) {
	req, _ := http.NewRequest("GET", googleUserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var info models.UserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse user info: %w", err)
	}
	return &info, nil
}

func grantPermissions(ctx context.Context, token, projectID, dlqSubName, topicName, saEmail string) error {
	member := "serviceAccount:" + saEmail

	dlqResource := dlqSubName
	if !strings.HasPrefix(dlqSubName, "projects/") {
		dlqResource = fmt.Sprintf("projects/%s/subscriptions/%s", projectID, dlqSubName)
	}
	if err := setPubSubIAM(ctx, token, dlqResource, "roles/pubsub.subscriber", member); err != nil {
		return fmt.Errorf("DLQ subscription: %w", err)
	}

	topicResource := topicName
	if !strings.HasPrefix(topicName, "projects/") {
		topicResource = fmt.Sprintf("projects/%s/topics/%s", projectID, topicName)
	}
	if err := setPubSubIAM(ctx, token, topicResource, "roles/pubsub.publisher", member); err != nil {
		return fmt.Errorf("main topic: %w", err)
	}

	if err := setProjectIAM(ctx, token, projectID, "roles/logging.viewer", member); err != nil {
		return fmt.Errorf("project logging: %w", err)
	}
	return nil
}

func setPubSubIAM(ctx context.Context, token, resource, role, member string) error {
	base := "https://pubsub.googleapis.com/v1/" + resource
	policy, err := getIAMPolicy(ctx, token, base+":getIamPolicy", "GET")
	if err != nil {
		return fmt.Errorf("getIamPolicy: %w", err)
	}
	addMemberToRole(policy, role, member)
	return postIAMPolicy(ctx, token, base+":setIamPolicy", policy)
}

func setProjectIAM(ctx context.Context, token, projectID, role, member string) error {
	base := "https://cloudresourcemanager.googleapis.com/v1/projects/" + projectID
	policy, err := getIAMPolicy(ctx, token, base+":getIamPolicy", "POST")
	if err != nil {
		return fmt.Errorf("getIamPolicy: %w", err)
	}
	addMemberToRole(policy, role, member)
	return postIAMPolicy(ctx, token, base+":setIamPolicy", policy)
}

func getIAMPolicy(ctx context.Context, token, endpoint, method string) (*iamPolicy, error) {
	var req *http.Request
	var err error
	if method == "GET" {
		req, err = http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBufferString("{}"))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}
	var policy iamPolicy
	if err := json.Unmarshal(body, &policy); err != nil {
		return nil, fmt.Errorf("parse policy: %w", err)
	}
	return &policy, nil
}

func postIAMPolicy(ctx context.Context, token, endpoint string, policy *iamPolicy) error {
	body, err := json.Marshal(map[string]interface{}{"policy": policy})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, b)
	}
	return nil
}

func addMemberToRole(policy *iamPolicy, role, member string) {
	for i, b := range policy.Bindings {
		if b.Role == role {
			for _, m := range b.Members {
				if m == member {
					return
				}
			}
			policy.Bindings[i].Members = append(b.Members, member)
			return
		}
	}
	policy.Bindings = append(policy.Bindings, iamBinding{Role: role, Members: []string{member}})
}
