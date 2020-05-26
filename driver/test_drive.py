import unittest
import socket
from threading import Thread
import time
from drive import drive


class TestDrive(unittest.TestCase):
    def test_drive(self):
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        t = Thread(target=drive)
        t.start()

        value = 0
        while True:
            if value == 255:
                value = 0

            data = bytes([value] * 4)
            s.sendto(data, ('127.0.0.1', 10008))
            print(data)

            value += 1
            time.sleep(0.2)
