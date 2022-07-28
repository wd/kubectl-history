package viewer

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	RevisionAnnotation = "deployment.kubernetes.io/revision"
)

type KindViewer interface {
	List(bool) (table.Writer, error)
	Diff(int64, int64) (*string, error)
}
