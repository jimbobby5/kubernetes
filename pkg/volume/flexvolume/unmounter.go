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

	"github.com/docker/docker/pkg/mount"

	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/util"
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
	klog.Warningf("GR: Doing TearDownAt for %v", dir)
	mounted, mountErr := mount.Mounted(dir)
	klog.Warningf("GR: mounted var is %v for %v", mounted, dir)
	if mountErr != nil {
		// only log warning here since plugins should anyways have to deal with errors
		klog.Warningf("Error checking path: %v", mountErr)
	}

	if mounted || mountErr != nil {
	    klog.Warningf("GR: about to run driver unmount for %v", dir)
		call := f.plugin.NewDriverCall(unmountCmd)
		call.Append(dir)
		_, err := call.Run()
		klog.Warningf("GR: Done unmount command for %v", dir)
		if isCmdNotSupportedErr(err) {
			klog.Warningf("GR: command not supported error for %v", dir)
			err = (*unmounterDefaults)(f).TearDownAt(dir)
		}
		if err != nil {
			klog.Warningf("GR: returning error %s for %v", err, dir)
			return err
		}
	} else {
		klog.Warningf("Warning: Unmount skipped because path is not mounted: %v", dir)
	}

	klog.Warningf("GR: About to check if path exists for %v", dir)

	// Flexvolume driver may remove the directory. Ignore if it does.
	if pathExists, pathErr := util.PathExists(dir); pathErr != nil {
		klog.Warningf("GR: Error checking if path exists for %v, %s", dir, pathErr)
		return fmt.Errorf("Error checking if path exists: %v", pathErr)
	} else if !pathExists {
		klog.Warningf("GR: Path does not exist for %v", dir)
		return nil
	}
	klog.Warningf("GR: About to remove dir for %v", dir)
	return os.Remove(dir)
}

