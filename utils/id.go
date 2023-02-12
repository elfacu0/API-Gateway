package utils

func ID(path, method string) string {
	return path + "_" + method
}

func IdCache(path, method string) string {
	return ID(path, method) + "_CACHE"
}

func IdRquests(path, method string) string {
	return ID(path, method) + "_REQUESTS"
}
