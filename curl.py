import requests

url = 'http://localhost:8789/'

res = requests.get(url)
print(res.text)

res = requests.get(url)
print(res.text)
