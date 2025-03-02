import pandas as pd
import raptor
from raptor.stub import *

# getting started code

@raptor.register(str, freshness='1m', staleness='10h', options={})
@raptor.connector("emails")
@raptor.builder("streaming")
@raptor.aggr([raptor.AggrFn.Count])
def emails_10h(**req: RaptorRequest):
    """email over 10 hours"""
    return 1, req["timestamp"], req['payload']['account_id']


@raptor.register(float, freshness='1m', staleness='10h', options={})
@raptor.connector("deals")
@raptor.builder("streaming")
@raptor.aggr([raptor.AggrFn.Sum, raptor.AggrFn.Avg, raptor.AggrFn.Max, raptor.AggrFn.Min])
def deals_10h(**req):
    """sum/avg/min/max of deal amount over 10 hours"""
    return req['payload']["amount"], req["timestamp"], req['payload']["account_id"]


@raptor.register('headless', freshness='-1', staleness='-1', options={})
def emails_deals(**req):
    """emails/deal[avg] rate over 10 hours"""
    e, _ = f("emails_10h.default[count]", req['entity_id'])
    d, _ = f("deals_10h.default[avg]", req['entity_id'])
    if e == None or d == None:
        return None
    return e / d


@raptor.feature_set(register=True)
def deal_prediction():
    return "emails_10h.default[count]", "deals_10h.default[sum]", emails_deals


@raptor.feature_set(register=True)
def deal_prediction():
    return "emails_10h.default[count]", "deals_10h.default[sum]", emails_deals


# first, calculate the "root" features
df = pd.read_parquet("https://gist.github.com/AlmogBaku/a1b331615eaf1284432d2eecc5fe60bc/raw/emails.parquet")
emails_10h.replay(df, entity_id_field="account_id")

df = pd.read_csv("https://gist.githubusercontent.com/AlmogBaku/a1b331615eaf1284432d2eecc5fe60bc/raw/deals.csv")
deals_10h.replay(df, entity_id_field="account_id")

# then, we can calculate the derrived features
emails_deals.replay(df, entity_id_field="account_id")

df = deal_prediction.historical_get(since=pd.to_datetime('2020-1-1'), until=pd.to_datetime('2022-12-31'))


## counters
@raptor.register(int, '-1', '-1')
def views(**req):
    incr_feature("views.default", req["entity_id"], 1)


df = pd.read_csv("https://gist.githubusercontent.com/AlmogBaku/a1b331615eaf1284432d2eecc5fe60bc/raw/deals.csv")
res = views.replay(df, entity_id_field="account_id")


## gong
crm_records_df = pd.DataFrame.from_records([
    {'event_at': '2022-01-01 12:00:00+00:00', 'salesman_id': 'ada',   'action': 'deal_assigned',  'opportunity_id': 15},
    {'event_at': '2022-02-01 13:10:00+00:00', 'salesman_id': 'ada',   'action': 'deal_removed',   'opportunity_id': 15},
    {'event_at': '2022-04-01 13:20:00+00:00', 'salesman_id': 'ada',   'action': 'deal_assigned',  'opportunity_id': 15},
    {'event_at': '2022-06-01 14:00:00+00:00', 'salesman_id': 'ada',   'action': 'deal_closed',    'opportunity_id': 25},
    {'event_at': '2022-06-01 14:10:00+00:00', 'salesman_id': 'ada',   'action': 'deal_assigned',  'opportunity_id': 17},
    {'event_at': '2022-07-01 14:20:00+00:00', 'salesman_id': 'ada',   'action': 'deal_removed',   'opportunity_id': 17},
    {'event_at': '2022-08-01 14:30:00+00:00', 'salesman_id': 'ada',   'action': 'deal_assigned',  'opportunity_id': 17},
    {'event_at': '2022-09-01 14:40:00+00:00', 'salesman_id': 'ada',   'action': 'deal_closed',    'opportunity_id': 17},
    {'event_at': '2022-11-01 15:30:00+00:00', 'salesman_id': 'ada',   'action': 'deal_removed',   'opportunity_id': 17},
    {'event_at': '2022-01-01 12:00:00+00:00', 'salesman_id': 'brian', 'action': 'deal_assigned',  'opportunity_id': 132},
    {'event_at': '2022-02-01 12:20:00+00:00', 'salesman_id': 'brian', 'action': 'deal_removed',   'opportunity_id': 132},
    {'event_at': '2022-02-01 13:40:00+00:00', 'salesman_id': 'brian', 'action': 'deal_assigned',  'opportunity_id': 132},
    {'event_at': '2022-04-01 15:00:00+00:00', 'salesman_id': 'brian', 'action': 'deal_closed',    'opportunity_id': 132},
    {'event_at': '2022-05-01 15:10:00+00:00', 'salesman_id': 'brian', 'action': 'deal_removed',   'opportunity_id': 132},
    {'event_at': '2022-06-01 15:20:00+00:00', 'salesman_id': 'brian', 'action': 'deal_assigned',  'opportunity_id': 544},
    {'event_at': '2022-07-01 15:30:00+00:00', 'salesman_id': 'brian', 'action': 'deal_removed',   'opportunity_id': 544},
    {'event_at': '2022-08-01 15:40:00+00:00', 'salesman_id': 'brian', 'action': 'deal_assigned',  'opportunity_id': 544},
    {'event_at': '2022-09-01 15:50:00+00:00', 'salesman_id': 'brian', 'action': 'deal_closed',    'opportunity_id': 544},
    {'event_at': '2022-10-01 16:00:00+00:00', 'salesman_id': 'brian', 'action': 'deal_assigned',  'opportunity_id': 233},
    {'event_at': '2022-11-01 16:10:00+00:00', 'salesman_id': 'brian', 'action': 'deal_closed',    'opportunity_id': 233},
    {'event_at': '2022-12-01 16:20:00+00:00', 'salesman_id': 'brian', 'action': 'deal_assigned',  'opportunity_id': 444},
    {'event_at': '2022-12-01 16:30:00+00:00', 'salesman_id': 'brian', 'action': 'deal_closed',    'opportunity_id': 444},
])

@raptor.register(int, freshness='24h', staleness='8760h')
@raptor.connector("crm_updates")
@raptor.aggr([raptor.AggrFn.DistinctCount])
def unique_deals_involvment_annualy(**req: RaptorRequest):
  if req['payload']['action'] == "deal_assigned":
    return req['payload']["opportunity_id"]
  return None

unique_deals_involvment_annualy.replay(crm_records_df, entity_id_field='salesman_id')

@raptor.register(int, freshness='24h', staleness='8760h')
@raptor.connector("crm_updates")
@raptor.aggr([raptor.AggrFn.Count])
def closed_deals_annualy(**req: RaptorRequest):
  if req['payload']['action'] == "deal_closed":
    return 1
  return None

closed_deals_annualy.replay(crm_records_df, entity_id_field='salesman_id')

@raptor.register(int, freshness='24h', staleness='8760h')
def salesperson_deals_closes_rate(**req: RaptorRequest):
  udia, _ = get_feature("unique_deals_involvment_annualy.default[distinct_count]", req['entity_id'])
  cda, _ = get_feature("closed_deals_annualy.default[count]", req['entity_id'])
  if udia == None or cda == None:
    return None
  return udia/cda

salesperson_deals_closes_rate.replay(crm_records_df, entity_id_field='salesman_id')

# other tests

@raptor.register(int, freshness='1m', staleness='10m', options={})
def simple(**req):
    age = req['payload']["age"]
    weight = req['payload']["weight"]
    if age == None or weight == None:
        return None
    return req['payload']["age"] / req['payload']["weight"]

## the above should fail

# df = pd.read_parquet("./nested.parquet")
# simple.replay(df, entity_id_field="name")


df = pd.DataFrame.from_records([
    {'event_at': '2022-01-01 12:00:00+00:00', 'account_id': 'ada', 'subject': 'wrote_code', 'commit_count': 1},
    {'event_at': '2022-01-01 13:10:00+00:00', 'account_id': 'ada', 'subject': 'wrote_code', 'commit_count': 1},
    {'event_at': '2022-01-01 13:20:00+00:00', 'account_id': 'ada', 'subject': 'fixed_bug', 'commit_count': 1},
    {'event_at': '2022-01-01 14:00:00+00:00', 'account_id': 'ada', 'subject': 'deployed', 'commit_count': 3},
    {'event_at': '2022-01-01 14:10:00+00:00', 'account_id': 'ada', 'subject': 'developed', 'commit_count': 1},
    {'event_at': '2022-01-01 14:20:00+00:00', 'account_id': 'ada', 'subject': 'built_model', 'commit_count': 4},
    {'event_at': '2022-01-01 14:30:00+00:00', 'account_id': 'ada', 'subject': 'wrote_code', 'commit_count': 3},
    {'event_at': '2022-01-01 14:40:00+00:00', 'account_id': 'ada', 'subject': 'experimented', 'commit_count': 2},
    {'event_at': '2022-01-01 15:30:00+00:00', 'account_id': 'ada', 'subject': 'wrote_code', 'commit_count': 1},
    {'event_at': '2022-01-01 12:00:00+00:00', 'account_id': 'brian', 'subject': 'developed', 'commit_count': 1},
    {'event_at': '2022-01-01 12:20:00+00:00', 'account_id': 'brian', 'subject': 'wrote_code', 'commit_count': 2},
    {'event_at': '2022-01-01 13:40:00+00:00', 'account_id': 'brian', 'subject': 'experimented', 'commit_count': 1},
    {'event_at': '2022-01-01 15:00:00+00:00', 'account_id': 'brian', 'subject': 'developed', 'commit_count': 1},
    {'event_at': '2022-01-01 15:10:00+00:00', 'account_id': 'brian', 'subject': 'wrote_code', 'commit_count': 4},
    {'event_at': '2022-01-01 15:20:00+00:00', 'account_id': 'brian', 'subject': 'developed', 'commit_count': 5},
    {'event_at': '2022-01-01 15:30:00+00:00', 'account_id': 'brian', 'subject': 'wrote_code', 'commit_count': 1},
    {'event_at': '2022-01-01 15:40:00+00:00', 'account_id': 'brian', 'subject': 'experimented', 'commit_count': 2},
    {'event_at': '2022-01-01 15:50:00+00:00', 'account_id': 'brian', 'subject': 'developed', 'commit_count': 1},
    {'event_at': '2022-01-01 16:00:00+00:00', 'account_id': 'brian', 'subject': 'wrote_code', 'commit_count': 2},
    {'event_at': '2022-01-01 16:10:00+00:00', 'account_id': 'brian', 'subject': 'built_model', 'commit_count': 1},
    {'event_at': '2022-01-01 16:20:00+00:00', 'account_id': 'brian', 'subject': 'built_model', 'commit_count': 1},
    {'event_at': '2022-01-01 16:30:00+00:00', 'account_id': 'brian', 'subject': 'experimented', 'commit_count': 3},
])

# convert `event_at` column from string to datetime
df['event_at'] = pd.to_datetime(df['event_at'])


@raptor.register(int, freshness='1m', staleness='10m', options={})
def simple(**req):
    pass


@raptor.register(str, freshness='10m', staleness='2h', options={})
@raptor.aggr([raptor.AggrFn.DistinctCount])
def unique_tasks_over_2h(**req):
    return req['payload']['subject']

unique_tasks_over_2h.replay(df, entity_id_field="account_id")

@raptor.register(int, freshness='1m', staleness='30m', options={})
@raptor.aggr([raptor.AggrFn.Sum, raptor.AggrFn.Count, raptor.AggrFn.Max])
def commits_30m(**req):
    """sum/max/count of commits over 30 minutes"""

    set_feature("simple.default", req["entity_id"], 55, req['timestamp'])
    incr_feature("simple.default", req["entity_id"], 55, req['timestamp'])
    update_feature("simple.default", req["entity_id"], 55, req['timestamp'])

    return req['payload']["commit_count"]


commits_30m.replay(df, entity_id_field="account_id")


@raptor.register(int, freshness='1m', staleness='30m', options={})
def commits_30m_greater_2(**req):
    res, ts = f("commits_30m.default[sum]", req['entity_id'])
    """sum/max/count of commits over 30 minutes"""
    return res > 2


commits_30m_greater_2.replay(df, entity_id_field='account_id')


@raptor.feature_set()
def newset():
    return "commits_30m.default[sum]", commits_30m_greater_2

print(raptor.manifests())

ret = newset.historical_get(since=pd.to_datetime('2019-12-04 00:00'), until=pd.to_datetime('2023-01-04 00:00'))
print(ret)
