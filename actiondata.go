package ig

func (c *Client) withDefaultData(data map[string]any) map[string]any {
	out := map[string]any{
		"_uuid":     c.uuids.UUID,
		"device_id": c.uuids.AndroidDeviceID,
	}
	for k, v := range data {
		out[k] = v
	}
	return out
}

func (c *Client) withActionData(data map[string]any) map[string]any {
	out := c.withDefaultData(map[string]any{"radio_type": "wifi-none"})
	for k, v := range data {
		out[k] = v
	}
	return out
}
