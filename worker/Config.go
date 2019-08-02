package worker

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	EtcdEndpoints       []string `json:"etcdEndPoints"`
	EtcdDialTimeout     int      `json:"etcdDialTimeout"`
	MongoUri            string   `json:"mongoUri"`
	MongoTimeout        int      `json:"mongoTimeout"`
	JobLogBatchSize     int      `json:"jobLogBatchSize"`
	JobLogCommitTimeout int      `json:"jobLogCommitTimeout"`
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
