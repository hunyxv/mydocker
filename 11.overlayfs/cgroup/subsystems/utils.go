package subsystems

import (
	"bufio"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	ErrCreateCgroupDirFail = errors.New("failed to create cgroup folder")
)

func GetCgrouppath(subsystem string, containerid string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)
	if _, err := os.Stat(path.Join(cgroupRoot, containerid)); err == nil || (autoCreate && os.IsNotExist(err)) {
		if os.IsNotExist(err) {
			if err := os.Mkdir(path.Join(cgroupRoot, containerid), 0755); err != nil {
				return "", ErrCreateCgroupDirFail
			}
		}
		return path.Join(cgroupRoot, containerid), nil
	}else {
		return "", err
	}
}

func FindCgroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		logrus.WithError(err).Error("open '/proc/self/mountinfo' fail")
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan(){
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		logrus.WithError(err).Error("Failed to read '/proc/self/mountinfo' file")
		return ""
	}
	return ""
}