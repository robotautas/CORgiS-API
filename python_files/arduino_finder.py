

# from influxdb import InfluxDBClient
from datetime import datetime
from serial import Serial
import serial.tools.list_ports
import re
from time import sleep

# def arduino_finder():
#     ''' Finds on which port is arduino '''
#     SysFs_objects = list(serial.tools.list_ports.comports())
#     ports = [str(port) for port in SysFs_objects]
#     for port in ports:
#         if "Nano 33 BLE" in port:
#             arduino_port = port.split(' - ')[0]
#             return arduino_port

def arduino_finder():
    ''' Finds on which port is arduino '''
    print('Looking for arduino...\n')
    while True:    
        SysFs_objects = list(serial.tools.list_ports.comports())
        ports = [str(port) for port in SysFs_objects]
        for obj in SysFs_objects:
            print(obj.pid)
        for port in ports:
            
            if "Nano 33 BLE" in port:
                arduino_port = port.split(' - ')[0]
                print(f'Found arduino on {arduino_port}.')
                return arduino_port
        else:
            print("\033[A                             \033[A")
            print("Can't find the device, waiting for connection!")
            sleep(1)


arduino = serial.Serial(port='/dev/ttyACM0', baudrate=115200, timeout=.1)

# print(arduino.)
arduino_finder()