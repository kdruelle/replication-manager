package server

import (
	"fmt"
	"strings"

	"github.com/signal18/replication-manager/cluster"
	"github.com/signal18/replication-manager/utils/misc"
)

func (repman *ReplicationManager) AcceptSubscription(userform cluster.UserForm, cl *cluster.Cluster) error {
	user := userform.Username
	if auser, ok := cl.APIUsers[user]; !ok {
		return fmt.Errorf("User %s does not exist ", user)
	} else {
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

		// log.Printf("User %s grants %s roles %s", user, userform.Grants, userform.Roles)

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
	}

	return nil
}

func (repman *ReplicationManager) RemoveSubscription(userform cluster.UserForm, isRemoveSponsor bool, cl *cluster.Cluster) error {
	user := userform.Username
	if auser, ok := cl.APIUsers[user]; !ok {
		return fmt.Errorf("User %s does not exist ", user)
	} else {
		grants := make([]string, 0)
		roles := make([]string, 0)
		for grant, v := range auser.Grants {
			if v {
				grants = append(grants, grant)
			}
		}
		userform.Grants = strings.Join(grants, " ")

		for role, v := range auser.Roles {
			if isRemoveSponsor {
				if v && role != "sponsor" {
					roles = append(roles, role)

					// If use has no other roles, remove grants
					if len(roles) == 0 {
						userform.Grants = ""
					}
				}
			} else {
				if v && role != "pending" {
					roles = append(roles, role)
				}
			}
		}
		userform.Roles = strings.Join(roles, " ")

		if userform.Grants == "" {
			cl.DropUser(userform)
		} else {
			cl.UpdateUser(userform, "admin")
		}
	}

	return nil
}
