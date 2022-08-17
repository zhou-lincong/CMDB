package utils

//
func PtrStrV(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}

//9.6
func PtrInt64(v *int64) int64 {
	if v == nil {
		return 0
	}

	return *v
}

//9.6
func PtrInt32(v *int32) int32 {
	if v == nil {
		return 0
	}

	return *v
}

//9.6
func PtrFloat64(v *float64) float64 {
	if v == nil {
		return 0
	}

	return *v
}

//9.6
func SlicePtrStrv(items []*string) []string {
	vs := []string{}
	for i := range items {
		v := PtrStrV(items[i])
		if v != "" {
			vs = append(vs, v)
		}
	}

	return vs
}
