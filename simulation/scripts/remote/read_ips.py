import paramiko
import sys
from utils import *

# Get the IPs given the linux machines names
if __name__ == '__main__':
    pwd = read_pwd()
    machines = getMachines()
    output_list = []

    for machine in machines:
        client = paramiko.SSHClient()
        client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

        try:
            client.connect(hostname=machine, username=USERNAME, password=pwd)

        except paramiko.AuthenticationException:
            print("authentication failed")
            sys.exit()

        stdin, stdout, stderr = client.exec_command('hostname -I')
        output = stdout.read().decode('utf-8')

        output_list.append(f'{machine}: {output}')

        client.close()

    with open('machines_ips.txt', 'w') as file:
        file.writelines(output_list)
