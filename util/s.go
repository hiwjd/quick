package util

func MaskMobile(mobile string) string {
	if mobile == "" {
		return ""
	}

	return mobile[0:3] + "****" + mobile[len(mobile)-4:]
}
