package server

import (
	"github.com/signal18/replication-manager/config"
	"github.com/signal18/replication-manager/utils/githelper"
)

func (repman *ReplicationManager) InitGitConfig(conf *config.Config) {
	acces_tok, err := githelper.GetGitLabTokenBasicAuth(conf.Cloud18GitUser, conf.GetDecryptedValue("cloud18-gitlab-password"), conf.LogGit)
	if err != nil {
		repman.LogModulePrintf(conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
		conf.Cloud18 = false
		return
	}

	uid, err := githelper.GetGitLabUserId(acces_tok, conf.LogGit)

	if err != nil {
		repman.LogModulePrintf(conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
		conf.Cloud18 = false
		return
	} else if uid == 0 {
		repman.LogModulePrintf(conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
		conf.Cloud18 = false
		return
	}

	_, err = githelper.GetGroupAccessLevel(acces_tok, conf.Cloud18Domain, conf.LogGit)
	if err != nil {
		repman.LogModulePrintf(conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
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
			repman.LogModulePrintf(conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
		}

		repman.LogModulePrintf(conf.Verbose, config.ConstLogModGit, config.LvlDbg, "Push to git : tok %s, dir %s, user %s, clustersList : %v\n", conf.PrintSecret(conf.GitAccesToken), conf.WorkingDir, conf.GitUsername, []string{})
		err = conf.PushConfigToGit(conf.GitUrl, conf.GitAccesToken, conf.GitUsername, conf.WorkingDir, []string{})
		if err != nil {
			repman.LogModulePrintf(conf.Verbose, config.ConstLogModGit, config.LvlErr, err.Error())
		}
	} else {
		repman.LogModulePrintf(conf.Verbose, config.ConstLogModGit, config.LvlInfo, "Could not get personal access token from gitlab")
	}
}
