//
// Copyright (c) 2019 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type oidcResponse struct {
	AccessToken string `json:"access_token"`
}

// GetServiceAccountToken retrieves service accout token from the OIDC provider
func GetServiceAccountToken(serviceAccountID, serviceAccountSecret, oidcProvider string) string {
	formURL := fmt.Sprintf("%s/token", oidcProvider)
	resp, err := http.PostForm(formURL,
		url.Values{
			"grant_type":    {"client_credentials"},
			"client_id":     {serviceAccountID},
			"client_secret": {serviceAccountSecret},
		})
	if err != nil {
		log.Fatalf("Could not get service account token: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Could not read response body when getting service account token")
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Fatalf("Failed to get service account token. Response: %s", body)
	}

	var token oidcResponse
	json.Unmarshal(body, &token)
	if token.AccessToken == "" {
		log.Fatalf("Failed to parse token from json response")
	}
	return token.AccessToken
}
