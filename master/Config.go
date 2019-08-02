package master

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ApiPort         int      `json:"apiPort"`
	ApiReadTimeout  int      `json:"ApiReadTimeout"`
	ApiWriteTimeout int      `json:"ApiWriteTimeout"`
	EtcdEndpoints   []string `json:"etcdEndPoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	WebRoot         string   `json:"webroot"`
	MongoUri        string   `json:"mongoUri"`
	MongoTimeout    int      `json:"mongoTimeout"`
}

var (
	G_config *Config
)

func InitConfig(filename string) (err error) {
	var (
		content []byte
		conf    Config
	)

	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}

	G_config = &conf
	return
}
