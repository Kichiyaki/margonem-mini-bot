package utils

import (
	"fmt"

	"github.com/denisbrodbeck/machineid"
)

func CheckMachineID(expected string) error {
	if expected == "*" {
		return nil
	}
	id, err := machineid.ProtectedID("margonem-mobile-app-bot")
	if err != nil {
		return err
	}
	if id != expected {
		return fmt.Errorf("Wrong machine id")
	}
	return nil
}
