package joystick

import (
	"HomeRover/utils"
	"fmt"
	"github.com/karalabe/hid"
	"strconv"
)

func GetJoystick() (*hid.Device, error) {
	devices := hid.Enumerate(0, 0)
	var controller *hid.Device

	for _, device := range devices {
		if device.ProductID == 654 {
			var err error
			controller, err = device.Open()
			if err != nil {
				return nil, err
			}
		}
	}

	if controller == nil {
		for index, device := range devices {
			fmt.Printf("%d. Manufacturer: %s, Product: %s;", index, device.Manufacturer, device.Product)
			fmt.Println()
		}
		var selectNumStr string
		_, err := fmt.Scanln(&selectNumStr)
		if err != nil {
			return nil, err
		}
		selectNum, err := strconv.Atoi(selectNumStr)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s select\n", devices[selectNum].Product)
		controller, err = (devices[selectNum]).Open()
		if err != nil {
			return nil, err
		}
	}

	return controller, nil
}

func ReadOnce(controller *hid.Device) ([]byte, error) {
	data := make([]byte, 14, 100)

	_, err := controller.Read(data)
	if err != nil {
		return nil, err
	}

	// left stick x aix
	leftStickX, err := utils.BytesToInt(data[6:8])
	if err != nil {
		return nil, err
	}

	// left stick y aix
	leftStickY, err := utils.BytesToInt(data[8:10])
	if err != nil {
		return nil, err
	}

	// right stick x aix
	rightStickX, err := utils.BytesToInt(data[10:12])
	if err != nil {
		return nil, err
	}

	// right stick y aix
	rightStickY, err := utils.BytesToInt(data[12:])
	if err != nil {
		return nil, err
	}

	return []byte{
		byte(int8(leftStickX >> 8)),
		byte(int8(leftStickY >> 8)),
		byte(int8(rightStickX >> 8)),
		byte(int8(rightStickY >> 8)),
	}, nil
}