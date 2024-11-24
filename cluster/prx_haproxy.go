// replication-manager - Replication Manager Monitoring and CLI for MariaDB and MySQL
// Copyright 2017-2021 SIGNAL18 CLOUD SAS
// Authors: Guillaume Lefranc <guillaume@signal18.io>
//          Stephane Varoqui  <svaroqui@gmail.com>
// This source code is licensed under the GNU General Public License, version 3.

package cluster

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/signal18/replication-manager/config"
	"github.com/signal18/replication-manager/router/haproxy"
	"github.com/signal18/replication-manager/utils/state"
	"github.com/spf13/pflag"
)

type HaproxyProxy struct {
	Proxy
}

func NewHaproxyProxy(placement int, cluster *Cluster, proxyHost string) *HaproxyProxy {
	conf := cluster.Conf
	prx := new(HaproxyProxy)
	prx.SetPlacement(placement, conf.ProvProxAgents, conf.SlapOSHaProxyPartitions, conf.HaproxyHostsIPV6, conf.HaproxyJanitorWeights)
	prx.Type = config.ConstProxyHaproxy
	prx.Port = strconv.Itoa(conf.HaproxyAPIPort)
	prx.ReadPort = conf.HaproxyReadPort
	prx.WritePort = conf.HaproxyWritePort
	prx.ReadWritePort = conf.HaproxyWritePort
	prx.Name = proxyHost
	prx.Host = proxyHost
	if conf.ProvNetCNI {
		prx.Host = prx.Host + "." + cluster.Name + ".svc." + conf.ProvOrchestratorCluster
	}
	prx.User = conf.HaproxyUser
	prx.Pass = cluster.Conf.GetDecryptedValue("haproxy-password")

	return prx
}

func (proxy *HaproxyProxy) AddFlags(flags *pflag.FlagSet, conf *config.Config) {
	flags.BoolVar(&conf.HaproxyOn, "haproxy", false, "Wrapper to use HAProxy on same host")
	flags.StringVar(&conf.HaproxyMode, "haproxy-mode", "runtimeapi", "HAProxy mode [standby|runtimeapi|dataplaneapi]")
	flags.BoolVar(&conf.HaproxyDebug, "haproxy-debug", true, "Extra info on monitoring backend")
	flags.IntVar(&conf.HaproxyLogLevel, "haproxy-log-level", 1, "Log level for debug")
	flags.StringVar(&conf.HaproxyUser, "haproxy-user", "admin", "HAProxy API user")
	flags.StringVar(&conf.HaproxyPassword, "haproxy-password", "admin", "HAProxy API password")
	flags.StringVar(&conf.HaproxyHosts, "haproxy-servers", "127.0.0.1", "HAProxy hosts")
	flags.StringVar(&conf.HaproxyJanitorWeights, "haproxy-janitor-weights", "100", "Weight of each HAProxy host inside janitor proxy")
	flags.IntVar(&conf.HaproxyAPIPort, "haproxy-api-port", 1999, "HAProxy runtime api port")
	flags.IntVar(&conf.HaproxyWritePort, "haproxy-write-port", 3306, "HAProxy read-write port to leader")
	flags.IntVar(&conf.HaproxyReadPort, "haproxy-read-port", 3307, "HAProxy load balancer read port to all nodes")
	flags.IntVar(&conf.HaproxyStatPort, "haproxy-stat-port", 1988, "HAProxy statistics port")
	flags.StringVar(&conf.HaproxyBinaryPath, "haproxy-binary-path", "/usr/sbin/haproxy", "HAProxy binary location")
	flags.StringVar(&conf.HaproxyReadBindIp, "haproxy-ip-read-bind", "0.0.0.0", "HAProxy input bind address for read")
	flags.StringVar(&conf.HaproxyWriteBindIp, "haproxy-ip-write-bind", "0.0.0.0", "HAProxy input bind address for write")
	flags.StringVar(&conf.HaproxyAPIReadBackend, "haproxy-api-read-backend", "service_read", "HAProxy API backend name used for read")
	flags.StringVar(&conf.HaproxyAPIWriteBackend, "haproxy-api-write-backend", "service_write", "HAProxy API backend name used for write")
	flags.StringVar(&conf.HaproxyHostsIPV6, "haproxy-servers-ipv6", "", "HAProxy IPv6 bind address ")
}

func (proxy *HaproxyProxy) Init() {
	cluster := proxy.ClusterGroup
	haproxydatadir := proxy.Datadir + "/var"

	if _, err := os.Stat(haproxydatadir); os.IsNotExist(err) {
		proxy.GetProxyConfig()
		os.Symlink(proxy.Datadir+"/init/data", haproxydatadir)
	}
	//haproxysockFile := "haproxy.stats.sock"

	haproxytemplateFile := "haproxy_config.template"
	haproxyconfigFile := "haproxy.cfg"
	haproxyjsonFile := "vamp_router.json"
	haproxypidFile := "haproxy.pid"
	haproxyerrorPagesDir := "error_pages"
	//	haproxymaxWorkDirSize := 50 // this value is based on (max socket path size - md5 hash length - pre and postfixes)

	haRuntime := haproxy.Runtime{
		Binary:   cluster.Conf.HaproxyBinaryPath,
		SockFile: filepath.Join(proxy.Datadir+"/var", "/haproxy.stats.sock"),
		Port:     proxy.Port,
		Host:     proxy.Host,
	}

	haConfig := haproxy.Config{
		TemplateFile:  filepath.Join(cluster.Conf.ShareDir, haproxytemplateFile),
		ConfigFile:    filepath.Join(haproxydatadir, "/", haproxyconfigFile),
		JsonFile:      filepath.Join(haproxydatadir, "/", haproxyjsonFile),
		ErrorPagesDir: filepath.Join(haproxydatadir, "/", haproxyerrorPagesDir, "/"),
		PidFile:       filepath.Join(haproxydatadir, "/", haproxypidFile),
		//	SockFile:      filepath.Join(haproxydatadir, "/", haproxysockFile),
		SockFile:   "/tmp/haproxy" + proxy.Id + ".sock",
		ApiPort:    proxy.Port,
		StatPort:   strconv.Itoa(proxy.ClusterGroup.Conf.HaproxyStatPort),
		Host:       proxy.Host,
		WorkingDir: filepath.Join(haproxydatadir + "/"),
	}

	cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy loading haproxy config at %s", haproxydatadir)
	err := haConfig.GetConfigFromDisk()
	if err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy did not find an haproxy config...initializing new config")
		haConfig.InitializeConfig()
	}
	few := haproxy.Frontend{Name: "my_write_frontend", Mode: "tcp", DefaultBackend: cluster.Conf.HaproxyAPIWriteBackend, BindPort: cluster.Conf.HaproxyWritePort, BindIp: cluster.Conf.HaproxyWriteBindIp}
	if err := haConfig.AddFrontend(&few); err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "Failed to add frontend write ")
	} else {
		if err := haConfig.AddFrontend(&few); err != nil {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "HAProxy should return nil on already existing frontend")
		}

	}
	if result, _ := haConfig.GetFrontend("my_write_frontend"); result.Name != "my_write_frontend" {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "HAProxy failed to add frontend write")
	}
	bew := haproxy.Backend{Name: cluster.Conf.HaproxyAPIWriteBackend, Mode: "tcp"}
	haConfig.AddBackend(&bew)

	if _, err := haConfig.GetServer(cluster.Conf.HaproxyAPIWriteBackend, "leader"); err != nil {
		// log.Printf("No leader")
	} else {
		// log.Printf("Found exiting leader removing")
	}

	if cluster.GetMaster() != nil {

		p, _ := strconv.Atoi(cluster.GetMaster().Port)
		s := haproxy.ServerDetail{Name: "leader", Host: cluster.GetMaster().Host, Port: p, Weight: 100, MaxConn: 2000, Check: true, CheckInterval: 1000}
		if err = haConfig.AddServer(cluster.Conf.HaproxyAPIWriteBackend, &s); err != nil {
			//	log.Printf("Failed to add server to service_write ")
		}
	} else {
		s := haproxy.ServerDetail{Name: "leader", Host: "unknown", Port: 3306, Weight: 100, MaxConn: 2000, Check: true, CheckInterval: 1000}
		if err = haConfig.AddServer(cluster.Conf.HaproxyAPIWriteBackend, &s); err != nil {
			//	log.Printf("Failed to add server to service_write ")
		}
	}
	fer := haproxy.Frontend{Name: "my_read_frontend", Mode: "tcp", DefaultBackend: cluster.Conf.HaproxyAPIReadBackend, BindPort: cluster.Conf.HaproxyReadPort, BindIp: cluster.Conf.HaproxyReadBindIp}
	if err := haConfig.AddFrontend(&fer); err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "HAProxy failed to add frontend read")
	} else {
		if err := haConfig.AddFrontend(&fer); err != nil {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "HAProxy should return nil on already existing frontend")
		}
	}
	if result, _ := haConfig.GetFrontend("my_read_frontend"); result.Name != "my_read_frontend" {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "HAProxy failed to get frontend")
	}
	/* End add front end */

	ber := haproxy.Backend{Name: cluster.Conf.HaproxyAPIReadBackend, Mode: "tcp"}
	if err := haConfig.AddBackend(&ber); err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "HAProxy failed to add backend for "+cluster.Conf.HaproxyAPIReadBackend)
	}

	//var checksum64 string
	//	crcHost := crc64.MakeTable(crc64.ECMA)
	for _, server := range cluster.Servers {
		if server.IsMaintenance == false {
			p, _ := strconv.Atoi(server.Port)
			//		checksum64 := fmt.Sprintf("%d", crc64.Checksum([]byte(server.Host+":"+server.Port), crcHost))
			s := haproxy.ServerDetail{Name: server.Id, Host: server.Host, Port: p, Weight: 100, MaxConn: 2000, Check: true, CheckInterval: 1000}
			if err := haConfig.AddServer(cluster.Conf.HaproxyAPIReadBackend, &s); err != nil {
				cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "Failed to add server in HAProxy for "+cluster.Conf.HaproxyAPIReadBackend)
			}
		}
	}

	err = haConfig.Render()
	if err != nil {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "Could not create haproxy config %s", err)
	}
	if cluster.Conf.HaproxyMode == "standby" {
		if err := haRuntime.SetPid(haConfig.PidFile); err != nil {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy set pid %s", err)
		} else {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy reload config on pid %s", haConfig.PidFile)
		}

		err = haRuntime.Reload(&haConfig)
		if err != nil {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "Can't reload haproxy config %s", err)
		}
	}
}

func (proxy *HaproxyProxy) Refresh() error {
	cluster := proxy.ClusterGroup
	// if proxy.ClusterGroup.Conf.HaproxyStatHttp {

	/*
		url := "http://" + proxy.Host + ":" + proxy.Port + "/stats;csv"
		client := &http.Client{
			Timeout: time.Duration(2 * time.Second),
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			cluster.SetState("ERR00052", state.State{ErrType: "WARNING", ErrDesc: fmt.Sprintf(clusterError["ERR00052"], err), ErrFrom: "MON"})
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			cluster.SetState("ERR00052", state.State{ErrType: "WARNING", ErrDesc: fmt.Sprintf(clusterError["ERR00052"], err), ErrFrom: "MON"})
			return err
		}
		defer resp.Body.Close()
		reader := csv.NewReader(resp.Body)

	*/
	//tcpAddr, err := net.ResolveTCPAddr("tcp4", proxy.Host+":"+proxy.Port)
	//cluster.LogModulePrintf(cluster.Conf.Verbose,config.ConstLogModHAProxy,config.LvlErr, "haproxy entering  refresh: ")

	haproxydatadir := proxy.Datadir + "/var"
	haproxysockFile := "haproxy.stats.sock"

	haRuntime := haproxy.Runtime{
		Binary:   cluster.Conf.HaproxyBinaryPath,
		SockFile: filepath.Join(haproxydatadir, "/", haproxysockFile),
		Port:     proxy.Port,
		Host:     proxy.Host,
	}

	backend_ip_host := make(map[string]string)
	if proxy.HasDNS() {
		// When using FQDN map server state host->IP to locate in show stats where it's only IPs
		cmd := "show servers state"

		showleaderstate, err := haRuntime.ApiCmd(cmd)
		if err != nil {
			cluster.SetState("ERR00052", state.State{ErrType: "WARNING", ErrDesc: fmt.Sprintf(clusterError["ERR00052"], err), ErrFrom: "MON"})
			return err
		}

		// API return a first row with return code make it as comment
		showleaderstate = "# " + showleaderstate

		// API return space sparator conveting to csv
		showleaderstate = strings.Replace(showleaderstate, " ", ",", -1)
		if cluster.Conf.HaproxyDebug {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "haproxy show servers state response :%s", showleaderstate)
		}
		showleaderstatereader := io.NopCloser(bytes.NewReader([]byte(showleaderstate)))

		defer showleaderstatereader.Close()
		reader := csv.NewReader(showleaderstatereader)
		reader.Comment = '#'
		for {
			line, error := reader.Read()
			if error == io.EOF {
				break
			} else if error != nil {
				cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "Could not read csv from haproxy response")
				return err
			}
			if len(line) > 17 {
				if cluster.Conf.HaproxyDebug {
					cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy adding IP map %s %s", line[4], line[17])
				}
				backend_ip_host[line[4]] = line[17]
			}
		}

	}

	if proxy.Version == "" {
		vstring, err := haRuntime.GetVersion()
		if err == nil {
			if vstring != "" {
				proxy.Version = vstring
			}
		}
	}

	result, err := haRuntime.ApiCmd("show stat")
	if err != nil {
		cluster.SetState("ERR00052", state.State{ErrType: "WARNING", ErrDesc: fmt.Sprintf(clusterError["ERR00052"], err), ErrFrom: "MON"})
		return err
	}
	if cluster.Conf.HaproxyDebug {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy show stat result: %s", result)
	}
	r := io.NopCloser(bytes.NewReader([]byte(result)))
	defer r.Close()
	reader := csv.NewReader(r)

	proxy.BackendsWrite = nil
	proxy.BackendsRead = nil
	foundMasterInStat := false
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "Could not read csv from haproxy response")
			return err
		}
		if len(line) < 73 {
			cluster.SetState("WARN0078", state.State{ErrType: "WARNING", ErrDesc: fmt.Sprintf(clusterError["WARN0078"], err), ErrFrom: "MON"})
			return errors.New(clusterError["WARN0078"])
		}
		if strings.Contains(strings.ToLower(line[0]), cluster.Conf.HaproxyAPIWriteBackend) {
			host := line[73]
			if proxy.HasDNS() {
				// After provisioning the stats may arrive with IP:Port while sometime not
				host = strings.Split(line[73], ":")[0]
				host = backend_ip_host[host]
			}

			srv := cluster.GetServerFromURL(host)
			// if cluster.Conf.HaproxyDebug {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy stat lookup writer: host %s translated to %s", line[73], host)
			// }
			if srv != nil {
				foundMasterInStat = true
				proxy.BackendsWrite = append(proxy.BackendsWrite, Backend{
					Host:           srv.Host,
					Port:           srv.Port,
					Status:         srv.State,
					PrxName:        line[73],
					PrxStatus:      line[17],
					PrxConnections: line[5],
					PrxByteIn:      line[8],
					PrxByteOut:     line[9],
					PrxLatency:     line[61], //ttime: average session time in ms over the 1024 last requests
				})
				if !srv.IsMaster() {
					master := cluster.GetMaster()
					if master != nil {
						cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "Detecting wrong master server in haproxy %s fixing it to master %s %s", proxy.Host+":"+proxy.Port, master.Host, master.Port)
						msg, err := haRuntime.SetMaster(master.Host, master.Port)
						if err != nil {
							cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModProxy, config.LvlErr, "%s: %s (master: %s)", proxy.Host+":"+proxy.Port, msg, master.Host+":"+master.Port)
						} else {
							cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModProxy, config.LvlDbg, "%s: %s (master: %s)", proxy.Host+":"+proxy.Port, msg, master.Host+":"+master.Port)
						}
					}
				}
			}
		}
		if strings.Contains(strings.ToLower(line[0]), cluster.Conf.HaproxyAPIReadBackend) {
			host := line[73]
			if proxy.HasDNS() {
				// After provisioning the stats may arrive with  IP:Port while sometime not
				host = strings.Split(line[73], ":")[0]
				host = backend_ip_host[host]
			}
			srv := cluster.GetServerFromURL(host)
			if cluster.Conf.HaproxyDebug {
				cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy stat lookup reader: host %s translated to %s", line[73], host)
			}
			if srv != nil {

				proxy.BackendsRead = append(proxy.BackendsRead, Backend{
					Host:           srv.Host,
					Port:           srv.Port,
					Status:         srv.State,
					PrxName:        line[73],
					PrxStatus:      line[17],
					PrxConnections: line[5],
					PrxByteIn:      line[8],
					PrxByteOut:     line[9],
					PrxLatency:     line[61],
				})
				if (srv.State == stateSlaveErr || srv.State == stateRelayErr || srv.State == stateSlaveLate || srv.State == stateRelayLate || srv.IsIgnored()) && line[17] == "UP" || srv.State == stateWsrepLate || srv.State == stateWsrepDonor {
					cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy detecting broken replication and UP state in haproxy %s drain  server %s", proxy.Host+":"+proxy.Port, srv.URL)
					msg, err := haRuntime.SetDrain(srv.Id, cluster.Conf.HaproxyAPIReadBackend)
					if err != nil {
						cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModProxy, config.LvlErr, "%s: %s (server: %s)", proxy.Host+":"+proxy.Port, msg, srv.Host+":"+srv.Port)
					} else {
						cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModProxy, config.LvlInfo, "%s: %s (server: %s)", proxy.Host+":"+proxy.Port, msg, srv.Host+":"+srv.Port)
					}
				}
				if (srv.State == stateSlave || srv.State == stateRelay || (srv.State == stateWsrep && !srv.IsLeader())) && line[17] == "DRAIN" && !srv.IsIgnored() {
					cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy valid replication and DRAIN state in haproxy %s enable traffic on server %s", proxy.Host+":"+proxy.Port, srv.URL)
					msg, err := haRuntime.SetReady(srv.Id, cluster.Conf.HaproxyAPIReadBackend)
					if err != nil {
						cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModProxy, config.LvlErr, "%s: %s (server: %s)", proxy.Host+":"+proxy.Port, msg, srv.Host+":"+srv.Port)
					} else {
						cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModProxy, config.LvlDbg, "%s: %s (server: %s)", proxy.Host+":"+proxy.Port, msg, srv.Host+":"+srv.Port)
					}
				}
				if srv.IsMaster() {
					if !cluster.Configurator.HasProxyReadLeader() && line[17] == "UP" {
						cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy master is not configured as reader but state is UP in haproxy %s for server %s", proxy.Host+":"+proxy.Port, srv.URL)
						msg, err := haRuntime.SetDrain(srv.Id, cluster.Conf.HaproxyAPIReadBackend)
						if err != nil {
							cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModProxy, config.LvlErr, "%s: %s (server: %s)", proxy.Host+":"+proxy.Port, msg, srv.Host+":"+srv.Port)
						} else {
							cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModProxy, config.LvlDbg, "%s: %s (server: %s)", proxy.Host+":"+proxy.Port, msg, srv.Host+":"+srv.Port)
						}
					}
					if cluster.Configurator.HasProxyReadLeader() && line[17] == "DRAIN" {
						cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy master is configured as reader but state is DRAIN in haproxy %s for server %s", proxy.Host+":"+proxy.Port, srv.URL)
						msg, err := haRuntime.SetReady(srv.Id, cluster.Conf.HaproxyAPIReadBackend)
						if err != nil {
							cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModProxy, config.LvlErr, "%s: %s (server: %s)", proxy.Host+":"+proxy.Port, msg, srv.Host+":"+srv.Port)
						} else {
							cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModProxy, config.LvlDbg, "%s: %s (server: %s)", proxy.Host+":"+proxy.Port, msg, srv.Host+":"+srv.Port)
						}
					}

				}
				if srv.IsMaintenance && line[17] == "UP" {
					cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy detecting server %s in maintenance but proxy %s reports UP  ", srv.URL, proxy.Host+":"+proxy.Port)
					proxy.SetMaintenance(srv)
				}
				if !srv.IsMaintenance && line[17] == "MAINT" {
					cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy detecting server %s UP but proxy %s reports in maintenance ", srv.URL, proxy.Host+":"+proxy.Port)
					proxy.SetMaintenance(srv)
				}
			}
		}
	}
	if !foundMasterInStat {
		master := cluster.GetMaster()
		if master != nil && master.IsLeader() {
			res, err := haRuntime.SetMaster(master.Host, master.Port)
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy has leader in cluster but not in %s fixing it to master %s return %s", proxy.Host+":"+proxy.Port, master.URL, res)
			if err != nil {
				cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "HAProxy cannot add leader %s in cluster but not in %s : %s", master.URL, proxy.Host+":"+proxy.Port, err)
			}
		}
	}
	return nil
}

func (cluster *Cluster) setMaintenanceHaproxy(pr *Proxy, server *ServerMonitor) {
	pr.SetMaintenance(server)
}

func (proxy *Proxy) SetMaintenance(server *ServerMonitor) {
	cluster := proxy.ClusterGroup
	if !cluster.Conf.HaproxyOn {
		return
	}
	if cluster.Conf.HaproxyMode == "standby" {
		proxy.Init()
		return
	}
	//if cluster.Conf.HaproxyDebug {
	cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy set maintenance for server %s ", server.URL)
	//}
	haRuntime := haproxy.Runtime{
		Binary:   cluster.Conf.HaproxyBinaryPath,
		SockFile: filepath.Join(proxy.Datadir+"/var", "/haproxy.stats.sock"),
		Port:     proxy.Port,
		Host:     proxy.Host,
	}

	if server.IsMaintenance {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy set server %s/%s state maint ", server.Id, cluster.Conf.HaproxyAPIReadBackend)
		res, err := haRuntime.SetMaintenance(server.Id, cluster.Conf.HaproxyAPIReadBackend)
		if err != nil {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "HAProxy can not set maintenance %s backend %s : %s", server.URL, cluster.Conf.HaproxyAPIReadBackend, err)
		}
		if cluster.Conf.HaproxyDebug {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy set maintenance %s backend %s result: %s", server.URL, cluster.Conf.HaproxyAPIReadBackend, res)
		}
	} else {
		cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy set server %s/%s state ready ", server.Id, cluster.Conf.HaproxyAPIReadBackend)
		res, err := haRuntime.SetReady(server.Id, cluster.Conf.HaproxyAPIReadBackend)
		if err != nil {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "HAProxy can not set ready %s backend %s : %s", server.URL, cluster.Conf.HaproxyAPIReadBackend, err)
		}
		if cluster.Conf.HaproxyDebug {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy set ready %s backend %s result: %s", server.URL, cluster.Conf.HaproxyAPIReadBackend, res)
		}

	}
	if server.IsMaster() {
		if server.IsMaintenance {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy set maintenance for server %s ", server.URL)

			res, err := haRuntime.SetMaintenance("leader", cluster.Conf.HaproxyAPIWriteBackend)
			if err != nil {
				cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "HAProxy can not set maintenance %s backend %s : %s", server.URL, cluster.Conf.HaproxyAPIReadBackend, err)
			}
			if cluster.Conf.HaproxyDebug {
				cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy set maintenance result: %s", res)
			}

		} else {
			cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy set ready for server %s ", server.URL)

			res, err := haRuntime.SetReady("leader", cluster.Conf.HaproxyAPIWriteBackend)
			if err != nil {
				cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlErr, "HAProxy can not set ready %s backend %s : %s", server.URL, cluster.Conf.HaproxyAPIWriteBackend, err)
			}
			if cluster.Conf.HaproxyDebug {
				cluster.LogModulePrintf(cluster.Conf.Verbose, config.ConstLogModHAProxy, config.LvlInfo, "HAProxy set ready %s backend %s result: %s", server.URL, cluster.Conf.HaproxyAPIWriteBackend, res)
			}
		}
	}
}

func (proxy *HaproxyProxy) Failover() {
	cluster := proxy.ClusterGroup
	if cluster.Conf.HaproxyMode == "runtimeapi" {
		proxy.Refresh()
	}
	if cluster.Conf.HaproxyMode == "standby" {
		proxy.Init()
	}
}

func (proxy *HaproxyProxy) BackendsStateChange() {
	proxy.Refresh()
}

func (proxy *HaproxyProxy) CertificatesReload() error {
	return nil
}
