from hash import jenkins_one_at_a_time as hash
from flask import Flask, jsonify, json, request
from werkzeug.exceptions import HTTPException
import sys

app = Flask(__name__)


# configura o flask para retornar JSON ao invés de HTML em erros
@app.errorhandler(HTTPException)
def handle_exception(e):
    response = e.get_response()
    response.data = json.dumps(
        {
            "code": e.code,
            "name": e.name,
            "description": e.description,
        }
    )
    response.content_type = "application/json"
    return response


with open("words.txt", "r") as f:
    data = [x.strip() for x in f]

db_open_positions = []
db_configs = []


# a nomenclatura das funções é inspirado no ruby on rails
@app.get("/db")
def db_index():
    print(db_configs, file=sys.stderr)
    return list(filter(lambda x: x is not None, db_configs))


@app.post("/db")
def db_create():
    """
    page_size:       uint | null
    number_of_pages: uint | null
    FR:              uint
    (mutualmente exclusivos)
    """

    try:
        body = request.get_json()
        print(f"Request body: {body}", file=sys.stderr)
        page_size = body["page_size"]
        number_of_pages = body["number_of_pages"]
        FR = body["FR"]
    except KeyError as error:
        return {
            "error": f"Expected {error} in body of request",
        }

    # garante que apenas 1 dos valores está definido
    if (page_size is None and number_of_pages is None) or (
        page_size is int and number_of_pages is int
    ):
        return {
            "error": "'page_size' and 'number_of_pages' are mutually exclusive",
        }

    new_db_config = {
        "id": len(db_configs),
        "page_size": page_size,
        "number_of_pages": number_of_pages,
        "FR": FR,
    }

    if len(db_open_positions) == 0:
        db_configs.append(new_db_config)
    else:
        position = db_open_positions.pop()
        new_db_config["id"] = position
        db_configs[position] = new_db_config

    print(f"Response body: {json.dumps(new_db_config)}")
    return new_db_config


@app.get("/db/<int:id>")
def db_show(id):
    if id >= len(db_configs):
        return {
            "error": "Config is not in system",
        }

    if db_configs[id] is None:
        return {
            "error": "This config was deleted",
        }

    return db_configs[id]


@app.delete("/db/<int:id>")
def db_destroy(id):
    if id >= len(db_configs):
        return {
            "error": "Config is not in system",
        }

    if db_configs[id] is None:
        return {
            "error": "This config was deleted",
        }

    deleted_config = db_configs[id]
    db_configs[id] = None
    if deleted_config["id"] < len(db_configs) - 1:
        db_open_positions.append(id)

    return deleted_config


print(app.url_map)
