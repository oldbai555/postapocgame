package gateway

import "postapocgame/server/service/gameserver/internel/iface"

// NOTE: 接口定义统一收归到 internel/iface，本包仅保留 type alias 以减少改造期的调用面。
type NetworkGateway = iface.NetworkGateway
type Session = iface.ISession
type SessionGateway = iface.SessionGateway
type ClientGateway = iface.ClientGateway
