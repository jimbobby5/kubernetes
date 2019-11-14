/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package flexvolume

import (
	"fmt"
	"os"

	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/mount"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/utils/exec"
)

// FlexVolumeUnmounter is the disk that will be cleaned by this plugin.
type flexVolumeUnmounter struct {
	*flexVolume
	// Runner used to teardown the volume.
	runner exec.Interface
}

var _ volume.Unmounter = &flexVolumeUnmounter{}

// Unmounter interface
func (f *flexVolumeUnmounter) TearDown() error {
	path := f.GetPath()
	return f.TearDownAt(path)
}

func (f *flexVolumeUnmounter) TearDownAt(dir string) error {
	mounts, mountCheckErr := f.flexVolume.mounter.GetMountRefs(dir)
	mounted := len(mounts) > 0
	if mountCheckErr != nil {
		// only log warning here since plugins should anyways have to deal with errors
		klog.Warningf("Error checking path: %v", mountCheckErr)
	} else {
		if !mounted {
			klog.Warningf("Warning: Unmount skipped because path is not mounted: %v", dir)
			return nil
		}
	}

	call := f.plugin.NewDriverCall(unmountCmd)
	call.Append(dir)
	_, err := call.Run()
	if isCmdNotSupportedErr(err) {
		err = (*unmounterDefaults)(f).TearDownAt(dir)
	}
	if err != nil {
		return err
	}

	// Flexvolume driver may remove the directory. Ignore if it does.
	if pathExists, pathErr := mount.PathExists(dir); pathErr != nil {
		return fmt.Errorf("Error checking if path exists: %v", pathErr)
	} else if !pathExists {
		return nil
	}
	return os.Remove(dir)
}
