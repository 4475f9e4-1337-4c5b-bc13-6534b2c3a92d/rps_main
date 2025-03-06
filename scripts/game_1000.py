import requests
import re
import websocket
import threading
import signal

URL = "http://localhost:9000/play"
DATA = {"bestOf": "1"}
stop = threading.Event()

def on_close(ws, close_status_code, close_msg):
    ws.close()

def connect_ws(id):
    print(id)
    ws = websocket.WebSocketApp("ws://localhost:9000/game/"+str(id), on_close=on_close)
    while not stop.is_set():
        ws.run_forever()

for i in range(1000):
    response = requests.post(URL, data=DATA)
    match = re.search(r'ws-connect="/game/([a-f0-9\-]+)"', response.text)
    if match:
        id = match.group(1)
        threading.Thread(target=connect_ws, args=(id,)).start()
    else:
        print("No match found")


def signal_handler(sig, frame):
    print('You pressed Ctrl+C!')
    stop.set()
    exit(0)

signal.signal(signal.SIGINT, signal_handler)
signal.pause()
