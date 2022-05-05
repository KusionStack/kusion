package version

func ReleaseVersion() string {
	return info.ReleaseVersion
}

func String() string {
	return info.String()
}

func ShortString() string {
	return info.ShortString()
}

func JSON() string {
	return info.JSON()
}

func YAML() string {
	return info.YAML()
}
