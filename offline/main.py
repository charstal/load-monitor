from prometheus_api_client import PrometheusConnect
import os
import logging
import pandas as pd

logging.basicConfig(level=logging.INFO)
promURL = os.environ.get("PROMETHEUS_ADDRESS", "http://10.214.241.226:39090/")
logging.info("promtheus address:" + promURL)
prom = PrometheusConnect(url=promURL, disable_ssl=True)


course_label = "instance"
node_label = "kubernetes_node"

result_dir = "result"
pod_info_csv = "pod_info.csv"
statistics_csv = "statistics.csv"


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
        if course_label in d:
            dd = dict()
            dd["label"] = d[course_label]
            if node_label in d:
                dd["node"] = d[node_label]
            data_total_dict[d["pod"]] = dd


def fetch_metrics(data_total_dict):
    sql_list = [
        ("sum(stddev_over_time(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate[15d])) by (pod)", "cpu_std/m", lambda x: x * 1000),
        ("sum(avg_over_time(node_namespace_pod_container:container_cpu_usage_seconds_total:sum_irate[15d])) by (pod)", "cpu_avg/m", lambda x: x * 1000),
        ("sum(last_over_time(cluster:namespace:pod_cpu:active:kube_pod_container_resource_requests[15d])) by (pod)", "cpu_request/m", lambda x: x * 1000),
        ("sum(last_over_time(cluster:namespace:pod_cpu:active:kube_pod_container_resource_limits[15d])) by (pod)", "cpu_limit/m", lambda x: x * 1000),


        ("sum(stddev_over_time(node_namespace_pod_container:container_memory_working_set_bytes[15d])) by (pod)", "mem_std/MiB", lambda x: x / 1024 / 1024),
        ("sum(avg_over_time(node_namespace_pod_container:container_memory_working_set_bytes[15d])) by (pod)", "mem_avg/MiB", lambda x: x / 1024 / 1024),
        ("sum(last_over_time(cluster:namespace:pod_memory:active:kube_pod_container_resource_requests[15d])) by (pod)", "mem_request/MiB", lambda x: x / 1024 / 1024),
        ("sum(last_over_time(cluster:namespace:pod_memory:active:kube_pod_container_resource_limits[15d])) by (pod)", "mem_limit/MiB", lambda x: x / 1024 / 1024),


    ]

    for sql in sql_list:
        add_data2dict1(data_total_dict, sql[0], sql[1], sql[2])


def to_file(data_total_dict):

    data = pd.DataFrame(data_total_dict)
    data = data.T
    data.to_csv(os.path.join(result_dir, pod_info_csv))

    statistics_label = ["cpu_std/m", "cpu_avg/m", "mem_std/MiB", "mem_avg/MiB"]

    new_data = data.groupby(data["label"])[statistics_label].mean()

    new_data.loc["all"] = data[statistics_label].mean()
    new_data.to_csv(os.path.join(result_dir, statistics_csv))


if __name__ == "__main__":

    data_total_dict = dict()
    fetch_label(data_total_dict)

    fetch_metrics(data_total_dict)
    to_file(data_total_dict)
