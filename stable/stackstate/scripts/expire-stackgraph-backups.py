#!/usr/bin/env python3

import os, re, subprocess, sys
from datetime import datetime, timedelta

def main():
  backup_datetime_re = re.compile(os.getenv('BACKUP_STACKGRAPH_BACKUP_NAME_PARSE_REGEXP'))
  backup_datetime_format = os.getenv('BACKUP_STACKGRAPH_BACKUP_DATETIME_PARSE_FORMAT')
  backup_retention_timedelta = eval('timedelta(' + os.getenv('BACKUP_STACKGRAPH_BACKUP_RETENTION_TIME_DELTA') + ')')
  bucket_name = os.getenv('BACKUP_STACKGRAPH_BUCKET_NAME')
  minio_endpoint = 'http://' + os.getenv('MINIO_ENDPOINT')
  now = datetime.now()

  print('Retrieving StackGraph backups from bucket "', bucket_name, '"', sep='')
  backup_files = subprocess.check_output(['aws', '--endpoint-url', minio_endpoint, 's3api', 'list-objects-v2', '--bucket', bucket_name, '--query', 'Contents[].[Key]' , '--output' ,'text'], text=True)
  for backup_file in backup_files.splitlines():
    m = backup_datetime_re.match(backup_file)
    if m:
      backup_datetime = datetime.strptime(m.group(1), backup_datetime_format)
      age = now - backup_datetime
      if age > backup_retention_timedelta:
        print('Purging StackGraph backup "', backup_file, '"', sep='')
        delete_backup = subprocess.run(['aws', '--endpoint-url', minio_endpoint, 's3api', 'delete-object', '--bucket', bucket_name, '--key', backup_file], check=True)
      else:
        print('Not purging StackGraph backup "', backup_file, '"', sep='')

if __name__=='__main__':
  sys.exit(main())
