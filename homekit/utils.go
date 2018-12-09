package homekit

func convertInterfaceToFloat64(v interface{}) float64 {
	switch v := v.(type) {
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case float64:
		return v
	case float32:
		return float64(v)
	}

	return 0
}
