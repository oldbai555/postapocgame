/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc: 副本服务器配置
**/

package conf

import (
	"os"
	"path"
	"postapocgame/server/pkg/tool"
)

type DungeonSrvConfSt struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

func LoadDungeonSrvConf(confPath string) (*DungeonSrvConfSt, error) {
	if confPath == "" {
		confPath = path.Join(tool.GetCurDir(), "dungeonsrv.json")
	}
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	var conf DungeonSrvConfSt
	err = tool.JsonUnmarshal(bytes, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
