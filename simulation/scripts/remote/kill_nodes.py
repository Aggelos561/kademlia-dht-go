import paramiko
from utils import *
import sys

if __name__ == '__main__':
    pwd = read_pwd()
    machines = getMachines()

    for machine in machines:
        print(f"Killing kademlia node: {machine}")
        client = paramiko.SSHClient()
        client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

        try:
            client.connect(hostname=machine, username=USERNAME, password=pwd)

        except paramiko.AuthenticationException:
            print(f"{machine}: authentication failed")
            sys.exit()

        command = f"screen -S mysession -X quit"

        stdin, stdout, stderr = client.exec_command(command)
        print(stderr.read().decode('utf-8'))

        client.close()
