import requests

session = requests.Session()

resp = session.post("http://localhost:8080/login", json={"username": "test", "password": "test"})

print(f"Login: {resp.json()}")

for i in range(10):
    resp = session.get("http://localhost:8080/profile")
    print(f"Request {i+1}: Server={resp.json()['server_id']}")