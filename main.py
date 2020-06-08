import pandas as pd
import hashlib 
import numpy as np
from operator import itemgetter


df = pd.read_csv('./wbqbiksz0m4y.csv')
df.columns = [c.strip() for c in df.columns]

df['traceID'] = df['traceID'].apply(lambda x: x.split(':')[0])

def f(x):
    com = str(str(x[1]) + str(x[2]) + str(x[3]) + str(x[4])).encode('utf-8')
    return hashlib.md5(com).hexdigest()
df['code'] = df.apply(f, axis=1)

mean_of_times = df.groupby('code')['timestamp'].mean().reset_index()

df.drop_duplicates(subset=['code'],keep='first',inplace=True)

df = df.drop('timestamp',1)

df = df.merge(mean_of_times, on='code')

df.sort_values(by='timestamp', inplace=True)

trace2packet = {}
req2traces = {}
for idx, row in df.iterrows():
    if not row['traceID'] in trace2packet:
        trace2packet[row['traceID']] = []
    trace2packet[row['traceID']].append({
        'req': row['req'],
        'src': row['src'],
        'dst': row['dst'],
        'time': row['timestamp']
    })
    
    if not row['req'] in req2traces:
        req2traces[row['req']] = set()
    req2traces[row['req']].add(row['traceID'])


for req_type in ['get_books', 'edit_books','auth_login']:
    print(req_type)
    values = [[],[],[],[]]
    for trace in req2traces[req_type]:
        packets = trace2packet[trace]
        packets = sorted(packets, key=itemgetter('time'), reverse=False)
        prev = packets[0]['time']
        i = 0
        for p in packets:
            values[i].append((p['time'] - prev)/1000000)
            i += 1
            prev = p['time']
    for v in values[1:]:
        print(np.percentile(v,95),'ms')
        print(np.max(v))
    print('=========')