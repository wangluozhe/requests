package url

// 解析map[string]interface{}为map[string][]string
func parseMapToStrings(paramsMap map[string]interface{}) (result map[string][]string) {
	v := Values{}
	for s, i := range paramsMap {
		switch i.(type) {
		case string:
			v.Add(s, i.(string))
		case []string:
			for _, s2 := range i.([]string) {
				v.Add(s, s2)
			}
		case int:
			v.Add(s, string(i.(int)))
		case []int:
			for _, s2 := range i.([]int) {
				v.Add(s, string(s2))
			}
		case float64:
			v.Add(s, string(int(i.(float64))))
		case []float64:
			for _, s2 := range i.([]float64) {
				v.Add(s, string(int(s2)))
			}
		}
	}
	return v.Values()
}
