import os
import sys
from flask import Flask, jsonify
from .models import User
from .config import settings

app = Flask(__name__)

class Application:
    def __init__(self):
        self.app = app

def create_app():
    return app

def _internal_helper():
    pass

if __name__ == "__main__":
    app.run()
