package joystick

import (
	"HomeRover/models/config"
	"HomeRover/utils"
	"fmt"
	"github.com/karalabe/hid"
	"strconv"
	"sync"
	"time"
)

type Joystick struct {
	Conf     *config.ControllerConfig
	confMu   sync.RWMutex
	deviceMu sync.RWMutex
	Data     *chan []byte
	device   *hid.Device
}

func NewJoystick(conf *config.ControllerConfig, data *chan []byte) (js *Joystick, err error) {
	js = &Joystick{
		Conf:     conf,
		Data:     data,
	}
	return
}

func (js *Joystick) Init() error {
	devices := hid.Enumerate(0, 0)
	var controller *hid.Device

	for _, device := range devices {
		if device.ProductID == 654 {
			var err error
			controller, err = device.Open()
			if err != nil {
				return err
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
			return err
		}
		selectNum, err := strconv.Atoi(selectNumStr)
		if err != nil {
			return err
		}
		fmt.Printf("%s select\n", devices[selectNum].Product)
		controller, err = (devices[selectNum]).Open()
		if err != nil {
			return err
		}
	}
	js.device = controller
	return nil
}

func (js *Joystick) ReadOnce() ([]byte, error) {
	data := make([]byte, 14, 100)

	js.deviceMu.Lock()
	_, err := js.device.Read(data)
	js.deviceMu.Unlock()

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

func (js *Joystick) Run()  {
	js.confMu.Lock()
	defer js.confMu.Unlock()
	freq := js.Conf.JoystickFreq
	for range time.Tick(time.Duration(1000000000 / freq)){
		deviceData, err := js.ReadOnce()
		if err != nil {
			continue
		}
		*js.Data <- deviceData
	}
}