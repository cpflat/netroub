import json
import editdistance

def calculate_levenshtein_distance(template1, template2):
    return editdistance.eval(template1, template2)

def compare_json_files(file_path1, file_path2):
    new_cluster = 0
    with open(file_path1, 'r') as file1, open(file_path2, 'r') as file2:
        json_data1 = json.load(file1)
        json_data2 = json.load(file2)

    data1 = json_data1['id_to_cluster']['_Cache__data']
    data2 = json_data2['id_to_cluster']['_Cache__data']

    for key1, value1 in data1.items():
        min_distance = float('inf')
        min_cluster_id = None
        min_cluster_text = None
        for key2, value2 in data2.items():
            distance = calculate_levenshtein_distance(value1['log_template_tokens']['py/tuple'], value2['log_template_tokens']['py/tuple'])
            if distance < min_distance:
                min_distance = distance
                min_cluster_id = value2['cluster_id']
                min_cluster_text = value2['log_template_tokens']['py/tuple']

        cluster_id = value1['cluster_id']
        cluster_text = value1['log_template_tokens']['py/tuple']

        if min_distance < 3:
            print(f"Cluster {cluster_id} - Minimum Distance: {min_distance} with Cluster {min_cluster_id}")
            print(f"Text of Cluster {cluster_id}: {cluster_text}")
            print(f"Text of Cluster {min_cluster_id}: {min_cluster_text}")
        else:
            print(f"Cluster {cluster_id} - Minimum Distance: {min_distance}, with Cluster {min_cluster_id}, Too different")
            print(f"Text of Cluster {cluster_id}: {cluster_text}")
            print(f"Closest matching Cluster {min_cluster_id}: {min_cluster_text}")
            new_cluster = new_cluster + 1 
        print("="*50)
    print(f"New clusters from this scenario: {new_cluster}")

if __name__ == "__main__":
    file_path2 = "/home/colin/Documents/colin/analysis/clusters_file/global_analysis/controlScenario.json"
    file_path1 = "/home/colin/Documents/colin/analysis/clusters_file/global_analysis/pause_corrupt_all_devices.json"
    compare_json_files(file_path1, file_path2)
