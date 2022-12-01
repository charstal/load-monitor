import etcd3
import logging
import requests
from config import ETCD_HOST, ETCD_PORT, ETCD_USER, ETCD_PASSWD
from config import LABEL_FILEMD5, LABEL_POD_INFO, LABEL_STATISTICS, LABEL_FILENAME
from config import LOAD_MONITOR_JOB_URL


pod_info_url = "/metric/pod_info_path"
pod_info_md5_url = "/metric/pod_info_md5"
statistics_url = "/metric/statistics_path"
statistics_md5_url = "/metric/statistics_path_md5"


def run(data_dict):
    save_md5_etcd(data_dict)
    send_finish_to_load_monitor(data_dict)


def save_md5_etcd(file_md5_dict):
    client = etcd3.client(host=ETCD_HOST, port=ETCD_PORT,
                          user=ETCD_USER, password=ETCD_PASSWD)

    client.put(statistics_url, file_md5_dict[LABEL_STATISTICS][LABEL_FILENAME])
    client.put(statistics_md5_url,
               file_md5_dict[LABEL_STATISTICS][LABEL_FILEMD5])
    client.put(pod_info_url, file_md5_dict[LABEL_POD_INFO][LABEL_FILENAME])
    client.put(pod_info_md5_url, file_md5_dict[LABEL_POD_INFO][LABEL_FILEMD5])
    logging.info("saved to etcd")


def send_finish_to_load_monitor(data_dict):
    data_dict["status"] = "finished"
    i = 1
    ok = False
    while i < 4:
        try:
            res = requests.post(url=LOAD_MONITOR_JOB_URL, data=data_dict)
            if res.ok():
                ok = True
                break
        except:
            logging.info("trying to send" + str(i) + " times")

        i = i + 1
    if ok:
        logging.info("sended to load monitor")
    else:
        logging.info("cannot send to load monitor")


def check_data():
    client = etcd3.client(host=ETCD_HOST, port=ETCD_PORT,
                          user=ETCD_USER, password=ETCD_PASSWD)

    print(client.get(pod_info_url))
    print(client.get(pod_info_md5_url))
    print(client.get(statistics_url))
    print(client.get(statistics_md5_url))


if __name__ == "__main__":
    check_data()
