package ctx

import "github.com/ontio/monitor_demo/scanners"

type Context struct {
	Scanners []*scanners.ScannerInterface
}