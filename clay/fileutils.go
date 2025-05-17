package clay

func FileChangeExtension(filename, newExt string) string {
	// Find the last dot in the filename
	lastDot := -1
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			lastDot = i
			break
		}
	}

	// If no dot is found, just append the new extension
	if lastDot == -1 {
		return filename + newExt
	}

	// Replace the old extension with the new one
	return filename[:lastDot] + newExt
}
