import json
import requests
from time import sleep

json_1 = '''
[
    {
        "Vxx": {"V08": [[7, 1]]},
        "Txx": {"T01": 100},
        "PUMP": "ON",
        "Sleep": 4
    }
]'''
json_2 = '''
[
    {
        "Vxx": {"V08": [[7, 0]]},
        "Txx": {"T01": 200},
        "PUMP": "OFF",
        "Sleep": 4
    }
]'''
json_3 = '''
[
    {
        "Vxx": {"V08": [[0, 1]]},
        "Txx": {"T05": 300},
        "PUMP": "OFF",
        "Sleep": 4
    }
]'''
json_4 = '''
[
    {
        "Txx": {"T05": 300},
        "PUMP": "ON",
        "Sleep": 4
    }
]'''

instructions = [json.loads(s) for s in [json_1, json_2, json_3, json_4]]

def send(instruction):
    print(instruction)
    r = requests.post("http://127.0.0.1:9999/start", json=instruction)
    print(r.text)


# Proceduuuura :)

send(instructions[3])
sleep(3)
send(instructions[2])
sleep(3)
send(instructions[3])
sleep(3)





