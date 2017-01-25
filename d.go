package main

import (
	"flag"

	"os"

	"github.com/eaciit/toolkit"

	. "eaciit/sebard/modules"
)

var (
	cfgFlag      = flag.String("c", "", "-c='path to config file' | config file")
	hostFlag     = flag.String("h", "", "-h=ip | ip")
	portFlag     = flag.Int("p", 6789, "-p=port | port number")
	joinFlag     = flag.String("j", "", "-j=joinaddress | address of join")
	dataPathFlag = flag.String("dp", "", "-dp=path | identify data path")

	sebarCfg = new(SebarConfig)
	sn       = new(SebarNode)
)

/*
Usage:
sebard                              //run sebard on local
sebard -port=9090                   //run sebard on port 9090
sebard -join=192.168.0.110:8080     //join sebar 110:8080

curl http://192.168.0.111:8080/token?userid=xxx&pass=yyy
curl -d anydata http://192.168.0.111:8080/set?token=111&key=d1
curl http://192.168.0.111:8080/get?token=111&key=d1
curl -d anydata http://192.168.0.111:8080/run?token=1111&service=fs.save

data key is build as:
owner.cluster.table.id


run service
- service need to be registered with pattern: servicename.servicemethod
- call the service /run?service=svcid&token=111
*/

func main() {
	defer sn.Close()
	log, _ := toolkit.NewLog(true, false, "", "", "")
	sn.SetLog(log)
	flag.Parse()

	sebarCfg = &SebarConfig{
		Port: 6789,
	}
	sn.Config = sebarCfg

	if *cfgFlag != "" {
		if eread := sn.ReadConfig(*cfgFlag); eread != nil {
			log.Error(eread.Error())
			os.Exit(100)
		} else {
			log.Info("Successfully reading config file " + *cfgFlag)
		}
	}

	if *hostFlag != "" {
		sebarCfg.Host = *hostFlag
	}

	if *portFlag != 6789 {
		sebarCfg.Port = *portFlag
	}

	if *joinFlag != "" {
		sebarCfg.Cluster, sebarCfg.ClusterUserID, sebarCfg.ClusterSecret = ParseUrlConnection(*joinFlag)
		toolkit.Println("Cluster: " + sebarCfg.Cluster + " User:" + sebarCfg.ClusterUserID + " Password:" + sebarCfg.ClusterSecret)
	}

	if es := sn.Start(); es != nil {
		return
	}
	sn.Wait()
}

func sendHearthbeat() error {
	return nil
}
