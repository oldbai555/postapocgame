/**
 * @Author: zjj
 * @Date: 2025/11/8
 * @Desc:
**/

package engine

var SendToClient func(sessionId string, msgId uint16, data []byte) error
