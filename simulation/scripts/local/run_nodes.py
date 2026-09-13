import argparse
import subprocess
import platform
import re
import shutil
from pathlib import Path

def get_pids(port):
    if platform.system() == "Windows":
        try:
            out = subprocess.check_output(f'netstat -ano | findstr :{port}', shell=True, text=True)
            return list({int(l.split()[-1]) for l in out.splitlines()})
        except subprocess.CalledProcessError:
            return []
    else:
        try:
            out = subprocess.check_output(['lsof', '-ti', f':{port}'], text=True)
            return list({int(p) for p in out.splitlines() if p.isdigit()})
        except subprocess.CalledProcessError:
            return []

def kill(pid):
    if platform.system() == "Windows":
        subprocess.run(['taskkill', '/PID', str(pid), '/F'])
    else:
        subprocess.run(['kill', '-9', str(pid)], check=True)

def build(exec_path):
    subprocess.run(['go', 'build', '.'], cwd=Path(exec_path).resolve().parent, check=True)

def extract_number(path):
    match = re.search(r'(\d+)', path.name)
    return int(match.group(1)) if match else -1

def infer_ports(boot_files):
    ports = []
    for file in boot_files:
        with file.open() as f:
            for line in f:
                match = re.search(r':(\d+)', line)
                if match:
                    ports.append(int(match.group(1)))
    return sorted(set(ports))

def clean_dbs():
    db_dir = Path('./dbs')
    if db_dir.exists() and db_dir.is_dir():
        shutil.rmtree(db_dir)

def run_nodes(exec_path, ip, ports, boot_files, log_dir, a):
    procs = []
    for i, port in enumerate(ports):
        log = log_dir / f'node_{i}.log'
        boot = boot_files[i] if i < len(boot_files) else None
        cmd = [str(exec_path), '-ip', ip, '-port', str(port), '-mode', 'node']
        if boot:
            cmd += ['-file', str(boot)]
        cmd += ['-a', a]
        with log.open('w') as f:
            procs.append(subprocess.Popen(cmd, stdout=f, stderr=f))
    return procs

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--exec', required=True)
    parser.add_argument('--ip', required=True)
    parser.add_argument('--boot-dir', required=True)
    parser.add_argument('--out-dir', required=True)
    parser.add_argument('--a', default='1')
    args = parser.parse_args()

    exec_path = Path(args.exec)
    node_boot_dir = Path(args.boot_dir)
    log_dir = Path(args.out_dir)
    log_dir.mkdir(parents=True, exist_ok=True)

    node_boot_files = sorted(node_boot_dir.glob('bootstrap*.txt'), key=extract_number)
    ports = infer_ports(node_boot_files)

    for port in ports:
        for pid in get_pids(port):
            kill(pid)

    clean_dbs()
    build(exec_path)

    procs = run_nodes(exec_path, args.ip, ports, node_boot_files, log_dir, args.a)
    print(f"Started {len(procs)} nodes. Logs are in {log_dir}")
    print(f"Nodes are running [{len(procs)}].")
    
    try:
        for p in procs:
            p.wait()
    except KeyboardInterrupt:
        print("\nStopping nodes...")
        for p in procs:
            if p.poll() is None:
                p.kill()
        

if __name__ == '__main__':
    main()
