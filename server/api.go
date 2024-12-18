// replication-manager - Replication Manager Monitoring and CLI for MariaDB and MySQL
// Copyright 2017-2021 SIGNAL18 CLOUD SAS
// Author: Stephane Varoqui  <svaroqui@gmail.com>
// License: GNU General Public License, version 3. Redistribution/Reuse of this code is permitted under the GNU v3 license, as an additional term ALL code must carry the original Author(s) credit in comment form.
// See LICENSE in this directory for the integral text.

package server

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/iancoleman/strcase"
	"github.com/klauspost/compress/zstd"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/codegangsta/negroni"
	jwt "github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt/request"
	"github.com/gorilla/mux"
	"github.com/signal18/replication-manager/cert"
	"github.com/signal18/replication-manager/cluster"
	"github.com/signal18/replication-manager/config"
	"github.com/signal18/replication-manager/regtest"
	"github.com/signal18/replication-manager/share"
	"github.com/signal18/replication-manager/utils/githelper"
)

//RSA KEYS AND INITIALISATION

var signingKey, verificationKey []byte
var apiPass string
var apiUser string

func (repman *ReplicationManager) initKeys() {
	repman.Lock()
	defer repman.Unlock()
	var (
		err         error
		privKey     *rsa.PrivateKey
		pubKey      *rsa.PublicKey
		pubKeyBytes []byte
	)

	privKey, err = rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		log.Fatal("Error generating private key")
	}
	pubKey = &privKey.PublicKey //hmm, this is stdlib manner...

	// Create signingKey from privKey
	// prepare PEM block
	var privPEMBlock = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey), // serialize private key bytes
	}
	// serialize pem
	privKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(privKeyPEMBuffer, privPEMBlock)
	//done
	signingKey = privKeyPEMBuffer.Bytes()

	//fmt.Println(string(signingKey))

	// create verificationKey from pubKey. Also in PEM-format
	pubKeyBytes, err = x509.MarshalPKIXPublicKey(pubKey) //serialize key bytes
	if err != nil {
		// heh, fatality
		log.Fatal("Error marshalling public key")
	}

	var pubPEMBlock = &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubKeyBytes,
	}
	// serialize pem
	pubKeyPEMBuffer := new(bytes.Buffer)
	pem.Encode(pubKeyPEMBuffer, pubPEMBlock)
	// done
	verificationKey = pubKeyPEMBuffer.Bytes()

	//	fmt.Println(string(verificationKey))
}

//STRUCT DEFINITIONS

type userCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ApiResponse struct {
	Data    string `json:"data"`
	Success bool   `json:"success"`
}

type token struct {
	Token string `json:"token"`
}

// Proxy function that forwards the request to the target URL
func (repman *ReplicationManager) proxyToURL(target string) http.Handler {
	url, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Modify the request as needed before forwarding
		r.Host = url.Host
		proxy.ServeHTTP(w, r)
	})
}

func (repman *ReplicationManager) SharedirHandler(folder string) http.Handler {
	sub, err := fs.Sub(share.EmbededDbModuleFS, folder)
	if err != nil {
		log.Printf("folder read error [%s]: %s", folder, err)
	}

	return http.FileServer(http.FS(sub))
}

func (repman *ReplicationManager) DashboardFSHandler() http.Handler {

	sub, err := fs.Sub(share.EmbededDbModuleFS, "dashboard")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}

func (repman *ReplicationManager) DashboardFSHandlerApp() http.Handler {
	sub, err := fs.Sub(share.EmbededDbModuleFS, "dashboard/index.html")
	if !repman.Conf.HttpUseReact {
		sub, err = fs.Sub(share.EmbededDbModuleFS, "dashboard/app.html")
	}
	if err != nil {
		panic(err)
	}

	return http.FileServer(http.FS(sub))
}

func (repman *ReplicationManager) rootHandler(w http.ResponseWriter, r *http.Request) {
	html, err := share.EmbededDbModuleFS.ReadFile("dashboard/index.html")
	if !repman.Conf.HttpUseReact {
		html, err = share.EmbededDbModuleFS.ReadFile("dashboard/app.html")
	}
	if err != nil {
		log.Printf("rootHandler read error : %s", err)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(html)
}

func (repman *ReplicationManager) apiserver() {
	var err error
	//PUBLIC ENDPOINTS
	router := mux.NewRouter()

	router.Use(repman.RecoveryMiddleware)
	//router.HandleFunc("/", repman.handlerApp)
	// page to view which does not need authorization
	graphiteHost := repman.Conf.GraphiteCarbonHost
	if repman.Conf.GraphiteEmbedded {
		graphiteHost = "127.0.0.1"
	}

	graphiteURL, err := url.Parse(fmt.Sprintf("http://%s:%d", graphiteHost, repman.Conf.GraphiteCarbonApiPort))
	if err == nil {
		// Set up the reverse proxy target for Graphite API
		graphiteProxy := httputil.NewSingleHostReverseProxy(graphiteURL)
		// Set up a route that forwards the request to the Graphite API
		router.PathPrefix("/graphite/").Handler(http.StripPrefix("/graphite/", graphiteProxy))
	}

	router.PathPrefix("/meet/").Handler(http.StripPrefix("/meet/", repman.proxyToURL("https://meet.signal18.io/api/v4")))

	// Define the dynamic proxy route with Base64-encoded peer URL and arbitrary route
	router.HandleFunc("/peer/{encodedpeer}/{route:.*}", repman.DynamicPeerHandler)

	if repman.Conf.Test {
		router.HandleFunc("/", repman.handlerApp)
		router.PathPrefix("/images/").Handler(http.FileServer(http.Dir(repman.Conf.HttpRoot)))
		router.PathPrefix("/assets/").Handler(http.FileServer(http.Dir(repman.Conf.HttpRoot)))

		router.PathPrefix("/static/").Handler(http.FileServer(http.Dir(repman.Conf.HttpRoot)))
		router.PathPrefix("/app/").Handler(http.FileServer(http.Dir(repman.Conf.HttpRoot)))
		router.PathPrefix("/grafana/").Handler(http.StripPrefix("/grafana/", http.FileServer(http.Dir(repman.Conf.ShareDir+"/grafana"))))
	} else {
		router.HandleFunc("/", repman.rootHandler)
		router.PathPrefix("/static/").Handler(repman.handlerStatic(repman.DashboardFSHandler()))
		router.PathPrefix("/app/").Handler(repman.DashboardFSHandler())
		router.PathPrefix("/images/").Handler(repman.handlerStatic(repman.DashboardFSHandler()))
		router.PathPrefix("/assets/").Handler(repman.DashboardFSHandler())
		router.PathPrefix("/grafana/").Handler(http.StripPrefix("/grafana/", repman.SharedirHandler("grafana")))
	}

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the path starts with "/api"
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			// Return 404 for /api paths
			http.NotFound(w, r)
		} else {
			// Redirect non /api paths to "/"
			http.Redirect(w, r, "/", http.StatusFound)
		}
	})

	router.HandleFunc("/api/login", repman.loginHandler)
	//router.Handle("/api", v3.NewHandler("My API", "/swagger.json", "/api"))

	router.Handle("/api/auth/callback", negroni.New(
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxAuthCallback)),
	))

	router.Handle("/api/clusters", negroni.New(
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxClusters)),
	))
	router.Handle("/api/clusters/peers", negroni.New(
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxPeerClusters)),
	))
	router.Handle("/api/clusters/{clusterName}/peer-register", negroni.New(
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxPeerRegister)),
	))
	router.Handle("/api/prometheus", negroni.New(
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxPrometheus)),
	))
	router.Handle("/api/status", negroni.New(
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxStatus)),
	))
	router.Handle("/api/timeout", negroni.New(
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxTimeout)),
	))
	router.Handle("/api/repocomp/current", negroni.New(
		negroni.Wrap(http.HandlerFunc(repman.handlerRepoComp)),
	))
	router.Handle("/api/configs/grafana", negroni.New(
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxGrafana)),
	))
	//UNPROTECTED ENDPOINTS FOR SETTINGS
	router.Handle("/api/monitor", negroni.New(
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxReplicationManager)),
	))
	router.Handle("/api/terms", negroni.New(
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxTerms)),
	))
	//PROTECTED ENDPOINTS FOR SETTINGS
	router.Handle("/api/monitor", negroni.New(
		negroni.HandlerFunc(repman.validateTokenMiddleware),
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxReplicationManager)),
	))

	router.Handle("/api/monitor/actions/adduser/{userName}", negroni.New(
		negroni.HandlerFunc(repman.validateTokenMiddleware),
		negroni.Wrap(http.HandlerFunc(repman.handlerMuxAddUser)),
	))

	repman.apiDatabaseUnprotectedHandler(router)
	repman.apiDatabaseProtectedHandler(router)
	repman.apiClusterUnprotectedHandler(router)
	repman.apiClusterProtectedHandler(router)
	repman.apiProxyProtectedHandler(router)

	tlsConfig := Repmanv3TLS{
		Enabled: false,
	}
	// Add default unsecure cert if not set
	if repman.Conf.MonitoringSSLCert == "" {
		host := repman.Conf.APIBind
		if host == "0.0.0.0" {
			host = "localhost," + host + ",127.0.0.1"
		}
		cert.Host = host
		cert.Organization = "Signal18 Replication-Manager"
		tmpKey, tmpCert, err := cert.GenerateTempKeyAndCert()
		if err != nil {
			log.Errorf("Cannot generate temporary Certificate and/or Key: %s", err)
		}
		log.Info("No TLS certificate provided using generated key (", tmpKey, ") and certificate (", tmpCert, ")")
		defer os.Remove(tmpKey)
		defer os.Remove(tmpCert)

		tlsConfig = Repmanv3TLS{
			Enabled:            true,
			CertificatePath:    tmpCert,
			CertificateKeyPath: tmpKey,
			SelfSigned:         true,
		}
	}

	if repman.Conf.MonitoringSSLCert != "" {
		log.Info("Starting HTTPS & JWT API on " + repman.Conf.APIBind + ":" + repman.Conf.APIPort)
		tlsConfig = Repmanv3TLS{
			Enabled:            true,
			CertificatePath:    repman.Conf.MonitoringSSLCert,
			CertificateKeyPath: repman.Conf.MonitoringSSLKey,
		}
	} else {
		log.Info("Starting HTTP & JWT API on " + repman.Conf.APIBind + ":" + repman.Conf.APIPort)
	}

	repman.SetV3Config(Repmanv3Config{
		Listen: Repmanv3ListenAddress{
			Address: repman.Conf.APIBind,
			Port:    repman.Conf.APIPort,
		},
		TLS: tlsConfig,
	})

	// pass the router to the V3 server that will multiplex the legacy API and the
	// new gRPC + JSON Gateway API.
	err = repman.StartServerV3(true, router)

	if err != nil {
		log.Errorf("JWT API can't start: %s", err)
	}
	repman.IsApiListenerReady = true
}

//////////////////////////////////////////
/////////////ENDPOINT HANDLERS////////////
/////////////////////////////////////////

func (repman *ReplicationManager) handleOriginValidator(origin string) bool {
	for _, cl := range repman.PeerClusters {
		if cl.ApiPublicUrl == origin {
			return true
		}
	}
	return false
}

func (repman *ReplicationManager) isValidRequest(r *http.Request) (bool, error) {

	_, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		vk, _ := jwt.ParseRSAPublicKeyFromPEM(verificationKey)
		return vk, nil
	})
	if err == nil {
		return true, nil
	}
	return false, err
}

func (repman *ReplicationManager) IsValidClusterACL(r *http.Request, cluster *cluster.Cluster) (bool, string) {

	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		vk, _ := jwt.ParseRSAPublicKeyFromPEM(verificationKey)
		return vk, nil
	})
	if err == nil {
		claims := token.Claims.(jwt.MapClaims)
		userinfo := claims["CustomUserInfo"]
		mycutinfo := userinfo.(map[string]interface{})
		meuser := mycutinfo["Name"].(string)
		mepwd := mycutinfo["Password"].(string)
		_, ok := mycutinfo["profile"]

		if ok {
			if strings.Contains(mycutinfo["profile"].(string), repman.Conf.OAuthProvider) /*&& strings.Contains(mycutinfo["email_verified"]*/ {
				meuser = mycutinfo["email"].(string)
				return cluster.IsValidACL(meuser, mepwd, r.URL.Path, "oidc"), meuser
			}
		}
		return cluster.IsValidACL(meuser, mepwd, r.URL.Path, "password"), meuser
	}
	return false, ""
}

func (repman *ReplicationManager) DecryptJWTPassword(r *http.Request) (string, error) {

	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		vk, _ := jwt.ParseRSAPublicKeyFromPEM(verificationKey)
		return vk, nil
	})
	if err == nil {
		claims := token.Claims.(jwt.MapClaims)
		userinfo := claims["CustomUserInfo"]
		mycutinfo := userinfo.(map[string]interface{})
		mepwd := mycutinfo["Password"].(string)
		_, ok := mycutinfo["profile"]

		if ok && strings.Contains(mycutinfo["profile"].(string), repman.Conf.OAuthProvider) /*&& strings.Contains(mycutinfo["email_verified"]*/ {
			return repman.Conf.GetDecryptedPassword("api-credentials", mepwd), nil
		}
		return "", fmt.Errorf("No Gitlab Profile in JWT")
	}
	return "", err
}

func (repman *ReplicationManager) GetJWTClaims(r *http.Request) (map[string]string, error) {
	UserInfoMap := make(map[string]string)
	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		vk, _ := jwt.ParseRSAPublicKeyFromPEM(verificationKey)
		return vk, nil
	})
	if err == nil {
		claims := token.Claims.(jwt.MapClaims)
		userinfo := claims["CustomUserInfo"]
		mycutinfo := userinfo.(map[string]interface{})
		UserInfoMap["Password"] = mycutinfo["Password"].(string)
		UserInfoMap["Role"] = mycutinfo["Role"].(string)
		_, ok := mycutinfo["profile"]
		if ok {
			profile := mycutinfo["profile"].(string)
			if strings.Contains(profile, repman.Conf.OAuthProvider) {
				UserInfoMap["User"] = mycutinfo["email"].(string)
				UserInfoMap["profile"] = profile
				return UserInfoMap, nil
			}
			return nil, fmt.Errorf("invalid oauth provider")
		}
		UserInfoMap["User"] = mycutinfo["Name"].(string)
		return UserInfoMap, nil
	}
	return nil, err
}

func (repman *ReplicationManager) GetUserFromRequest(r *http.Request) string {

	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		vk, _ := jwt.ParseRSAPublicKeyFromPEM(verificationKey)
		return vk, nil
	})

	if err == nil {
		claims := token.Claims.(jwt.MapClaims)
		userinfo := claims["CustomUserInfo"]
		mycutinfo := userinfo.(map[string]interface{})
		meuser := mycutinfo["Name"].(string)
		_, ok := mycutinfo["profile"]

		if ok {
			if strings.Contains(mycutinfo["profile"].(string), repman.Conf.OAuthProvider) /*&& strings.Contains(mycutinfo["email_verified"]*/ {
				return mycutinfo["email"].(string)
			}
		}
		return meuser
	}

	return ""
}

func (repman *ReplicationManager) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var user userCredentials

	//decode request into UserCredentials struct
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Error in request")
		return
	}

	if v, ok := repman.UserAuthTry.Load(user.Username); ok {
		auth_try := v.(authTry)
		if auth_try.Try == 3 {
			if time.Now().Before(auth_try.Time.Add(3 * time.Minute)) {
				fmt.Println("Time until last auth try : " + time.Until(auth_try.Time).String())
				fmt.Println("3 authentication errors for the user " + user.Username + ", please try again in 3 minutes")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			} else {
				auth_try.Try = 1
				auth_try.Time = time.Now()
				repman.UserAuthTry.Store(user.Username, auth_try)
			}
		} else {

			auth_try.Try += 1
			repman.UserAuthTry.Store(user.Username, auth_try)
		}
	} else {
		var auth_try authTry = authTry{
			User: user.Username,
			Try:  1,
			Time: time.Now(),
		}
		repman.UserAuthTry.Store(user.Username, auth_try)
	}

	var tok string
	var userInfo interface{}

	if repman.Conf.Cloud18 && strings.Contains(user.Username, "@") {
		tok, _ = githelper.GetGitLabTokenBasicAuth(user.Username, user.Password, false)
		if tok == "" {
			http.Error(w, "Error logging in to gitlab: Token is empty", http.StatusUnauthorized)
			return
		}

		userInfo = struct {
			Name     string
			Role     string
			Password string
			Email    string `json:"email"`
			Profile  string `json:"profile"`
		}{user.Username, "Member", repman.Conf.GetEncryptedString(user.Password), user.Username, repman.Conf.OAuthProvider}

	} else {
		loggedIn := false
		for _, cl := range repman.Clusters {
			//validate user credentials
			if !cl.IsValidACL(user.Username, user.Password, r.URL.Path, "password") {
				continue
			}
			loggedIn = true
			userInfo = struct {
				Name     string
				Role     string
				Password string
			}{user.Username, "Member", repman.Conf.GetEncryptedString(user.Password)}
		}

		if !loggedIn {
			http.Error(w, "Error logging in: Invalid credentials", http.StatusUnauthorized)
			return
		}
	}

	var auth_try authTry = authTry{
		User: user.Username,
		Try:  1,
		Time: time.Now(),
	}

	repman.UserAuthTry.Store(user.Username, auth_try)

	signer := jwt.New(jwt.SigningMethodRS256)
	claims := signer.Claims.(jwt.MapClaims)
	//set claims
	claims["iss"] = "https://api.replication-manager.signal18.io"
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(repman.Conf.TokenTimeout)).Unix()
	claims["jti"] = "1"   // should be user ID(?)
	claims["token"] = tok // store gitlab token
	claims["CustomUserInfo"] = userInfo
	signer.Claims = claims
	sk, _ := jwt.ParseRSAPrivateKeyFromPEM(signingKey)

	tokenString, err := signer.SignedString(sk)
	if err != nil {
		fmt.Fprintln(w, "Error while signing the token")
		http.Error(w, "Error signing token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//create a token instance using the token string
	specs := r.Header.Get("Accept")
	resp := token{tokenString}
	if strings.Contains(specs, "text/html") {
		w.Write([]byte(tokenString))
		return
	}

	repman.jsonResponse(resp, w)
	return
}

func (repman *ReplicationManager) handlerMuxAuthCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	OAuthContext := context.Background()
	Provider, err := oidc.NewProvider(OAuthContext, repman.Conf.OAuthProvider)
	if err != nil {
		log.Printf("OAuth callback Failed to init oidc from gitlab:%s %v\n", repman.Conf.OAuthProvider, err)
		return
	}
	OAuthConfig := oauth2.Config{
		ClientID:     repman.Conf.OAuthClientID,
		ClientSecret: repman.Conf.GetDecryptedPassword("api-oauth-client-secret", repman.Conf.OAuthClientSecret),
		Endpoint:     Provider.Endpoint(),
		RedirectURL:  repman.Conf.APIPublicURL + "/api/auth/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "read_api", "api"},
	}
	log.Printf("OAuth oidc to gitlab: %v\n", OAuthConfig)
	oauth2Token, err := OAuthConfig.Exchange(OAuthContext, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	repman.OAuthAccessToken = oauth2Token

	userInfo, err := Provider.UserInfo(OAuthContext, oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	r.Header.Get("Accept")

	for _, cluster := range repman.Clusters {
		//validate user credentials
		if cluster.IsValidACL(userInfo.Email, cluster.APIUsers[userInfo.Email].Password, r.URL.Path, "oidc") {
			apiuser := cluster.APIUsers[userInfo.Email]
			apiuser.GitToken = oauth2Token.AccessToken
			tmp := strings.Split(userInfo.Profile, "/")
			apiuser.GitUser = tmp[len(tmp)-1]
			cluster.APIUsers[userInfo.Email] = apiuser

			if cluster.Conf.Cloud18 {
				tokenName := conf.Cloud18Domain + "-" + conf.Cloud18SubDomain + "-" + conf.Cloud18SubDomainZone
				new_token, user_id := githelper.GetGitLabTokenOAuth(oauth2Token.AccessToken, tokenName, cluster.Conf.LogGit)
				//vault_aut_url := vaulthelper.GetVaultOIDCAuth()
				//vaulthelper.GetVaultOIDCAuth()
				//http.Redirect(w, r, vault_aut_url, http.StatusSeeOther)
				if new_token != "" {
					//to create project for user if not exist
					path := cluster.Conf.Cloud18Domain + "/" + cluster.Conf.Cloud18SubDomain + "-" + cluster.Conf.Cloud18SubDomainZone
					name := cluster.Conf.Cloud18SubDomain + "-" + cluster.Conf.Cloud18SubDomainZone
					githelper.GitLabCreateProject(new_token, name, path, cluster.Conf.Cloud18Domain, user_id, cluster.Conf.LogGit)
					//to store new gitlab token
					cluster.Conf.GitUrl = repman.Conf.OAuthProvider + "/" + cluster.Conf.Cloud18Domain + "/" + cluster.Conf.Cloud18SubDomain + "-" + cluster.Conf.Cloud18SubDomainZone + ".git"
					cluster.Conf.GitUsername = tmp[len(tmp)-1]
					newSecret := cluster.Conf.Secrets["git-acces-token"]
					newSecret.OldValue = newSecret.Value
					newSecret.Value = new_token
					cluster.Conf.Secrets["git-acces-token"] = newSecret
					//cluster.Conf.GitAccesToken = tokenInfo.token
					cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModGit, config.LvlDbg, "Clone from git : url %s, tok %s, dir %s\n", cluster.Conf.GitUrl, cluster.Conf.PrintSecret(cluster.Conf.Secrets["git-acces-token"].Value), cluster.Conf.WorkingDir)
					cluster.Conf.CloneConfigFromGit(cluster.Conf.GitUrl, cluster.Conf.GitUsername, cluster.Conf.Secrets["git-acces-token"].Value, cluster.Conf.WorkingDir)
				} else {
					log.Printf("Failed to get token from gitlab: %v\n", err)
				}

			}

			signer := jwt.New(jwt.SigningMethodRS256)
			claims := signer.Claims.(jwt.MapClaims)
			//set claims
			claims["iss"] = "https://api.replication-manager.signal18.io"
			claims["iat"] = time.Now().Unix()
			claims["exp"] = time.Now().Add(time.Hour * time.Duration(repman.Conf.TokenTimeout)).Unix()
			claims["jti"] = "1" // should be user ID(?)
			claims["CustomUserInfo"] = struct {
				Name     string
				Role     string
				Password string
			}{userInfo.Email, "Member", repman.Conf.GetEncryptedString(cluster.APIUsers[userInfo.Email].Password)}
			password := cluster.APIUsers[userInfo.Email].Password
			signer.Claims = claims
			sk, _ := jwt.ParseRSAPrivateKeyFromPEM(signingKey)

			tokenString, err := signer.SignedString(sk)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Error while signing the token")
				log.Printf("Error signing token: %v\n", err)
			}
			//create a token instance using the token string
			specs := r.Header.Get("Accept")
			resp := token{tokenString}
			if strings.Contains(specs, "text/html") {
				http.Redirect(w, r, repman.Conf.APIPublicURL+"/#!/dashboard?token="+tokenString+"&user="+userInfo.Email+"&pass="+password, http.StatusTemporaryRedirect)
				return
			}
			repman.jsonResponse(resp, w)
			return
		}

	}

	w.WriteHeader(http.StatusForbidden)
	fmt.Println("Error logging in")
	fmt.Fprint(w, "Invalid credentials")
	return
}

//AUTH TOKEN VALIDATION

func (repman *ReplicationManager) handlerMuxReplicationManager(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	mycopy := repman
	var cl []string

	for _, cluster := range repman.Clusters {

		if valid, _ := repman.IsValidClusterACL(r, cluster); valid {
			cl = append(cl, cluster.Name)
		}
	}

	mycopy.ClusterList = cl

	res, err := json.Marshal(mycopy)
	if err != nil {
		http.Error(w, "Error Marshal", 500)
		return
	}

	for crkey, _ := range mycopy.Conf.Secrets {
		res, err = jsonparser.Set(res, []byte(`"*:*" `), "config", strcase.ToLowerCamel(crkey))
	}

	if err != nil {
		http.Error(w, "Encoding error", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func (repman *ReplicationManager) handlerMuxTerms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(repman.Terms)
}

func (repman *ReplicationManager) handlerMuxAddUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var userform cluster.UserForm
	//decode request into UserCredentials struct
	err := json.NewDecoder(r.Body).Decode(&userform)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error in request")
		return
	}

	for _, cluster := range repman.Clusters {
		if valid, _ := repman.IsValidClusterACL(r, cluster); valid {
			cluster.AddUser(userform)
		}
	}

}

func (repman *ReplicationManager) handlerMuxAddClusterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vars := mux.Vars(r)

	var userform cluster.UserForm
	//decode request into UserCredentials struct
	err := json.NewDecoder(r.Body).Decode(&userform)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error in request")
		return
	}

	mycluster := repman.getClusterByName(vars["clusterName"])
	if mycluster != nil {
		err := mycluster.AddUser(userform)
		if err != nil {
			http.Error(w, "Error adding new user: "+err.Error(), 500)
			return
		}
	} else {
		http.Error(w, "No valid cluster", 500)
		return
	}
}

func (repman *ReplicationManager) handlerMuxUpdateClusterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vars := mux.Vars(r)

	var userform cluster.UserForm
	//decode request into UserCredentials struct
	err := json.NewDecoder(r.Body).Decode(&userform)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error in request")
		return
	}

	mycluster := repman.getClusterByName(vars["clusterName"])
	if mycluster != nil {
		err := mycluster.UpdateUser(userform)
		if err != nil {
			http.Error(w, "Error updating user: "+err.Error(), 500)
			return
		}
	} else {
		http.Error(w, "No valid cluster", 500)
		return
	}
}

// swagger:route GET /api/clusters clusters
//
// This will show all the available clusters
//
//	Responses:
//	  200: clusters
func (repman *ReplicationManager) handlerMuxClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if ok, err := repman.isValidRequest(r); ok {

		var clusters []*cluster.Cluster

		for _, cluster := range repman.Clusters {
			if valid, _ := repman.IsValidClusterACL(r, cluster); valid {
				clusters = append(clusters, cluster)
			}
		}

		sort.Sort(cluster.ClusterSorter(clusters))

		cl, err := json.MarshalIndent(clusters, "", "\t")
		if err != nil {
			http.Error(w, "Error Marshal", 500)
			return
		}

		for i, cluster := range clusters {
			for crkey, _ := range cluster.Conf.Secrets {
				cl, err = jsonparser.Set(cl, []byte(`"*:*" `), fmt.Sprintf("[%d]", i), "config", strcase.ToLowerCamel(crkey))
				if err != nil {
					http.Error(w, "Encoding error", 500)
					return
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(cl)

	} else {
		http.Error(w, "Unauthenticated resource: "+err.Error(), 401)
		return
	}
}

func (repman *ReplicationManager) handlerMuxPeerClusters(w http.ResponseWriter, r *http.Request) {
	ok, err := repman.isValidRequest(r)
	if !ok {
		http.Error(w, "Unauthenticated resource: "+err.Error(), 401)
		return
	}

	cl, err := json.MarshalIndent(repman.PeerClusters, "", "\t")
	if err != nil {
		http.Error(w, "Error Marshal", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(cl)
}

func (repman *ReplicationManager) handlerMuxPeerRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vars := mux.Vars(r)

	var userform cluster.UserForm
	//decode request into UserCredentials struct
	err := json.NewDecoder(r.Body).Decode(&userform)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error in request")
		return
	}

	mycluster := repman.getClusterByName(vars["clusterName"])
	if mycluster == nil {
		http.Error(w, "No valid cluster", 500)
		return
	}

	uinfomap, err := repman.GetJWTClaims(r)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Error parsing JWT: "+err.Error())
		return
	}

	if _, ok := uinfomap["profile"]; !ok {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Current user is not logged in via Gitlab!")
		return
	}

	tok, _ := githelper.GetGitLabTokenBasicAuth(uinfomap["User"], repman.Conf.GetDecryptedPassword("peer-login", uinfomap["Password"]), false)
	if tok == "" {
		http.Error(w, "Error logging in to gitlab: Token credentials is not valid", http.StatusUnauthorized)
		return
	}

	if repman.Conf.Cloud18GitUser == "" || repman.Conf.Cloud18GitPassword == "" || !repman.Conf.Cloud18 {
		http.Error(w, "Peer does not have cloud18 setup!", 500)
		return
	}

	_, ok := mycluster.APIUsers[userform.Username]
	if ok {
		http.Error(w, "User already registered on peer cluster!", http.StatusConflict)
		return
	}

	userform.Roles = "pending dbops sponsor"
	mycluster.AddUser(userform)

	err = repman.SendCloud18ClusterSubscriptionMail(mycluster.Name, userform)
	if err != nil {
		http.Error(w, "Error sending email :"+err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent to admin!"))
}

func (repman *ReplicationManager) SendCloud18ClusterSubscriptionMail(clustername string, userform cluster.UserForm) error {
	err := repman.SendOwnerCloud18SubscriptionMail(clustername, userform)
	if err != nil {
		return fmt.Errorf("Cluster admin : %v", err)
	}

	err = repman.SendSponsorCloud18SubscriptionMail(clustername, userform)
	if err != nil {
		return fmt.Errorf("Cluster sponsor : %v", err)
	}
	return nil
}

func (repman *ReplicationManager) SendOwnerCloud18SubscriptionMail(clustername string, userform cluster.UserForm) error {
	to := repman.Conf.Cloud18GitUser
	subj := fmt.Sprintf("Subscription Request for Cluster %s: %s", clustername, userform.Username)
	msg := fmt.Sprintf(`Dear Admin,

A new user has requested to register for the cluster service.

Details:
- User Email: %s
- Cluster: %s
- Monitoring Node: %s
- Registration Request Time: %s

Please review the registration request and take the necessary actions.

Best regards,
Replication Manager
`, userform.Username, clustername, repman.Conf.APIPublicURL, time.Now().Format("2006-01-02 15:04:05"))

	return repman.Mailer.SendEmailMessage(msg, subj, repman.Conf.MailFrom, to, "", repman.Conf.MailSMTPTLSSkipVerify)
}

func (repman *ReplicationManager) SendSponsorCloud18SubscriptionMail(clustername string, userform cluster.UserForm) error {
	to := userform.Username

	subj := fmt.Sprintf("Subscription Request for Cluster %s: %s", clustername, userform.Username)
	msg := fmt.Sprintf(`Dear Sponsor,

Thank you for submitting your request. We have successfully received it and are currently preparing to process it.

To proceed further, we kindly request you to make the payment to the bank account details provided below. Once the payment has been completed, please allow us time to verify it, and we will follow up with the next steps via email.

Registration Details:
- User Email: %s
- Cluster: %s
- Registration Request Time: %s

Bank Account Details:
Account Name: %s
Bank Name: %s
Account Number: %s
IFSC/Swift Code: %s
Reference: %s

Kindly ensure the payment reference matches the request/invoice ID to help us track your payment efficiently.

If you have any questions or need assistance, feel free to reply to this email.

We appreciate your cooperation and look forward to assisting you further.

Best regards,

Signal18
`, userform.Username, clustername, time.Now().Format("2006-01-02 15:04:05"), "", "", "", "", "")

	return repman.Mailer.SendEmailMessage(msg, subj, repman.Conf.MailFrom, to, "", repman.Conf.MailSMTPTLSSkipVerify)
}

func (repman *ReplicationManager) validateTokenMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//validate token
	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			vk, _ := jwt.ParseRSAPublicKeyFromPEM(verificationKey)
			return vk, nil
		})

	if err == nil {
		if token.Valid {
			next(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Token is not valid")
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Unauthorised access to this resource: "+err.Error())
	}
}

//HELPER FUNCTIONS

func (repman *ReplicationManager) jsonResponse(apiresponse interface{}, w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json, err := json.Marshal(apiresponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func (repman *ReplicationManager) handlerMuxClusterAdd(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vars := mux.Vars(r)

	username := repman.GetUserFromRequest(r)
	if username == "" {
		http.Error(w, "User is not valid", http.StatusInternalServerError)
		return
	}

	var cForm cluster.ClusterForm

	//decode request into UserCredentials struct
	err := json.NewDecoder(r.Body).Decode(&cForm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error in request")
		return
	}

	if cname, ok := vars["clusterName"]; !ok || cname == "" {
		vars["clusterName"] = cForm.ClusterName
	}

	cl := repman.getClusterByName(vars["clusterName"])
	if cl != nil {
		http.Error(w, "Cluster already exists", http.StatusBadRequest)
		return
	}

	repman.AddCluster(vars["clusterName"], "")
	// Create user and grant for new cluster
	cl = repman.getClusterByName(vars["clusterName"])
	if cl != nil {
		if u, ok := cl.APIUsers[username]; !ok {
			// Create user and grant for new cluster
			userform := cluster.UserForm{
				Username: username,
				Roles:    strings.Join(([]string{config.RoleSponsor, config.RoleDBOps, config.RoleSysOps}), " "),
				Grants:   "cluster db proxy prov",
			}

			cl.AddUser(userform)
		} else {
			// Update grant for new cluster
			cl.SetNewUserGrants(&u, "cluster db proxy prov")
			cl.SetNewUserRoles(&u, strings.Join(([]string{config.RoleSponsor, config.RoleDBOps, config.RoleSysOps}), " "))
			cl.APIUsers[u.User] = u
		}

		// Adjust cluster based on selected orchestrator
		if cForm.Orchestrator != "" && cForm.Orchestrator != cl.Conf.ProvOrchestrator {
			cl.SetProvOrchestrator(cForm.Orchestrator)
		}

		// Cluster will auto set service when plan is not empty
		if cForm.Plan != "" {
			cl.SetServicePlan(cForm.Plan)
		}
	}
}

func (repman *ReplicationManager) handlerMuxClusterDelete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	vars := mux.Vars(r)
	repman.DeleteCluster(vars["clusterName"])

}

// swagger:operation GET /api/prometheus prometheus
// Returns the Prometheus metrics for all database instances on the server
// in the Prometheus text format
//
// ---
// produces:
//   - text/plain; version=0.0.4
//
// responses:
//
//	'200':
//	  description: Prometheus file format
//	  schema:
//	    type: string
//	  headers:
//	    Access-Control-Allow-Origin:
//	      type: string
func (repman *ReplicationManager) handlerMuxPrometheus(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	for _, cluster := range repman.Clusters {
		for _, server := range cluster.Servers {
			res := server.GetPrometheusMetrics()
			w.Write([]byte(res))
		}
	}
}

func (repman *ReplicationManager) handlerMuxClustersOld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	s := new(Settings)
	s.Clusters = repman.ClusterList
	regtest := new(regtest.RegTest)
	s.RegTests = regtest.GetTests()
	e := json.NewEncoder(w)
	e.SetIndent("", "\t")
	err := e.Encode(s)
	if err != nil {
		http.Error(w, "Encoding error", 500)
		return
	}
}

// The Status contains string value for the alive status.
// Possible values are: running, starting, errors
//
// swagger:response status
type StatusResponse struct {
	// Example: *
	AccessControlAllowOrigin string `json:"Access-Control-Allow-Origin"`
	// The status message
	// in: body
	Body struct {
		// Example: running
		// Example: starting
		// Example: errors
		Alive string `json:"alive"`
	}
}

// swagger:route GET /api/status status
//
// This will show the status of the cluster
//
//     Responses:
//       200: status

func (repman *ReplicationManager) handlerMuxStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if repman.isStarted {
		io.WriteString(w, `{"alive": "running"}`)
	} else {
		io.WriteString(w, `{"alive": "starting"}`)
	}
}

// swagger:route GET /api/timeout timeout
//
//     Responses:
//       200: status

func (repman *ReplicationManager) handlerMuxTimeout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	time.Sleep(1200 * time.Second)
	io.WriteString(w, `{"alive": "running"}`)
}

// swagger:route GET /api/heartbeat heartbeat
//
//     Responses:
//       200: heartbeat

func (repman *ReplicationManager) handlerMuxMonitorHeartbeat(w http.ResponseWriter, r *http.Request) {
	var send Heartbeat
	send.UUID = repman.UUID
	send.UID = repman.Conf.ArbitrationSasUniqueId
	send.Secret = repman.Conf.ArbitrationSasSecret
	send.Status = repman.Status
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(send); err != nil {
		panic(err)
	}
}

func (repman *ReplicationManager) handlerStatic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", repman.Conf.CacheStaticMaxAge))
		w.Header().Set("Etag", repman.Version)

		h.ServeHTTP(w, r)
	})
}

func (repman *ReplicationManager) handlerMuxGrafana(w http.ResponseWriter, r *http.Request) {
	var entries []fs.DirEntry
	var list []string = make([]string, 0)
	var err error
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if repman.Conf.Test {
		entries, err = os.ReadDir(conf.ShareDir + "/grafana")
	} else {
		entries, err = share.EmbededDbModuleFS.ReadDir("grafana")
	}
	if err != nil {
		http.Error(w, "Encoding reading directory", 500)
		return
	}
	for _, b := range entries {
		if !b.IsDir() {
			list = append(list, b.Name())
		}
	}

	err = json.NewEncoder(w).Encode(list)
	if err != nil {
		http.Error(w, "Encoding error", 500)
		return
	}
}

func (repman *ReplicationManager) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logrus.WithFields(logrus.Fields{
					"error":      err,
					"stacktrace": string(debug.Stack()),
					"url":        r.URL.String(),
				}).Error("Recovered from panic")

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// isAllowedPeer checks if a given peer URL is in the allowed list
func (repman *ReplicationManager) IsAllowedPeer(peerURL string) bool {
	for _, pcl := range repman.PeerClusters {
		if peerURL == pcl.ApiPublicUrl {
			return true
		}
	}
	return false
}

// isAllowedPeer checks if a given peer URL is in the allowed list
func (repman *ReplicationManager) IsAllowedPeerRoute(route string) bool {
	prefixes := []string{"api"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(route, prefix) {
			return true
		}
	}
	return false
}

// DynamicPeerHandler forwards requests to the specified peer URL
func (repman *ReplicationManager) DynamicPeerHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the Base64-encoded peer URL and the route from the path
	vars := mux.Vars(r)
	encodedPeer := vars["encodedpeer"]
	route := vars["route"]

	// logRequest(r)

	// Decode the Base64-encoded peer URL
	peer, err := base64.StdEncoding.DecodeString(encodedPeer)
	if err != nil {
		http.Error(w, "Invalid peer URL encoding", http.StatusBadRequest)
		log.Printf("Error decoding peer URL: %v", err)
		return
	}

	// Convert the decoded URL to string
	peerURL := string(peer)

	// Validate the peer URL
	if !repman.IsAllowedPeer(peerURL) {
		http.Error(w, "Peer URL not allowed", http.StatusForbidden)
		log.Printf("Blocked forwarding to disallowed peer: %s", peerURL)
		return
	}

	// Parse the peer URL
	parsedPeerURL, err := url.Parse(peerURL)
	if err != nil {
		http.Error(w, "Invalid peer URL", http.StatusBadRequest)
		log.Printf("Error parsing peer URL: %v", err)
		return
	}

	// Attach the specific route from the URL to the peer URL
	parsedPeerURL.Path = parsedPeerURL.Path + "/" + route

	var user userCredentials
	if route == "api/login" {
		//decode request into UserCredentials struct
		err = json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Error in request")
			return
		}

		uinfomap, err := repman.GetJWTClaims(r)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Error parsing JWT: "+err.Error())
			return
		}

		if _, ok := uinfomap["profile"]; !ok {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Current user is not logged in via Gitlab!")
			return
		}

		user.Password = repman.Conf.GetDecryptedPassword("peer-login", uinfomap["Password"])

		// Marshal the modified JSON back to a byte slice
		loginBody, err := json.Marshal(user)
		if err != nil {
			http.Error(w, "Failed to marshal modified JSON", http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(loginBody))
		r.ContentLength = int64(len(loginBody)) // Update content length
	}

	// Log the forwarding request
	log.Printf("Forwarding request to: %s", parsedPeerURL.String())

	// Create a new request to forward to Peer
	req, err := http.NewRequest(r.Method, parsedPeerURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy Content-Type and other headers from the original request
	req.Header = r.Header.Clone()

	// logForwardedRequest(req)

	// Send the request to GoApp 2
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error forwarding request: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// logResponse(resp)

	var body []byte
	// Check if the response is compressed with zstd
	if resp.Header.Get("Content-Encoding") == "zstd" {
		// Decompress the zstd response
		decoder, err := zstd.NewReader(resp.Body)
		if err != nil {
			http.Error(w, "Failed to create zstd decoder: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer decoder.Close()

		// Read the decompressed data
		body, err = io.ReadAll(decoder)
		if err != nil {
			http.Error(w, "Failed to read decompressed body: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// fmt.Printf("Decompressed Response: %s\n", body)
	} else {
		// Handle uncompressed response
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read response body: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// fmt.Printf("Response: %s\n", body)
	}

	// Forward the response back to the React client
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// logRequest logs details of the incoming HTTP request
func logRequest(r *http.Request) {
	log.Printf("Incoming Request -> Method: %s, URL: %s", r.Method, r.URL.String())
	log.Printf("Incoming Headers: %v", r.Header)
	if r.Body != nil {
		body, _ := io.ReadAll(r.Body)
		log.Printf("Incoming Body: %s", string(body))
		r.Body = io.NopCloser(bytes.NewReader(body)) // Reset the body for further use
	}
}

// logForwardedRequest logs details of the request sent to GoApp 2
func logForwardedRequest(req *http.Request) {
	log.Printf("Forwarding Request -> Method: %s, URL: %s", req.Method, req.URL.String())
	log.Printf("Forwarding Headers: %v", req.Header)
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		log.Printf("Forwarding Body: %s", string(body))
		req.Body = io.NopCloser(bytes.NewReader(body)) // Reset the body for sending
	}
}

// logResponse logs details of the HTTP response received from GoApp 2
func logResponse(resp *http.Response) {
	log.Printf("Response Status: %s", resp.Status)
	log.Printf("Response Headers: %v", resp.Header)
	if resp.Body != nil {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Response Body: %s", string(body))
		resp.Body = io.NopCloser(bytes.NewReader(body)) // Reset the body for further use
	}
}
