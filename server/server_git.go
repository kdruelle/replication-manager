package server

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/signal18/replication-manager/config"
	"github.com/signal18/replication-manager/utils/githelper"
)

func (repman *ReplicationManager) SetIsGitPush(val bool) {
	repman.IsGitPush = val
	for _, cl := range repman.Clusters {
		cl.IsGitPush = val
	}
}

func (repman *ReplicationManager) SetIsGitPull(val bool) {
	repman.IsGitPull = val
	for _, cl := range repman.Clusters {
		cl.IsGitPull = val
	}
}

func (repman *ReplicationManager) InitGitConfig(conf *config.Config) error {
	repman.SetIsGitPush(true)
	defer repman.SetIsGitPush(false)

	if conf.GitUrl != "" && conf.GitAccesToken != "" && !conf.Cloud18 {
		var tok string

		if conf.IsVaultUsed() && conf.IsPath(conf.GitAccesToken) {
			conn, err := conf.GetVaultConnection()
			if err != nil {
				repman.Logrus.Printf("Error vault connection %v", err)
			}
			tok, err = conf.GetVaultCredentials(conn, conf.GitAccesToken, "git-acces-token")
			if err != nil {
				repman.Logrus.Printf("Error get vault git-acces-token value %v", err)
				tok = conf.GetDecryptedValue("git-acces-token")
			} else {
				var Secrets config.Secret
				Secrets.Value = tok
				conf.Secrets["git-acces-token"] = Secrets
			}

		} else {
			tok = conf.GetDecryptedValue("git-acces-token")
		}

		conf.CloneConfigFromGit(conf.GitUrl, conf.GitUsername, tok, conf.WorkingDir)
	}

	if conf.Cloud18GitUser != "" && conf.Cloud18GitPassword != "" && conf.Cloud18 {
		acces_tok, err := githelper.GetGitLabTokenBasicAuth(conf.Cloud18GitUser, conf.GetDecryptedValue("cloud18-gitlab-password"), conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlDbg))
		if err != nil {
			if conf.Verbose || conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlErr) {
				repman.Logrus.Errorf(err.Error() + conf.GetDecryptedValue("cloud18-gitlab-password") + "\n")
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
			path := conf.Cloud18Domain + "/" + conf.Cloud18SubDomain + "-" + conf.Cloud18SubDomainZone
			name := conf.Cloud18SubDomain + "-" + conf.Cloud18SubDomainZone
			githelper.GitLabCreateProject(personal_access_token, name, path, conf.Cloud18Domain, uid, conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlDbg))
			githelper.GitLabCreatePullProject(personal_access_token, name, path, conf.Cloud18Domain, uid, conf.IsEligibleForPrinting(config.ConstLogModGit, config.LvlDbg))

			conf.GitUrl = conf.OAuthProvider + "/" + conf.Cloud18Domain + "/" + conf.Cloud18SubDomain + "-" + conf.Cloud18SubDomainZone + ".git"
			conf.GitUrlPull = conf.OAuthProvider + "/" + conf.Cloud18Domain + "/" + conf.Cloud18SubDomain + "-" + conf.Cloud18SubDomainZone + "-pull.git"
			conf.GitUsername = conf.Cloud18GitUser
			conf.GitAccesToken = personal_access_token
			conf.ImmuableFlagMap["git-url"] = conf.GitUrl
			conf.ImmuableFlagMap["git-url-pull"] = conf.GitUrlPull
			conf.ImmuableFlagMap["git-username"] = conf.GitUsername
			conf.ImmuableFlagMap["git-acces-token"] = personal_access_token

			if conf.ConfRestoreOnStart {
				conf.ConfRestoreOnStart = false
				conf.ImmuableFlagMap["monitoring-restore-config-on-start"] = false
				os.RemoveAll(conf.WorkingDir)
				conf.CloneConfigFromGit(conf.GitUrl, conf.GitUsername, conf.GitAccesToken, conf.WorkingDir)
				conf.CloneConfigFromGit(conf.GitUrlPull, conf.GitUsername, conf.GitAccesToken, conf.WorkingDir+"/.pull")
			}
			//conf.GitAddReadMe(conf.GitUrl, conf.GitAccesToken, conf.GitUsername, conf.WorkingDir)

		} else if conf.LogGit {
			err := fmt.Errorf("Could not get personal access token from gitlab")
			repman.Logrus.WithField("group", repman.ClusterList[cfgGroupIndex]).Infof(err.Error())
			return err
		}

	}

	return nil
}

func (repman *ReplicationManager) PushAllConfigsToGit() {
	// Set Flag as Git Push, prevent new cluster save is processed
	repman.SetIsGitPush(true)
	defer repman.SetIsGitPush(false)

	// Wait if any cluster is saving config
	for _, cl := range repman.Clusters {
		for cl.IsSavingConfig {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if repman.Conf.GitUrl != "" {
		repman.Logrus.Infof("Pushing All Configs To Git")
		repman.Conf.PushConfigToGit(repman.Conf.GitUrl, repman.Conf.Secrets["git-acces-token"].Value, repman.Conf.GitUsername, repman.Conf.WorkingDir, repman.ClusterList)
	}
}

func (repman *ReplicationManager) PullCloud18Configs() {
	// Set Flag as Git Pull, prevent new cluster save / push is processed
	repman.SetIsGitPull(true)
	defer repman.SetIsGitPull(false)

	// Wait if any cluster is saving config
	for _, cl := range repman.Clusters {
		for cl.IsSavingConfig {
			time.Sleep(100 * time.Millisecond)
		}
	}

	pullDir := repman.Conf.WorkingDir + "/.pull"
	filePath := pullDir + "/cloud18.toml"

	if repman.Conf.GitUrlPull != "" {
		repman.Conf.CloneConfigFromGit(repman.Conf.GitUrlPull, repman.Conf.GitUsername, repman.Conf.Secrets["git-acces-token"].Value, pullDir)

		//to check cloud18.toml for the first time
		if repman.cloud18CheckSum == nil && repman.Conf.Cloud18 {
			new_h := md5.New()
			repman.ReadCloud18Config()
			file, err := os.Open(filePath)
			if err != nil {
				if os.IsPermission(err) {
					repman.Logrus.Infof("File permission denied: %s", filePath)
				}
			} else {
				if _, err := io.Copy(new_h, file); err != nil {
					repman.Logrus.Infof("Error during computing cloud18.toml hash: %s", err)
				} else {
					repman.cloud18CheckSum = new_h
				}
			}
			defer file.Close()

		} else if repman.Conf.Cloud18 {
			//to check whether new parameters have been injected into the cloud18.toml config file
			file, err := os.Open(filePath)
			if err != nil {
				if os.IsPermission(err) {
					repman.Logrus.Infof("File permission denied: %s", filePath)
				}
			} else {
				new_h := md5.New()
				if _, err := io.Copy(new_h, file); err != nil {
					repman.Logrus.Infof("Error during computing cloud18.toml hash: %s", err)
				} else if !bytes.Equal(repman.cloud18CheckSum.Sum(nil), new_h.Sum(nil)) {
					repman.ReadCloud18Config()
					repman.cloud18CheckSum = new_h
				}
			}
			defer file.Close()

		}
	}

	if repman.Conf.Cloud18 {
		//then to check new file pulled in working dir
		files, err := os.ReadDir(repman.Conf.WorkingDir)
		if err != nil {
			repman.Logrus.Infof("No working directory %s ", repman.Conf.WorkingDir)
		}
		//check all dir of the datadir to check if a new cluster has been pull by git
		for _, f := range files {
			new_cluster_discover := true
			if f.IsDir() && f.Name() != "graphite" && f.Name() != "backups" && f.Name() != ".git" && f.Name() != "cloud18.toml" && !strings.Contains(f.Name(), ".json") && !strings.Contains(f.Name(), ".csv") {
				for name, _ := range repman.Clusters {
					if name == f.Name() {
						new_cluster_discover = false
					}
				}
			} else {
				new_cluster_discover = false
			}
			//find a dir that is not in the cluster list (and diff from backups and graphite)
			//so add the to new cluster to the repman
			if new_cluster_discover {
				//check if this there is a config file in the dir
				if _, err := os.Stat(repman.Conf.WorkingDir + "/" + f.Name() + "/" + f.Name() + ".toml"); !os.IsNotExist(err) {
					//init config, start the cluster and add it to the cluster list
					repman.ViperConfig.SetConfigName(f.Name())
					repman.ViperConfig.SetConfigFile(repman.Conf.WorkingDir + "/" + f.Name() + "/" + f.Name() + ".toml")
					err := repman.ViperConfig.MergeInConfig()
					if err != nil {
						repman.Logrus.Errorf("Config error in " + repman.Conf.WorkingDir + "/" + f.Name() + "/" + f.Name() + ".toml" + ":" + err.Error())
					}
					repman.Confs[f.Name()] = repman.GetClusterConfig(repman.ViperConfig, repman.Conf.ImmuableFlagMap, repman.Conf.DynamicFlagMap, f.Name(), repman.Conf)
					repman.StartCluster(f.Name())
					repman.Clusters[f.Name()].IsGitPull = true
					for _, cluster := range repman.Clusters {
						cluster.SetClusterList(repman.Clusters)
					}
					repman.ClusterList = append(repman.ClusterList, f.Name())
				}
			}
		}
	}
}

func (repman *ReplicationManager) ReadCloud18Config() {
	filePath := conf.WorkingDir + "/.pull/cloud18.toml"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		repman.Conf.ReadCloud18Config(repman.ViperConfig, filePath)
	}
}
