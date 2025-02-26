from hash import jenkins_one_at_a_time as hash
from flask import Flask

app = Flask(__name__)


@app.route("/")
def hello_world():
    return {
        "message": "hiiii"
    }
