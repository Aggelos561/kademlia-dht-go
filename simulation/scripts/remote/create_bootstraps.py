import os

# Create cyclic bootsrap files for remote kademlia nodes given an input file
def generate_bootstrap_files(input_file, output_dir, port=40000):
    with open(input_file, "r") as f:
        lines = [line.strip() for line in f if line.strip()]
    
    ips = [line.split(":")[1].strip() for line in lines]
    total_nodes = len(ips)

    os.makedirs(output_dir, exist_ok=True)

    for i in range(total_nodes):
        next1 = ips[(i + 1) % total_nodes]
        next2 = ips[(i + 2) % total_nodes]

        filename = os.path.join(output_dir, f"node_{i+1}.txt")
        with open(filename, "w") as out_file:
            out_file.write(f"{next1}:{port}\n")
            out_file.write(f"{next2}:{port}\n")

    print(f"Generated {total_nodes} bootstrap files with port {port} in '{output_dir}'.")


if __name__ == '__main__':
    generate_bootstrap_files("./output/machines_ips.txt", "remote_bootstraps")
