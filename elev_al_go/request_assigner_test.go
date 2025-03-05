package elevalgo

import (
	"testing"
)

func TestTimeToIdle(t *testing.T) {
	c := config{1, 3}

	e := Elevator{
		floor:     0,
		direction: stop, /// ShouldStop func gjør at døren åpnes en unødvendig gang. Døren åpnes derfor 2 ganger i test 2
		behaviour: idle,
		config:    c,
	}
	time := timeToIdle(e)
	if time != 0 {
		t.Errorf("Expected timeToIdle to be 0, got %d", time)
	}

	e.Requests[2][1] = true // Request at floor 2, hall down
	time = timeToIdle(e)
	if time != 10 {
		t.Errorf("Expected timeToIdle to be 10, got %d", time)
	}
}

func TestPreferredOrder(t *testing.T) {
	c := config{1, 3}
	testCases := []struct {
		name          string
		elevators     []Elevator
		expectedOrder []int
	}{
		{
			name: "Two elevators, one idle and one moving, order in the middle.",
			elevators: []Elevator{
				{
					floor:     2,
					direction: down,
					behaviour: moving,
					config:    c,
				},
				{
					floor:     0,
					direction: stop,
					behaviour: idle,
					config:    c,
				},
			},
			expectedOrder: []int{0, 1},
		},
		{
			name: "Two elevators, one with a cab call and one idle, double hall order.",
			elevators: []Elevator{
				{
					floor:     0,
					direction: stop,
					behaviour: idle,
					config:    c,
				},
				{
					floor:     1,
					direction: up,
					behaviour: moving,
					config:    c,
				},
			},
			expectedOrder: []int{0, 1},
		},
	}
	// Case 0:
	// Incoming Order:
	testCases[0].elevators[0].Requests[1][0] = true // Request at floor 1, hall up
	testCases[0].elevators[1].Requests[1][0] = true // Request at floor 1, hall up

	// Case 1:
	// Start State:
	testCases[1].elevators[1].Requests[3][2] = true // Cab call at floor 4 (index3)
	testCases[1].elevators[1].Requests[2][0] = true // Request at floor 3, hall up
	//Incoming Order:
	testCases[1].elevators[0].Requests[2][1] = true // Request at floor 2, hall down
	testCases[1].elevators[1].Requests[2][1] = true // Request at floor 2, hall down

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pOrder := preferredOrder(tc.elevators)
			for i, v := range pOrder {
				if v != tc.expectedOrder[i] {
					t.Errorf("Expected preferredOrder[%d] to be %d, got %d", i, tc.expectedOrder[i], v)
				}
			}
		})
	}
}
