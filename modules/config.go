package modules

import "time"

type SebarConfig struct {
	Host                         string
	Port                         int
	Cluster                      string
	ClusterUserID, ClusterSecret string
	DataPath                     string

	CheckRate time.Duration
	User      string
	Secret    string

	AuthServer string
}
