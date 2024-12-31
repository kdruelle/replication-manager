package server

import (
	"fmt"
	"time"

	"github.com/signal18/replication-manager/cluster"
	"github.com/signal18/replication-manager/config"
)

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
	to := repman.Conf.MailTo
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

func (repman *ReplicationManager) SendSponsorActivationMail(cl *cluster.Cluster, userform cluster.UserForm) error {
	to := userform.Username

	subj := fmt.Sprintf("Subscription Active for Cluster %s: %s", cl.Name, userform.Username)
	msg := fmt.Sprintf(`Dear Sponsor,

We’re excited to let you know that your subscription is now active!

As part of your subscription, you’ll soon receive an email containing your database credentials. 

You can use these credentials to access your cluster resources after the provisioning complete.

If you have any questions in the meantime, feel free to contact our support team by replying to this email.

Thank you for choosing Cloud18!

Best regards,

Signal18
`)

	return repman.Mailer.SendEmailMessage(msg, subj, repman.Conf.MailFrom, to, "", repman.Conf.MailSMTPTLSSkipVerify)
}

func (repman *ReplicationManager) SendSponsorCredentialsMail(cl *cluster.Cluster) error {
	var user cluster.APIUser
	for _, u := range cl.APIUsers {
		if u.Roles[config.RoleSponsor] {
			user = u
			break
		}
	}

	if user.User == "" {
		return fmt.Errorf("No sponsor found for cluster %s", cl.Name)
	}

	to := user.User

	subj := fmt.Sprintf("DB Credentials for Cluster %s: %s", cl.Name, user.User)
	msg := fmt.Sprintf(`Dear Sponsor,

We are pleased to provide you with the necessary credentials to access your database. Please find the connection details below:

- Cloud18 DB Read-Write Split Server: %s
- Cloud18 DB Read-Write Server: %s
- Cloud18 DB Read-Only Server: %s
- Username: %s
- Password: %s

If you require assistance with connecting to the database, please do not hesitate to contact our support team.

This email contains confidential information. Please do not share it with unauthorized individuals. If you are not the intended recipient, kindly delete this email immediately and notify us.

Thank you for choosing Cloud18. We are committed to supporting your success.

Best regards,

Signal18
`, cl.Conf.Cloud18DatabaseReadWriteSplitSrvRecord, cl.Conf.Cloud18DatabaseReadWriteSrvRecord, cl.Conf.Cloud18DatabaseReadSrvRecord, cl.GetSponsorUser(), cl.GetSponsorPass())

	return repman.Mailer.SendEmailMessage(msg, subj, repman.Conf.MailFrom, to, "", repman.Conf.MailSMTPTLSSkipVerify)
}

func (repman *ReplicationManager) SendDBACredentialsMail(cl *cluster.Cluster, to, delegator string) error {
	subj := fmt.Sprintf("DB Credentials for Cluster %s", cl.Name)
	msg := fmt.Sprintf(`Dear DBA,

User %s has delegated you to access the database.

Please find below the credentials required to connect to the database:

- Cloud18 DB Read-Write Split Server: %s
- Cloud18 DB Read-Write Server: %s
- Cloud18 DB Read-Only Server: %s
- Username: %s
- Password: %s

Please treat this information as confidential and do not share it with unauthorized individuals. If you are not the intended recipient, please delete this email immediately and notify us.

Thank you for your attention to this matter.

Best regards,

Signal18
`, delegator, cl.Conf.Cloud18DatabaseReadWriteSplitSrvRecord, cl.Conf.Cloud18DatabaseReadWriteSrvRecord, cl.Conf.Cloud18DatabaseReadSrvRecord, cl.GetDbaUser(), cl.GetDbaPass())

	return repman.Mailer.SendEmailMessage(msg, subj, repman.Conf.MailFrom, to, "", repman.Conf.MailSMTPTLSSkipVerify)
}

func (repman *ReplicationManager) SendSysAdmCredentialsMail(cl *cluster.Cluster, to, delegator string) error {
	subj := fmt.Sprintf("DB Credentials for Cluster %s", cl.Name)
	msg := fmt.Sprintf(`Dear System Admin,

User %s has delegated you to access the database server.

Please find below the credentials required to connect to the server:

- Cloud18 DB Read-Write Split Server: %s
- Cloud18 DB Read-Write Server: %s
- Cloud18 DB Read-Only Server: %s
- Username: %s
- Password: %s

Please treat this information as confidential and do not share it with unauthorized individuals. If you are not the intended recipient, please delete this email immediately and notify us.

Thank you for your attention to this matter.

Best regards,

Signal18
`, delegator, cl.Conf.Cloud18DatabaseReadWriteSplitSrvRecord, cl.Conf.Cloud18DatabaseReadWriteSrvRecord, cl.Conf.Cloud18DatabaseReadSrvRecord, cl.GetDbaUser(), cl.GetDbaPass())

	return repman.Mailer.SendEmailMessage(msg, subj, repman.Conf.MailFrom, to, "", repman.Conf.MailSMTPTLSSkipVerify)
}
