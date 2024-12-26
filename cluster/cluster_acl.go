// replication-manager - Replication Manager Monitoring and CLI for MariaDB and MySQL
// Copyright 2017 Signal 18 Cloud SAS
// Authors: Guillaume Lefranc <guillaume@signal18.io>
//          Stephane Varoqui  <svaroqui@gmail.com>
// This source code is licensed under the GNU General Public License, version 3.

package cluster

import (
	"fmt"
	"slices"
	"strings"

	"github.com/signal18/replication-manager/config"
	v3 "github.com/signal18/replication-manager/repmanv3"
	"github.com/signal18/replication-manager/utils/misc"
	"google.golang.org/grpc/codes"
)

type APIUser struct {
	User       string          `json:"user"`
	Password   string          `json:"-"`
	GitToken   string          `json:"-"`
	GitUser    string          `json:"-"`
	IsExternal bool            `json:"-"`
	Roles      map[string]bool `json:"roles"`
	Grants     map[string]bool `json:"grants"`
}

func (cluster *Cluster) SetNewUserGrants(u *APIUser, grant string) {
	acls := strings.Split(grant, " ")
	for key, value := range cluster.Grants {
		found := false
		for _, acl := range acls {
			if strings.HasPrefix(key, acl) && acl != "" {
				found = true
				break
			}
		}
		u.Grants[value] = found
	}
}

func (cluster *Cluster) SetNewUserRoles(u *APIUser, roles string) {
	list := strings.Split(roles, " ")

	if u.Grants[config.GrantGlobalGrant] && roles == "" {
		u.Roles[config.RoleSysOps] = true
		u.Roles[config.RoleDBOps] = true
		return
	}

	for key, value := range cluster.Roles {
		found := false
		if slices.Contains(list, key) {
			found = true
		}
		u.Roles[value] = found
	}
}

func (u *APIUser) Granted(grant string) error {
	if value, ok := u.Grants[grant]; ok {
		if !value {
			return v3.NewErrorResource(codes.PermissionDenied, v3.ErrUserNotGranted, "user", u.User).Err()
		}
		return nil
	}

	return v3.NewErrorResource(codes.PermissionDenied, v3.ErrGrantNotFound, "grant not found", "").Err()
}

func (cluster *Cluster) IsValidACL(strUser string, strPassword string, URL string, AuthMethod string) bool {
	if user, ok := cluster.APIUsers[strUser]; ok {
		//		fmt.Printf("password :" + user.Password)
		if user.Password == cluster.Conf.GetDecryptedPassword("api-credentials", strPassword) || AuthMethod == "oidc" {
			return cluster.IsURLPassACL(strUser, URL, true)
		}
		return false
	}
	//	for key, value := range cluster.Grants {

	return false
}

func (cluster *Cluster) GetAPIUser(strUser string, strPassword string) (APIUser, error) {
	if user, ok := cluster.APIUsers[strUser]; ok {
		if user.Password == strPassword {
			return user, nil
		}
		return APIUser{}, fmt.Errorf("incorrect password")
	}

	return APIUser{}, fmt.Errorf("user not found")
}

func (cluster *Cluster) SaveUserAcls(user string) (string, string) {
	granted, discarded := cluster.Conf.GetCompactGrants(cluster.APIUsers[user].Grants)
	return strings.Join(granted, " "), strings.Join(discarded, " ")
}

func (cluster *Cluster) SaveUserRoles(user string) string {
	var aEnabledRoles []string
	for grant, value := range cluster.APIUsers[user].Roles {
		if value {
			aEnabledRoles = append(aEnabledRoles, grant)
		}
	}
	return strings.Join(aEnabledRoles, " ")
}

func (cluster *Cluster) SaveAcls() {
	credentials := strings.Split(cluster.Conf.GetDecryptedValue("api-credentials")+","+cluster.Conf.GetDecryptedValue("api-credentials-external"), ",")
	aUserAcls := make([]string, 0)
	aUserDiscardAcls := make([]string, 0)
	for _, credential := range credentials {
		user, _ := misc.SplitPair(credential)
		enabledAcls, discardedAcls := cluster.SaveUserAcls(user)
		enabledRoles := cluster.SaveUserRoles(user)
		aUserAcls = append(aUserAcls, user+":"+enabledAcls+":"+enabledRoles+":"+cluster.Name)
		aUserDiscardAcls = append(aUserDiscardAcls, user+":"+discardedAcls)
	}
	cluster.Conf.APIUsersACLAllowExternal = strings.Join(aUserAcls, ",")
	cluster.Conf.APIUsersACLDiscard = strings.Join(aUserDiscardAcls, ",")
}

// func (cluster *Cluster) SetGrant(user string, grant string, enable bool) {
// 	if _, ok := cluster.APIUsers[user].Grants[grant]; ok {
// 		cluster.APIUsers[user].Grants[grant] = enable
// 	} else {
// 		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModGeneral, config.LvlErr, "Failed grant not found for user %s, grant %s ", user, grant)
// 	}

//		cluster.SaveAcls()
//	}
type ListUserACL struct {
	User  string
	ACLs  string
	Roles string
}

func (cluster *Cluster) GetClusterUserAllowACLs(acls string) map[string]ListUserACL {
	results := make(map[string]ListUserACL)
	usersAllowACL := strings.Split(acls, ",")

	for _, userACL := range usersAllowACL {
		if userACL == "" {
			continue
		}

		acl := ListUserACL{}
		useracl, listacls, listroles, listcluster := misc.SplitAcls(userACL)
		cluster_acls := strings.Split(listcluster, " ")

		if listcluster == "" || slices.Contains(cluster_acls, cluster.Name) {
			acl.User = useracl
			acl.ACLs = listacls
			acl.Roles = listroles

			results[useracl] = acl
		}
	}

	return results
}

func (cluster *Cluster) GetClusterUserDiscardACLs(acls string) map[string]ListUserACL {
	results := make(map[string]ListUserACL)
	usersDiscardACL := strings.Split(cluster.Conf.APIUsersACLDiscard, ",")

	for _, userACL := range usersDiscardACL {
		if userACL == "" {
			continue
		}

		acl := ListUserACL{}
		useracl, listacls, _, listcluster := misc.SplitAcls(userACL)
		cluster_acls := strings.Split(listcluster, " ")

		if listcluster == "" || slices.Contains(cluster_acls, cluster.Name) {
			acl.User = useracl
			acl.ACLs = listacls

			results[useracl] = acl
		}
	}

	return results
}

func (cluster *Cluster) LoadAPIUsers() error {
	meUsers := make(map[string]APIUser)
	credentials := strings.Split(cluster.Conf.Secrets["api-credentials"].Value+","+cluster.Conf.Secrets["api-credentials-external"].Value, ",")
	listACLs := cluster.GetClusterUserAllowACLs(cluster.Conf.APIUsersACLAllow)
	listDiscard := cluster.GetClusterUserDiscardACLs(cluster.Conf.APIUsersACLDiscard)
	listACLsExt := cluster.GetClusterUserAllowACLs(cluster.Conf.APIUsersACLAllowExternal)

	for _, credential := range credentials {
		// Prevent empty credentials
		if credential == "" {
			continue
		}

		// Assign User Credentials
		var newapiuser APIUser
		newapiuser.User, newapiuser.Password = misc.SplitPair(credential)
		if _, ok := meUsers[newapiuser.User]; ok {
			continue
		}

		newapiuser.Password = cluster.Conf.GetDecryptedPassword("api-credentials", newapiuser.Password)
		newapiuser.Grants = make(map[string]bool)
		newapiuser.Roles = make(map[string]bool)

		// Assign Roles and ACLs
		if userACL, ok := listACLs[newapiuser.User]; ok {
			cluster.SetNewUserGrants(&newapiuser, userACL.ACLs)
			cluster.SetNewUserRoles(&newapiuser, userACL.Roles)
		}

		if discardACL, ok := listDiscard[newapiuser.User]; ok {
			acls := strings.Split(discardACL.ACLs, " ")
			for _, acl := range acls {
				newapiuser.Grants[acl] = false
			}
		}

		// Assign Roles and ACLs
		if userACL, ok := listACLsExt[newapiuser.User]; ok {
			cluster.SetNewUserGrants(&newapiuser, userACL.ACLs)
			cluster.SetNewUserRoles(&newapiuser, userACL.Roles)
		}

		meUsers[newapiuser.User] = newapiuser
	}

	cluster.APIUsers = meUsers
	return nil
}

func (cluster *Cluster) IsURLPassDatabasesACL(strUser string, URL string) bool {
	if cluster.APIUsers[strUser].Grants[config.GrantClusterProcess] {
		if strings.Contains(URL, "/actions/run-jobs") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantProvDBProvision] {
		if strings.Contains(URL, "/actions/provision") {
			return true
		}
		if strings.Contains(URL, "/service-opensvc") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantProvDBUnprovision] {
		if strings.Contains(URL, "/actions/unprovision") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBStart] {
		if strings.Contains(URL, "/actions/start") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBStop] {
		if strings.Contains(URL, "/actions/stop") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterSwitchover] {
		if strings.Contains(URL, "/actions/switchover") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterFailover] {
		if strings.Contains(URL, "/actions/set-prefered") {
			return true
		}
		if strings.Contains(URL, "/actions/set-unrated") {
			return true
		}
		if strings.Contains(URL, "/actions/set-ignored") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBKill] {
		if strings.Contains(URL, "/actions/kill") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBOptimize] {
		if strings.Contains(URL, "/actions/analyze-pfs") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBAnalyse] {
		if strings.Contains(URL, "/actions/analyze-pfs") {
			return true
		}
		if strings.Contains(URL, "/actions/analyze-slowlog") {
			return true
		}
		if strings.Contains(URL, "/actions/reset-pfs-queries") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBReplication] {
		if strings.Contains(URL, "/all-slaves-status") {
			return true
		}
		if strings.Contains(URL, "/master-status") {
			return true
		}
		if strings.Contains(URL, "actions/start-slave") {
			return true
		}
		if strings.Contains(URL, "actions/stop-slave") {
			return true
		}
		if strings.Contains(URL, "actions/skip-replication-event") {
			return true
		}
		if strings.Contains(URL, "actions/reset-master") {
			return true
		}
		if strings.Contains(URL, "actions/reset-slave-all") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBBackup] {
		if strings.Contains(URL, "/actions/backup-logical") {
			return true
		}
		if strings.Contains(URL, "/actions/backup-error-log") {
			return true
		}
		if strings.Contains(URL, "/actions/backup-physical") {
			return true
		}
		if strings.Contains(URL, "/actions/backup-slowquery-log") {
			return true
		}
		if strings.Contains(URL, "/actions/flush-logs") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBRestore] {
		if strings.Contains(URL, "/actions/reseed/") {
			return true
		}
		if strings.Contains(URL, "/actions/pitr") {
			return true
		}
		if strings.Contains(URL, "/actions/reseed-cancel") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterProcess] {
		if strings.Contains(URL, "/actions/job-cancel/") {
			return true
		}
		if strings.Contains(URL, "/actions/reseed-cancel") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBReadOnly] {
		if strings.Contains(URL, "actions/toogle-read-only") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantProxyConfigFlag] {
		if strings.Contains(URL, "/config") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBLogs] {
		if strings.Contains(URL, "/processlist") {
			return true
		}
		if strings.Contains(URL, "/status-innodb") {
			return true
		}
		if strings.Contains(URL, "/errorlog") {
			return true
		}
		if strings.Contains(URL, "/slow-queries") {
			return true
		}
		if strings.Contains(URL, "/query-response-time") {
			return true
		}
		if strings.Contains(URL, "/meta-data-locks") {
			return true
		}
		if strings.Contains(URL, "/digest-statements-pfs") {
			return true
		}
		if strings.Contains(URL, "/digest-statements-slow") {
			return true
		}
		if strings.Contains(URL, "/actions/toogle-sql-error-log") {
			return true
		}
		if strings.Contains(URL, "/actions/toogle-sql-error-log") {
			return true
		}
		if strings.Contains(URL, "/actions/toogle-query-response-time") {
			return true
		}
		if strings.Contains(URL, "/actions/toogle-meta-data-locks") {
			return true
		}
		if strings.Contains(URL, "/actions/toogle-slow-query-table") {
			return true
		}
		if strings.Contains(URL, "/actions/toogle-slow-query-capture") {
			return true
		}
		if strings.Contains(URL, "/actions/toogle-slow-query") {
			return true
		}
		if strings.Contains(URL, "/actions/set-long-query-time") {
			return true
		}
		if strings.Contains(URL, "/actions/toogle-pfs-slow-query") {
			return true
		}
		if strings.Contains(URL, "/actions/toogle-slow-query") {
			return true
		}
		if strings.Contains(URL, "actions/toogle-innodb-monitor") {
			return true
		}
		if strings.Contains(URL, "/actions/explain-pfs") {
			return true
		}
		if strings.Contains(URL, "/actions/explain-slowlog") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBCapture] {
		if strings.Contains(URL, "/actions/toogle-slow-query-capture") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBMaintenance] {
		if strings.Contains(URL, "/actions/optimize") {
			return true
		}
		if strings.Contains(URL, "/actions/maintenance") {
			return true
		}
		if strings.Contains(URL, "/actions/set-maintenance") {
			return true
		}
		if strings.Contains(URL, "/actions/del-maintenance") {
			return true
		}
		if strings.Contains(URL, "/actions/wait-innodb-purge") {
			return true
		}
	}
	/*	if cluster.APIUsers[strUser].Grants[config.GrantDBConfigCreate] {
			if strings.Contains(URL, "/kill") {
				return true
			}
		}
		if cluster.APIUsers[strUser].Grants[config.GrantDBConfigGet] {
			if strings.Contains(URL, "/kill") {
				return true
			}
		}
		if cluster.APIUsers[strUser].Grants[config.GrantDBConfigFlag] {
			if strings.Contains(URL, "/kill") {
				return true
			}
		}*/
	if cluster.APIUsers[strUser].Grants[config.GrantDBShowVariables] {
		if strings.Contains(URL, "/variables") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBShowSchema] {
		if strings.Contains(URL, "/tables") {
			return true
		}
		if strings.Contains(URL, "/vtables") {
			return true
		}
		if strings.Contains(URL, "/tables") {
			return true
		}
		if strings.Contains(URL, "/schemas") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBShowStatus] {
		if strings.Contains(URL, "/status") {
			return true
		}
		if strings.Contains(URL, "/status-delta") {
			return true
		}
	}
	cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModGeneral, config.LvlInfo, "ACL check failed for user %s : %s ", strUser, URL)
	return false
}

func (cluster *Cluster) IsURLPassProxiesACL(strUser string, URL string) bool {

	if cluster.APIUsers[strUser].Grants[config.GrantProvProxyProvision] {
		if strings.Contains(URL, "/actions/provision") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantProvProxyUnprovision] {
		if strings.Contains(URL, "/actions/unprovision") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantProxyStart] {
		if strings.Contains(URL, "/actions/start") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantProxyStop] {
		if strings.Contains(URL, "/actions/stop") {
			return true
		}
	}
	cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModGeneral, config.LvlInfo, "ACL proxy check failed for user %s : %s ", strUser, URL)

	return false
}

func (cluster *Cluster) IsURLPassACL(strUser string, URL string, errorPrint bool) bool {
	switch URL {
	case "/api/login":
		return true
	case "/api/auth/callback":
		return true
	case "/api/clusters":
		return true
	case "/api/monitor":
		return true
	case "/api/clusters/" + cluster.Name + "/actions/waitdatabases":
		return true
	case "/api/clusters/" + cluster.Name:
		return true
	case "/api/clusters/" + cluster.Name + "/diffvariables":
		return true
	}

	if strings.Contains(URL, "/api/clusters/settings/actions/switch") {
		return cluster.APIUsers[strUser].Grants[config.GrantGlobalSettings]
	}
	if strings.Contains(URL, "/api/clusters/settings/actions/set") {
		return cluster.APIUsers[strUser].Grants[config.GrantGlobalSettings]
	}
	if strings.Contains(URL, "/api/clusters/settings/actions/clear") {
		return cluster.APIUsers[strUser].Grants[config.GrantGlobalSettings]
	}
	if strings.Contains(URL, "/api/clusters/settings/actions/reload-clusters-plans") {
		return cluster.APIUsers[strUser].Grants[config.GrantGlobalSettings]
	}

	if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/servers") {
		return cluster.IsURLPassDatabasesACL(strUser, URL)
	}
	if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/proxies") {
		return cluster.IsURLPassProxiesACL(strUser, URL)
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterSharding] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/schema") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/shardclusters") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterProcess] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/jobs") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterProcess] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/top") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterShowBackups] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/backups") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterShowRoutes] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/queryrules") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterShowCertificates] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/certificates") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterCertificatesReload] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/certificates-reload") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterCertificatesRotate] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/certificates-rotate") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterResetSLA] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/reset-sla") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterCreateMonitor] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/addserver") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterDropMonitor] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/dropserver") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterSwitchover] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/switchover") {
			return true
		}
	}

	if cluster.APIUsers[strUser].Grants[config.GrantClusterTraffic] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/stop-traffic") {
			return true

		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/start-traffic") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBBackup] {
		if strings.Contains(URL, "/actions/master-logical-backup") {
			return true
		}
		if strings.Contains(URL, "/actions/master-physical-backup") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterBench] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/sysbench") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterTest] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/sysbench") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/tests/") {
			return true
		}

	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterFailover] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/failover") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterReplication] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/replication/bootstrap") {
			return true
		}

		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/replication/cleanup") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterRolling] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/optimize") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/rolling") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/cancel-rolling-restart") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/cancel-rolling-reprov") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterRotatePasswords] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/rotate-passwords") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantDBConfigFlag] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/drop-db-tag") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/add-db-tag") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/apply-dynamic-config") {
			return true
		}

	}
	if cluster.APIUsers[strUser].Grants[config.GrantProxyConfigFlag] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/drop-proxy-tag") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/add-proxy-tag") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterSettings] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/reload") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/switch") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/set") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/clear") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/discover") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/reset-failover-control") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterChecksum] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/actions/checksum-all-tables") {
			return true
		}
	}

	if cluster.APIUsers[strUser].Grants[config.GrantProvCluster] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/services/actions/provision") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/actions/add") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantProvClusterUnprovision] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/services/actions/unprovision") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterCreate] {
		if strings.Contains(URL, "/api/clusters/actions/add") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterDelete] {
		if strings.Contains(URL, "/api/clusters/actions/delete") {
			return true
		}
	}
	if cluster.APIUsers[strUser].Grants[config.GrantClusterConfigGraphs] {
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/set-graphite-filterlist") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/reload-graphite-filterlist") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/settings/actions/reset-graphite-filterlist") {
			return true
		}
	}

	if cluster.APIUsers[strUser].Grants[config.GrantClusterGrant] {
		if strings.Contains(URL, "/api/monitor/actions/adduser/") {
			return true
		}
		if strings.Contains(URL, "/api/clusters/"+cluster.Name+"/users/add") {
			return true
		}
	}

	/*	case cluster.APIUsers[strUser].Grants[config.GrantClusterGrant] == true:
			return false
		case cluster.APIUsers[strUser].Grants[config.GrantClusterDropMonitor] == true:
			return false
		case cluster.APIUsers[strUser].Grants[config.GrantClusterCreate] == true:
			return false
	*/

	// Print error with no valid ACL
	if errorPrint {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModGeneral, config.LvlInfo, "ACL check failed for user %s : %s ", strUser, URL)
	}
	return false
}
