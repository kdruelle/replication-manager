package cluster

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/helloyi/go-sshclient"
	sshcli "github.com/helloyi/go-sshclient"
	"github.com/signal18/replication-manager/config"
	"github.com/signal18/replication-manager/utils/misc"
)

func (cluster *Cluster) OnPremiseGetSSHKey(user string) string {

	// repmanuser := os.Getenv("HOME")
	// if repmanuser == "" {
	// 	repmanuser = "/root"
	// 	if user != "root" {
	// 		repmanuser = "/home/" + user
	// 	}
	// }
	key := cluster.OsUser.HomeDir + "/.ssh/id_rsa"

	if cluster.Conf.OnPremiseSSHPrivateKey != "" {
		key = cluster.Conf.OnPremiseSSHPrivateKey
	}
	return key
}

func (cluster *Cluster) OnPremiseConnect(server *ServerMonitor) (*sshclient.Client, error) {
	if cluster.IsInFailover() {
		return nil, errors.New("OnPremise provisioning cancel during failover")
	}
	if !cluster.Conf.OnPremiseSSH {
		return nil, errors.New("onpremise-ssh disable ")
	}
	user, password := misc.SplitPair(cluster.Conf.GetDecryptedValue("onpremise-ssh-credential"))

	key := cluster.OnPremiseGetSSHKey(user)
	if password != "" {
		client, err := sshcli.DialWithPasswd(misc.Unbracket(server.Host)+":"+strconv.Itoa(cluster.Conf.OnPremiseSSHPort), user, password)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("OnPremise Provisioning via SSH %s %s", err.Error(), key))
		}
		return client, nil
	} else {
		client, err := sshcli.DialWithKey(misc.Unbracket(server.Host)+":"+strconv.Itoa(cluster.Conf.OnPremiseSSHPort), user, key)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("OnPremise Provisioning via SSH %s %s", err.Error(), key))
		}
		return client, nil
	}
	// return nil, errors.New("onpremise-ssh no key no password ")
}

func (cluster *Cluster) OnPremiseProvisionDatabaseService(server *ServerMonitor) {
	client, err := cluster.OnPremiseConnect(server)
	if err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModOrchestrator, config.LvlErr, "OnPremise provision database failed to connect : %s", err)
		cluster.errorChan <- err
		return
	}
	defer client.Close()
	err = cluster.OnPremiseSetEnv(client, server)
	if err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModOrchestrator, config.LvlErr, "OnPremise provision database failed in env setup : %s", err)
		cluster.errorChan <- err
	}
	dbtype := "mariadb"
	cmd := "wget --no-check-certificate -q -O- $REPLICATION_MANAGER_URL/static/configurator/onpremise/repository/debian/" + dbtype + "/bootstrap | sh"
	if cluster.Configurator.HaveDBTag("rpm") {
		cmd = "wget --no-check-certificate -q -O- $REPLICATION_MANAGER_URL/static/configurator/onpremise/repository/redhat/" + dbtype + "/bootstrap | sh"
	}
	if cluster.Configurator.HaveDBTag("package") {
		cmd = "wget --no-check-certificate -q -O- $REPLICATION_MANAGER_URL/static/configurator/onpremise/package/linux/" + dbtype + "/bootstrap | sh"
	}

	out, err := client.Cmd(cmd).SmartOutput()
	if err != nil {
		cluster.errorChan <- err
	}
	cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModOrchestrator, config.LvlInfo, "OnPremise Provisioning  : %s", string(out))
	cluster.errorChan <- nil
}

func (cluster *Cluster) OnPremiseUnprovisionDatabaseService(server *ServerMonitor) {

	cluster.errorChan <- nil

}

func (cluster *Cluster) OnPremiseStopDatabaseService(server *ServerMonitor) error {
	//s.JobServerStop() need an agent or ssh to trigger this
	server.Shutdown()
	return nil
}

func (cluster *Cluster) OnPremiseGetNodes() ([]Agent, error) {
	//cat proc/cpuinfo | grep "cpu cores" | wc -l
	//nb de cpu

	var Agents []Agent
	for id, server := range cluster.Servers {
		var agent Agent
		client, err := cluster.OnPremiseConnect(server)
		if err != nil {
			cluster.errorChan <- err
		}
		defer client.Close()

		agent.Id = strconv.Itoa(id)

		if cluster.Configurator.HaveDBTag("linux") {
			agent.OsName = "linux"
			//get cpu core
			cmd := "cat /proc/cpuinfo | grep 'cpu cores' | wc -l | awk '{print $1}'"
			out, err := client.Cmd(cmd).SmartOutput()
			if err != nil {
				cluster.errorChan <- err
			}
			re := regexp.MustCompile("[0-9]+")
			i, _ := strconv.ParseInt(re.FindAllString(string(out), -1)[0], 10, 64)
			agent.CpuCores = i

			//uname -m
			//get arch
			cmd = "uname -r"
			out, err = client.Cmd(cmd).SmartOutput()
			if err != nil {
				cluster.errorChan <- err
			}
			agent.OsKernel = string(out)

			//cat proc/cpuinfo | grep "cache size" | head -n 1
			//get mem
			cmd = "cat /proc/meminfo | grep -i  MemTotal | awk '{print $2}'"
			out, err = client.Cmd(cmd).SmartOutput()
			if err != nil {
				cluster.errorChan <- err
			}

			re = regexp.MustCompile("[0-9]+")
			i, _ = strconv.ParseInt(re.FindAllString(string(out), -1)[0], 10, 64)

			agent.MemBytes = i / 1024

			//cat proc/cpuinfo | grep 'cpu MHz'
			//get cpu freq
			cmd = "cat /proc/cpuinfo | grep 'cpu MHz' | head -n 1 | awk '{print $4}'"
			out, err = client.Cmd(cmd).SmartOutput()
			if err != nil {
				cluster.errorChan <- err
			}
			re = regexp.MustCompile("[0-9]+")
			i, _ = strconv.ParseInt(re.FindAllString(string(out), -1)[0], 10, 64)
			agent.CpuFreq = int64(i)
		}

		agent.HostName = server.Name
		Agents = append(Agents, agent)

	}

	return Agents, nil
}

func (cluster *Cluster) OnPremiseSetEnv(client *sshclient.Client, server *ServerMonitor) error {

	buf := strings.NewReader(server.GetSshEnv())
	/*
		  REPLICATION_MANAGER_USER
			REPLICATION_MANAGER_PASSWORD
			REPLICATION_MANAGER_URL
			REPLICATION_MANAGER_CLUSTER_NAME
			REPLICATION_MANAGER_HOST_NAME
			REPLICATION_MANAGER_HOST_USER
			REPLICATION_MANAGER_HOST_PASSWORD
			REPLICATION_MANAGER_HOST_PORT

	*/
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	var err error
	if client.Shell().SetStdio(buf, &stdout, &stderr).Start(); err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModOrchestrator, config.LvlWarn, "OnPremise start ssh setup env %s", stderr.String())
		return err
	}
	cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModOrchestrator, config.LvlInfo, "OnPremise start database install secret env: %s", stdout.String())

	return nil
}

func (cluster *Cluster) OnPremiseStartDatabaseService(server *ServerMonitor) error {

	server.SetWaitStartCookie()
	cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModOrchestrator, config.LvlInfo, "OnPremise start database via ssh script")
	client, err := cluster.OnPremiseConnect(server)
	if err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModOrchestrator, config.LvlErr, "OnPremise start database via ssh failed : %s", err)
		return err
	}
	defer client.Close()

	cmd := cluster.Configurator.GetSshStartDBScript()

	filerc, err := os.Open(cmd)
	if err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModOrchestrator, config.LvlErr, "OnPremise start database via ssh script %%s failed : %s ", cmd, err)
		return errors.New("can't open script")
	}

	defer filerc.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(filerc)

	buf2 := strings.NewReader(server.GetSshEnv())
	r := io.MultiReader(buf2, buf)

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	if client.Shell().SetStdio(r, &stdout, &stderr).Start(); err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModOrchestrator, config.LvlWarn, "OnPremise start database via ssh script %s", stderr.String())
	}
	out := stdout.String()

	cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModOrchestrator, config.LvlInfo, "OnPremise start script: %s ,out: %s ,err: %s", cmd, out, stderr.String())

	return nil
}
