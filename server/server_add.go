// replication-manager - Replication Manager Monitoring and CLI for MariaDB and MySQL
// Copyright 2017-2021 SIGNAL18 CLOUD SAS
// Author: Guillaume Lefranc <guillaume@signal18.io>
// License: GNU General Public License, version 3. Redistribution/Reuse of this code is permitted under the GNU v3 license, as an additional term ALL code must carry the original Author(s) credit in comment form.
// See LICENSE in this directory for the integral text.

package server

import (
	"fmt"

	"github.com/signal18/replication-manager/cluster"
	"github.com/signal18/replication-manager/config"
)

func (repman *ReplicationManager) AddCluster(clusterName string, clusterHead string) error {
	var myconf = make(map[string]config.Config)
	myconf[clusterName] = repman.Conf
	repman.Lock()
	repman.ClusterList = append(repman.ClusterList, clusterName)
	repman.Confs[clusterName] = repman.Conf

	repman.VersionConfs[clusterName] = new(config.ConfVersion)
	repman.VersionConfs[clusterName].ConfInit = repman.Conf

	repman.ImmuableFlagMaps[clusterName] = repman.ImmuableFlagMaps["default"]
	repman.DynamicFlagMaps[clusterName] = repman.DynamicFlagMaps["default"]

	repman.Unlock()
	cluster, _ := repman.StartCluster(clusterName)

	for _, cluster := range repman.Clusters {
		cluster.SetClusterList(repman.Clusters)
	}
	//fmt.Printf("ADD CLUSTER def map :\n")
	//fmt.Printf("%s\n", repman.ImmuableFlagMaps)
	//cluster.Conf.PrintConf()

	//cluster.SetClusterHead(clusterHead)
	//cluster.SetClusterHead(clusterName)
	//cluster.SetClusterList(repman.Clusters)
	cluster.Save()
	return nil

}

func (repman *ReplicationManager) CreateAdminUserForm(username string) cluster.UserForm {
	return cluster.UserForm{
		Username: username,
		Roles:    "sysops dbops",
		Grants:   "cluster db proxy prov global grant show sale extrole",
	}
}

func (repman *ReplicationManager) AddCloud18GitUser(cl *cluster.Cluster) error {
	username := repman.Conf.Cloud18GitUser

	// Create user and grant for new cluster
	userform := repman.CreateAdminUserForm(username)

	if _, ok := cl.APIUsers[username]; ok {
		return cl.UpdateUser(userform)
	} else {
		return cl.AddUser(userform)
	}
}

func (repman *ReplicationManager) AddLocalAdminUserACL(cl *cluster.Cluster) error {
	username := "admin"
	// Create user and grant for new cluster
	userform := repman.CreateAdminUserForm(username)

	if _, ok := cl.APIUsers[username]; ok {
		return cl.UpdateUser(userform)
	}

	return fmt.Errorf("User %s not found in cluster %s", username, cl.Name)
}
