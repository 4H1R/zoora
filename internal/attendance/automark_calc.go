package attendance

import "github.com/google/uuid"

// computePresentByPercent returns the user IDs whose accumulated seconds meet the
// threshold percent of the total room seconds. ok=false means the denominator is
// non-positive and attendance cannot be computed (caller should skip).
func computePresentByPercent(totalRoomSeconds int, userSeconds map[uuid.UUID]int, percent int) ([]uuid.UUID, bool) {
	if totalRoomSeconds <= 0 {
		return nil, false
	}
	var present []uuid.UUID
	for userID, secs := range userSeconds {
		// secs/total >= percent/100  ->  secs*100 >= percent*total
		if secs*100 >= percent*totalRoomSeconds {
			present = append(present, userID)
		}
	}
	return present, true
}
