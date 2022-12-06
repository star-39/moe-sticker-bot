import emoji
from flask import Flask, request

# This simple python microservice utilises
# 'emoji' package in PyPI which has great funcionality in
# parsing and extracting complicated emojis from string.

# This service should be served on port 5000.

app = Flask(__name__)

@app.route("/", methods=['POST'])
def get_emoji():
    body = request.data.decode("utf-8")
    print(body)
    e = emoji.distinct_emoji_list(body)
    print(e)
    return ''.join(e)

@app.route("/status",methods=['GET'])
def rep_status():
    return 'OK'
