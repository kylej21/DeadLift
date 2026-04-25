from fastapi import FastAPI, status

app = FastAPI()

@app.get("/health", tags=["health"])
def health_check():
    """
    Returns a 200 OK status to indicate the server is running.
    """
    return {"status": "ok", "message": "server is healthy"}

@app.get("/")
def root():
    return {"message": "Hello World"}

