package server

import (
	"fmt"

	"github.com/signal18/replication-manager/config"
	"github.com/signal18/replication-manager/utils/githelper"
)

func (repman *ReplicationManager) InitGitConfig(conf *config.Config) error {
	acces_tok, err := githelper.GetGitLabTokenBasicAuth(conf.Cloud18GitUser, conf.GetDecryptedValue("cloud18-gitlab-password"), conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlDbg))
	if err != nil {
		if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlErr) {
			repman.Logrus.Errorf(err.Error() + "\n")
		}
		conf.Cloud18 = false
		return err
	}

	uid, err := githelper.GetGitLabUserId(acces_tok, conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlDbg))
	if err != nil {
		if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlErr) {
			repman.Logrus.Errorf(err.Error() + "\n")
		}
		conf.Cloud18 = false
		return err
	} else if uid == 0 {
		if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlErr) {
			repman.Logrus.Errorf("Invalid user Id \n")
		}
		conf.Cloud18 = false
		return fmt.Errorf("Invalid user Id")
	}

	_, err = githelper.InitGroupAccessLevel(acces_tok, conf.Cloud18Domain, uid, conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlDbg))
	if err != nil {
		if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlErr) {
			repman.Logrus.Errorf(err.Error() + "\n")
		}
		conf.Cloud18 = false
		return err
	}

	personal_access_token, _ := githelper.GetGitLabTokenOAuth(acces_tok, conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlDbg))
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

		if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlDbg) {
			repman.Logrus.Debugf("Clone from git : url %s, tok %s, dir %s\n", conf.GitUrl, conf.PrintSecret(conf.GitAccesToken), conf.WorkingDir)
		}
		err := conf.CloneConfigFromGit(conf.GitUrl, conf.GitUsername, conf.GitAccesToken, conf.WorkingDir)
		if err != nil {
			if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlErr) {
				repman.Logrus.Errorf(err.Error() + "\n")
			}
		}

		if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlDbg) {
			repman.Logrus.Debugf("Push to git : tok %s, dir %s, user %s, clustersList : %v\n", conf.PrintSecret(conf.GitAccesToken), conf.WorkingDir, conf.GitUsername, []string{})
		}
		err = conf.PushConfigToGit(conf.GitUrl, conf.GitAccesToken, conf.GitUsername, conf.WorkingDir, []string{})
		if err != nil {
			if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlErr) {
				repman.Logrus.Errorf(err.Error() + "\n")
			}
		}
	} else {
		if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlErr) {
			repman.Logrus.Errorf("Could not get personal access token from gitlab" + "\n")
			conf.Cloud18 = false
			return fmt.Errorf("Could not get personal access token from gitlab")
		}
	}

	return nil
}
