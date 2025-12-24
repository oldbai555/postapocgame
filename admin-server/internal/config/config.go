// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import "github.com/zeromicro/go-zero/rest"

// Config 聚合服务配置，RestConf 内嵌以支持 go-zero HTTP 配置。
type Config struct {
	rest.RestConf `json:",inline" yaml:",inline" mapstructure:",squash"`
	Database      DatabaseConf `json:"database" yaml:"database" mapstructure:"database"`
	Redis         RedisConf    `json:"redis" yaml:"redis" mapstructure:"redis"`
	JWT           JWTConf      `json:"jwt" yaml:"jwt" mapstructure:"jwt"`
	Bcrypt        BcryptConf   `json:"bcrypt" yaml:"bcrypt" mapstructure:"bcrypt"`
}

type DatabaseConf struct {
	DSN     string `json:"dsn" yaml:"dsn" mapstructure:"dsn"`
	MaxOpen int    `json:"maxOpen" yaml:"maxOpen" mapstructure:"maxOpen"`
	MaxIdle int    `json:"maxIdle" yaml:"maxIdle" mapstructure:"maxIdle"`
}

type RedisConf struct {
	Address     string `json:"address" yaml:"address" mapstructure:"address"`
	Password    string `json:"password" yaml:"password" mapstructure:"password"`
	DB          int    `json:"db" yaml:"db" mapstructure:"db"`
	Timeout     int    `json:"timeout" yaml:"timeout" mapstructure:"timeout"`             // 连接超时（秒），默认 5
	DialTimeout int    `json:"dialTimeout" yaml:"dialTimeout" mapstructure:"dialTimeout"` // 拨号超时（秒），默认 5
}

type JWTConf struct {
	AccessSecret  string `json:"accessSecret" yaml:"accessSecret" mapstructure:"accessSecret"`
	RefreshSecret string `json:"refreshSecret" yaml:"refreshSecret" mapstructure:"refreshSecret"`
	AccessExpire  int64  `json:"accessExpire" yaml:"accessExpire" mapstructure:"accessExpire"`
	RefreshExpire int64  `json:"refreshExpire" yaml:"refreshExpire" mapstructure:"refreshExpire"`
	Issuer        string `json:"issuer" yaml:"issuer" mapstructure:"issuer"`
}

type BcryptConf struct {
	Cost int `json:"cost" yaml:"cost" mapstructure:"cost"`
}
