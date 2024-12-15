package utfpr

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
)

type ProfileType string
type ProfileTypes []ProfileType

const (
	API_URL                       = "https://coensapp.dv.utfpr.edu.br/siacoes/service"
	ProfileType_Student           = "STUDENT"
	ProfileType_Professor         = "PROFESSOR"
	ProfileType_CompanySupervisor = "COMPANYSUPERVISOR" // não utilizado, mas pode retornar na API
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnexpectedStatus   = errors.New("unexpected status code")
)

func (p ProfileTypes) Contains(profileType ProfileType) bool {
	return slices.Contains(p, profileType)
}

type Profile struct {
	Name        string `json:"name"`
	Login       string `json:"login"`
	Email       string `json:"email"`
	StudentCode string `json:"studentCode"`
	Active      bool   `json:"active"`

	ProfileTypes ProfileTypes `json:"-"`
}

func Login(ctx context.Context, username, password string) (token string, err error) {
	b, err := json.Marshal(map[string]string{
		"login":    username,
		"password": password,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", API_URL+"/login", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrInvalidCredentials
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	}

	// body é o token
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func GetProfile(ctx context.Context, token string) (profile *Profile, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", API_URL+"/user/profile", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	profile = &Profile{}
	err = json.Unmarshal(body, profile)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequestWithContext(ctx, "GET", API_URL+"/user/list/profiles", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrUnexpectedStatus, resp.StatusCode)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &profile.ProfileTypes)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func GetProfileFromLogin(ctx context.Context, username, password string) (profile *Profile, err error) {
	token, err := Login(ctx, username, password)
	if err != nil {
		return nil, err
	}

	return GetProfile(ctx, token)
}
