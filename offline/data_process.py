from prometheus_api_client import PrometheusConnect
# import os
import logging
import pandas as pd
from config import PROMETHEUS_URL,  LABEL_POD_INFO, LABEL_STATISTICS

# import datetime

CPU_STD_UNIT_LABEL = "cpu_std/m"
CPU_AVG_UNIT_LABEL = "cpu_avg/m"
CPU_REQUEST_UNIT_LABEL = "cpu_request/m"
CPU_LIMIT_UNIT_LABEL = "cpu_limit/m"

MEM_STD_UNIT_LABEL = "mem_std/MiB"
MEM_AVG_UNIT_LABEL = "mem_avg/MiB"
MEM_REQUEST_UNIT_LABEL = "mem_request/m"
MEM_LIMIT_UNIT_LABEL = "mem_limit/m"


COURSE_LABEL = "course_label"
ALL_COURSE_LABEL = "all"
NODE_LABEL = "node"

# Todo(label)
# need change label
prom_course_label = "label_course_id"
prom_node_label = "node"


logging.info("promtheus address:" + PROMETHEUS_URL)
prom = PrometheusConnect(url=PROMETHEUS_URL, disable_ssl=True)


# result_dir = RESULT_PATH
# pod_info_csv = "pod_info.csv"
# statistics_csv = "statistics.csv"


# if not os.path.exists(result_dir):
#     os.makedirs(result_dir)


def range_data_query(sql):
    logging.info("query: " + sql)
    data = prom.custom_query(sql)
    return data


def add_data2dict1(data, sql, tag, fc):
    dd = range_data_query(sql)

    for d in dd:
        pod_name = d["metric"]["pod"]
        value = float(d["value"][1])
        if pod_name in data:
            data[pod_name][tag] = fc(value)


def fetch_label(data_total_dict):
    data = range_data_query("last_over_time(kube_pod_labels[15d])")
    for d in data:
        d = d["metric"]
        # only record the pod that has prom_course_label
        if prom_course_label in d:
            dd = dict()
            dd[COURSE_LABEL] = d[prom_course_label]
            # record the node name
            # if prom_node_label in d:
            #     dd["node"] = d[prom_node_label]
            data_total_dict[d["pod"]] = dd


def fetch_node_for_pod(data_total_dict):
    data = range_data_query("kube_pod_info")
    for d in data:
        d = d["metric"]
        if prom_node_label in d:
            pod_name = d["pod"]
            if pod_name in data_total_dict:
                node_name = d[prom_node_label]
                data_total_dict[pod_name][NODE_LABEL] = node_name


def fetch_metrics(data_total_dict):
    sql_list = [
        ("sum(stddev_over_time(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate[15d])) by (pod)", CPU_STD_UNIT_LABEL, lambda x: x * 1000),
        ("sum(avg_over_time(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate[15d])) by (pod)", CPU_AVG_UNIT_LABEL, lambda x: x * 1000),
        ("sum(last_over_time(cluster:namespace:pod_cpu:active:kube_pod_container_resource_requests[15d])) by (pod)", CPU_REQUEST_UNIT_LABEL, lambda x: x * 1000),
        ("sum(last_over_time(cluster:namespace:pod_cpu:active:kube_pod_container_resource_limits[15d])) by (pod)", CPU_LIMIT_UNIT_LABEL, lambda x: x * 1000),


        ("sum(stddev_over_time(node_namespace_pod_container:container_memory_working_set_bytes[15d])) by (pod)", MEM_STD_UNIT_LABEL, lambda x: x / 1024 / 1024),
        ("sum(avg_over_time(node_namespace_pod_container:container_memory_working_set_bytes[15d])) by (pod)", MEM_AVG_UNIT_LABEL, lambda x: x / 1024 / 1024),
        ("sum(last_over_time(cluster:namespace:pod_memory:active:kube_pod_container_resource_requests[15d])) by (pod)", MEM_REQUEST_UNIT_LABEL, lambda x: x / 1024 / 1024),
        ("sum(last_over_time(cluster:namespace:pod_memory:active:kube_pod_container_resource_limits[15d])) by (pod)", MEM_LIMIT_UNIT_LABEL, lambda x: x / 1024 / 1024),


    ]

    for sql in sql_list:
        add_data2dict1(data_total_dict, sql[0], sql[1], sql[2])


# def to_file(data_total_dict):
#     time = datetime.datetime.now().strftime("%Y-%m-%d-%H-%M-%S")
#     file_path_dict = {}
#     data = pd.DataFrame(data_total_dict)
#     data = data.T
#     newfileName = time + "-" + pod_info_csv
#     p = os.path.join(result_dir, newfileName)
#     file_path_dict[LABEL_POD_INFO] = {LABEL_FILENAME: newfileName}
#     data.to_csv(p)

#     statistics_label = ["cpu_std/m", "cpu_avg/m", "mem_std/MiB", "mem_avg/MiB"]
#     new_data = data.groupby(data["label"])[statistics_label].mean()
#     new_data.loc["all"] = data[statistics_label].mean()
#     newfileName = time + "-" + statistics_csv
#     p = os.path.join(result_dir, newfileName)
#     file_path_dict[LABEL_STATISTICS] = {LABEL_FILENAME: newfileName}
#     new_data.to_csv(p)

#     return file_path_dict


def analysis(data_total_dict):
    total_dict = dict()
    data = pd.DataFrame(data_total_dict)
    data = data.T

    statistics_label = [CPU_STD_UNIT_LABEL, CPU_AVG_UNIT_LABEL,
                        MEM_STD_UNIT_LABEL, MEM_AVG_UNIT_LABEL]
    # ["cpu_std/m", "cpu_avg/m", "mem_std/MiB", "mem_avg/MiB"]
    new_data = data.groupby(data[COURSE_LABEL])[statistics_label].mean()
    new_data.loc[ALL_COURSE_LABEL] = data[statistics_label].mean()

    total_dict[LABEL_POD_INFO] = data.T
    total_dict[LABEL_STATISTICS] = new_data.T

    return total_dict


def run():
    source_data_dict = dict()

    fetch_label(source_data_dict)
    fetch_node_for_pod(source_data_dict)
    fetch_metrics(source_data_dict)
    data_total = analysis(source_data_dict)
    # file_paths = to_file(data_total_dict)

    return data_total
