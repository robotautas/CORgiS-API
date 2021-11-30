import json
import requests

def send_post():
    all_instructions = []
    
    while True:
        instruction = {}    
        instruction['param'] = input("ivesk parametra: ")
        instruction['value'] = input("ivesk verte: ")
        instruction['sleep'] = input("ivesk pauze: ")
        all_instructions.append(instruction)
        end = input('dar? ')
        if end == 'ne':
            break
        
    to_json = json.dumps(all_instructions)
    print(to_json)
    r = requests.post("http://127.0.0.1:9999/start", json=all_instructions)
    print(r.text)
send_post()


    