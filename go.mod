module github.com/Channel-3-Eugene/tribd

go 1.22.2

require github.com/stretchr/testify v1.9.0

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/Channel-3-Eugene/tribd/mpegts => ../mpegts
replace github.com/Channel-3-Eugene/tribd/config => ../config
replace github.com/Channel-3-Eugene/tribd/channels => ../channels

