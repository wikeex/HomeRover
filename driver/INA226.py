import smbus


class INA226:
    # Registers/etc.
    __configuration_register = 0x00
    __shunt_voltage_register = 0x01
    __bus_voltage_register = 0x02
    __power_register = 0x03
    __current_register = 0x04
    __calibration_register = 0x05
    __enable_register = 0x06
    __alert_limit_register = 0x07
    __manufacturer_id_register = 0xfe
    __die_id_register = 0xff

    def __init__(self, address=0x44, debug=False):
        self.bus = smbus.SMBus(1)
        self.address = address
        self.debug = debug
        if self.debug:
            print("Reseting INA226")
        self.write(self.__configuration_register, [0x48, 0x4f])

    def write(self, reg, value):
        """Writes an 8-bit value to the specified register/address"""
        self.bus.write_i2c_block_data(self.address, reg, value)
        if self.debug:
            print("I2C: Write 0x%02X to register 0x%02X" % (value, reg))

    def read(self, reg):
        """Read an unsigned byte from the I2C device"""
        result = self.bus.read_i2c_block_data(self.address, reg, 2)
        if self.debug:
            print("I2C: Device 0x%02X returned 0x%02X from reg 0x%02X" % (self.address, result & 0xFF, reg))
        return result


if __name__ == '__main__':
    ina226 = INA226()
    current_register = ina226.read(0x01)
    print(current_register)

