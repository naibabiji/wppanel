//go:build !linux

package handlers

func readDiskStats() (int64, int64) {
	return 0, 0
}
