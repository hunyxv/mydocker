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
	imagePath         string
	containerID       string
	containerFS       string
	readonlyLayerPath string
	writelayerpath    string
	mergeDirPath      string
	workDirPath       string
	volumes           []string
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
		imagePath:         imagePath,
		containerID:       containerid,
		containerFS:       containerFS,
		readonlyLayerPath: containerFS + "/readonly-layer",
		writelayerpath:    containerFS + "/write-layer",
		mergeDirPath:      containerFS + "/merged",
		workDirPath:       containerFS + "/work",
	}, nil
}

func (ws *WorkSpace) CreateWorkSpace() error {
	if err := ws.createReadOnlyLayer(); err != nil {
		return err
	}

	if err := ws.createWriteLayer(); err != nil {
		return err
	}

	if err := ws.createMergedAndMount(); err != nil {
		return err
	}

	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		ws.readonlyLayerPath, ws.writelayerpath, ws.workDirPath)
	return syscall.Mount("overlay", ws.mergeDirPath, "overlay", syscall.MS_NOSUID, options)
}

func (ws *WorkSpace) createReadOnlyLayer() error {
	exist, err := PathExists(ws.imagePath)
	if err != nil {
		return err
	}
	if !exist {
		return os.ErrNotExist
	}

	err = os.Mkdir(ws.readonlyLayerPath, 0710)
	if err != nil {
		if err == os.ErrExist {
			return nil
		}
		return err
	}

	if _, err := exec.Command("tar", "-xvf", ws.imagePath, "-C", ws.readonlyLayerPath).CombinedOutput(); err != nil {
		logrus.WithError(err).Errorf("'tar -xvf %s -C %s' fail", ws.imagePath, ws.readonlyLayerPath)
		return err
	}
	return nil
}

func (ws *WorkSpace) createWriteLayer() error {
	return os.Mkdir(ws.writelayerpath, 0710)
}

func (ws *WorkSpace) createMergedAndMount() error {
	if err := os.Mkdir(ws.mergeDirPath, 0710); err != nil {
		return err
	}

	return os.Mkdir(ws.workDirPath, 0710)
}

func (ws *WorkSpace) mountVolume(source, target string) error {
	exist, err := PathExists(source)
	if err != nil {
		return err
	}

	if !exist {
		err = os.MkdirAll(source, 0755)
		if err != nil {
			return fmt.Errorf("craete %s dir fail", source)
		}
	}

	targetpath := ws.mergeDirPath + target
	exist, err = PathExists(targetpath)
	if err != nil {
		return err
	}

	if !exist {
		err = os.MkdirAll(targetpath, 0755)
		if err != nil {
			return fmt.Errorf("craete %s dir fail", targetpath)
		}
	}

	err = syscall.Mount(source, targetpath, "bind", syscall.MS_BIND, "")
	if err != nil {
		return err
	}

	ws.volumes = append(ws.volumes, target)
	return nil
}

func (ws *WorkSpace) RootPath() string {
	return ws.mergeDirPath
}

func (ws *WorkSpace) Remove() error {
	err := syscall.Unmount(ws.mergeDirPath, syscall.MS_NOSUID)
	if err != nil {
		logrus.WithError(err).Error("unmount oberlayfs fail, %s", ws.mergeDirPath)
	}
	for _, volumePath := range ws.volumes {
		err = syscall.Unmount(volumePath, 0)
		if err != nil {
			logrus.WithError(err).Errorf("unmount volume fail, %s", volumePath)
		}
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
