package viewer

import (
	"fmt"

	"math"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	appsv1 "k8s.io/api/apps/v1"
)

func resourceDiff[P *appsv1.ControllerRevision | *appsv1.ReplicaSet](
	resList []P,
	origOldRev int64,
	origNewRev int64,
	getRev func(res P) int64,
	getDiffString func(res P) string,
) (*string, error) {
	var oldRs, newRs P
	var oldRev, newRev int64 = math.MaxInt, math.MaxInt
	if len(resList) == 0 {
		return nil, fmt.Errorf("Versions for %s not found.", "")
	}

	if origOldRev < 0 && origOldRev*int64(-1) <= int64(len(resList)-1) {
		oldRs = resList[origOldRev*-1]
		oldRev = getRev(oldRs)
	}

	if origNewRev < 0 && origNewRev*int64(-1) <= int64(len(resList)-1) {
		newRs = resList[origNewRev*-1]
		newRev = getRev(newRs)
	}

	if origNewRev == 0 {
		newRs = resList[0]
		newRev = getRev(newRs)
	}

	for _, rs := range resList {
		rev := getRev(rs)

		if origOldRev == rev {
			oldRs = rs
			oldRev = rev
		}
		if origNewRev == rev {
			newRs = rs
			newRev = rev
		}
	}
	if oldRev == math.MaxInt {
		return nil, fmt.Errorf("Old reversion %d not found", origOldRev)
	} else if newRev == math.MaxInt {
		return nil, fmt.Errorf("New reversion %d not found", origNewRev)
	} else if oldRev >= newRev {
		return nil, fmt.Errorf("Old reversion %d(%d) should less than new reversion %d(%d)", oldRev, origOldRev, newRev, origNewRev)
	}

	oldYaml := getDiffString(oldRs)
	newYaml := getDiffString(newRs)
	diff := myDiff(oldYaml, newYaml)
	return &diff, nil
}

func myDiff(text1, text2 string) string {
	edits := myers.ComputeEdits("OLD", text1, text2)
	diff := fmt.Sprint(gotextdiff.ToUnified("OLD", "NEW", text1, edits))
	return diff
}
