/**
 * @Author: zjj
 * @Date: 2025/11/7
 * @Desc:
**/

package event

import "context"

type Type uint32

type Handler func(ctx context.Context, event *Event)
