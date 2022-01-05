import requests

def stop(number):
    r = requests.get(f'http://127.0.0.1:9999/stop?id={number}')

stop(input("id?: "))

