FROM python:3.12-alpine

RUN pip install requests sseclient-py

WORKDIR /app
COPY main.py ./
ENTRYPOINT [ "python", "main.py" ]