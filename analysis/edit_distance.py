import os
import json
import editdistance
from Levenshtein import distance


def calculate_levenshtein_distance(template1, template2):
    return distance(template1, template2)

def find_scenario_with_min_clusters(file_path):
    min_clusters = float('inf')
    min_scenario = None

    with open(file_path, 'r') as file:
        lines = file.readlines()
        for line in lines[2:]:
            parts = line.strip().split('|')
            scenario_name = parts[0].strip()
            num_clusters = int(parts[1].strip())

            if num_clusters < min_clusters:
                min_clusters = num_clusters
                min_scenario = scenario_name

    return min_scenario, min_clusters

def compare_clusters_with_min_scenario(min_scenario, folder_path):
    new_cluster = 0
    json_files = [f for f in os.listdir(folder_path) if f.endswith('.json')]

    with open(os.path.join(folder_path, min_scenario + '.json'), 'r') as min_file:
        min_json_data = json.load(min_file)
        min_data = min_json_data['id_to_cluster']['_Cache__data']

        with open(os.path.join(folder_path, 'distance.txt'), 'a') as output_file:
            output_file.write(f"Scenario with minimum clusters: {min_scenario}\n")
            for json_file in json_files:
                if json_file != min_scenario + '.json':
                    new_cluster = 0
                    output_file.write(f"\n \n \n \nComparing with scenario -> {json_file[:-5]}\n")
                    with open(os.path.join(folder_path, json_file), 'r') as file:
                        json_data = json.load(file)
                        data = json_data.get('id_to_cluster', {}).get('_Cache__data', {})  # Check if 'clusters' exists in JSON data

                        for _, value1 in data.items():
                            min_distance = float('inf')
                            min_cluster_id = None
                            min_cluster_text = None
                            for _, value2 in min_data.items():
                                distance = calculate_levenshtein_distance(value1['log_template_tokens']['py/tuple'], value2['log_template_tokens']['py/tuple'])
                                if distance < min_distance:
                                    min_distance = distance
                                    min_cluster_id = value2['cluster_id']
                                    min_cluster_text = value2['log_template_tokens']['py/tuple']
                            if min_distance < 3:
                                output_file.write(f"Cluster {value1['cluster_id']} - Minimum Distance: {min_distance} with Cluster {min_cluster_id}\n")
                                output_file.write(f"Text of Cluster {value1['cluster_id']}: {value1['log_template_tokens']['py/tuple']}\n")
                                output_file.write(f"Text of Cluster {min_cluster_id}: {min_cluster_text}\n")
                            else:
                                output_file.write(f"Cluster {value1['cluster_id']} - Minimum Distance: {min_distance}, with Cluster {min_cluster_id}, Too different\n")
                                output_file.write(f"Text of Cluster {value1['cluster_id']}: {value1['log_template_tokens']['py/tuple']}\n")
                                output_file.write(f"Text of Cluster {min_cluster_id}: {min_cluster_text}\n")
                                new_cluster = new_cluster + 1
                            output_file.write("==================================================\n")
                    output_file.write(f"New clusters from this scenario: {new_cluster}\n")

def write_summary_to_file(file_path, summary_data):
    with open(file_path, 'w') as summary_file:
        summary_file.write("Scenario Name                  | Number of Clusters  |  New clusters (compared to controlScenario)\n")
        summary_file.write("-----------------------------------------------------------------------------------------------\n")
        for scenario_name, num_clusters, new_clusters in summary_data:
            summary_file.write(f"{scenario_name: <30} | {num_clusters: <19} | +{new_clusters}\n")

if __name__ == "__main__":
    folder_path = "/home/colin/Documents/colin/analysis/clusters_file/global_analysis"
    file_path = os.path.join(folder_path, "results.txt")

    min_scenario, min_clusters = find_scenario_with_min_clusters(file_path)

    # Clear the content of the existing distance.txt file before running the comparison
    with open(os.path.join(folder_path, 'distance.txt'), 'w') as output_file:
        output_file.write('')

    # Run the comparison and get the summary data
    summary_data = []
    scenario_clusters = {}  # Dictionnaire pour stocker le nombre de clusters pour chaque scénario
    json_files = [f for f in os.listdir(folder_path) if f.endswith('.json')]
    for json_file in json_files:
        with open(os.path.join(folder_path, json_file), 'r') as file:
            json_data = json.load(file)
            scenario_name = json_file[:-5]
            num_clusters = len(json_data.get('id_to_cluster', {}).get('_Cache__data', {}))  # Check if 'clusters' exists in JSON data
            scenario_clusters[scenario_name] = num_clusters

    compare_clusters_with_min_scenario(min_scenario, folder_path)
    with open(os.path.join(folder_path, 'distance.txt'), 'r') as output_file:
        lines = output_file.readlines()
        for line in lines:
            if line.startswith("Scenario with minimum clusters:"):
                continue
            elif line.startswith("Comparing with scenario"):
                parts = line.strip().split('->')
                scenario_name = parts[1].strip()
            elif line.startswith("New clusters from this scenario:"):
                new_clusters = int(line.split(':')[1].strip())
                #print(scenario_clusters, scenario_name)
                num_clusters = scenario_clusters[scenario_name]  # Utiliser le nombre réel de clusters pour chaque scénario
                summary_data.append((scenario_name, num_clusters, new_clusters))

    # Sort the summary data by "Scenario Name" in alphabetical order
    summary_data.sort(key=lambda x: x[0])

    # Write the summary to results_test.txt
    results_test_file_path = os.path.join(folder_path, "results_test.txt")
    write_summary_to_file(results_test_file_path, summary_data)