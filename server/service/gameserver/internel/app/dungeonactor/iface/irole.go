/**
 * @Author: zjj
 * @Date: 2025/11/25
 * @Desc:
**/

package iface

type IRole interface {
	IEntity

	GetSessionId() string
}
