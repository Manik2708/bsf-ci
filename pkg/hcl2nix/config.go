package hcl2nix

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"

	bstrings "github.com/buildsafedev/bsf/pkg/strings"
	"github.com/buildsafedev/bsf/pkg/update"
)

// Config for hcl2nix
type Config struct {
	Packages       Packages        `hcl:"packages,block"`
	GoModule       *GoModule       `hcl:"gomodule,block"`
	PoetryApp      *PoetryApp      `hcl:"poetryapp,block"`
	RustApp        *RustApp        `hcl:"rustapp,block"`
	JsNpmApp       *JsNpmApp       `hcl:"jsnpmapp,block"`
	OCIArtifact    []OCIArtifact   `hcl:"oci,block"`
	ConfigFiles    []ConfigFiles   `hcl:"config,block"`
	GitHubReleases []GitHubRelease `hcl:"githubRelease,block"`
}

// GitHubRelease holds github release parameters
type GitHubRelease struct {
	Owner string `hcl:"owner"`
	Repo  string `hcl:"repo"`
	App   string `hcl:"app,label"`
	Dir   string `hcl:"dir,optional"`
}

// Packages holds package parameters
type Packages struct {
	// Maybe these should be of type Set? https://github.com/deckarep/golang-set
	Development []string `hcl:"development"`
	Runtime     []string `hcl:"runtime"`
}

// ReadHclFile reads an HCL file
func ReadHclFile(fileName string) (*Config, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("error: %s", err.Error())
	}

	var dstErr bytes.Buffer
	conf, err := ReadConfig(data, &dstErr)
	if err != nil {
		return nil, fmt.Errorf(dstErr.String())
	}
	return conf, nil
}

// WriteConfig writes packages to writer
func WriteConfig(config Config, wr io.Writer) error {
	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(&config, f.Body())
	_, err := f.WriteTo(wr)
	if err != nil {
		return err
	}
	return nil
}

// ModifyConfig modifes the config
func ModifyConfig(oldName string, artifact OCIArtifact, config *Config) error {
	updated := false
	for i, existingArtifact := range config.OCIArtifact {
		if existingArtifact.Name == oldName {
			config.OCIArtifact[i] = artifact
			updated = true
			break
		}
	}

	if updated {
		var buf bytes.Buffer
		err := WriteConfig(*config, &buf)
		if err != nil {
			return err
		}

		err = os.WriteFile("bsf.hcl", buf.Bytes(), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// ReadConfig reads config from bytes and returns Config. If any errors are encountered, they are written to dstErr
func ReadConfig(src []byte, dstErr io.Writer) (*Config, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCL(src, "bsf.hcl")
	if diags.HasErrors() {
		wr := hcl.NewDiagnosticTextWriter(
			dstErr,
			parser.Files(),
			78,
			true,
		)
		if err := wr.WriteDiagnostics(diags); err != nil {
			return nil, fmt.Errorf("error writing diagnostics: %w", err)
		}
		return nil, diags
	}

	var config Config
	diags = gohcl.DecodeBody(f.Body, nil, &config)
	if diags.HasErrors() {
		wr := hcl.NewDiagnosticTextWriter(
			dstErr,
			parser.Files(),
			78,
			true,
		)
		if err := wr.WriteDiagnostics(diags); err != nil {
			return nil, fmt.Errorf("error writing diagnostics: %w", err)
		}
		return nil, diags
	}

	return &config, nil
}

// AddPackages updates config with new packages. It appends new packages to existing packages
func AddPackages(src []byte, packages Packages, wr io.Writer) error {
	existingConfig, err := ReadConfig(src, io.Discard)
	if err != nil {
		return err
	}

	parse := func(s string) string {
		name, _ := update.ParsePackage(s)
		return name
	}

	existingConfig.Packages.Development = bstrings.PreferNewSliceElements(existingConfig.Packages.Development, packages.Development, parse)
	existingConfig.Packages.Runtime = bstrings.PreferNewSliceElements(existingConfig.Packages.Runtime, packages.Runtime, parse)

	err = WriteConfig(*existingConfig, wr)
	if err != nil {
		return err
	}

	return nil
}

// SetPackages updates config with new packages. It replaces existing packages with new packages
func SetPackages(src []byte, packages Packages, wr io.Writer) error {
	existingConfig, err := ReadConfig(src, io.Discard)
	if err != nil {
		return err
	}

	existingConfig.Packages = packages

	err = WriteConfig(*existingConfig, wr)
	if err != nil {
		return err
	}

	return nil

}
