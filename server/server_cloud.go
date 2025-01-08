package server

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/signal18/replication-manager/cluster"
	"github.com/signal18/replication-manager/config"
	"github.com/signal18/replication-manager/utils/misc"
)

func (repman *ReplicationManager) AcceptSubscription(userform cluster.UserForm, cl *cluster.Cluster) error {
	user := userform.Username
	auser, ok := cl.APIUsers[user]
	if !ok {
		return fmt.Errorf("User %s does not exist ", user)
	}

	if v, ok := auser.Roles[config.RolePending]; !ok || !v {
		return fmt.Errorf("User %s does not have 'pending' role", user)
	}

	grants := strings.Split("db show proxy grant extrole sales-unsubscribe", " ")
	roles := strings.Split("sponsor", " ")
	for grant, v := range auser.Grants {
		if v {
			grants = append(grants, grant)
		}
	}
	userform.Grants = strings.Join(grants, " ")

	for role, v := range auser.Roles {
		if v && role != "pending" {
			roles = append(roles, role)
		}
	}
	userform.Roles = strings.Join(roles, " ")

	// If external sysops different from cloud18 git user
	if cl.Conf.Cloud18ExternalSysOps != "" && cl.Conf.Cloud18ExternalSysOps != cl.Conf.Cloud18GitUser {
		esys := repman.CreateExtSysopsForm(cl.Conf.Cloud18ExternalSysOps)
		if euser, ok := cl.APIUsers[cl.Conf.Cloud18ExternalSysOps]; !ok {
			cl.AddUser(esys, cl.Conf.Cloud18GitUser, false)
		} else {
			esys.Grants = cl.AppendGrants(esys.Grants, &euser)
			esys.Roles = cl.AppendRoles(esys.Roles, &euser)
			cl.UpdateUser(esys, cl.Conf.Cloud18GitUser, false)
		}
	}

	// If external dbops different from cloud18 dbops
	if cl.Conf.Cloud18ExternalDbOps != "" && cl.Conf.Cloud18ExternalDbOps != cl.Conf.Cloud18DbOps {
		edbops := repman.CreateExtDBOpsForm(cl.Conf.Cloud18ExternalDbOps)
		if edbuser, ok := cl.APIUsers[cl.Conf.Cloud18ExternalDbOps]; !ok {
			cl.AddUser(edbops, cl.Conf.Cloud18GitUser, false)
		} else {
			edbops.Grants = cl.AppendGrants(edbops.Grants, &edbuser)
			edbops.Roles = cl.AppendRoles(edbops.Roles, &edbuser)
			cl.UpdateUser(edbops, cl.Conf.Cloud18GitUser, false)
		}
	}

	new_acls := make([]string, 0)

	acls := strings.Split(cl.Conf.APIUsersACLAllowExternal, ",")
	for _, acl := range acls {
		useracl, listgrants, _, listroles := misc.SplitAcls(acl)
		// log.Printf("ACL: %s", acl)
		if useracl == user {
			acl = useracl + ":" + userform.Grants + ":" + cl.Name
			if userform.Roles != "" {
				acl = acl + ":" + userform.Roles
			}
			new_acls = append(new_acls, acl)
		} else {
			old_roles := strings.Split(listroles, " ")
			new_roles := make([]string, 0)
			for _, role := range old_roles {
				if role == "pending" {
					continue
				}
				new_roles = append(new_roles, role)
			}
			acl = useracl + ":" + listgrants + ":" + cl.Name
			if len(new_roles) > 0 {
				acl = acl + ":" + strings.Join(new_roles, " ")
			}
			new_acls = append(new_acls, acl)
		}
		// log.Printf("New ACL: %s", acl)
	}

	cl.Conf.APIUsersACLAllowExternal = strings.Join(new_acls, ",")
	// log.Printf("APIUsersACLAllowExternal: %s", cl.Conf.APIUsersACLAllowExternal)

	cl.LoadAPIUsers()
	cl.SaveAcls()
	cl.Save()

	return nil
}

func (repman *ReplicationManager) CancelSubscription(userform cluster.UserForm, cl *cluster.Cluster) error {
	user := userform.Username
	auser, ok := cl.APIUsers[user]
	if !ok {
		return fmt.Errorf("User %s does not exist ", user)
	}
	grants := make([]string, 0)
	roles := make([]string, 0)
	for grant, v := range auser.Grants {
		if v {
			grants = append(grants, grant)
		}
	}
	userform.Grants = strings.Join(grants, " ")

	for role, v := range auser.Roles {
		if v && role != "pending" {
			roles = append(roles, role)
		}
	}

	userform.Roles = strings.Join(roles, " ")

	if userform.Grants == "" {
		cl.DropUser(userform, true)
	} else {
		cl.UpdateUser(userform, "admin", true)
	}

	return nil
}

func (repman *ReplicationManager) EndSubscription(userform cluster.UserForm, cl *cluster.Cluster) error {
	user := userform.Username
	auser, ok := cl.APIUsers[user]

	if !ok {
		return fmt.Errorf("User %s does not exist ", user)
	}

	if v, ok := auser.Roles[config.RoleSponsor]; !ok || !v {
		return fmt.Errorf("User %s does not have 'sponsor' role", user)
	}

	grants := make([]string, 0)
	roles := make([]string, 0)
	for grant, v := range auser.Grants {
		if v {
			grants = append(grants, grant)
		}
	}
	userform.Grants = strings.Join(grants, " ")

	for role, v := range auser.Roles {
		if v && role != "sponsor" {
			roles = append(roles, role)
		}
	}

	// If use has no other roles, remove grants
	if len(roles) == 0 {
		userform.Grants = ""
	}

	roles = append(roles, config.RoleUnsubscribed)

	userform.Roles = strings.Join(roles, " ")

	cl.UpdateUser(userform, "admin", true)

	return nil
}

func (repman *ReplicationManager) BashScriptSalesSubscribe(mycluster *cluster.Cluster, subscriber string) error {
	if repman.Conf.Cloud18SalesSubscriptionScript != "" {
		mycluster.LogModulePrintf(mycluster.Conf.Verbose, config.ConstLogModGeneral, "INFO", "Calling cluster subscription script")
		var out []byte
		out, err := exec.Command(repman.Conf.Cloud18SalesSubscriptionScript, mycluster.Name, subscriber).CombinedOutput()
		if err != nil {
			mycluster.LogModulePrintf(mycluster.Conf.Verbose, config.ConstLogModGeneral, "ERROR", "%s", err)
		}

		mycluster.LogModulePrintf(mycluster.Conf.Verbose, config.ConstLogModGeneral, "INFO", "Cluster subscription script complete %s:", string(out))
	}
	return nil
}

func (repman *ReplicationManager) BashScriptSalesSubscriptionValidate(mycluster *cluster.Cluster, subscriber, operator string) error {
	if repman.Conf.Cloud18SalesSubscriptionValidateScript != "" {
		mycluster.LogModulePrintf(mycluster.Conf.Verbose, config.ConstLogModGeneral, "INFO", "Calling cluster subscription script")
		var out []byte
		out, err := exec.Command(repman.Conf.Cloud18SalesSubscriptionValidateScript, mycluster.Name, subscriber, operator).CombinedOutput()
		if err != nil {
			mycluster.LogModulePrintf(mycluster.Conf.Verbose, config.ConstLogModGeneral, "ERROR", "%s", err)
		}

		mycluster.LogModulePrintf(mycluster.Conf.Verbose, config.ConstLogModGeneral, "INFO", "Cluster subscription script complete %s:", string(out))
	}
	return nil
}

func (repman *ReplicationManager) BashScriptSalesUnsubscribe(mycluster *cluster.Cluster, subscriber, operator string) error {
	if repman.Conf.Cloud18SalesUnsubscribeScript != "" {
		repman.LogModulePrintf(repman.Conf.Verbose, config.ConstLogModGeneral, "INFO", "Calling cluster subscription script")
		var out []byte
		out, err := exec.Command(repman.Conf.Cloud18SalesUnsubscribeScript, mycluster.Name, subscriber, operator).CombinedOutput()
		if err != nil {
			mycluster.LogModulePrintf(mycluster.Conf.Verbose, config.ConstLogModGeneral, "ERROR", "%s", err)
		}

		mycluster.LogModulePrintf(mycluster.Conf.Verbose, config.ConstLogModGeneral, "INFO", "Cluster subscription script complete %s:", string(out))
	}
	return nil
}

func (repman *ReplicationManager) GetPartnerFromDomain(domain string) config.Partner {
	tmpPartner := config.Partner{}
	for _, partner := range repman.Partners {
		domains := strings.Split(partner.Domains, ",")
		if slices.Contains(domains, domain) {
			return partner
		} else if slices.Contains(domains, "cloud18") {
			tmpPartner = partner
		}
	}

	return tmpPartner
}
