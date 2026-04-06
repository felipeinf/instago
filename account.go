package ig

import (
	"fmt"
)

// AccountInfo returns the logged-in account from accounts/current_user/?edit=true.
func (c *Client) AccountInfo() (Account, error) {
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "accounts/current_user/?edit=true",
		WithSignature: true,
	})
	if err != nil {
		return Account{}, err
	}
	um, ok := res["user"].(map[string]any)
	if !ok {
		return Account{}, fmt.Errorf("ig: account missing user")
	}
	return extractAccount(um), nil
}
