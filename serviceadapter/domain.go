package serviceadapter

import (
	"errors"

	"github.com/pivotal-cf/on-demand-service-broker-sdk/bosh"

	"gopkg.in/go-playground/validator.v8"
)

//go:generate counterfeiter -o fake_service_adapter/fake_service_adapter.go . ServiceAdapter
type ServiceAdapter interface {
	GenerateManifest(boshInfo BoshInfo, serviceReleases ServiceReleases, plan Plan, arbitraryParams map[string]interface{}, previousManifest *bosh.BoshManifest) (bosh.BoshManifest, error)
	CreateBinding(bindingID string, deploymentTopology bosh.BoshVMs, manifest bosh.BoshManifest, arbitraryParams map[string]interface{}) (map[string]interface{}, error)
	DeleteBinding(bindingID string, deploymentTopology bosh.BoshVMs, manifest bosh.BoshManifest) error
}

var validate *validator.Validate

func init() {
	config := &validator.Config{TagName: "validate"}
	validate = validator.New(config)
}

type ServiceRelease struct {
	Name    string   `json:"name" validate:"required"`
	Version string   `json:"version" validate:"required"`
	Jobs    []string `json:"jobs" validate:"required,min=1"`
}

type ServiceReleases []ServiceRelease

func (r ServiceReleases) Validate() error {
	if len(r) < 1 {
		return errors.New("no releases specified")
	}

	for _, serviceRelease := range r {
		if err := validate.Struct(serviceRelease); err != nil {
			return err
		}
	}

	return nil
}

type BoshInfo struct {
	Name            string `json:"name" validate:"required"`
	StemcellOS      string `json:"stemcell_os" validate:"required"`
	StemcellVersion string `json:"stemcell_version" validate:"required"`
}

func (b BoshInfo) Validate() error {
	return validate.Struct(b)
}

type Properties map[string]interface{}

type Plan struct {
	Properties     Properties      `json:"properties"`
	InstanceGroups []InstanceGroup `json:"instance_groups" validate:"required,dive"`
}

func (p Plan) Validate() error {
	if err := validate.Struct(p); err != nil {
		return err
	}

	for _, instanceGroup := range p.InstanceGroups {
		if instanceGroup.Jobs == nil {
			continue
		}
		for _, job := range instanceGroup.Jobs {
			if err := validate.Struct(job); err != nil {
				return err
			}
		}
	}

	return nil
}

type InstanceGroup struct {
	Name           string   `json:"name" validate:"required"`
	VMType         string   `yaml:"vm_type" json:"vm_type" validate:"required"`
	PersistentDisk string   `yaml:"persistent_disk,omitempty" json:"persistent_disk_type,omitempty"`
	Instances      int      `json:"instances" validate:"min=1"`
	Networks       []string `json:"networks" validate:"required"`
	AZs            []string `json:"azs,omitempty"`
	Lifecycle      string   `yaml:"lifecycle,omitempty" json:"lifecycle,omitempty"`
	Jobs           []Job    `json:"jobs,omitempty"`
}

type Job struct {
	Name       string     `json:"name" validate:"required"`
	Release    string     `json:"release" validate:"required"`
	Properties Properties `json:"properties" validate:"required"`
}
