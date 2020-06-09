import pandas as pd
import hashlib 
import numpy as np
from operator import itemgetter
import sys

np.set_printoptions(suppress=True)
df = pd.read_csv(sys.argv[1])
df.columns = [c.strip() for c in df.columns]

# We are not using traceID for now, we are using debugID
df = df.drop('traceID',1)

# create md5 code to remove duplicate packets
def f(x):
    com = str(str(x[1]) + str(x[2]) + str(x[3]) + str(x[4])).encode('utf-8')
    return hashlib.md5(com).hexdigest()
df['code'] = df.apply(f, axis=1)

# TODO: why there are two packets?!
print('95 percentile of difference between duplicates:', np.percentile(df.groupby('code')['ts'].diff().fillna(0).values, 95))
print('99 percentile of difference between duplicates:', np.percentile(df.groupby('code')['ts'].diff().fillna(0).values, 99))

# we consider the mean of every two duplicate packets
mean_of_times = df.groupby('code')['ts'].mean().reset_index()
df.drop_duplicates(subset=['code'],keep='first',inplace=True)
df = df.drop('ts',1) #there will be 2 ts so we drop one

# merge the mean values of ts and get rid of code, we don't need it anymore
df = df.merge(mean_of_times, on='code')
df = df.drop('code', 1)

# some of the packets don't know their request type, we find them based on debugID
debugID2req = {}
for idx, row in df.iterrows():
    if not pd.isnull(row['req']):
        debugID2req[row['debugID']] = row['req']

df['req'] = df.apply(lambda row: debugID2req[row['debugID']] if pd.isnull(row['req']) and row['debugID'] in debugID2req else row['req'], axis=1)

# final analyse on data
data = dict(tuple(df.groupby('req')))
for key, df in data.items():
    byDebugID = list(tuple(df.groupby('debugID')))
    if len(byDebugID) == 0 : continue
    diffs = np.zeros((len(byDebugID), 9))
    for i, tmp in enumerate(byDebugID):
        debugID = tmp[0]
        request = tmp[1]
        request = request.sort_values(by="ts")
        if request.shape[0] != 10:
            continue
        diffs[i] = np.diff(request.ts)
    diffs = np.percentile(diffs, 95, axis=0)
    print(key, len(byDebugID), np.round(diffs, 3))