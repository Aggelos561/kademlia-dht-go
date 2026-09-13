import paramiko
from utils import *
import os
import sys
import re

PORT = 40000
ALPHA = 8
DELAY = 120
NODES = 20
BOOTSTRAP = './remote_bootstraps_20'


def extract_number(name):
    match = re.search(r'(\d+)', name)
    return int(match.group(1)) if match else -1


if __name__ == '__main__':
    pwd = read_pwd()
    machines = getMachines()
    bootstrap_files = sorted(os.listdir(BOOTSTRAP), key=extract_number)

    for machine, bootstrap in zip(machines[:NODES], bootstrap_files[:NODES]):
        client = paramiko.SSHClient()
        client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

        print(f"running node {machine}")

        try:
            client.connect(hostname=machine, username=USERNAME, password=pwd)

        except paramiko.AuthenticationException:
            print(f"{machine}: authentication failed")
            sys.exit()

        screen = "screen -dmS mysession bash -c"
        command = f"'./simulation -ip `hostname -I` -port {PORT} -file " + f'{BOOTSTRAP}/{bootstrap}' + f" -a {ALPHA} -delay {DELAY}"
        command = 'chmod +x simulation; '+ screen + ' ' + command + f" > {machine}_output.log 2>&1'"

        stdin, stdout, stderr = client.exec_command(command)        
        print(stderr.read().decode('utf-8'))

        client.close()
