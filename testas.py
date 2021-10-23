import requests
from time import sleep
import json

fail_counter = 0
for i in range(256):
    r = requests.get(f'http://127.0.0.1:9999/set?param=V00&value={i}')
    results = json.loads(r.text)
    if results["V00"] == i:
        print(f'Request #{i} OK. "V00"={results["V00"]}')
    else:
        print(f'Request #{i} FAILED. Response begins with {r.text[:20]}...')
        fail_counter += 1
    sleep(1)
print(f'Total {fail_counter} fails of 256 requests.')


    