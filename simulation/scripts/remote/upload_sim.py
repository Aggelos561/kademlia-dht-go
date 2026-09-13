import os
import paramiko
from utils import *

REMOTE_DIR = f'/home/users/{USERNAME}'

def upload_project(sftp, local_dir, remote_dir):
    for root, _, files in os.walk(local_dir):
        rel_path = os.path.relpath(root, local_dir)
        remote_path = os.path.join(remote_dir, rel_path).replace("\\", "/")
        
        try:
            sftp.mkdir(remote_path)
        except IOError:
            pass
    
        for file in files:
            local_file = os.path.join(root, file)
            remote_file = os.path.join(remote_path, file).replace("\\", "/")
            sftp.put(local_file, remote_file)


if __name__ == '__main__':
    pwd = read_pwd()
    machine = 'linux01.di.uoa.gr'

    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())

    client.connect(hostname=machine, username=USERNAME, password=pwd)

    sftp = client.open_sftp()
    upload_project(sftp, local_dir='./remote_sim', remote_dir=REMOTE_DIR)

    sftp.close()
    client.close()
