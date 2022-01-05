import json
import requests

def send_post():
    all_instructions = []
    
    while True:
        instruction = {}
        while True:
            param = input('parameter: ')
            if param.startswith("V"):
                bit_changes = []
                while True:
                    bit_changes.append(
                        [
                            int(input("Bit index: ")),
                            int(input("Bit value: ")),
                        ]
                    )
                    if input('more bits?: ') == 'n':
                        instruction[param] = bit_changes
                        break
            else:
                instruction[param] = int(input("Value: "))
            if input('more parameters?: ') == 'n':
                instruction['sleep'] = int(input('sleep?: '))
                all_instructions.append(instruction)
                break
        if input('more instructions?: ') == 'n':
                break





    to_json = json.dumps(all_instructions, indent=2)
    print(to_json)

    r = requests.post("http://127.0.0.1:9999/start", json=all_instructions)
    print(r.text)
# send_post()


fast_json = '''
[
    {
        "Vxx": {
            "V00": [[0, 1], [3, 0], [7, 1]],
            "V01": [[0, 0]]
        },
        "Txx": {
            "T01": 255
        },
        "PUMP": "ON",
        "Sleep": 10
    },
    {
        "Vxx":{"V00": [[3, 1]]},
        "Txx": {"T03": 300},
        "PUMP": "OFF",
        "Sleep": 30
    } 
]'''
fast_dict = json.loads(fast_json)
print(fast_dict)
r = requests.post("http://127.0.0.1:9999/start", json=fast_dict)
print(r.text)