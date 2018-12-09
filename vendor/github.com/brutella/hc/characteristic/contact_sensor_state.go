// THIS FILE IS AUTO-GENERATED
package characteristic

const (
	ContactSensorStateContactDetected    int = 0
	ContactSensorStateContactNotDetected int = 1
)

const TypeContactSensorState = "0000006A-0000-1000-8000-0026BB765291"

type ContactSensorState struct {
	*Int
}

func NewContactSensorState() *ContactSensorState {
	char := NewInt(TypeContactSensorState)
	char.Format = FormatUInt8
	char.Perms = []string{PermRead, PermEvents}

	char.SetValue(0)

	return &ContactSensorState{char}
}
