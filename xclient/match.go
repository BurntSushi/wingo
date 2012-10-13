package xclient

import (
	"strings"
)

func (c *Client) matchWmClass(haystack []string) bool {
	instance := strings.ToLower(c.class.Instance)
	class := strings.ToLower(c.class.Class)
	for _, s := range haystack {
		if s == instance || s == class {
			return true
		}
	}
	return false
}
