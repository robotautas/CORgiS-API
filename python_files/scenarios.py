import json
import requests
from time import sleep

json_1 = '''
[
    {
        "Vxx": {"V08": [[7, 1], [6, 1]],
                "V07": [[7, 1], [6, 1]]},
        "Txx": {"T01": 100},
        "PUMP": "ON",
        "Sleep": 10
    }
]'''
json_2 = '''
[
    {
        "Vxx": {"V08": [[6, 1]]},
        "Txx": {"T02": 200},
        "PUMP": "OFF",
        "Sleep": 6
    }
]'''
json_3 = '''
[
    {
        "Vxx": {"V08": [[5, 1]]},
        "Txx": {"T03": 300},
        "PUMP": "OFF",
        "Sleep": 6
    }
]'''
json_4 = '''
[
    {
        "Txx": {"T04": 400, "T05": 400, "T06": 400, "T07": 400},
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

send(instructions[0])
# sleep(3)
# send(instructions[1])
# sleep(3)
# send(instructions[2])
# sleep(3)
# send(instructions[3])
# sleep(3)





