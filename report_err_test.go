package tc2rc

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
)

func TestConvertReportErr(t *testing.T) {
	rep := Convert(context.Background(), filepath.Join(t.TempDir(), "missing"), t.TempDir(), Options{})
	if rep.OK || rep.Err == nil || rep.Error != rep.Err.Error() {
		t.Fatalf("report = OK %v, Error %q, Err %v; want the failure in both", rep.OK, rep.Error, rep.Err)
	}
	if !errors.Is(rep.Err, fs.ErrNotExist) {
		t.Errorf("Err = %v, want it to wrap fs.ErrNotExist so callers can tell what failed", rep.Err)
	}
}
