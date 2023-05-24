package cfg

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	SERVER_ERROR     = "{SE}"
	VERIFY_FAILED    = "{VF}"
	VERIFY_SUCCESSED = "{VS}"
	// length of SERVER_ERROR\VERIFY_FAILED\VERIFY_SUCCESSED must be 4
)

type Tunnel struct {
	Domain string `yaml:"domain"`
	Proto  string `yaml:"proto"`
}

type Config struct {
	ServerAddr    string   `yaml:"server-addr"`
	Token         string   `yaml:"token"`
	MaxConn       int      `yaml:"max-conn"`
	DomainAsProto bool     `yaml:"domain-as-proto"`
	Tunnel        []Tunnel `yaml:"tunnel"`
}

func Init(path string, conf interface{}) {
	c, err := os.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}
	err = yaml.UnmarshalStrict(c, conf)
	if err != nil {
		log.Fatalln(err)
	}
}
