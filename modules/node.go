package modules

import (
	"os"
	"strings"

	"github.com/eaciit/appconfig"
	"github.com/eaciit/appserver"
	"github.com/eaciit/toolkit"
)

type SebarConfig struct {
	Host                         string
	Port                         int
	Cluster                      string
	ClusterUserID, ClusterSecret string
	DataPath                     string

	User   string
	Secret string

	AuthServer string
}

type SebarNode struct {
	Listener *appserver.Server
	Client   *appserver.Client

	Done chan bool

	Config *SebarConfig
	log    *toolkit.LogEngine
}

func (sn *SebarNode) Close() {
	if sn.Client != nil {
		sn.Client.Close()
	}

	if sn.Listener != nil {
		sn.Listener.Log.Info(toolkit.Sprintf("Closing node"))
		sn.Listener.Stop()
	}

	sn.Log().Info("Server " + sn.Config.HostAddress() + " is stopped")
}

func (sn *SebarNode) SetLog(l *toolkit.LogEngine) {
	sn.log = l
}

func (sn *SebarNode) Log() *toolkit.LogEngine {
	if sn.log == nil {
		l, _ := toolkit.NewLog(true, false, "", "", "")
		sn.log = l
	}
	return sn.log
}

func (sc *SebarConfig) HostAddress() string {
	return sc.Host + ":" + toolkit.ToString(sc.Port)
}

/*
Convert arief:darmawan@host:3000 to:
host = host:3000
user = arief
pass = darmawan
*/
func ParseUrlConnection(url string) (host, userid, password string) {
	credhosts := strings.Split(url, "@")
	if len(credhosts) == 1 {
		host = credhosts[0]
	} else {
		creds := strings.Split(credhosts[0], ":")
		host = credhosts[1]
		userid = creds[0]
		if len(creds) > 1 {
			password = creds[1]
		}
	}
	return
}

func (sn *SebarNode) Start() error {
	listener := new(appserver.Server)
	listener.AllowMultiLogin = true
	listener.Log = sn.Log()
	log := sn.Log()

	sebarCfg := sn.Config

	autoSecretCode := false
	if sebarCfg.Secret == "" {
		sebarCfg.Secret = toolkit.RandomString(32)
		autoSecretCode = true
	}
	listener.SetSecret(sebarCfg.Secret)

	if sebarCfg.Cluster != "" && sebarCfg.Cluster != sebarCfg.HostAddress() {
		if econnect := joinCluster(sn); econnect != nil {
			log.Error(econnect.Error())
			os.Exit(100)
		} else {
			log.Info("Successfully join cluster " + sebarCfg.Cluster)
		}
	} else if sebarCfg.Cluster == "" {
		sebarCfg.Cluster = sebarCfg.HostAddress()
	}

	if autoSecretCode {
		log.Info(toolkit.Sprintf("Preparing server to run on [%s], secret: %s", sebarCfg.HostAddress(), sebarCfg.Secret))
	} else {
		log.Info(toolkit.Sprintf("Preparing server to run on [%s]", sebarCfg.HostAddress()))
	}

	listener.AddUser("root", sebarCfg.Secret)
	if estart := listener.Start(sebarCfg.HostAddress()); estart != nil {
		log.Error(estart.Error())
		os.Exit(100)
	}

	//sn.Listener = listener
	/*
		for {
			select {
			case b := <-sn.Done:
				if b {
					sn.Close()
					os.Exit(100)
				}

			case <-time.After(1 * time.Second):
				//-- do nothing
				//log.Info(toolkit.Sprintf("Time tock: %v", time.Now()))
				if ehealth := sendHearthbeat(); ehealth != nil {
					log.Error("Hearthbeat: " + ehealth.Error())
					os.Exit(100)
				}
			}
		}
	*/

	return nil
}

func (sn *SebarNode) Wait() {
}

func joinCluster(sn *SebarNode) error {
	sc := sn.Config
	client := new(appserver.Client)
	if ejoin := client.Connect(sc.Cluster, sc.ClusterSecret, sc.ClusterUserID); ejoin != nil {
		return ejoin
	}
	sn.Client = client
	return nil
}

func (sn *SebarNode) ReadConfig(cfgfile string) error {
	cfg := new(appconfig.Config)
	cfg.SetConfigFile(cfgfile)

	if sn.Config == nil {
		sn.Config = new(SebarConfig)
	}

	sn.Config.Host = cfg.GetDefault("host", sn.Config.Host).(string)
	sn.Config.Port = toolkit.ToInt(cfg.GetDefault("port", sn.Config.Port).(float64), toolkit.RoundingAuto)
	sn.Config.User = cfg.GetDefault("user", sn.Config.User).(string)
	sn.Config.Secret = cfg.GetDefault("secret", sn.Config.Secret).(string)
	sn.Config.Cluster = cfg.GetDefault("cluster", sn.Config.Cluster).(string)
	sn.Config.ClusterUserID = cfg.GetDefault("clusteruserid", sn.Config.ClusterUserID).(string)
	sn.Config.ClusterSecret = cfg.GetDefault("clustersecret", sn.Config.ClusterSecret).(string)

	return nil
}
