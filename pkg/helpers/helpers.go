package helpers

func IsAvailable(hosts []string, self string) (string, bool) {
	flag := false
	available := ""
	for _, host := range hosts {
		if host == self {
			flag = true
		} else {
			available = host
		}
	}
	return available, flag
}
