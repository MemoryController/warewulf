package image

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

func TestFindInitramfs(t *testing.T) {
	conf := warewulfconf.Get()
	temp, err := os.MkdirTemp(os.TempDir(), "ww-conf-*")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(temp) }()
	conf.Paths.WWChrootdir = temp

	assert.NoError(t, os.MkdirAll(filepath.Join(RootFsDir("image"), "boot"), 0700))

	tests := map[string]struct {
		initramfs []string
		release   string
		ver       string
		path      string
	}{
		"full debug release": {
			initramfs: []string{"/boot/initrd.img-6.12.0-1.x86_64", "/boot/initrd.img-6.12.0-1.x86_64+debug"},
			release:   "6.12.0-1.x86_64+debug",
			ver:       "6.12.0-1",
			path:      "/boot/initrd.img-6.12.0-1.x86_64+debug",
		},
		"ok case 1": {
			initramfs: []string{"/boot/initramfs-1.1.1.aarch64.img"},
			ver:       "1.1.1",
			path:      "/boot/initramfs-1.1.1.aarch64.img",
		},
		"ok case 2": {
			initramfs: []string{"/boot/initrd-1.1.1.aarch64"},
			ver:       "1.1.1",
			path:      "/boot/initrd-1.1.1.aarch64",
		},
		"ok case 3": {
			initramfs: []string{"/boot/initramfs-1.1.1.aarch64"},
			ver:       "1.1.1",
			path:      "/boot/initramfs-1.1.1.aarch64",
		},
		"ok case 4": {
			initramfs: []string{"/boot/initrd-1.1.1.aarch64.img"},
			ver:       "1.1.1",
			path:      "/boot/initrd-1.1.1.aarch64.img",
		},
		"ok case 5 (ubuntu)": {
			initramfs: []string{"/boot/initrd.img-6.11.0-18-generic"},
			ver:       "6.11.0-18",
			path:      "/boot/initrd.img-6.11.0-18-generic",
		},
		"exact match in later search path": {
			initramfs: []string{"/boot/initramfs-1.1.1.x86_64+debug.img", "/boot/initrd-1.1.1.x86_64"},
			release:   "1.1.1.x86_64",
			ver:       "1.1.1",
			path:      "/boot/initrd-1.1.1.x86_64",
		},
		"prefix match": {
			release:   "1.1.1-custom",
			initramfs: []string{"/boot/initrd-1.1.1.aarch64.img"},
			ver:       "1.1",
			path:      "/boot/initrd-1.1.1.aarch64.img",
		},
		"error case, wrong init name": {
			initramfs: []string{"/boot/initrr-1.1.1.aarch64.img"},
			ver:       "1.1.1",
			path:      "",
		},
		"error case, wrong ver": {
			initramfs: []string{"/boot/initrd-1.1.1.aarch64.img"},
			ver:       "1.1.2",
			path:      "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			for _, init := range tt.initramfs {
				env.CreateFile(filepath.Join("/var/lib/warewulf/chroots/image/rootfs", init))
			}

			initramfs := FindInitramfs("image", tt.release, tt.ver)
			if tt.path == "" {
				assert.Nil(t, initramfs)
			} else {
				require.NotNil(t, initramfs)
				assert.Equal(t, tt.path, initramfs.Path)
			}
		})
	}
}
