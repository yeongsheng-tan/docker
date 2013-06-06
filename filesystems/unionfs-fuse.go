package filesystems

import (
	"fmt"
	"os/exec"
)

func MountUnionFuseFS(ro []string, rw string, target string) error {
	// FIXME: Now mount the layers
	rwBranch := fmt.Sprintf("%v=RW", rw)
	roBranches := ""
	for _, layer := range ro {
		roBranches += fmt.Sprintf("%v=RO:", layer)
	}
	branches := fmt.Sprintf("%v:%v", rwBranch, roBranches)

	if err := exec.Command("unionfs-fuse", "-o", "cow", "-o", "dev", branches, target).Run(); err != nil {
		fmt.Println(err.Error())
		return fmt.Errorf("Unable to mount using UnionFS-Fuse")
	}
	return nil
}

func UnmountUnionFuseFS(target string) error {
	return Unmount(target)
}