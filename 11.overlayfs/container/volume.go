package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

const (
	rootpath = "/var/lib/mydocker"
)

type WorkSpace struct {
	imagePath   string
	containerID string
	containerFS string
}

func NewWorkSpace(imagePath, containerid string) (*WorkSpace, error) {
	containerFS := rootpath + "/" + containerid
	exist, err := PathExists(containerFS)
	if err != nil {
		return nil, err
	}

	if !exist {
		os.MkdirAll(containerFS, 0710)
	}
	return &WorkSpace{
		imagePath:   imagePath,
		containerID: containerid,
		containerFS: containerFS,
	}, nil
}

func (ws *WorkSpace) CreateWorkSpace() error {
	readonlyLayer, err := ws.createReadOnlyLayer()
	if err != nil {
		return err
	}

	writeLayer, err := ws.createWriteLayer()
	if err != nil {
		return err
	}

	merged, work, err := ws.createMergedAndMount()
	if err != nil {
		return err
	}

	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", readonlyLayer, writeLayer, work)
	return syscall.Mount("overlay", merged, "overlay", syscall.MS_NOSUID, options)
}

func (ws *WorkSpace) createReadOnlyLayer() (string, error) {
	exist, err := PathExists(ws.imagePath)
	if err != nil {
		return "", err
	}
	if !exist {
		return "", os.ErrNotExist
	}

	path := ws.containerFS + "/readonly-layer"
	err = os.Mkdir(path, 0710)
	if err != nil {
		if err == os.ErrExist {
			return path, nil
		}
		return "", err
	}

	if _, err := exec.Command("tar", "-xvf", ws.imagePath, "-C", path).CombinedOutput(); err != nil {
		logrus.WithError(err).Errorf("'tar -xvf %s -C %s' fail", ws.imagePath, path)
		return "", err
	}
	return path, nil
}

func (ws *WorkSpace) createWriteLayer() (string, error) {
	writeLayerPath := ws.containerFS + "/write-layer"
	return writeLayerPath, os.Mkdir(writeLayerPath, 0710)
}

func (ws *WorkSpace) createMergedAndMount() (string, string, error) {
	mergedPath := ws.containerFS + "/merged"
	if err := os.Mkdir(mergedPath, 0710); err != nil {
		return "", "", err
	}

	workPath := ws.containerFS + "/work"
	if err := os.Mkdir(workPath, 0710); err != nil {
		return "", "", err
	}

	return mergedPath, workPath, nil
}

func (ws *WorkSpace) Remove() error {
	err := syscall.Unmount(ws.containerFS+"/merged", syscall.MS_NOSUID)
	if err != nil {
		return err
	}

	return os.RemoveAll(ws.containerFS)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
