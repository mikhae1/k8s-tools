#!/usr/bin/env python3

import boto3
import os
from datetime import datetime, timedelta, timezone
from prettytable import PrettyTable
from kubernetes import client, config
import subprocess
import json

AWS_REGION = os.getenv("AWS_REGION", "eu-central-1")

def main():
    config.load_kube_config()
    api = client.CustomObjectsApi()
    core_api = client.CoreV1Api()

    now = datetime.utcnow()
    one_day_ago = now - timedelta(days=1)
    one_week_ago = now - timedelta(weeks=1)

    table = PrettyTable(
        [
            "Instance Name",
            "Instance Type",
            "Private IP",
            "Instance ID",
            "CPU Avg (1d)",
            "CPU Avg (7d)",
            "Max CPU (7d)",
            "Mem Util",
            "Eph. Usage",
            "Disk Util",
            "AGE",
        ]
    )

    count = 0
    for instance in get_all_instances():
        avg_cpu_utilization_1d = get_average_cpu_utilization(
            instance["InstanceId"], one_day_ago, now
        )
        avg_cpu_utilization_7d = get_average_cpu_utilization(
            instance["InstanceId"], one_week_ago, now
        )
        max_cpu_utilization = get_maximum_cpu_utilization(
            instance["InstanceId"], one_week_ago, now
        )
        mem_utilization = get_memory_utilization(
            api, core_api, instance["PrivateIpAddress"]
        )
        eph_usage = get_eph_usage(api, core_api, instance["PrivateIpAddress"])
        disk_usage = get_disk_usage(instance["PrivateIpAddress"])
        instance_age = calculate_age(instance["LaunchTime"])

        count += 1
        table.add_row(
            [
                instance["Name"],
                instance["InstanceType"],
                instance["PrivateIpAddress"],
                instance["InstanceId"],
                f"{round(avg_cpu_utilization_1d, 1)}%"
                if avg_cpu_utilization_1d
                else "No Data",
                f"{round(avg_cpu_utilization_7d, 1)}%"
                if avg_cpu_utilization_7d
                else "No Data",
                f"{round(max_cpu_utilization, 1)}%"
                if max_cpu_utilization
                else "No Data",
                f"{mem_utilization}%" if mem_utilization else "No Data",
                f"{eph_usage}%" if eph_usage else "No Data",
                f"{disk_usage}%" if disk_usage else "No Data",
                instance_age,
            ]
        )

    table.sortby = "Instance Name"
    table.align["Instance Name"] = "l"
    print(table)

def get_disk_usage(private_ip):
    try:
        node_name = f"ip-{private_ip.replace('.', '-')}.{AWS_REGION}.compute.internal"
        cmd = f"kubectl get --raw /api/v1/nodes/{node_name}/proxy/stats/summary"
        response = subprocess.check_output(cmd, shell=True)
        response = response.decode("utf-8")
        json_response = json.loads(response)

        if "node" in json_response:
            fs_info = json_response["node"].get("fs", {})
            if fs_info:
                capacity_bytes = fs_info.get("capacityBytes", 0)
                used_bytes = fs_info.get("usedBytes", 0)
                usage_percent = (used_bytes / capacity_bytes) * 100
                return round(usage_percent, 1)
        return None
    except Exception as e:
        print(f"ERROR: {e}")
        return None

def get_eph_usage(api, core_api, private_ip):
    try:
        node_list = api.list_cluster_custom_object("metrics.k8s.io", "v1beta1", "nodes")
        node_name = f"ip-{private_ip.replace('.', '-')}.{AWS_REGION}.compute.internal"
        for node in node_list["items"]:
            if node["metadata"]["name"] == node_name:
                node_status = core_api.read_node_status(node_name)
                allocatable_keys = node_status.status.allocatable.keys()
                capacity_keys = node_status.status.capacity.keys()

                total_disk_key = next((key for key in capacity_keys if "storage" in key), None)
                used_disk_key = next((key for key in allocatable_keys if "ephemeral-storage" in key), None)

                if total_disk_key and used_disk_key:
                    allocatable = int(node_status.status.allocatable[used_disk_key].strip("Ki"))
                    capacity = int(node_status.status.capacity[total_disk_key].strip("Ki"))

                    allocatable_mb = round(allocatable / 1024 / 1024 / 1024)
                    capacity_mb = round(capacity / 1024 / 1024)
                    usage = capacity_mb - allocatable_mb
                    usage_percent = (usage / capacity_mb) * 100
                    return round(usage_percent, 1)

        return None
    except Exception as e:
        print(f"ERROR: {e}")
        return None


def calculate_age(instance_launch_time):
    now = datetime.now(timezone.utc)  # Make 'now' offset-aware
    instance_age = now - instance_launch_time
    return f"{instance_age.days}d"


def get_memory_utilization(api, core_api, private_ip):
    try:
        node_list = api.list_cluster_custom_object("metrics.k8s.io", "v1beta1", "nodes")
        node_name = f"ip-{private_ip.replace('.', '-')}.{AWS_REGION}.compute.internal"
        for node in node_list["items"]:
            if node["metadata"]["name"] == node_name:
                try:
                    node_status = core_api.read_node_status(node_name)
                    total_memory = int(
                        node_status.status.capacity["memory"].strip("Ki")
                    )  # Assuming the value is in Kibibytes
                    used_memory = int(
                        node["usage"]["memory"].strip("Ki")
                    )  # Assuming the value is in Kibibytes
                    memory_utilization_percentage = (used_memory / total_memory) * 100
                    return round(memory_utilization_percentage, 1)
                except KeyError as e:
                    # Handle KeyError if the expected keys are missing in the node_status or node objects
                    print(f"KeyError: {e}")
                    return None
        return None
    except Exception as e:
        print(f"ERROR: {e}")
        return None


def get_all_instances():
    ec2 = boto3.resource("ec2")
    eks = boto3.client("eks")
    autoscaling = boto3.client("autoscaling")

    instances = []

    clusters = eks.list_clusters()["clusters"]
    for cluster_name in clusters:
        cluster_info = eks.describe_cluster(name=cluster_name)
        nodegroup_names = eks.list_nodegroups(clusterName=cluster_name)["nodegroups"]
        for nodegroup_name in nodegroup_names:
            nodegroup_info = eks.describe_nodegroup(
                clusterName=cluster_name, nodegroupName=nodegroup_name
            )
            autoscaling_group_name = nodegroup_info["nodegroup"]["resources"][
                "autoScalingGroups"
            ][0]["name"]
            response = autoscaling.describe_auto_scaling_groups(
                AutoScalingGroupNames=[autoscaling_group_name]
            )
            for instance in response["AutoScalingGroups"][0]["Instances"]:
                ec2_instance = ec2.Instance(instance["InstanceId"])
                instance_name = ""
                for tag in ec2_instance.tags:
                    if tag["Key"] == "Name":
                        instance_name = tag["Value"]
                        break
                launch_time = (
                    ec2_instance.launch_time
                )  # Fetch the launch time of the instance
                instances.append(
                    {
                        "InstanceId": ec2_instance.id,
                        "Name": instance_name,
                        "PrivateIpAddress": ec2_instance.private_ip_address,
                        "InstanceType": ec2_instance.instance_type,
                        "LaunchTime": launch_time,
                    }
                )

    return instances


def get_average_cpu_utilization(instance_id, start_time, end_time, period=86400):
    return get_cpu_utilization_statistic(
        instance_id, start_time, end_time, period, "Average"
    )


def get_maximum_cpu_utilization(instance_id, start_time, end_time, period=86400):
    return get_cpu_utilization_statistic(
        instance_id, start_time, end_time, period, "Maximum"
    )


def get_cpu_utilization_statistic(instance_id, start_time, end_time, period, statistic):
    cloudwatch = boto3.client("cloudwatch")

    try:
        response = cloudwatch.get_metric_statistics(
            Namespace="AWS/EC2",
            MetricName="CPUUtilization",
            Dimensions=[
                {"Name": "InstanceId", "Value": instance_id},
            ],
            StartTime=start_time,
            EndTime=end_time,
            Period=period,
            Statistics=[statistic],
        )

        datapoints = response.get("Datapoints", [])
        if datapoints:
            return sum(d[statistic] for d in datapoints) / len(datapoints)

        return None
    except Exception as e:
        print(f"ERROR: {e}")
        return None


if __name__ == "__main__":
    main()
