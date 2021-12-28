from random import randint, choice
import requests
import json
from time import sleep

# Vs = ["V00", "V01", "V02", "V03", "V04", "V05", "V06", "V07", "V08"]
Ts = ["T01", "T02", "T03", "T04", "T05", "T06", "T07", "T08"]
Ps = ["PUMP_ON", "PUMP_OFF"]

def generate_json():
    number_of_tasks = randint(1, 3)
    instruction = []
    for i in range(number_of_tasks):
        task = {}
        task['Vxx'] = {}
        for i in range(randint(1, 2)):
            Vs = ["V00", "V01", "V02", "V03", 
            "V04", "V05", "V06", "V07", "V08"]
            param = choice(Vs)
            Vs.remove(param)
            task['Vxx'][param] = []
            channels = list(range(8))
            for i in range(randint(1,3)):
                channel = choice(channels)
                channels.remove(channel)
                task['Vxx'][param].append([channel, randint(0,1)])
        task['Txx'] = {'T01': 100}
        if randint(1, 5) == 1:
            task["PUMP"] = choice(['ON', 'OFF'])
        task['Sleep'] = randint(3, 20)    
        instruction.append(task)
    # json_string = json.loads(instruction)
    for task in instruction:
        for k in task:
            if k in ["PUMP", "Sleep"]:
                print(k, ': ', task[k])
            else:
                print(task[k])
        print("")
    # print(json_string)
    return instruction




for i in range(100):
    send = generate_json()
    r = requests.post("http://127.0.0.1:9999/start", json=send)
    
    print(r.text)
    sleep(3)