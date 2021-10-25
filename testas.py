import requests
from time import sleep
import json
from random import randint

# fail_counter = 0
# for i in range(256):
#     r = requests.get(f'http://127.0.0.1:9999/set?param=V00&value={i}')
#     results = json.loads(r.text)
#     if results["V00"] == i:
#         print(f'Request #{i} OK. "V00"={results["V00"]}')
#     else:
#         print(f'Request #{i} FAILED. Response begins with {r.text[:20]}...')
#         fail_counter += 1
#     sleep(1)
# print(f'Total {fail_counter} fails of 256 requests.')

PARAMS = ["V00", "V01", "V02", "V03", "V04", "V05", "V06", "V07", "V08", 
          "T01", "T02", "T03", "T04", "T05", "T06", "T07", "T08",
          "PUMP_ON", "PUMP_OFF"]

URL = "http://127.0.0.1:9999/set"

def generate_params():
    param = PARAMS[randint(0, len(PARAMS))-1]
    if param.startswith("V"):
        value = randint(0, 255)
        command = f'<SET_{param}_{value};>'
    elif param.startswith("T"):
        value = randint(0, 999)
        command = f'<SET_{param}_{value};>'
    else:
        value = None
        command = f'<{param}>'

    return param, value, command


def send_request():
    param, value, command = generate_params()
    print(param, value)
    payload = {'param': param, 'value': value} if value != None else {'param': param}
    
    r = requests.get(URL, params=payload)

    results = json.loads(r.text)
    
    if param.startswith("V"):
        if results[param] == value:
            print(f'OK. Command {command} sent. {param}={value}')
        else:
            print(f'FAIL! Command {command} sent. {param}!={value}, actual value is {results[param]}')
    
    elif param.startswith("T"):
        acceptable_vals = range(value, value + 3)
        if results[param] in acceptable_vals:
            print(f'OK. Command {command} sent. {param}={results[param]}')
        else:
            print(f'FAIL! Command {command} sent.{param}!={value}, actual value is {results[param]}')
    else:
        acceptable_value = 1 if param.endswith('ON') else 0
        if results["PUMP"] == acceptable_value:
            print(f'OK. Command {command} sent. PUMP={acceptable_value}')
        else:
            print(f'FAIL! Command {command} sent. PUMP!={acceptable_value}, actual value is {results["PUMP"]}')



for i in range(1200):
    send_request()
    sleep(1)
