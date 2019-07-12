import csv
import sys
import os

def read_config(path):
  d = {}
  with open(os.path.join(path, 'config.csv'), 'r') as csvfile:
    reader = csv.reader(csvfile)
    n, duration = next(reader)
    next(reader)

    for row in reader:
      name, id = row
      d[id] = name
  return (n, duration, d)

def read_all_data(path, config):
  ins = {}
  outs = {}
  error_counts = 0
  for id in config:
    # print("read data: ", id)
    in_dict, out_dict, error_count = read_data(path, id)
    error_counts += error_count
    outs = {**outs, **out_dict}
    for msg_id, timestamp in in_dict.items():
      ins.setdefault(msg_id, []).append(timestamp)
  return (ins, outs, error_counts)

def read_data(path, id):
  in_dict = {}
  out_dict = {}
  error_count = 0
  try:
    with open(os.path.join(path, f'{id}.csv'),'r') as csvfile:
      reader = csv.reader(csvfile)
      # if reader.line_num == 0:
        # return ({}, {}, 1) 
      next(reader)
      for row in reader:
        direction, timestamp, entity_id, digest, x, y, z = row
        msg_id = entity_id+'-'+digest
        if direction == "IN":
          in_dict[msg_id] = int(timestamp)
        else:
          out_dict[msg_id] = int(timestamp)
  except EnvironmentError:
    # print(f'missing: {id}.csv')
    return ({}, {}, 1)
  return (in_dict, out_dict, error_count)

def analyze(path):
  path = os.path.join(path, 'csv')
  print('-'*80)
  print('analyze: ', path)
  n, duration, config = read_config(path)
  ins,outs,error_count = read_all_data(path, config)
  total = 0
  count = 0
  missing_msg_count = 0
  with open(os.path.join(path,'result.csv'),'w') as csvfile:
    writer = csv.writer(csvfile)
    for msg_id, send_timestamp in outs.items():
      if msg_id in ins:
        n = len(ins[msg_id])
        avg_mdl = (sum(ins[msg_id]) - n*send_timestamp) / n
        writer.writerow([send_timestamp, msg_id, avg_mdl])
        total += avg_mdl
        count += 1
      else:
        writer.writerow([send_timestamp, msg_id, 'MISSING'])
        missing_msg_count += 1
  print("path: ", path)
  if count!=0:
    print('avg delivery latency:', total/count,'ms')
  print('error count: ', error_count)
  print('missing message count: ', missing_msg_count)

if __name__=="__main__":
  path = ''
  if len(sys.argv) == 2:
    path = sys.argv[1]

  if not path:
    print("please give a path")
    sys.exit(1)
  else:
    print('start to analyze: ', path)

  for file in os.listdir(path):
    base = os.path.join(path,file)
    if os.path.isdir(os.path.join(path, file)):
      analyze(base)

