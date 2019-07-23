module github.com/xeals/signal-back

go 1.12

require (
	github.com/golang/protobuf v1.1.0
	github.com/h2non/filetype v1.0.5
	github.com/pkg/errors v0.8.1
	github.com/urfave/cli v1.20.0
	github.com/xeals/signal-back/signal v0.0.0
	golang.org/x/crypto v0.0.0-20180808211826-de0752318171
	golang.org/x/sys v0.0.0-20181122145206-62eef0e2fa9b // indirect
	gopkg.in/h2non/filetype.v1 v1.0.5 // indirect
)

replace github.com/xeals/signal-back/signal => ./signal
