package set

type LightStatus string

const (
	On        LightStatus = "on"
	Off       LightStatus = "off"
	Undefined LightStatus = "undefined"
)

// ToBool converts a LightStatus to its bool value. Undefined is considered as false.
func ToBool(status LightStatus) *bool {

	onValue, offValue := true, false

	if status == On {
		return &onValue
	} else if status == Off {
		return &offValue
	} else {
		return &offValue
	}
}
