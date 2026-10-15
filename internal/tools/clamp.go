package tools

func Clamp(value, minValue, maxValue int64) int64 {
	if value < minValue {
		return minValue
	}

	if value > maxValue {
		return maxValue
	}

	return value
}
