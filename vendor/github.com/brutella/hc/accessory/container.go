package accessory

// Container manages a list of accessories.
type Container struct {
	Accessories []*Accessory `json:"accessories"`

	idCount int64
}

// NewContainer returns a container.
func NewContainer() *Container {
	return &Container{
		Accessories: make([]*Accessory, 0),
		idCount:     1,
	}
}

// AddAccessory adds an accessory to the container.
// This method ensures that the accessory ids are valid and unique withing the container.
func (m *Container) AddAccessory(a *Accessory) {
	a.SetID(m.idCount)
	m.idCount++
	m.Accessories = append(m.Accessories, a)
}

// RemoveAccessory removes an accessory from the container.
func (m *Container) RemoveAccessory(a *Accessory) {
	for i, accessory := range m.Accessories {
		if accessory == a {
			m.Accessories = append(m.Accessories[:i], m.Accessories[i+1:]...)
		}
	}
}

// Equal returns true when receiver has the same accessories as the argument.
func (m *Container) Equal(other interface{}) bool {
	if container, ok := other.(*Container); ok == true {
		if len(m.Accessories) != len(container.Accessories) {
			return false
		}

		for i, a := range m.Accessories {
			if a.Equal(container.Accessories[i]) == false {
				return false
			}
		}
		return true
	}

	return false
}

// AccessoryType returns the accessory type identifier for the accessories inside the container.
func (m *Container) AccessoryType() AccessoryType {
	if as := m.Accessories; len(as) > 0 {
		if len(as) > 1 {
			return TypeBridge
		}

		return as[0].Type
	}

	return TypeOther
}
