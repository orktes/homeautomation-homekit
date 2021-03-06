// THIS FILE IS AUTO-GENERATED
package characteristic

const TypeCurrentVerticalTiltAngle = "0000006E-0000-1000-8000-0026BB765291"

type CurrentVerticalTiltAngle struct {
	*Int
}

func NewCurrentVerticalTiltAngle() *CurrentVerticalTiltAngle {
	char := NewInt(TypeCurrentVerticalTiltAngle)
	char.Format = FormatInt32
	char.Perms = []string{PermRead, PermEvents}
	char.SetMinValue(-90)
	char.SetMaxValue(90)
	char.SetStepValue(1)
	char.SetValue(-90)
	char.Unit = UnitArcDegrees

	return &CurrentVerticalTiltAngle{char}
}
