package filesystems

import (
	"fmt"
	"log"
	"github.com/dotcloud/docker/utils"
	"os/exec"
	"syscall"
)
func MountOverlayFS(ro []string, rw string, target string) error {

	mountFunc := func (ro []string, rw string, target string) error {
		//
		// unlike aufs, you can't mount several branches with
			// overlayfs, so we do recursive mounts
		// for the ro layers (A, B, C, D) and RW and TARGET
		// we mount A as lowerdir, B as upperdir on B
		//          B as lowerdir, C as upperdir on C
		//          ----
		//          D as lowerdir, RW as upperdir on TARGET
		//
		// The layers are in reverse, greater index is lower
			// on the stack (rw is on top)
		prevLayer := ro[len(ro) - 1]
		for i := len(ro) - 2; i >= 0; i-- {
			layer := ro[i]
			options := fmt.Sprintf("lowerdir=%v,upperdir=%v", prevLayer, layer)
			if err := mount("overlayfs", layer, "overlayfs", syscall.MS_RDONLY, options); err != nil {
				return fmt.Errorf("Unable to mount %v on %v using overlayfs (ro)", prevLayer, layer)
			}
			utils.Debugf("MountOverlayfs %v/%v -> %v", prevLayer, layer, layer)
			prevLayer = layer
		}

		options := fmt.Sprintf("lowerdir=%v,upperdir=%v", prevLayer, rw)
		if err := mount("overlayfs", target, "overlayfs", 0, options); err != nil {
			return fmt.Errorf("Unable to mount %v on %v using overlayfs", prevLayer, target)
		}
		utils.Debugf("MountOverlayfs %v/%v -> %v", prevLayer, rw, target)
		return nil
	}

	if err := mountFunc(ro, rw, target); err != nil {
		log.Printf("Kernel does not support overlayfs, trying to load the overlayfs module with modprobe...")
		if err := exec.Command("modprobe", "overlayfs").Run(); err != nil {
			return fmt.Errorf("Unable to load the overlayfs module")
		}
		log.Printf("...module loaded.")
		if err := mountFunc(ro, rw, target); err != nil {
			return fmt.Errorf("Unable to mount using overlayfs")
		}
	}
	return nil
}

func UnmountOverlayFS(target string, layers []string) error {		
	if err := Unmount(target); err != nil {
		return err
	}
	
	for _, layer := range layers {
		if err := Unmount(layer); err != nil {
			return err
		}
	}
	return nil
}