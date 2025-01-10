package cluster

import (
	"fmt"
	"slices"
	"strings"

	"github.com/signal18/replication-manager/config"
)

type Alert struct {
	From        string
	To          string
	Instance    string
	State       string
	PrevState   string
	Cluster     string
	Host        string
	Destination string
	User        string
	Password    string
	TlsVerify   bool
	Resolved    bool
}

// PrepareMail prepares the mail message
// if the cloud18 flag is set to true
//   - if sendDbOps is true and the external dbops is set, the mail will be sent to the external dbops
//   - if sendSysOps is true and the external sysops is set, the mail will be sent to the external sysops
func (cluster *Cluster) PrepareMail(sendDbOps, sendSysOps bool) (string, string, []string) {
	address := cluster.Conf.MonitorAddress
	from := cluster.Conf.MailFrom
	to := strings.Split(cluster.Conf.MailTo, ",")
	if cluster.Conf.Cloud18 {
		address = fmt.Sprintf("%s (%s)", cluster.Conf.APIPublicURL, address)
		if cluster.Conf.Cloud18GitUser != "" && !slices.Contains(to, cluster.Conf.Cloud18GitUser) {
			to = append(to, cluster.Conf.Cloud18GitUser)
		}
		if sendDbOps && cluster.Conf.Cloud18ExternalDbOps != "" && !slices.Contains(to, cluster.Conf.Cloud18ExternalDbOps) {
			to = append(to, cluster.Conf.Cloud18ExternalDbOps)
		}
		if sendSysOps && cluster.Conf.Cloud18ExternalSysOps != "" && !slices.Contains(to, cluster.Conf.Cloud18ExternalSysOps) {
			to = append(to, cluster.Conf.Cloud18ExternalSysOps)
		}
	}

	return address, from, to
}

// SendMailFromAlert sends an email from an alert
// if the cloud18 flag is set to true
//   - if sendDbOps is true and the external dbops is set, the mail will be sent to the external dbops
//   - if sendSysOps is true and the external sysops is set, the mail will be sent to the external sysops
//
// the subject will be "Replication-Manager@<monitor_address> Alert - Cluster <cluster_name> state change detected"
func (cluster *Cluster) SendMailFromAlert(a Alert, sendDbOps, sendSysOps bool) error {
	address, from, to := cluster.PrepareMail(sendDbOps, sendSysOps)
	host := ""
	if a.Host != "" {
		host = "Host: " + a.Host + "\n"
	}

	subj := fmt.Sprintf("Replication-Manager@%s Alert - Cluster %s state change detected", address, cluster.Name)
	msg := fmt.Sprintf("Alert: State changed from %s to %s\nMonitor: %s\nCluster: %s\n%s", a.PrevState, a.State, address, a.Cluster, host)
	if a.PrevState == "" {
		if a.Resolved {
			msg = fmt.Sprintf("Resolved: %s\nMonitor: %s\nCluster: %s\n%s", a.State, address, a.Cluster, host)
		} else {
			msg = fmt.Sprintf("Alert: %s\nMonitor: %s\nCluster: %s\n%s", a.State, address, a.Cluster, host)
		}
	}

	err := cluster.Mailer.SendEmailMessage(msg, subj, from, strings.Join(to, ","), "", cluster.Conf.MailSMTPTLSSkipVerify)
	if err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModGeneral, config.LvlErr, "Error sending email for alert %s on %s: %v", a.State, a.Host, err)
		return err
	}

	return nil
}

// SendMail sends an email
// if the cloud18 flag is set to true
//   - if sendDbOps is true and the external dbops is set, the mail will be sent to the external dbops
//   - if sendSysOps is true and the external sysops is set, the mail will be sent to the external sysops
//
// if isAlert is true, the message will be prepended with "Alert: "
func (cluster *Cluster) SendMail(msg, subj string, isAlert, sendDbOps, sendSysOps bool) error {
	address, from, to := cluster.PrepareMail(sendDbOps, sendSysOps)

	if isAlert {
		msg = fmt.Sprintf("Alert: %s\nMonitor: %s\nCluster: %s\n", msg, address, cluster.Name)
	}

	err := cluster.Mailer.SendEmailMessage(msg, subj, from, strings.Join(to, ","), "", cluster.Conf.MailSMTPTLSSkipVerify)
	if err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModGeneral, config.LvlErr, "Error sending email for with subject %s. Err: %v", subj, err)
		return err
	}

	return nil
}
