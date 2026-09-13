import subprocess
import platform
import argparse
from pathlib import Path

def read_ports(file_path):
    ports = set()
    with open(file_path, 'r') as f:
        for line in f:
            line = line.strip()
            if line.isdigit():
                ports.add(int(line))
    return sorted(ports)

def get_pids(port):
    system = platform.system()
    pids = set()

    if system == "Windows":
        try:
            result = subprocess.check_output('netstat -ano', shell=True, text=True)
            for line in result.splitlines():
                if f":{port}" in line:
                    parts = line.strip().split()
                    if parts and parts[-1].isdigit():
                        pids.add(int(parts[-1]))
        except subprocess.CalledProcessError:
            pass
    else:
        try:
            result = subprocess.check_output(['lsof', '-i', f'TCP:{port}'], text=True)
            for line in result.splitlines():
                parts = line.split()
                if len(parts) > 1 and parts[1].isdigit():
                    pids.add(int(parts[1]))
        except subprocess.CalledProcessError:
            pass

    return list(pids)

def kill(pid):
    system = platform.system()
    try:
        if system == "Windows":
            subprocess.run(['taskkill', '/PID', str(pid), '/F'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        else:
            subprocess.run(['kill', '-9', str(pid)], check=True)
        return True
    except subprocess.SubprocessError:
        return False

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--file', required=True, help='File with one port number per line')
    args = parser.parse_args()

    file_path = Path(args.file)
    if not file_path.exists():
        print(f"File '{file_path}' does not exist.")
        return

    ports = read_ports(file_path)
    if not ports:
        print("No valid ports found.")
        return

    print(f"Scanning and killing processes on ports: {ports}")
    total_killed = 0
    for port in ports:
        pids = get_pids(port)
        if not pids:
            print(f"  No process found on port {port}")
            continue
        for pid in pids:
            if kill(pid):
                print(f"Killed PID {pid} on port {port}")
                total_killed += 1
            else:
                print(f"Failed to kill PID {pid} on port {port}")
    print(f"Done. Total processes killed: {total_killed}")

if __name__ == '__main__':
    main()
