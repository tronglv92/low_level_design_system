// Designing a Parking Lot System

// The parking lot should have multiple levels, each level with a certain number of parking spots.
// The parking lot should support different types of vehicles, such as cars, motorcycles, and trucks.
// Each parking spot should be able to accommodate a specific type of vehicle.
// The system should assign a parking spot to a vehicle upon entry and release it when the vehicle exits.
// The system should track the availability of parking spots and provide real-time information to customers.
// The system should handle multiple entry and exit points and support concurrent access.

package main

import (
	"fmt"
	"sync"
	"time"
)

type VehicleType int

const (
	Car VehicleType = iota
	Motorcycle
	Truck
)

func (vt VehicleType) String() string {
	switch vt {
	case Car:
		return "Car"
	case Motorcycle:
		return "Motorcycle"
	case Truck:
		return "Truck"
	default:
		return "Unknown"
	}
}

type Vehicle interface {
	GetType() VehicleType
}
type CarVehicle struct{}

func (c CarVehicle) GetType() VehicleType {
	return Car
}

type TruckVehicle struct{}

func (tv TruckVehicle) GetType() VehicleType {
	return Truck
}

type MotorcycleVehicle struct{}

func (m MotorcycleVehicle) GetType() VehicleType {
	return Motorcycle
}

type ParkingSpot struct {
	SpotID     int
	SpotType   VehicleType
	IsOccupied bool
	mu         sync.Mutex // Lock for spot-level concurrency
}

func (ps *ParkingSpot) CanPark(vehicle Vehicle) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return !ps.IsOccupied && ps.SpotType == vehicle.GetType()
}

func (ps *ParkingSpot) Park() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.IsOccupied = true
}

func (ps *ParkingSpot) Leave() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.IsOccupied = false
}

// Observer interface
type Observer interface {
	Update(spot *ParkingSpot)
}
type ParkingLot struct {
	Levels    []*ParkingLevel
	observers []Observer
}

func (pl *ParkingLot) RegisterObserver(observer Observer) {

	pl.observers = append(pl.observers, observer)
}

func (pl *ParkingLot) RemoveObserver(observer Observer) {

	for i, obs := range pl.observers {
		if obs == observer {
			pl.observers = append(pl.observers[:i], pl.observers[i+1:]...)
			break
		}
	}
}

func (pl *ParkingLot) NotifyObservers(spot *ParkingSpot) {

	for _, observer := range pl.observers {
		observer.Update(spot)
	}
}

func (pl *ParkingLot) FindAndParkVehicle(vehicle Vehicle) (*ParkingSpot, *ParkingLevel) {
	for _, level := range pl.Levels {
		spot := level.FindAvailableSpot(vehicle)
		if spot != nil {
			level.ParkVehicle(spot, &vehicle)
			pl.NotifyObservers(spot)
			return spot, level
		}

	}
	return nil, nil
}

type ParkingLevel struct {
	LevelID int
	Spots   []*ParkingSpot
}

func (pl *ParkingLevel) FindAvailableSpot(vehicle Vehicle) *ParkingSpot {

	for _, sport := range pl.Spots {
		if sport.CanPark(vehicle) {
			return sport
		}
	}
	return nil
}
func (pl *ParkingLevel) ParkVehicle(spot *ParkingSpot, vehicle *Vehicle) {

	spot.Park()
}

func (pl *ParkingLevel) ReleaseVehicle(spot *ParkingSpot) {

	spot.Leave()
}

type EntryExitGate struct {
	ParkingLot *ParkingLot
}

func (gate *EntryExitGate) Enter(vehicle Vehicle) {
	spot, level := gate.ParkingLot.FindAndParkVehicle(vehicle)
	if spot != nil {
		fmt.Printf("Vehicle %v parked at Level %d, Spot %d \n", vehicle.GetType(), level.LevelID, spot.SpotID)
	} else {
		fmt.Printf("No available sport for this vehicle type")
	}
}

func (gate *EntryExitGate) Exit(level *ParkingLevel, spot *ParkingSpot) {
	level.ReleaseVehicle(spot)
	fmt.Printf("Vehicle exited from Level %d, Spot %d \n", level.LevelID, spot.SpotID)
}

// Concrete Observer
type ParkingStatusObserver struct {
	name string
}

func (observer *ParkingStatusObserver) Update(spot *ParkingSpot) {
	status := "Occupied"
	if !spot.IsOccupied {
		status = "Available"
	}
	fmt.Printf("%s - Spot %d:%s\n", observer.name, spot.SpotID, status)
}

// Real-time Parking Status
func (pl *ParkingLot) GetParkingStatus() {
	for _, level := range pl.Levels {
		fmt.Printf("Level %d:\n", level.LevelID)
		for _, spot := range level.Spots {
			spotStatus := "Available"
			if spot.IsOccupied {
				spotStatus = "Occupied"
			}
			fmt.Printf("Spot %d: %s\n", spot.SpotID, spotStatus)
		}
	}
}

func main() {
	level1 := &ParkingLevel{
		LevelID: 1,
		Spots: []*ParkingSpot{
			{SpotID: 1, SpotType: Car, IsOccupied: false},
			{SpotID: 2, SpotType: Motorcycle, IsOccupied: false},
			{SpotID: 3, SpotType: Car, IsOccupied: false},
			{SpotID: 4, SpotType: Truck, IsOccupied: false},
		},
	}

	parkingLot := &ParkingLot{
		Levels: []*ParkingLevel{
			level1,
		},
	}
	// Create entry and exit gates
	gate := &EntryExitGate{ParkingLot: parkingLot}

	// Create observers for parking status
	observer1 := &ParkingStatusObserver{name: "Observer 1"}

	// Register observers to the parking lot
	parkingLot.RegisterObserver(observer1)

	// Show parking status before vehicles enter
	fmt.Println("Parking Lot Status Before Vehicles Enter:")
	parkingLot.GetParkingStatus()

	// Create a WaitGroup to handle concurrent operations
	var wg sync.WaitGroup

	// Test Case 1: Multiple Vehicles Entering Simultaneously
	fmt.Println("Test Case 1: Multiple Vehicles Entering Simultaneously")
	var wg1 sync.WaitGroup
	wg1.Add(3)

	go func() {
		defer wg1.Done()
		car := CarVehicle{}
		gate.Enter(car)
	}()

	go func() {
		defer wg1.Done()
		motorcycle := MotorcycleVehicle{}
		gate.Enter(motorcycle)
	}()

	go func() {
		defer wg1.Done()
		truck := TruckVehicle{}
		gate.Enter(truck)
	}()
	wg1.Wait()
	parkingLot.GetParkingStatus()

	// Test Case 2: Multiple Vehicles Exiting Simultaneously
	fmt.Println("\nTest Case 2: Multiple Vehicles Exiting Simultaneously")
	var wg2 sync.WaitGroup
	wg2.Add(2)

	go func() {
		defer wg2.Done()
		gate.Exit(level1, level1.Spots[0]) // Assume car is in Spot 1
	}()

	go func() {
		defer wg2.Done()
		gate.Exit(level1, level1.Spots[1]) // Assume motorcycle is in Spot 2
	}()
	wg2.Wait()
	parkingLot.GetParkingStatus()

	// Test Case 3: Mixed Entries and Exits
	fmt.Println("\nTest Case 3: Mixed Entries and Exits")
	var wg3 sync.WaitGroup
	wg3.Add(4)

	go func() {
		defer wg3.Done()
		car := CarVehicle{}
		gate.Enter(car)
	}()

	go func() {
		defer wg3.Done()
		motorcycle := MotorcycleVehicle{}
		gate.Enter(motorcycle)
	}()

	go func() {
		defer wg3.Done()
		gate.Exit(level1, level1.Spots[2]) // Assume truck is in Spot 3
	}()

	go func() {
		defer wg3.Done()
		gate.Exit(level1, level1.Spots[3]) // Assume car is in Spot 4
	}()
	wg3.Wait()
	parkingLot.GetParkingStatus()

	// Test concurrent exits for occupied spots
	fmt.Println("\nSimulating Concurrent Vehicle Exits:")
	wg.Add(2)

	// Test Case 4: High Concurrency with More Vehicles than Spots
	fmt.Println("\nTest Case 4: High Concurrency with More Vehicles than Spots")
	var wg4 sync.WaitGroup
	for i := 0; i < 10; i++ { // 10 vehicles trying to enter
		wg4.Add(1)
		go func(vehicleType VehicleType) {
			defer wg4.Done()
			if vehicleType == Car {
				gate.Enter(CarVehicle{})
			} else if vehicleType == Motorcycle {
				gate.Enter(MotorcycleVehicle{})
			} else {
				gate.Enter(TruckVehicle{})
			}
		}(VehicleType(i % 3)) // Cycling through Car, Motorcycle, and Truck
	}
	wg4.Wait()
	parkingLot.GetParkingStatus()

	// Test Case 5: Rapid Entry and Exit on the Same Spot
	fmt.Println("\nTest Case 5: Rapid Entry and Exit on the Same Spot")
	var wg5 sync.WaitGroup
	wg5.Add(2)

	go func() {
		defer wg5.Done()
		for i := 0; i < 5; i++ {
			car := CarVehicle{}
			gate.Enter(car)
			time.Sleep(100 * time.Millisecond) // Short delay to simulate real scenario
			gate.Exit(level1, level1.Spots[0]) // Assume the car parked in Spot 1
		}
	}()

	go func() {
		defer wg5.Done()
		for i := 0; i < 5; i++ {
			motorcycle := MotorcycleVehicle{}
			gate.Enter(motorcycle)
			time.Sleep(100 * time.Millisecond)
			gate.Exit(level1, level1.Spots[1]) // Assume the motorcycle parked in Spot 2
		}
	}()
	wg5.Wait()
	parkingLot.GetParkingStatus()
	// Show parking status after concurrent entries
	fmt.Println("\nParking Lot Status After Concurrent Entries:")
	parkingLot.GetParkingStatus()

	// Remove observers after usage
	parkingLot.RemoveObserver(observer1)

}
