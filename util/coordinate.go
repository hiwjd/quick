package util

// LongitudePan 计算经度偏移meter米后的经度
func LongitudePan(longtitude float32, meter int) float32 {
	return longtitude + (float32(meter)/1.1)*0.00001
}

// LatitudePan 计算纬度偏移meter米后的纬度
func LatitudePan(latitude float32, meter int) float32 {
	return latitude + float32(meter)*0.00001
}
