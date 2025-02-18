package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	jwt "github.com/golang-jwt/jwt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/signal18/replication-manager/cluster"
	"github.com/signal18/replication-manager/config"
	v3 "github.com/signal18/replication-manager/repmanv3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	log "github.com/sirupsen/logrus"
)

type Repmanv3Config struct {
	Listen Repmanv3ListenAddress
	TLS    Repmanv3TLS
}

type Repmanv3ListenAddress struct {
	Address string
	Port    string
}

func (l *Repmanv3ListenAddress) AddressWithPort() string {
	var str strings.Builder
	str.WriteString(l.Address)
	str.WriteString(`:`)
	str.WriteString(l.Port)
	return str.String()
}

type Repmanv3TLS struct {
	Enabled            bool
	CertificatePath    string
	CertificateKeyPath string
	SelfSigned         bool
}

func (s *ReplicationManager) SetV3Config(config Repmanv3Config) {
	s.v3Config = config
}

func (s *ReplicationManager) StartServerV3(debug bool, router *mux.Router) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := net.Listen("tcp", s.v3Config.Listen.AddressWithPort())
	if err != nil {
		return err
	}

	var serverOpts []grpc.ServerOption
	var dopts []grpc.DialOption
	var tlsConfig *tls.Config
	serverOpts, dopts, tlsConfig, err = s.getCredentials()
	if err != nil {
		return err
	}

	if debug {
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				s.unaryInterceptor,
				// grpc_zap.UnaryServerInterceptor(s.log),
			),
		))
		serverOpts = append(serverOpts, grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				s.streamInterceptor,
				// grpc_zap.StreamServerInterceptor(s.log),
			),
		))
	} else {
		serverOpts = append(serverOpts,
			grpc.UnaryInterceptor(s.unaryInterceptor),
			grpc.StreamInterceptor(s.streamInterceptor),
		)
	}

	s.grpcServer = grpc.NewServer(serverOpts...)
	v3.RegisterClusterPublicServiceServer(s.grpcServer, s)
	v3.RegisterClusterServiceServer(s.grpcServer, s)

	/* Bootstrap the Muxed connection */
	httpmux := http.NewServeMux()
	gwmux := runtime.NewServeMux()

	err = v3.RegisterClusterPublicServiceHandlerFromEndpoint(ctx,
		gwmux,
		s.v3Config.Listen.AddressWithPort(),
		dopts,
	)
	if err != nil {
		return fmt.Errorf("could not register service ClusterPublicService: %w", err)
	}

	err = v3.RegisterClusterServiceHandlerFromEndpoint(ctx,
		gwmux,
		s.v3Config.Listen.AddressWithPort(),
		dopts,
	)

	if err != nil {
		return fmt.Errorf("could not register service ClusterService: %w", err)
	}

	httpmux.Handle("/", gwmux)

	srv := &http.Server{
		Addr: s.v3Config.Listen.AddressWithPort(),
		Handler: grpcHandlerFunc(s,
			httpmux,
			handlers.CORS(
				handlers.AllowCredentials(),
				handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
				handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
				handlers.AllowedOriginValidator(s.handleOriginValidator),
			)(router),
		),

		// ErrorLog: zap.NewStdLog(s.log),
	}

	s.grpcWrapped = grpcweb.WrapServer(s.grpcServer, grpcweb.WithOriginFunc(func(origin string) bool {
		return true
	}))

	// s.V3Up <- true
	if s.v3Config.TLS.Enabled {
		log.Info("starting multiplexed TLS HTTP/2.0 and HTTP/1.1 Gateway server: ", s.v3Config.Listen.AddressWithPort())
		srv.TLSConfig = tlsConfig
		s.IsApiListenerReady = true
		err = srv.Serve(tls.NewListener(conn, srv.TLSConfig))
	} else {
		log.Info("starting multiplexed HTTP/2.0 and HTTP/1.1 Gateway server: ", s.v3Config.Listen.AddressWithPort())
		// we need to wrap the non-tls connection inside h2c because http2 in Go enforces TLS
		srv.Handler = h2c.NewHandler(srv.Handler, &http2.Server{})
		s.IsApiListenerReady = true
		err = srv.Serve(conn)
	}

	if err != nil {
		return err
	}

	return nil
}

func (s *ReplicationManager) streamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if strings.Contains(info.FullMethod, "Public") {
		return handler(srv, stream)
	}

	// handle ACL
	log.Infof("grpc stream srv: %v", srv)
	return handler(srv, stream)
}

func (s *ReplicationManager) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// check the fullmethod if public
	if strings.Contains(info.FullMethod, "Public") {
		return handler(ctx, req)
	}

	log.Infof("grpc unary req: %v", req)

	if cMsg, ok := req.(v3.ContainsClusterMessage); ok {
		// mycluster, err := s.getClusterFromFromRequest(cMsg)
		// if err != nil {
		// 	return nil, err
		// }
		// ctx, err = s.authorize(ctx, mycluster)
		// if err != nil {
		// 	return nil, v3.NewError(codes.Unauthenticated, err).Err()
		// }

		// log.Infof("new ctx: %v", ctx)
		return handler(ctx, cMsg)
	}
	return nil, v3.NewError(codes.InvalidArgument, fmt.Errorf("no message sent with a cluster property")).Err()

	// handle ACL

}

func (s *ReplicationManager) getClusterFromFromRequest(req v3.ContainsClusterMessage) (*cluster.Cluster, error) {
	c, err := req.GetClusterMessage()
	if err != nil {
		return nil, err
	}

	if c.Name == "" {
		return nil, v3.NewErrorResource(codes.NotFound, v3.ErrClusterNotSet, "Name", c.Name).Err()
	}

	mycluster := s.getClusterByName(c.Name)
	if mycluster == nil {
		return nil, v3.NewErrorResource(codes.NotFound, v3.ErrClusterNotFound, "Name", c.Name).Err()
	}

	return mycluster, nil
}

type ContextKey string

// getClusterAndUser checks if the cluster exists and if the token has a valid user
func (s *ReplicationManager) getClusterAndUser(ctx context.Context, req v3.ContainsClusterMessage) (cluster.APIUser, *cluster.Cluster, error) {
	mycluster, err := s.getClusterFromFromRequest(req)
	if err != nil {
		return cluster.APIUser{}, nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return cluster.APIUser{}, nil, fmt.Errorf("metadata missing")
	}
	log.Info("md", md)

	auth := md.Get("authorization")
	if len(auth) == 0 {
		return cluster.APIUser{}, nil, fmt.Errorf("authorization header missing")
	}

	if len(auth[0]) > 6 && strings.ToUpper(auth[0][0:7]) == "BEARER " {
		token, err := jwt.Parse(auth[0][7:], func(token *jwt.Token) (interface{}, error) {
			vk, _ := jwt.ParseRSAPublicKeyFromPEM(verificationKey)
			return vk, nil
		})
		if err != nil {
			return cluster.APIUser{}, nil, fmt.Errorf("failed to parse jwt token: %w", err)
		}

		claims := token.Claims.(jwt.MapClaims)
		userinfo := claims["CustomUserInfo"]
		mycutinfo := userinfo.(map[string]interface{})

		user, err := mycluster.GetAPIUser(mycutinfo["Name"].(string), mycutinfo["Password"].(string))
		if err != nil {
			return cluster.APIUser{}, nil, err
		}

		return user, mycluster, nil
	}

	return cluster.APIUser{}, nil, fmt.Errorf("bearer is missing")
}

func (s *ReplicationManager) getCredentials() (opts []grpc.ServerOption, dopts []grpc.DialOption, tlsConfig *tls.Config, err error) {
	if s.v3Config.TLS.Enabled {
		// TLS for the gRPC server
		grpcCreds, err := credentials.NewServerTLSFromFile(s.v3Config.TLS.CertificatePath, s.v3Config.TLS.CertificateKeyPath)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error configuring credentials for TLS: %w", err)
		}

		opts = append(opts, grpc.Creds(grpcCreds))

		cer, err := tls.LoadX509KeyPair(s.v3Config.TLS.CertificatePath, s.v3Config.TLS.CertificateKeyPath)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("error loading certificates for TLS: %w", err)
		}
		//	log.Warning("Ici :" + s.v3Config.Listen.Address)
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cer},
			// declare that the listener supports http/2.0
			NextProtos:               []string{"h2"},
			ServerName:               s.v3Config.Listen.Address, // this is critical
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: false,
			CipherSuites: []uint16{
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
				tls.TLS_CHACHA20_POLY1305_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,

				// only supported in TLS1.2
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
			},
		}
		// in case the certificate is self-signed we must add the certificate to the TLS' known pool of CA's
		// else the local dialing will not function for the REST Gateway
		if s.v3Config.TLS.SelfSigned {
			certPEMBlock, err := os.ReadFile(s.v3Config.TLS.CertificatePath)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("could not read self-signed cert for root-ca: %w", err)
			}

			rootCa := x509.NewCertPool()
			if !rootCa.AppendCertsFromPEM(certPEMBlock) {
				return nil, nil, nil, fmt.Errorf("could not append self-signed cert for root-ca")
			}

			tlsConfig.RootCAs = rootCa
		}

		dopts = []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))}
	} else {
		dopts = []grpc.DialOption{grpc.WithInsecure()}
	}

	return opts, dopts, tlsConfig, nil
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Adapter from cockroachdb.
func grpcHandlerFunc(s *ReplicationManager, otherHandler http.Handler, legacyHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.ProtoMajor == 2 && req.Header.Get("Content-Type") == "application/grpc" {
			s.grpcServer.ServeHTTP(resp, req)
		} else {
			if s.grpcWrapped.IsAcceptableGrpcCorsRequest(req) || s.grpcWrapped.IsGrpcWebRequest(req) {
				s.grpcWrapped.ServeHTTP(resp, req)
			}

			// check if we need to serve the new API or the old one
			if strings.Contains(req.URL.Path, "/v3") {
				otherHandler.ServeHTTP(resp, req)
			} else {
				legacyHandler.ServeHTTP(resp, req)
			}
		}
	})
}

func (s *ReplicationManager) GetCluster(ctx context.Context, in *v3.Cluster) (*structpb.Struct, error) {
	user, mycluster, err := s.getClusterAndUser(ctx, in)
	if err != nil {
		return nil, err
	}

	if err = user.Granted(config.GrantClusterGrant); err != nil {
		return nil, err
	}

	// TODO: note we are not scrubbing the passwords here
	b, err := json.Marshal(mycluster)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not marshal cluster")
	}

	out := &structpb.Struct{}
	err = protojson.Unmarshal(b, out)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not unmarshal json config to struct")
	}

	return out, nil
}

// ClusterStatus is a public endpoint so it doesn't need to verify a user
func (s *ReplicationManager) ClusterStatus(ctx context.Context, in *v3.Cluster) (*v3.StatusMessage, error) {
	mycluster, err := s.getClusterFromFromRequest(in)
	if err != nil {
		return nil, err
	}

	if mycluster.GetStatus() {
		return &v3.StatusMessage{
			Alive: v3.ServiceStatus_RUNNING,
		}, nil
	}
	return &v3.StatusMessage{
		Alive: v3.ServiceStatus_ERRORS,
	}, nil

}

// MasterPhysicalBackup is a public endpoint
func (s *ReplicationManager) MasterPhysicalBackup(ctx context.Context, in *v3.Cluster) (*emptypb.Empty, error) {
	mycluster, err := s.getClusterFromFromRequest(in)
	if err != nil {
		return nil, err
	}

	m := mycluster.GetMaster()
	if m == nil {
		return nil, v3.NewErrorResource(codes.InvalidArgument, v3.ErrClusterMasterNotSet, "cluster", in.Name).Err()
	}
	_, err = m.JobBackupPhysical()
	return &emptypb.Empty{}, err
}

func (s *ReplicationManager) GetSettingsForCluster(ctx context.Context, in *v3.Cluster) (*structpb.Struct, error) {
	user, mycluster, err := s.getClusterAndUser(ctx, in)
	if err != nil {
		return nil, err
	}
	if err = user.Granted(config.GrantClusterSettings); err != nil {
		return nil, err
	}

	b, err := json.Marshal(mycluster.Conf)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not marshal config")
	}

	out := &structpb.Struct{}
	err = protojson.Unmarshal(b, out)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not unmarshal json config to struct")
	}

	return out, nil
}

func (s *ReplicationManager) SetActionForClusterSettings(ctx context.Context, in *v3.ClusterSetting) (res *emptypb.Empty, err error) {
	user, mycluster, err := s.getClusterAndUser(ctx, in)
	if err != nil {
		return nil, err
	}

	if strings.Contains(in.Action.String(), "TAG") {
		if in.TagValue == "" {
			return nil, v3.NewErrorResource(codes.InvalidArgument, v3.ErrFieldNotSet, "TagValue", "").Err()
		}
	}

	log.Printf("incoming: %v", in)

	// check if we are doing a set or switch
	if in.Action == v3.ClusterSetting_UNSPECIFIED {
		if in.Setting != nil {
			in.Action = v3.ClusterSetting_SET
		}
		if in.Switch != nil {
			in.Action = v3.ClusterSetting_SWITCH
		}
	}

	switch in.Action {
	case v3.ClusterSetting_UNSPECIFIED:
		return nil, v3.NewErrorResource(codes.InvalidArgument, v3.ErrEnumNotSet, "action", "").Err()

	case v3.ClusterSetting_DISCOVER:
		if err = user.Granted(config.GrantClusterSettings); err != nil {
			return nil, err
		}
		mycluster.ConfigDiscovery()

	case v3.ClusterSetting_RELOAD:
		if err = user.Granted(config.GrantClusterSettings); err != nil {
			return nil, err
		}
		s.InitConfig(s.Conf, true)
		mycluster.ReloadConfig(s.Confs[in.Cluster.Name])

	case v3.ClusterSetting_ADD_PROXY_TAG:
		if err = user.Granted(config.GrantProxyConfigFlag); err != nil {
			return nil, err
		}
		mycluster.AddProxyTag(in.TagValue)

	case v3.ClusterSetting_DROP_PROXY_TAG:
		if err = user.Granted(config.GrantProxyConfigFlag); err != nil {
			return nil, err
		}
		mycluster.DropProxyTag(in.TagValue)

	case v3.ClusterSetting_SET:
		if err = user.Granted(config.GrantClusterSettings); err != nil {
			return nil, err
		}
		if in.Setting.Name == v3.ClusterSetting_Setting_UNSPECIFIED {
			return nil, v3.NewErrorResource(codes.InvalidArgument, v3.ErrFieldNotSet, "setting.name", "").Err()
		}

		if in.Setting.Value == "" {
			return nil, v3.NewErrorResource(codes.InvalidArgument, v3.ErrFieldNotSet, "setting.value", "").Err()
		}

		s.setClusterSetting(mycluster, in.Setting.Name.Legacy(), in.Setting.Value)

	case v3.ClusterSetting_SWITCH:
		if err = user.Granted(config.GrantClusterSettings); err != nil {
			return nil, err
		}

		if in.Switch.Name == v3.ClusterSetting_Switch_UNSPECIFIED {
			return nil, v3.NewErrorResource(codes.InvalidArgument, v3.ErrEnumNotSet, "switch", "").Err()
		}

		s.switchClusterSettings(mycluster, in.Switch.Name.Legacy())

	case v3.ClusterSetting_APPLY_DYNAMIC_CONFIG:
		if err = user.Granted(config.GrantDBConfigFlag); err != nil {
			return nil, err
		}
		go mycluster.SetDBDynamicConfig()

	case v3.ClusterSetting_ADD_DB_TAG:
		if err = user.Granted(config.GrantDBConfigFlag); err != nil {
			return nil, err
		}
		mycluster.AddDBTag(in.TagValue)

	case v3.ClusterSetting_DROP_DB_TAG:
		if err = user.Granted(config.GrantDBConfigFlag); err != nil {
			return nil, err
		}
		mycluster.DropDBTag(in.TagValue)
	}

	return res, nil
}

func (s *ReplicationManager) PerformClusterAction(ctx context.Context, in *v3.ClusterAction) (res *emptypb.Empty, err error) {
	// WARNING: this one cannot be validated for ACL, as there is no cluster to validate against
	// special case, the clustername doesn't exist yet
	if in.Cluster.ClusterShardingName == "" {
		if in.Action == v3.ClusterAction_ADD {
			err = s.AddCluster(in.Cluster.Name, "")
			if err != nil {
				return nil, v3.NewError(codes.Unknown, err).Err()
			}
			return
		}
	}

	user, mycluster, err := s.getClusterAndUser(ctx, in)
	if err != nil {
		return nil, err
	}

	switch in.Action {
	case v3.ClusterAction_ADD:
		if err = user.Granted(config.GrantProvCluster); err != nil {
			return nil, err
		}
		err = s.AddCluster(in.Cluster.ClusterShardingName, in.Cluster.Name)
		if err != nil {
			return nil, v3.NewError(codes.Unknown, err).Err()
		}
		err = mycluster.RollingRestart()
	case v3.ClusterAction_ADDSERVER:
		switch in.Server.Type {
		case v3.ClusterAction_Server_TYPE_UNSPECIFIED:
			err = mycluster.AddSeededServer(in.Server.GetURI())
		case v3.ClusterAction_Server_PROXY:
			if in.Server.Proxy == v3.ClusterAction_Server_PROXY_UNSPECIFIED {
				return nil, v3.NewErrorResource(codes.InvalidArgument, v3.ErrEnumNotSet, "Proxy", v3.ClusterAction_Server_PROXY_UNSPECIFIED.String()).Err()
			}
			err = mycluster.AddSeededProxy(
				strings.ToLower(in.Server.Proxy.String()),
				in.Server.Host,
				fmt.Sprintf("%d", in.Server.Port), "", "")
		case v3.ClusterAction_Server_DATABASE:
			switch in.Server.Database {
			case v3.ClusterAction_Server_DATABASE_UNSPECIFIED:
				return nil, v3.NewErrorResource(codes.InvalidArgument, v3.ErrEnumNotSet, "Database", v3.ClusterAction_Server_DATABASE_UNSPECIFIED.String()).Err()
			case v3.ClusterAction_Server_MARIADB:
				mycluster.Conf.ProvDbImg = "mariadb:latest"
			case v3.ClusterAction_Server_PERCONA:
				mycluster.Conf.ProvDbImg = "percona:latest"
			case v3.ClusterAction_Server_MYSQL:
				mycluster.Conf.ProvDbImg = "mysql:latest"
				// TODO: Postgres is an option but previous code doesn't mention it
			}
			err = mycluster.AddSeededServer(in.Server.GetURI())
		}
	case v3.ClusterAction_REPLICATION_BOOTSTRAP:
		if in.Topology == v3.ClusterAction_RT_UNSPECIFIED {
			return nil, v3.NewErrorResource(codes.InvalidArgument, v3.ErrEnumNotSet, "Topology", v3.ClusterAction_RT_UNSPECIFIED.String()).Err()
		}
		mycluster.BootstrapTopology(in.Topology.Legacy())
		err = mycluster.BootstrapReplication(true)
	case v3.ClusterAction_CANCEL_ROLLING_REPROV:
		err = mycluster.CancelRollingReprov()
	case v3.ClusterAction_CANCEL_ROLLING_RESTART:
		err = mycluster.CancelRollingRestart()
	case v3.ClusterAction_CHECKSUM_ALL_TABLES:
		go mycluster.CheckAllTableChecksum()
	case v3.ClusterAction_FAILOVER:
		mycluster.MasterFailover(true)
	case v3.ClusterAction_MASTER_PHYSICAL_BACKUP:
		m := mycluster.GetMaster()
		if m == nil {
			return nil, v3.NewErrorResource(codes.InvalidArgument, v3.ErrClusterMasterNotSet, "cluster", in.Cluster.Name).Err()
		}
		_, err = m.JobBackupPhysical()
	case v3.ClusterAction_OPTIMIZE:
		mycluster.RollingOptimize()
	case v3.ClusterAction_RESET_FAILOVER_CONTROL:
		mycluster.ResetFailoverCtr()
	case v3.ClusterAction_RESET_SLA:
		mycluster.SetEmptySla()
	case v3.ClusterAction_ROLLING:
		err = mycluster.RollingRestart()
	case v3.ClusterAction_ROTATEKEYS:
		mycluster.KeyRotation()
	case v3.ClusterAction_START_TRAFFIC:
		mycluster.SetTraffic(true)
	case v3.ClusterAction_STOP_TRAFFIC:
		mycluster.SetTraffic(false)
	case v3.ClusterAction_SWITCHOVER:
		mycluster.LogModulePrintf(mycluster.Conf.Verbose, config.ConstLogModGeneral, "INFO", "API force for prefered master: %s", in.Server.GetURI())
		if mycluster.IsInHostList(in.Server.GetURI()) {
			mycluster.SetPrefMaster(in.Server.GetURI())
			mycluster.MasterFailover(false)
			return
		} else {
			return nil, v3.NewErrorResource(codes.NotFound, v3.ErrServerNotFound, "Server", in.Server.GetURI()).Err()
		}
	case v3.ClusterAction_SYSBENCH:
		go mycluster.RunSysbench()
	case v3.ClusterAction_WAITDATABASES:
		err = mycluster.WaitDatabaseCanConn()
	case v3.ClusterAction_REPLICATION_CLEANUP:
		err = mycluster.BootstrapReplicationCleanup()
	}

	if err != nil {
		mycluster.LogModulePrintf(mycluster.Conf.Verbose, config.ConstLogModGeneral, "ERROR", "API Error: %s", err)
		return nil, v3.NewError(codes.Unknown, err).Err()
	}

	return
}

func (s *ReplicationManager) RetrieveFromTopology(in *v3.TopologyRetrieval, stream v3.ClusterService_RetrieveFromTopologyServer) error {
	user, mycluster, err := s.getClusterAndUser(stream.Context(), in.Cluster)
	if err != nil {
		return err
	}

	// TODO: introduce new Grants for this type of endpoint
	if err = user.Granted(config.GrantClusterSettings); err != nil {
		return err
	}

	if in.Retrieve == v3.TopologyRetrieval_RETRIEVAL_UNSPECIFIED {
		return v3.NewErrorResource(codes.InvalidArgument, v3.ErrEnumNotSet, "retrieve", "").Err()
	}

	if in.Retrieve == v3.TopologyRetrieval_ALERTS {
		a := new(cluster.Alerts)
		a.Errors = mycluster.GetStateMachine().GetOpenErrors()
		a.Warnings = mycluster.GetStateMachine().GetOpenWarnings()

		return marshalAndSend(a, stream.Send)
	}

	if in.Retrieve == v3.TopologyRetrieval_CRASHES {
		cr := mycluster.GetCrashes()
		if cr == nil {
			return nil
		}
		data, err := json.Marshal(cr)
		if err != nil {
			return status.Error(codes.Internal, "could not marshal crashes list")
		}
		var crashes []*cluster.Crash
		err = json.Unmarshal(data, &crashes)
		if err != nil {
			return status.Error(codes.Internal, "could not unmarshal crashes list")
		}
		return marshalAndSend(crashes, stream.Send)
	}

	if in.Retrieve == v3.TopologyRetrieval_LOGS {
		var clusterlogs []string
		for _, slog := range s.tlog.Buffer {
			if strings.Contains(slog, mycluster.Name) {
				clusterlogs = append(clusterlogs, slog)
			}
		}

		return marshalAndSend(clusterlogs, stream.Send)
	}

	if in.Retrieve == v3.TopologyRetrieval_MASTER {
		m := mycluster.GetMaster()
		if m == nil {
			return v3.NewErrorResource(codes.InvalidArgument, v3.ErrClusterMasterNotSet, "cluster", in.Cluster.Name).Err()
		}

		// note we do a double marshal and unmarshal to prevent dereferencing objects
		data, err := json.Marshal(m)
		if err != nil {
			return status.Error(codes.Internal, "could not marshal master")
		}
		var srv *cluster.ServerMonitor
		srv.Pass = "XXXXXXXX"
		err = json.Unmarshal(data, &srv)
		if err != nil {
			return status.Error(codes.Internal, "could not unmarshal master")
		}

		return marshalAndSend(srv, stream.Send)
	}

	if in.Retrieve == v3.TopologyRetrieval_PROXIES {
		// note we do a double marshal and unmarshal to prevent dereferencing objects
		data, err := json.Marshal(mycluster.GetProxies())
		if err != nil {
			return status.Error(codes.Internal, "could not marshal proxy list")
		}
		var prxs []*cluster.Proxy

		err = json.Unmarshal(data, &prxs)
		if err != nil {
			return status.Error(codes.Internal, "could not unmarshal proxy list")
		}

		for _, prx := range prxs {
			if prx != nil {
				prx.Pass = "XXXXXXXX"
				if err := marshalAndSend(prx, stream.Send); err != nil {
					return err
				}
			}
		}
	}

	if in.Retrieve == v3.TopologyRetrieval_SERVERS || in.Retrieve == v3.TopologyRetrieval_SLAVES {
		// note we do a double marshal and unmarshal to prevent dereferencing objects
		var data []byte
		if in.Retrieve == v3.TopologyRetrieval_SERVERS {
			data, err = json.Marshal(mycluster.GetServers())
		}

		if in.Retrieve == v3.TopologyRetrieval_SLAVES {
			data, err = json.Marshal(mycluster.GetServers())
		}
		if err != nil {
			return status.Error(codes.Internal, "could not marshal server list")
		}
		var srvs []*cluster.ServerMonitor

		err = json.Unmarshal(data, &srvs)
		if err != nil {
			return status.Error(codes.Internal, "could not unmarshal server list")
		}

		for _, sm := range srvs {
			if sm != nil {
				sm.Pass = "XXXXXXXX"
				if err := marshalAndSend(sm, stream.Send); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func marshalAndSend(in interface{}, send func(*structpb.Struct) error) error {
	type String struct {
		String string
	}
	var data []byte
	var err error
	if s, ok := in.(string); ok {
		var str String
		str.String = s
		data, err = json.Marshal(str)
		if err != nil {
			return status.Error(codes.Internal, "could not marshal String to json")
		}
	}

	if sl, ok := in.([]string); ok {
		type Strings struct {
			Data []string
		}
		var strs Strings
		strs.Data = sl
		data, err = json.Marshal(strs)
		if err != nil {
			return status.Error(codes.Internal, "could not marshal Strings to json")
		}
	}

	if len(data) == 0 {
		data, err = json.Marshal(in)
		if err != nil {
			return status.Error(codes.Internal, "could not marshal to json")
		}
	}

	out := &structpb.Struct{}
	err = protojson.Unmarshal(data, out)
	if err != nil {
		return status.Error(codes.Internal, "could not unmarshal json to struct")
	}

	if err := send(out); err != nil {
		return err
	}

	return nil
}

func (s *ReplicationManager) GetClientCertificates(ctx context.Context, in *v3.Cluster) (res *v3.Certificate, err error) {
	user, mycluster, err := s.getClusterAndUser(ctx, in)
	if err != nil {
		return nil, err
	}

	if err = user.Granted(config.GrantClusterShowCertificates); err != nil {
		return nil, err
	}

	certs, err := mycluster.GetClientCertificates()
	if err != nil {
		return nil, err
	}

	res = &v3.Certificate{
		ClientCertificate: certs["clientCert"],
		ClientKey:         certs["clientKey"],
		Authority:         certs["caCert"],
	}

	return
}

func (s *ReplicationManager) GetBackups(in *v3.Cluster, stream v3.ClusterService_GetBackupsServer) error {
	user, mycluster, err := s.getClusterAndUser(stream.Context(), in)
	if err != nil {
		return err
	}

	if err = user.Granted(config.GrantClusterShowBackups); err != nil {
		return err
	}

	for _, backup := range mycluster.GetBackups() {
		if err := stream.Send(&backup); err != nil {
			return err
		}
	}

	return nil
}

func (s *ReplicationManager) GetTags(in *v3.Cluster, stream v3.ClusterService_GetTagsServer) error {
	user, mycluster, err := s.getClusterAndUser(stream.Context(), in)
	if err != nil {
		return err
	}

	if err = user.Granted(config.GrantClusterShowBackups); err != nil {
		return err
	}

	for _, tag := range mycluster.Configurator.GetDBModuleTags() {
		if err := stream.Send(&tag); err != nil {
			return err
		}
	}

	return nil
}

func (s *ReplicationManager) GetQueryRules(in *v3.Cluster, stream v3.ClusterService_GetQueryRulesServer) error {
	user, mycluster, err := s.getClusterAndUser(stream.Context(), in)
	if err != nil {
		return err
	}

	// TODO: introduce new Grants for this type of endpoint
	if err = user.Granted(config.GrantClusterGrant); err != nil {
		return err
	}

	for _, queryrule := range mycluster.GetQueryRules() {
		if err := marshalAndSend(queryrule, stream.Send); err != nil {
			return err
		}
	}

	return nil
}

func (s *ReplicationManager) GetSchema(in *v3.Cluster, stream v3.ClusterService_GetSchemaServer) error {
	user, mycluster, err := s.getClusterAndUser(stream.Context(), in)
	if err != nil {
		return err
	}

	if err = user.Granted(config.GrantDBShowSchema); err != nil {
		return err
	}

	m := mycluster.GetMaster()
	if m == nil {
		return v3.NewErrorResource(codes.InvalidArgument, v3.ErrClusterMasterNotSet, "cluster", in.Name).Err()
	}

	for _, table := range m.GetDictTables() {
		if err := stream.Send(table); err != nil {
			return err
		}
	}

	return nil
}
