// +build linux,cgo

package generate

import (
	"context"
	"io/ioutil"

	"github.com/containers/libpod/v2/libpod/image"
	"github.com/containers/libpod/v2/pkg/seccomp"
	"github.com/containers/libpod/v2/pkg/specgen"
	spec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	goSeccomp "github.com/seccomp/containers-golang"
	"github.com/sirupsen/logrus"
)

func getSeccompConfig(s *specgen.SpecGenerator, configSpec *spec.Spec, img *image.Image) (*spec.LinuxSeccomp, error) {
	var seccompConfig *spec.LinuxSeccomp
	var err error
	scp, err := seccomp.LookupPolicy(s.SeccompPolicy)
	if err != nil {
		return nil, err
	}

	if scp == seccomp.PolicyImage {
		if img == nil {
			return nil, errors.New("cannot read seccomp profile without a valid image")
		}
		labels, err := img.Labels(context.Background())
		if err != nil {
			return nil, err
		}
		imagePolicy := labels[seccomp.ContainerImageLabel]
		if len(imagePolicy) < 1 {
			return nil, errors.New("no seccomp policy defined by image")
		}
		logrus.Debug("Loading seccomp profile from the security config")
		seccompConfig, err = goSeccomp.LoadProfile(imagePolicy, configSpec)
		if err != nil {
			return nil, errors.Wrap(err, "loading seccomp profile failed")
		}
		return seccompConfig, nil
	}

	if s.SeccompProfilePath != "" {
		logrus.Debugf("Loading seccomp profile from %q", s.SeccompProfilePath)
		seccompProfile, err := ioutil.ReadFile(s.SeccompProfilePath)
		if err != nil {
			return nil, errors.Wrapf(err, "opening seccomp profile (%s) failed", s.SeccompProfilePath)
		}
		seccompConfig, err = goSeccomp.LoadProfile(string(seccompProfile), configSpec)
		if err != nil {
			return nil, errors.Wrapf(err, "loading seccomp profile (%s) failed", s.SeccompProfilePath)
		}
	} else {
		logrus.Debug("Loading default seccomp profile")
		seccompConfig, err = goSeccomp.GetDefaultProfile(configSpec)
		if err != nil {
			return nil, errors.Wrapf(err, "loading seccomp profile (%s) failed", s.SeccompProfilePath)
		}
	}

	return seccompConfig, nil
}
