FROM python:3.8.9-slim
RUN mkdir /metric

copy . /metric

WORKDIR /metric
RUN pip3 install -r requirements.txt

ENTRYPOINT ["python3", "main.py"]
