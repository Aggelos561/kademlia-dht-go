import os
import argparse

def gen_and_split(ip, start_port, count, output_dir, chunk_size, overlap):
    os.makedirs(output_dir, exist_ok=True)

    total_generated = 0
    chunk_id = 0

    while total_generated < count:
        chunk_start_port = start_port + total_generated
        lines_left = count - total_generated
        current_chunk_size = min(chunk_size, lines_left)
        
        chunk_lines = [f"{ip}:{chunk_start_port + i}\n" for i in range(current_chunk_size)]
        output_path = os.path.join(output_dir, f"bootstrap{chunk_id}.txt")

        with open(output_path, "w") as f:
            f.writelines(chunk_lines)

        total_generated += chunk_size - overlap
        chunk_id += 1


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate IP:port lines and split into bootstrap chunks.")
    parser.add_argument("--ip", type=str, required=True)
    parser.add_argument("--start-port", type=int, required=True)
    parser.add_argument("--count", type=int, required=True)
    parser.add_argument("--output-dir", type=str, required=True)
    parser.add_argument("--chunk-size", type=int, required=True)
    parser.add_argument("--overlap", type=int, default=1)
    args = parser.parse_args()

    gen_and_split(args.ip, args.start_port, args.count, args.output_dir, args.chunk_size, args.overlap)
