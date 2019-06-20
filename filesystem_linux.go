package main

import (
	"os"
	"path/filepath"
	"strings"
)

func getDeviceMapping() map[string]string {
	m := make(map[string]string)
	fid, err := os.Open("/dev/disk/by-id")
	if err != nil {
		return nil
	}
	defer fid.Close()

	devices, err := fid.Readdir(1024)
	if err != nil {
		return nil
	}

	for _, device := range devices {
		parts := strings.Split(strings.TrimPrefix(device.Name(), "virtio-"), "-")
		uuid := strings.Join(parts[:len(parts)-1], "-")
		src, err := os.Readlink(filepath.Join(fid.Name(), device.Name()))
		if err != nil {
			continue
		}
		m[filepath.Base(src)] = uuid
	}

	return m
}
