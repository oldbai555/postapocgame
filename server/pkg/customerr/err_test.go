/**
 * @Author: zjj
 * @Date: 2025/1/13
 * @Desc:
**/

package customerr

import "testing"

func TestWrapByCall(t *testing.T) {
	err := NewInvalidArg("111")
	err = Wrap(err)
	t.Logf("customerr:%v", err)
}
