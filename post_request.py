import json
import requests

def send_post():
    all_instructions = []
    
    while True:
        instruction = {}
        commands = {}
        while True:
            param = input('parameter: ')
            commands[param] = int(input('value: '))
            if input('more?: ') == 'n':
                instruction['commands'] = commands
                instruction['sleep'] = int(input('sleep: '))
                break
        all_instructions.append(instruction)
        if input("one more instruction? ") == 'n':
            break




    to_json = json.dumps(all_instructions, indent=2)
    print(to_json)
    r = requests.post("http://127.0.0.1:9999/start", json=all_instructions)
    print(r.text)
send_post()


    