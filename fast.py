from fastapi import FastAPI
import uvicorn
import time

app = FastAPI()

@app.get("/")
def read_root():
    time.sleep(5)
    return {"Hello": "World"}

if __name__ == "__main__":
    uvicorn.run(app=app, host='0.0.0.0', port=8789)
