package warewulfd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
)

var initramfsHandlerTests = []struct {
	description string
	url         string
	status      int
	ip          string
	suffix      string
}{
	{"find initramfs", "/initramfs/00:00:00:ff:ff:ff", 200, "10.10.10.10:9873", ""},
	{"explicit debug kernel", "/initramfs/00:00:00:ff:ff:fe", 200, "10.10.10.11:9873", "+debug"},
}

func Test_HandleInitramfs(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	const release = "6.12.0-211.34.1.el10_2.x86_64"
	env.WriteFile("/etc/warewulf/nodes.conf", `nodeprofiles:
  default:
    image name: suse
nodes:
  n1:
    network devices:
      default:
        hwaddr: 00:00:00:ff:ff:ff
    profiles:
    - default
  n2:
    profiles:
    - default
    network devices:
      default:
        hwaddr: 00:00:00:ff:ff:fe
    kernel:
      version: /boot/vmlinuz-`+release+"+debug.gz")

	const root = "/var/lib/warewulf/chroots/suse/rootfs"
	env.CreateFile(root + "/boot/vmlinuz-" + release + "+debug.gz")
	for _, suffix := range []string{"", "+debug"} {
		env.CreateFile(root + "/lib/modules/" + release + suffix + "/vmlinuz")
		env.WriteFile(root+"/boot/initramfs-"+release+suffix+".img", release+suffix)
	}

	dbErr := LoadNodeDB()
	assert.NoError(t, dbErr)

	conf := warewulfconf.Get()
	secureFalse := false
	conf.Warewulf.SecureP = &secureFalse

	for _, tt := range initramfsHandlerTests {
		t.Run(tt.description, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.RemoteAddr = tt.ip
			w := httptest.NewRecorder()
			HandleInitramfs(w, req)
			res := w.Result()
			defer func() { _ = res.Body.Close() }()

			assert.Equal(t, tt.status, res.StatusCode)
			assert.Equal(t, release+tt.suffix, w.Body.String())
		})
	}
}
