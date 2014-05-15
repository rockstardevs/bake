package main

import (
	"encoding/json"
	"fmt"
	"github.com/singhsaysdotcom/shlog"
	"io/ioutil"
	"os"
)

type Version struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Build int `json:"build"`
}

func GetVersion(versionFile *string) (bool, *Version, error) {
	logger.Message("Current Version")
	var current_version Version
	version_data, err := ioutil.ReadFile(*versionFile)
	if os.IsNotExist(err) {
		logger.Status(shlog.Orange, "unversioned")
		return false, nil, nil
	}
	if err != nil {
		logger.Status(shlog.Orange, "unknown")
		return false, nil, err
	}
	err = json.Unmarshal(version_data, &current_version)
	if err != nil {
		logger.Status(shlog.Orange, "unknown")
		return false, nil, err
	}
	logger.Status(shlog.Green, current_version.String())
	return true, &current_version, nil
}

func SaveVersion(versionFile *string) (bool, error) {
	logger.Message("saving new version ...")
	version_data, err := json.Marshal(current_version)
	if err != nil {
		logger.Err()
		return false, err
	}
	err = ioutil.WriteFile(*versionFile, version_data, 0644)
	if err != nil {
		logger.Err()
		return false, err
	}
	logger.Ok()
	return true, nil
}

func NewVersion() *Version {
	return &Version{Major: 0, Minor: 1, Build: 0}
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Build)
}

func (v *Version) IncMajor() {
	logger.Message("new major version")
	v.Major += 1
	v.Minor = 0
	v.Build = 1
	logger.Status(shlog.Green, current_version.String())
}

func (v *Version) IncMinor() {
	logger.Message("new minor version")
	v.Minor += 1
	v.Build = 1
	logger.Status(shlog.Green, current_version.String())
}

func (v *Version) IncBuild() {
	logger.Message("new build version")
	v.Build += 1
	logger.Status(shlog.Green, current_version.String())
}
