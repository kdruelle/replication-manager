package server

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/signal18/replication-manager/cluster"
	"github.com/signal18/replication-manager/config"
	"github.com/signal18/replication-manager/utils/githelper"
	"github.com/signal18/replication-manager/utils/state"
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

		repman.GitRepo, err = githelper.CloneConfigFromGit(conf.GitUrl, conf.GitUsername, conf.GitAccesToken, conf.WorkingDir)
		if err != nil {
			if strings.Contains(err.Error(), git.ErrNonFastForwardUpdate.Error()) {
				for _, cl := range repman.Clusters {
					if cl != nil {
						cl.SetState("WARN0132", state.State{ErrType: config.LvlWarn, ErrDesc: fmt.Sprintf(config.ClusterError["WARN0132"], conf.GitUrl, err.Error()), ErrFrom: "GIT"})
					}
				}
			} else {
				if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlErr) {
					repman.Logrus.Errorf(err.Error() + "\n")
				}
			}
		} else {
			for _, cl := range repman.Clusters {
				if cl != nil && cl.GetStateMachine() != nil && cl.GetStateMachine().IsInState("WARN0132") {
					cl.GetStateMachine().DeleteState("WARN0132")
				}

				if _, ok := cl.APIUsers[conf.GitUsername]; !ok {
					cl.AddUser(cluster.UserForm{Username: conf.GitUsername, Roles: "dbops sysops", Grants: "db cluster prov proxy global"})
				}
			}
		}

		if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlDbg) {
			repman.Logrus.Debugf("Push to git : tok %s, dir %s, user %s, clustersList : %v\n", conf.PrintSecret(conf.GitAccesToken), conf.WorkingDir, conf.GitUsername, []string{})
		}

		err = repman.PushConfigToGit(conf.GitUrl, conf.GitAccesToken, conf.GitUsername, conf.WorkingDir)
		if err != nil {
			if strings.Contains(err.Error(), git.ErrNonFastForwardUpdate.Error()) {
				for _, cl := range repman.Clusters {
					if cl != nil {
						cl.SetState("WARN0132", state.State{ErrType: config.LvlWarn, ErrDesc: fmt.Sprintf(config.ClusterError["WARN0132"], conf.GitUrl, err.Error()), ErrFrom: "GIT"})
					}
				}
			} else if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlErr) {
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
