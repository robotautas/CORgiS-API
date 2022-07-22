import json
import requests
from time import sleep

json_1 = '''
[
    {
        "Vxx": {"V08": [[7, 1]]},
        "Txx": {"T21": 100},
        "PUMP": "ON",
        "Sleep": 20
    },
    {
        "Vxx": {"V08": [[7, 0], [6, 1]]},
        "PUMP": "ON",
        "Sleep": 20
    }
]'''
json_2 = '''
[
    {
        "Vxx": {"V08": [[7, 1]]},
        "PUMP": "ON",
        "Sleep": 20
    },
        {
        "Vxx": {"V08": [[6, 0]]},
        "PUMP": "ON",
        "Sleep": 20
    }
]'''
json_3 = '''
[
    {
        "Vxx": {"V08": [[4, 1]]},
        "Txx": {"T03": 300},
        "PUMP": "OFF",
        "Sleep": 20
    },
        {
        "Vxx": {"V08": [[4, 0], [6, 1]]},
        "Txx": {"T03": 300},
        "PUMP": "OFF",
        "Sleep": 20
    }
]'''
json_4 = '''
[
    {
        "Vxx": {"V08": [[5, 1]]},
        "Txx": {"T03": 300},
        "PUMP": "OFF",
        "Sleep": 20
    },
        {
        "Vxx": {"V08": [[7, 1]]},
        "Txx": {"T03": 300},
        "PUMP": "OFF",
        "Sleep": 20
    }
]'''

instructions = [json.loads(s) for s in [json_1, json_2, json_3, json_4]]

def send(instruction):
    print(instruction)
    r = requests.post("http://127.0.0.1:9999/start", json=instruction)
    print(r.text)


# Proceduuuura :)

send(instructions[0])
sleep(8)
# send(instructions[1])
# sleep(8)
# send(instructions[2])
# sleep(8)
# send(instructions[3])
# sleep(3)





