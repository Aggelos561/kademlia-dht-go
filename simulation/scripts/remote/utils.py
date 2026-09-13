import subprocess

MACHINES_FILE = './output/machines.txt'
USERNAME = ''

def read_pwd() -> str:
    return subprocess.check_output(["pass", "linux/lab"], text=True).strip()

def getMachines() -> list:
    machines = []
    with open(MACHINES_FILE) as file:
        for line in file:
            machines.append(line.rstrip())

    return machines
