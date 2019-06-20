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
		// After removing "virtio-" prefix, device id has form 36ae2fbb-3619-4933-...,
		// get first 3 element only.
		uuid := strings.Join(parts[:3], "-")
		src, err := filepath.EvalSymlinks(filepath.Join(fid.Name(), device.Name()))
		if err != nil {
			continue
		}
		m[src] = uuid
	}

	return m
}
