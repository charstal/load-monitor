FROM python:3.8.9-slim
RUN mkdir /metric

copy . /metric

WORKDIR /metric
RUN pip3 config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple
RUN pip3 install -r requirements.txt 
ENV PROTOCOL_BUFFERS_PYTHON_IMPLEMENTATION=python
ENTRYPOINT ["python3", "main.py"]
