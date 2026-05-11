module github.com/sverrehu/spacegame

go 1.26.1

require (
	// gogpu/gg v0.45.4 and onwards break on Retina displays. do not upgrade without testing.
	github.com/gogpu/gg v0.45.3
	github.com/gogpu/gogpu v0.34.3
	github.com/gogpu/gpucontext v0.18.0
	github.com/sverrehu/goutils v1.0.3
)

require (
	github.com/go-text/typesetting v0.3.4 // indirect
	github.com/go-webgpu/goffi v0.5.0 // indirect
	github.com/go-webgpu/webgpu v0.4.3 // indirect
	github.com/gogpu/gputypes v0.5.0 // indirect
	github.com/gogpu/naga v0.17.13 // indirect
	github.com/gogpu/wgpu v0.27.3 // indirect
	golang.org/x/image v0.39.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/text v0.36.0 // indirect
)
