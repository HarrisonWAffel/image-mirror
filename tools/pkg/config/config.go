package config

import (
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/yaml"
)

type Config struct {
	Images       []*Image
	Repositories []Repository
}

type Image struct {
	// The source image without any tags.
	SourceImage            string
	defaultTargetImageName string
	// Used to specify the desired name of the target image if it differs
	// from default. This field would be private if it was convenient for
	// marshalling to JSON/YAML, but it is not. This field should not be
	// accessed directly - instead, use the TargetImageName() and
	// SetTargetImageName() methods.
	SpecifiedTargetImageName string `json:"TargetImageName,omitempty"`
	// The tags that we want to mirror.
	Tags []string
}

type Repository struct {
	// BaseUrl is used exclusively for building the target image ref
	// for a given image for a repository. For example, a target
	// image name of "mirrored-rancher-cis-operator" and a BaseUrl
	// of "docker.io/rancher" produce a target image ref of
	// "docker.io/rancher/mirrored-rancher-cis-operator".
	BaseUrl string
	// Whether the repository should have images mirrored to it.
	Target bool
	// Password is what goes into the "pass" field of regsync.yaml
	// for this repository. For more information please see
	// https://github.com/regclient/regclient/blob/main/docs/regsync.md
	Password string
	// Registry is what goes into the "registry" field of regsync.yaml
	// for this repository. For more information please see
	// https://github.com/regclient/regclient/blob/main/docs/regsync.md
	Registry string
	// Username is what goes into the "user" field of regsync.yaml
	// for this repository. For more information please see
	// https://github.com/regclient/regclient/blob/main/docs/regsync.md
	Username string
}

func Parse(fileName string) (Config, error) {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read: %w", err)
	}

	config := Config{}
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal as JSON: %w", err)
	}

	for _, image := range config.Images {
		if err := image.SetDefaults(); err != nil {
			return Config{}, fmt.Errorf("failed to set defaults for image %q: %w", image, err)
		}
	}

	return config, nil
}

func (image *Image) SetDefaults() error {
	parts := strings.Split(image.SourceImage, "/")
	if len(parts) < 2 {
		return fmt.Errorf("source image split into %d parts (>=2 parts expected)", len(parts))
	}
	repoName := parts[len(parts)-2]
	imageName := parts[len(parts)-1]
	image.defaultTargetImageName = "mirrored-" + repoName + "-" + imageName
	return nil
}

func (image Image) TargetImageName() string {
	if image.SpecifiedTargetImageName != "" {
		return image.SpecifiedTargetImageName
	}
	return image.defaultTargetImageName
}

func (image *Image) SetTargetImageName(value string) {
	if value == image.defaultTargetImageName {
		image.SpecifiedTargetImageName = ""
	} else {
		image.SpecifiedTargetImageName = value
	}
}
