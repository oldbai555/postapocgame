/**
 * @Author: zjj
 * @Date: 2025/11/5
 * @Desc: 游戏业务服务配置
**/

package conf

import (
	"os"
	"path"
	"postapocgame/server/pkg/tool"
)

type GameSrvConfSt struct {
	Address    string `json:"address"`
	Port       int    `json:"port"`
	Gateway    string `json:"gateway"`
	DungeonSrv string `json:"dungeonSrv"`
}

func LoadGameSrvConfConf(confPath string) (*GameSrvConfSt, error) {
	if confPath == "" {
		confPath = path.Join(tool.GetCurDir(), "gamesrv.json")
	}
	bytes, err := os.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	var conf GameSrvConfSt
	err = tool.JsonUnmarshal(bytes, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
