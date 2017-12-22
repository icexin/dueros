package auth

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

var (
	// 百度token服务器url
	tokenUrl = "https://openapi.baidu.com/oauth/2.0/token"
	// 百度oauth服务器url
	oauthUrl = "https://openapi.baidu.com/oauth/2.0/authorize"

	redirectUri = "http://pi.local:8080/authresponse"

	tokenFile = "token.json"
)

var (
	clientID     = flag.String("client_id", "", "client id of oauth")
	clientSecret = flag.String("client_secret", "", "client secret of oauth")
	accessToken  = flag.String("access_token", "", "access token of oauth, if not empty, client_id and client_secret can leave empty")
)

type token struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

func (t *token) Save(f string) error {
	buf, _ := json.Marshal(t)
	return ioutil.WriteFile(f, buf, 0755)
}

func (t *token) Refresh() error {
	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", t.RefreshToken)
	values.Set("client_id", *clientID)
	values.Set("client_secret", *clientSecret)
	values.Set("scope", "basic")
	uri := fmt.Sprintf("%s?%s", tokenUrl, values.Encode())
	v, err := httpjson(uri)
	if err != nil {
		return err
	}
	t.AccessToken = v["access_token"].(string)
	t.RefreshToken = v["refresh_token"].(string)
	t.Expiry = time.Now().Add(time.Duration(int(v["expires_in"].(float64))) * time.Second)
	return nil
}

func loadToken(f string) (*token, error) {
	buf, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}
	t := token{}
	err = json.Unmarshal(buf, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func GetToken() (string, error) {
	if *accessToken != "" {
		return *accessToken, nil
	}
	t, err := loadToken(tokenFile)
	if err != nil {
		return "", err
	}
	if time.Now().After(t.Expiry) {
		log.Print("token expire, refresh")
		err = t.Refresh()
		if err != nil {
			return "", err
		}
		err = t.Save(tokenFile)
		if err != nil {
			return "", err
		}
		return t.AccessToken, nil
	}
	return t.AccessToken, nil
}

func httpjson(uri string) (map[string]interface{}, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	v := make(map[string]interface{})
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func login(w http.ResponseWriter, r *http.Request) {
	if *clientID == "" || *clientSecret == "" {
		fmt.Fprintf(w, "missing client_id or client_secret flag")
		return
	}
	_, err := GetToken()
	if err == nil {
		fmt.Fprint(w, "token ok")
		return
	}
	log.Print(err)
	values := url.Values{}
	values.Set("client_id", *clientID)
	values.Set("scope", "basic")
	values.Set("response_type", "code")
	values.Set("redirect_uri", redirectUri)
	uri := fmt.Sprintf("%s?%s", oauthUrl, values.Encode())
	http.Redirect(w, r, uri, 302)
}

func authResponse(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	code := r.FormValue("code")
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("client_id", *clientID)
	values.Set("client_secret", *clientSecret)
	values.Set("redirect_uri", redirectUri)
	uri := fmt.Sprintf("%s?%s", tokenUrl, values.Encode())
	v, err := httpjson(uri)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var t token
	t.AccessToken = v["access_token"].(string)
	t.RefreshToken = v["refresh_token"].(string)
	t.Expiry = time.Now().Add(time.Duration(int(v["expires_in"].(float64))) * time.Second)
	err = t.Save(tokenFile)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "token ok")
}

func init() {
	http.HandleFunc("/login", login)
	http.HandleFunc("/authresponse", authResponse)
}
