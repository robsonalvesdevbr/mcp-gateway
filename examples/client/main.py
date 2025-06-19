import requests
import sseclient
import json

payload = {
    "tool": "search",
    "input": {"query": "Docker"}
}

response = requests.post("http://gateway:8811/mcp", json=payload, stream=True)

client = sseclient.SSEClient(response)

for event in client.events():
    print(f"Event: {event.event}, Data: {event.data}")
    if event.event == "done":
        break