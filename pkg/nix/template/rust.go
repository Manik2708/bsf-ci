package template

import (
	"html/template"
	"io"

	"github.com/buildsafedev/bsf/pkg/hcl2nix"
)

const (
	rustTmpl = `
	{ pkgs}:
    rustPkgs = pkgs: pkgs.rustBuilder.makePackageSet {
		packageFun = import ./Cargo.nix;
		release = {{ .Release }}
		{{ if ne .RustVersion ""}}
		rustVersion = {{ .RustVersion }}; {{ end }}
		{{ if ne .RustToolChain ""}}
		rustToolchain = {{ .RustToolChain }}; {{ end }}
		{{ if ne .RustChannel ""}}
		rustChannel = {{ .RustChannel }}; {{ end }}
		{{ if ne .RustProfile ""}}
		rustProfile = {{ .RustProfile }}; {{ end }}
		{{ if gt (len .ExtraRustComponents) 0}}
		extraRustComponenets = {{ .ExtraRustComponents }} {{ end }}
	};
	
	default = (rustPkgs pkgs).workspace.{{.Name}} {};
    }
    `
)

type rustApp struct {
	CrateName           string
	RustVersion         string
	RustToolChain       string
	RustChannel         string
	RustProfile         string
	ExtraRustComponents []string
	Release             bool
}

// GenerateRustApp generates default flake
func GenerateRustApp(fl *hcl2nix.RustApp, wr io.Writer) error {
	data := rustApp{
		CrateName:           fl.CrateName,
		RustVersion:         fl.RustVersion,
		RustToolChain:       fl.RustToolChain,
		RustChannel:         fl.RustChannel,
		RustProfile:         fl.RustProfile,
		ExtraRustComponents: fl.ExtraRustComponents,
	}

	if !fl.Release {
		data.Release = fl.Release
	} else {
		data.Release = true
	}

	t, err := template.New("rust").Parse(rustTmpl)
	if err != nil {
		return err
	}

	err = t.Execute(wr, data)
	if err != nil {
		return err
	}

	return nil
}
