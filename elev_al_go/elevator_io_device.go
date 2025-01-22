package elevalgo

const (
	NUM_FLOORS  = 4
	NUM_BUTTONS = 3
)

type Dir int

const (
	DIR_DOWN Dir = iota - 1
	DIR_STOP
	DIR_UP
)

type Button int

const (
	BTN_HALLUP Button = iota
	BTN_HALLDOWN
	BTN_HALLCAB
)

func ElevioDirToString(d Dir) string {
	switch d {
	case DIR_UP:
		return "D_Up"
	case DIR_DOWN:
		return "D_Down"
	case DIR_STOP:
		return "D_Stop"
	default:
		return "D_UNDEFINED"
	}
}

func ElevioButtonToString(b Button) string {
	switch b {
	case BTN_HALLUP:
		return "B_HallUp"
	case BTN_HALLDOWN:
		return "B_HallDown"
	case BTN_HALLCAB:
		return "B_Cab"
	default:
		return "B_UNDEFINED"
	}
}
