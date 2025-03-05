from hash import jenkins_one_at_a_time as hash
from flask import Flask, jsonify, json
from werkzeug.exceptions import HTTPException

app = Flask(__name__)


# configura o flask para retornar JSON ao invés de HTML em erros
@app.errorhandler(HTTPException)
def handle_exception(e):
    response = e.get_response()
    response.data = json.dumps({
        "code": e.code,
        "name": e.name,
        "description": e.description,
    })
    response.content_type = "application/json"
    return response


db_collection = [
    123
]


# a nomenclatura das funções é inspirado no ruby on rails
@app.get('/db')
def db_index():
    return db_collection


@app.post('/db')
def db_create():
    new_db = {
        "db_id": len(db_collection)
    }


@app.get('/db/<int:db_id>')
def db_show(db_id):
    if db_id >= len(db_collection):
        return {
            "error": "db is not in collection"
        }
    return db_collection[db_id]


@app.delete('/db/<int:db_id>')
def db_destroy(db_id):
    if db_id >= len(db_collection):
        return {
            "error": "db is not in collection"
        }
    return db_collection.pop(db_id)
