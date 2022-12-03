import os
import logging

logging.basicConfig(level=logging.INFO)
PROMETHEUS_URL = os.environ.get(
    "PROMETHEUS_ADDRESS", "http://prometheus-k8s.monitoring.svc.cluster.local:9090/")
RESULT_PATH = os.environ.get("RESULT_PATH", "offline")


ETCD_HOST = os.environ.get("ETCD_HOST", "etcd-dev.default.svc.cluster.local")
ETCD_PORT = os.environ.get("ETCD_PROT", "2379")
ETCD_USER = os.environ.get("ETCD_USER", "root")
ETCD_PASSWD = os.environ.get("ETCD_PASSWD", "CrHkL98Ryr")

logging.info("PROMETHEUS_URL:" + PROMETHEUS_URL)
logging.info("ETCD_HOST:" + ETCD_HOST + " ETCD_PORT" + ETCD_PORT)


LABEL_POD_INFO = "POD_INFO"
LABEL_STATISTICS = "STATISTICS"
LABEL_FILENAME = "NAME"
LABEL_FILEMD5 = "MD5"


LOAD_MONITOR_JOB_URL = os.environ.get(
    "LOAD_MONITOR_JOB_URL", "http://load-monitor-svc.default.svc.cluster.local:2020/job")
