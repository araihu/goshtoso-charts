package pages

func navItemClass(active bool) string {
	base := "shrink-0 rounded-radius px-3 py-2 text-sm font-medium"
	if active {
		return base + " bg-primary text-on-primary dark:bg-primary-dark dark:text-on-primary-dark"
	}
	return base + " text-on-surface hover:bg-surface-alt dark:text-on-surface-dark dark:hover:bg-surface-dark-alt"
}

func navCurrent(active bool) string {
	if active {
		return "page"
	}
	return ""
}
