package main

import (
	"os"
	"path/filepath"
	"strings"
)

func getDeviceMapping() (m map[string]string) {
	fid, err := os.Open("/dev/disk/by-id")
	if err != nil {
		return
	}

	devices, err := fid.Readdir(1024)
	if err != nil {
		return
	}

	for _, device := range devices {
		parts := strings.Split(strings.TrimPrefix(device.Name(), "virtio-"), "-")
		uuid := strings.Join(parts[:len(parts)], "-")
		src, err := os.Readlink(filepath.Join(fid.Name(), device.Name()))
		if err != nil {
			continue
		}

		m[filepath.Base(src)] = uuid
	}

	return
}
