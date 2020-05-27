from PCA9685 import PCA9685
import socket

BUF_SIZE = 548

CAM_X = 0
CAM_Y = 1
LEFT_MOTOR = 2
RIGHT_MOTOR = 3


def electric_differential(x, y, max_difference=500):
    base_pulse = y / 128.0 * 1000 + 1500
    difference = x / 128.0 * max_difference

    return base_pulse - difference, base_pulse + difference


def drive():
    pwm = PCA9685()
    pwm.setPWMFreq(50)
    pwm.setRotationAngle(CAM_X, 180)
    pwm.setRotationAngle(CAM_Y, 180)
    pwm.setServoPulse(3000 - LEFT_MOTOR, 1501)
    pwm.setServoPulse(RIGHT_MOTOR, 1501)

    listen_addr = ('127.0.0.1', 10008)
    conn = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    conn.bind(listen_addr)

    while True:
        data, _ = conn.recvfrom(BUF_SIZE)
        left_x = int.from_bytes(data[:1], byteorder='little', signed=True)
        left_y = int.from_bytes(data[1:2], byteorder='little', signed=True)
        right_x = int.from_bytes(data[2:3], byteorder='little', signed=True)
        right_y = int.from_bytes(data[3:], byteorder='little', signed=True)

        left_motor, right_motor = electric_differential(left_x, left_y)
        pwm.setServoPulse(LEFT_MOTOR, left_motor)
        pwm.setServoPulse(RIGHT_MOTOR, right_motor)

        pwm.setRotationAngle(CAM_X, 160 - (right_x / 128.0 * 70 + 90))
        pwm.setRotationAngle(CAM_Y, min(right_y / 128.0 * 70 + 150, 180))


if __name__ == '__main__':
    drive()
