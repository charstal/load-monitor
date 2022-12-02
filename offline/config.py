import os
import logging

logging.basicConfig(level=logging.INFO)
PROMETHEUS_URL = os.environ.get(
    "PROMETHEUS_ADDRESS", "http://10.214.241.226:39090/")
RESULT_PATH = os.environ.get("RESULT_PATH", "metrics_result")


ETCD_HOST = os.environ.get("ETCD_HOST", "10.214.241.226")
ETCD_PORT = os.environ.get("ETCD_PROT", "32379")
ETCD_USER = os.environ.get("ETCD_USER", "root")
ETCD_PASSWD = os.environ.get("ETCD_PASSWD", "Y4b5EAwMlQ")

logging.info("PROMETHEUS_URL:" + PROMETHEUS_URL)
logging.info("ETCD_HOST:" + ETCD_HOST + " ETCD_PORT" + ETCD_PORT)


LABEL_POD_INFO = "POD_INFO"
LABEL_STATISTICS = "STATISTICS"
LABEL_FILENAME = "NAME"
LABEL_FILEMD5 = "MD5"


LOAD_MONITOR_JOB_URL = os.environ.get(
    "LOAD_MONITOR_JOB_URL", "http://localhost:2020/job")
