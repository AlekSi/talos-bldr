module github.com/talos-systems/bldr

go 1.15

require (
	github.com/AlekSi/pointer v1.1.0
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/alessio/shellescape v1.4.1
	github.com/containerd/containerd v1.4.1-0.20201117152358-0edc412565dc
	github.com/emicklei/dot v0.15.0
	github.com/google/go-github/v33 v33.0.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/mitchellh/copystructure v1.1.1 // indirect
	github.com/moby/buildkit v0.8.1
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/otiai10/copy v1.4.2
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.3.1-0.20200512144102-f13ba8f2f2fd
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200310163718-4634ce647cf2+incompatible
)
