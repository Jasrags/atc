package aircraft

// NewDeparture creates a departure aircraft at a gate position.
func NewDeparture(callsign string, gateX, gateY int, gate string) Aircraft {
	return Aircraft{
		Callsign:     callsign,
		X:            float64(gateX),
		Y:            float64(gateY),
		Heading:      0,
		Altitude:     0, // on the ground
		Speed:        0,
		State:        AtGate,
		AssignedGate: gate,
	}
}

const (
	takeoffRollTicks = 15 // ticks on runway before liftoff (~1.5s)
	departAltitude   = 5  // climb to 5000ft after takeoff
	departSpeed      = 3  // departure speed
)

// TakeoffTick handles the OnRunway → Departing transition.
// Aircraft accelerates on the runway, then lifts off and transitions to Departing.
// Returns a new Aircraft.
func (a Aircraft) TakeoffTick() Aircraft {
	if a.State != OnRunway {
		return a
	}

	next := a
	next.tickCount++

	// Wait for takeoff roll to complete
	if next.tickCount < takeoffRollTicks {
		return next
	}

	// Liftoff — transition to Departing (airborne)
	next.State = Departing
	next.Altitude = 1
	next.TargetAltitude = departAltitude
	next.Speed = departSpeed
	next.TargetSpeed = departSpeed
	// tickCount resets for airborne Tick()
	next.tickCount = 0
	return next
}
