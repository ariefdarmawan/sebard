package modules

import (
	"os"
	"strings"

	"time"

	"github.com/eaciit/appconfig"
	"github.com/eaciit/appserver"
	"github.com/eaciit/toolkit"
)

type SebarNode struct {
	Listener *appserver.Server
	Client   *appserver.Client

	Done chan bool

	Config *SebarConfig

	log       *toolkit.LogEngine
	closeNode chan bool

	events map[string]func(toolkit.M) *toolkit.Result
}

func (sn *SebarNode) Close() {
	if sn.Client != nil {
		sn.Client.Close()
	}

	if sn.Listener != nil {
		sn.Log().Info("Stop listener on " + sn.Config.HostAddress())
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
	sn.Listener = listener

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

func (sn *SebarNode) SendCloseSignal() {
	sn.closeNode <- true
	sn.Close()
}

func (sn *SebarNode) Wait() {
	if sn.closeNode == nil {
		sn.closeNode = make(chan bool)
	}
	for {
		select {
		case b := <-sn.closeNode:
			if b == true {
				return
			}

		case <-time.After(sn.Config.HealthCheckRate):
			sn.Event("healthcheck", nil)
			//-- do nothing
		}
	}
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
	sn.Config.HealthCheckRate = time.Duration(toolkit.ToInt(cfg.GetDefault("healthcheckrate",
		float64(int(1*time.Second))).(float64), toolkit.RoundingAuto))

	return nil
}

func (sn *SebarNode) AddEvent(eventname string, fnevent func(toolkit.M) *toolkit.Result) {
	if sn.events == nil {
		sn.events = make(map[string]func(toolkit.M) *toolkit.Result)
	}

	eventname = strings.ToLower(strings.Trim(eventname, " "))
	sn.events[eventname] = fnevent
}

func (sn *SebarNode) RemoveEvent(eventname string) {
	if sn.events == nil {
		return
	}

	eventname = strings.ToLower(strings.Trim(eventname, " "))
	delete(sn.events, eventname)
}

func (sn *SebarNode) Event(eventname string, in toolkit.M) *toolkit.Result {
	if sn.events == nil {
		return toolkit.NewResult().SetErrorTxt("Event " + eventname + " is not exists")
	}

	if in == nil {
		in = toolkit.M{}
	}
	in.Set("server", sn)

	eventnameLower := strings.ToLower(eventname)
	if event, b := sn.events[eventnameLower]; b == false {
		return toolkit.NewResult().SetErrorTxt("Event " + eventname + " is not exists")
	} else {
		ret := event(in)
		if ret == nil {
			ret = toolkit.NewResult()
		}
	}

	return nil
}
