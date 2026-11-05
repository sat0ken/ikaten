package inputstate

const Count = 128

var raw [Count]bool
var duration [Count]int
var released [Count]bool

func Reset() { raw = [Count]bool{}; duration = [Count]int{}; released = [Count]bool{} }
func Set(key int, pressed bool) {
	if key >= 0 && key < Count {
		raw[key] = pressed
	}
}
func Advance() {
	for i := range duration {
		released[i] = !raw[i] && duration[i] > 0
		if raw[i] {
			if duration[i] < int(^uint(0)>>1) {
				duration[i]++
			}
		} else {
			duration[i] = 0
		}
	}
}
func Duration(key int) int {
	if key < 0 || key >= Count {
		return 0
	}
	return duration[key]
}
func Released(key int) bool { return key >= 0 && key < Count && released[key] }
