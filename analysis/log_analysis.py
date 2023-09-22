# SPDX-License-Identifier: MIT

import json
import logging
import os
import glob

import subprocess
import sys
import time
from os.path import dirname

from drain3 import TemplateMiner
from drain3.template_miner_config import TemplateMinerConfig
from drain3.file_persistence import FilePersistence

logger = logging.getLogger(__name__)
logging.basicConfig(stream=sys.stdout, level=logging.INFO, format='%(message)s')


def process_logs(log_dir, json_file, results_dict):
    persistence = FilePersistence(json_file)

    config = TemplateMinerConfig()
    config.load(f"{dirname(__file__)}/drain3.ini")
    config.profiling_enabled = True
    template_miner = TemplateMiner(persistence, config=config)

    line_count = 0

    # Use the parent log directory to search for all log files recursively
    file_paths = glob.glob(os.path.join(log_dir, "**/frr.log"), recursive=True)

    for path in file_paths:
        with open(path) as f:
            lines = f.readlines()

        start_time = time.time()
        batch_start_time = start_time
        batch_size = 10000

        for line in lines:
            line = line.rstrip()
            line = line.partition(": ")[2]
            result = template_miner.add_log_message(line)
            line_count += 1
            if line_count % batch_size == 0:
                time_took = time.time() - batch_start_time
                rate = batch_size / time_took
                logger.info(f"Processing line: {line_count}, rate {rate:.1f} lines/sec, "
                            f"{len(template_miner.drain.clusters)} clusters so far.")
                batch_start_time = time.time()
            if result["change_type"] != "none":
                result_json = json.dumps(result)
                logger.info(f"Input ({line_count}): {line}")
                logger.info(f"Result: {result_json}")

        time_took = time.time() - start_time
        rate = line_count / time_took
        logger.info(f"--- Done processing file in {time_took:.2f} sec. Total of {line_count} lines, rate {rate:.1f} lines/sec, "
                    f"{len(template_miner.drain.clusters)} clusters")

        # Update the results dictionary with the number of clusters and the number of lines for this scenario
        scenario_name = os.path.basename(log_dir)
        results_dict[scenario_name] = {
            "num_clusters": len(template_miner.drain.clusters),
            "num_lines": line_count
        }

        # sorted_clusters = sorted(template_miner.drain.clusters, key=lambda it: it.cluster_id, reverse=False)
        # for cluster in sorted_clusters:
        #     logger.info(cluster)

    with open(json_file, 'r') as file:
        json_data = file.read()
        data = json.loads(json_data)

    indented_json = json.dumps(data, indent=4)

    with open(json_file, 'w') as file:
        file.write(indented_json)


def write_results_to_file(results_dict, output_file):
    sorted_results = sorted(results_dict.items(), key=lambda x: x[0])

    with open(output_file, 'w') as file:
        max_scenario_length = max(len(scenario) for scenario, _ in sorted_results)
        max_clusters_length = max(len(str(data["num_clusters"])) for _, data in sorted_results)
        max_lines_length = max(len(str(data["num_lines"])) for _, data in sorted_results)

        header_format = "{{:<{}}} | {{:>{}}} | {{:>{}}}\n".format(
            max_scenario_length, max_clusters_length, max_lines_length)
        row_format = "{{:<{}}} | {{:>{}}} | {{:>{}}}\n".format(
            max_scenario_length, max_clusters_length, max_lines_length)

        file.write(header_format.format("Scenario Name", "Number of Clusters", "Number of Lines"))
        file.write("-" * (max_scenario_length + max_clusters_length + max_lines_length + 10) + "\n")
        for scenario, data in sorted_results:
            file.write(row_format.format(scenario, data["num_clusters"], data["num_lines"]))



# Specify the output directory for JSON files
output_dir = "/home/colin/Documents/colin/analysis/clusters_file/diff_controlScenario"

# Dictionary to store the results
results_dict = {}

# Iterate over each subdirectory (scenario) within the "log" directory
for root, subdirs, files in os.walk("/home/colin/Documents/colin/log/controlScenario"):
    for subdir in subdirs:
        # Form the log directory path for the current scenario
        log_dir = os.path.join(root, subdir)
        
        # Form the JSON file path for the current scenario
        json_file_path = os.path.join(output_dir, f"{subdir}.json")
        
        # Process the logs for the current scenario
        process_logs(log_dir, json_file_path, results_dict)

    # After processing the first level of subdirectories, break the loop to stop further traversal
    break

output_file = output_dir + "/results.txt"

# Triez les résultats par nom de scénario avant de les écrire dans le fichier results.txt
sorted_results_dict = dict(sorted(results_dict.items(), key=lambda x: x[0]))

# Appelez la fonction avec le dictionnaire trié
write_results_to_file(sorted_results_dict, output_file)
