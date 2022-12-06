# import etcd3
import logging
import requests
# from config import ETCD_HOST, ETCD_PORT, ETCD_USER, ETCD_PASSWD
# from config import LABEL_FILEMD5, LABEL_POD_INFO, LABEL_STATISTICS, LABEL_FILENAME
from config import LOAD_MONITOR_JOB_URL, MONGODB_URL, LABEL_POD_INFO, LABEL_STATISTICS
from pymongo import MongoClient
import datetime

pod_info_url = "/metric/pod_info_path"
pod_info_md5_url = "/metric/pod_info_md5"
statistics_url = "/metric/statistics_path"
statistics_md5_url = "/metric/statistics_path_md5"


def run(data_dict):
    # save_md5_etcd(data_dict)
    msg_dict = save_to_mongodb(data_dict)
    send_finish_to_load_monitor(msg_dict)


def save_to_mongodb(data_dict):
    client = MongoClient(MONGODB_URL)
    db = client.mo
    pod_info_collection = db.pod_info
    statistics_collection = db.statistics_collection

    time = datetime.datetime.utcnow()
    pod_info_dict = data_dict[LABEL_POD_INFO]

    pod_info_dict["time"] = time

    msg_pod_info_id = pod_info_collection.insert_one(pod_info_dict.to_dict())

    statistic_dict = data_dict[LABEL_STATISTICS]
    statistic_dict["time"] = time
    msg_statistic_id = statistics_collection.insert_one(
        statistic_dict.to_dict())

    return {LABEL_POD_INFO: msg_pod_info_id, LABEL_STATISTICS: msg_statistic_id}
# def save_md5_etcd(file_md5_dict):
#     client = etcd3.client(host=ETCD_HOST, port=ETCD_PORT,
#                           user=ETCD_USER, password=ETCD_PASSWD)

#     client.put(statistics_url, file_md5_dict[LABEL_STATISTICS][LABEL_FILENAME])
#     client.put(statistics_md5_url,
#                file_md5_dict[LABEL_STATISTICS][LABEL_FILEMD5])
#     client.put(pod_info_url, file_md5_dict[LABEL_POD_INFO][LABEL_FILENAME])
#     client.put(pod_info_md5_url, file_md5_dict[LABEL_POD_INFO][LABEL_FILEMD5])
#     logging.info("saved to etcd")


def send_finish_to_load_monitor(msg_dict):
    msg_dict["status"] = "finished"
    i = 1
    ok = False
    while i < 4:
        try:
            requests.get(url=LOAD_MONITOR_JOB_URL, data=msg_dict)
            break
        except:
            logging.info("trying to send" + str(i) + " times")
            i = i + 1
    if i < 4:
        logging.info("sended to load monitor")
    else:
        logging.info("error: cannot send to load monitor")


# def check_data():
#     client = etcd3.client(host=ETCD_HOST, port=ETCD_PORT,
#                           user=ETCD_USER, password=ETCD_PASSWD)

#     print(client.get(pod_info_url))
#     print(client.get(pod_info_md5_url))
#     print(client.get(statistics_url))
#     print(client.get(statistics_md5_url))


if __name__ == "__main__":
    check_data()
