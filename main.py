import pandas as pd
import hashlib 
import numpy as np
from operator import itemgetter

np.set_printoptions(suppress=True)
df = pd.read_csv('/home/vahid/Desktop/temp1/tcpdumps/yy93peyghh47/http_packets.csv')
df.columns = [c.strip() for c in df.columns]
df = df.drop('traceID',1)

def f(x):
    com = str(str(x[1]) + str(x[2]) + str(x[3]) + str(x[4])).encode('utf-8')
    return hashlib.md5(com).hexdigest()
df['code'] = df.apply(f, axis=1)

mean_of_times = df.groupby('code')['ts'].mean().reset_index()

df.drop_duplicates(subset=['code'],keep='first',inplace=True)

df = df.drop('ts',1)

df = df.merge(mean_of_times, on='code')
df = df.drop('code', 1)

df['req'] = df.apply(lambda row: df[(df.debugID == row.debugID) & (df.req.notnull())].values[0][2] if pd.isnull(row['req']) else row['req'], axis=1)

# data = df.groupby('req')['ts','src','dst','debugID'].apply(list)
data = dict(tuple(df.groupby('req')))
for key, df in data.items():
    
    byDebugID = list(tuple(df.groupby('debugID')))
    if len(byDebugID) == 0 : continue
    print(key)
    diffs = np.zeros((len(byDebugID), 9))
    for i, tmp in enumerate(byDebugID):
        debugID = tmp[0]
        request = tmp[1]
        request = request.sort_values(by="ts")
        diffs[i] = np.diff(request.ts)
    diffs = np.percentile(diffs, 95, axis=0)
    print(np.round(diffs, 3))
# df.sort_values(by='timestamp', inplace=True)

# trace2packet = {}
# req2traces = {}
# for idx, row in df.iterrows():
#     if not row['traceID'] in trace2packet:
#         trace2packet[row['traceID']] = []
#     trace2packet[row['traceID']].append({
#         'req': row['req'],
#         'src': row['src'],
#         'dst': row['dst'],
#         'time': row['timestamp']
#     })
    
#     if not row['req'] in req2traces:
#         req2traces[row['req']] = set()
#     req2traces[row['req']].add(row['traceID'])


# for req_type in ['get_books', 'edit_books','auth_login']:
#     print(req_type)
#     values = [[],[],[],[]]
#     for trace in req2traces[req_type]:
#         packets = trace2packet[trace]
#         packets = sorted(packets, key=itemgetter('time'), reverse=False)
#         prev = packets[0]['time']
#         i = 0
#         for p in packets:
#             values[i].append((p['time'] - prev)/1000000)
#             i += 1
#             prev = p['time']
#     for v in values[1:]:
#         print(np.percentile(v,95),'ms')
#         print(np.max(v))
#     print('=========')