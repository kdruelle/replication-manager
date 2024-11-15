package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/signal18/replication-manager/config"
	"github.com/signal18/replication-manager/utils/githelper"
)

type GroupId struct {
	ID          int `json:"id"`
	AccessLevel int `json:"access_level"`
}

func (repman *ReplicationManager) InitGitConfig(conf *config.Config) {
	acces_tok, err := githelper.GetGitLabTokenBasicAuth(conf.Cloud18GitUser, conf.GetDecryptedValue("cloud18-gitlab-password"), conf.LogGit)
	if err != nil {
		repman.LogModulePrintf(repman.Conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
		conf.Cloud18 = false
		return
	}

	uid, body, err := githelper.GetGitLabUserId(acces_tok)
	repman.LogModulePrintf(repman.Conf.Verbose, config.ConstLogModGit, config.LvlDbg, "Init Git Config - Get User response: %s", string(body))
	if err != nil {
		repman.LogModulePrintf(repman.Conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
		conf.Cloud18 = false
		return
	} else if uid == 0 {
		repman.LogModulePrintf(repman.Conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
		conf.Cloud18 = false
		return
	}

	access_level, err := repman.GetGroupAccessLevel(acces_tok, conf.Cloud18Domain)
	if err != nil {
		repman.LogModulePrintf(repman.Conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
	}
	// If access level is 0
	if access_level == 0 {
		conf.Cloud18 = false
		// Send Request to Domain Owner
		return
	}

	personal_access_token, _ := githelper.GetGitLabTokenOAuth(acces_tok, conf.LogGit)
	if personal_access_token != "" {
		var Secrets config.Secret
		Secrets.Value = personal_access_token
		conf.Secrets["git-acces-token"] = Secrets
		conf.GitUrl = conf.OAuthProvider + "/" + conf.Cloud18Domain + "/" + conf.Cloud18SubDomain + "-" + conf.Cloud18SubDomainZone + ".git"
		conf.GitUsername = conf.Cloud18GitUser
		conf.GitAccesToken = personal_access_token
		conf.ImmuableFlagMap["git-url"] = conf.GitUrl
		conf.ImmuableFlagMap["git-username"] = conf.GitUsername
		conf.ImmuableFlagMap["git-acces-token"] = personal_access_token

		repman.LogModulePrintf(conf.Verbose, config.ConstLogModGit, config.LvlDbg, "Clone from git : url %s, tok %s, dir %s\n", conf.GitUrl, conf.PrintSecret(conf.GitAccesToken), conf.WorkingDir)
		err := conf.CloneConfigFromGit(conf.GitUrl, conf.GitUsername, conf.GitAccesToken, conf.WorkingDir)
		if err != nil {
			repman.LogModulePrintf(repman.Conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
		}

		repman.LogModulePrintf(conf.Verbose, config.ConstLogModGit, config.LvlDbg, "Push to git : tok %s, dir %s, user %s, clustersList : %v\n", conf.PrintSecret(conf.GitAccesToken), conf.WorkingDir, conf.GitUsername, []string{})
		err = conf.PushConfigToGit(conf.GitUrl, conf.GitAccesToken, conf.GitUsername, conf.WorkingDir, []string{})
		if err != nil {
			repman.LogModulePrintf(repman.Conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
		}
	} else {
		repman.LogModulePrintf(conf.Verbose, config.ConstLogModGit, config.LvlInfo, "Could not get personal access token from gitlab")
	}
}

// Get User Access Level
func (repman *ReplicationManager) GetGroupAccessLevel(acces_token, domain string) (int, error) {
	var body = make([]byte, 0)
	var err error

	req, err := http.NewRequest("GET", "https://gitlab.signal18.io/api/v4/groups/"+domain, nil)
	if err != nil {
		return 0, fmt.Errorf("Gitlab User API Error: ", err)
	}
	req.Header.Set("Authorization", "Bearer "+acces_token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Gitlab User API Error: ", err)
	}
	defer resp.Body.Close()

	// If 404 error, create the group
	if resp.StatusCode == http.StatusNotFound {
		_, createErr := repman.CreateCloud18Domain(acces_token, domain)
		if createErr != nil {
			return 0, fmt.Errorf("error creating GitLab group: %w", createErr)
		}
		// Try to get the access level again
		return repman.GetGroupAccessLevel(acces_token, domain)
	}

	body, _ = io.ReadAll(resp.Body)

	repman.LogModulePrintf(repman.Conf.Verbose, config.ConstLogModGit, config.LvlDbg, "Get User Access response: %s", string(body))

	var groupId GroupId
	err = json.Unmarshal(body, &groupId)
	if err != nil {
		return 0, fmt.Errorf("Gitlab User API Unmarshall Error: ", err)
	}

	return groupId.AccessLevel, nil
}

func (repman *ReplicationManager) CreateCloud18Domain(acces_token, domain string) (int, error) {
	data := "name=" + domain + "&path=" + domain
	req, err := http.NewRequest("POST", "https://gitlab.signal18.io/api/v4/groups", strings.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("Gitlab User API Error: ", err)
	}
	req.Header.Set("Authorization", "Bearer "+acces_token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("Gitlab User API Error: ", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	repman.LogModulePrintf(repman.Conf.Verbose, config.ConstLogModGit, config.LvlDbg, "Create Cloud18 Domain response: %s", string(body))

	var groupId GroupId

	err = json.Unmarshal(body, &groupId)
	if err != nil {
		return 0, fmt.Errorf("Gitlab User API Unmarshall Error: ", err)
	}

	return groupId.ID, nil

}
