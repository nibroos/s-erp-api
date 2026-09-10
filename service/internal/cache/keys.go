package cache

import "fmt"

// keyPrefix namespaces every key this service writes into the shared Redis
// instance, so keys never collide with other services (s-erp-auth, etc.).
const keyPrefix = "s-erp:api:"

// CustomerKey is the cache key for a single customer detail record.
func CustomerKey(id uint) string {
	return fmt.Sprintf("%scustomer:%d", keyPrefix, id)
}
