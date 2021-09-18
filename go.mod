module github.com/kuredoro/cptest

go 1.17

require (
	github.com/alexflint/go-arg v1.3.0
	github.com/atomicgo/cursor v0.0.1
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jonboulle/clockwork v0.2.2
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/mattn/go-colorable v0.1.8
	github.com/maxatome/go-testdeep v1.10.0
	github.com/sanity-io/litter v1.3.0
	github.com/shettyh/threadpool v0.0.0-20200323115144-b99fd8aaa945
)

require (
	github.com/alexflint/go-scalar v1.0.0 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
)

retract (
	v1.2.0
	v1.1.0
	v1.0.0
)
