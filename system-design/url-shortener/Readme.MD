# URL Shortener Service

## Step 1: Business case

URL shortener that will take URL and create shorted version of URL

Requirements:
- User redirected to the original URL when accessed to the short version
- User can add custom URL path, like: short.io/custom-name
- User can track simple analytics(Number of views), using admin-key 
generated with URL like: short.io/custom-name?adminkey=Dd3*Hf92
- User can add up to 10 URL/day
- Service should be able to work with 100M requests/day
- Service should be able to simply scale up/down if needed
- Service will save shortened URL for 2 month
 
## Step 2: KPM & Constraints

Key Performance Metrics
- Redirect response latency < 50ms
- Up time 99,99%
- Severs load 80%
- Analytics lag time up to 10min

### Constraints Start
Requirements for project launch

Traffic:
- Reads/Writes = 100/1
- Reads/Day = 99M/req
- Writes/Day = 1M/req
- Linear Load
- Read RPS = 1145
- Write RPS = 12

Bandwidth:
- Write = 12*500B = 20KB/s 
- Read = 1145*100B + 1145*500B = 0.7MB/s

Storage:
- DB 1M * 60days * 500B = 27GB/2 month (add some coef. for growth for few years)
- Cache 30% =  27*0.3 = 8.1GB/2 month

### Constraints Base
Requirements for project after some time of work
------//------

### Constraints Load
Requirements for project when best business scenario met
------//------

## Step 3: What data do we save
Here we describing the data with which we will be working.
It's not about DB schema, etc. It just to understand better what 
we working with, so to understand how this data can be handled
in the best way
 
Short URL object:
- original URL
- short URL
- admin key
- created_at

## Step 4: What 3rd party components we will use

Here we need to choose what components we can use for our solution:
DBs, Caches, Balancers, Reverse Proxies, Programming language, etc..

The main idea here is to pick what you are familiar with if it fit good for solution.
In case if you don't have a good fit, new component that you don't familiar with 
should be hard tested before use!

DB: 
As our data is small, have no relations and used only for read(only 1* time write)
we can use Key/Value or Column DB.
AWS DynamoDB, ScylaDB, Cassandra, Redis, Couchbase, etc

Cache:
Redis, AWS ElastiCache, Memcahced

Now check prices/features for each
For example Redis vs AWS ElastiCache
AWS EC2 Instance RAM 48GB, 24CPU, 25Gb/s - 870$/month
AWS ElastiCache RAM 52GB, 24CPU, 10Gb/s - 870$/month

Let's assume that for db and cache we will use Redis on EC2 instance.

**Why?**

This will minimize amount of different software that we will use, which is good,
as it minimize human resources needed and payment for licences if needed.

Also redis is good fit as K/V store and Cache, and also using some idea farther
we will drastically minimize amount of stored data. So our cache will be our DB.

## Step 5: High-level design

Here we build helicopter view on our architecture, describing the most 
top blocks and processes of our architecture.

As we need minimum latency we will need to use geo replication, to move servers,
to the location near to client. But we will do this only for pages with redirect.
For managing urls we will use one geo location, as we dont need such speed on write,
and this will simplify and make less expensive our solution.

That's why we have two types of services: for redirecting and for managing urls

We also will need service that will generate short urls for us and will store
(used/free/else urls data). So we have some batch already pregenerated urls

![High-level design](high-level-design.png)

## Step 6: Thinking and pre-testing

At this stage we already now what we want and from what blocks it may be build.
Here we trying to understand deeper how all these things will work, what and how
we will be using. 

Also we need to test some of our statements that may arise, like: 
- Can Redis handle such load with good latency on read?
- How we will replicate redis cluster to different geo-location, is it possible?
- Etc.

### Simplify Key Generation Service
Most of the internet for such URL shortener saying that we need KG Service with
additional DB for saving keys. But it will cost more and i dont like it. So
we will make it simpler.

Instead of managing used/free urls, we will just save one counter and thats all.
So instead saving 2-30GBs we will need to save few bytes only.

We have 7 symbols in our URL, from `aaaaaaa` to `zzzzzzz`. 
That's 8B combinations. Let's assume that `aaaaaaa` == 1 and 
`zzzzzzz` == 8B, `aaaaaab` == 2.
So what we need is just increment counter for a new url and translate counter 
to string based on our algorithm.
To minimize load on DB we can generate in batches by 100-1000, depends on load.
Also batch will give some randomness in url names.

We have 1M new urls per day, so 60M per 2 month. Each url lives only 2 month.
So when we will achieve 8B on our counter, url under index 1 will be already free to use.
So we have something like round-robing algorithm here.

The only thing that you will need to monitor here is that url adding rate was 
less than capacity of the 7symbols url.

### Redis TTL
We use redis as it gives grate functionality for our use case.
One of which is Key TTL, which removing from us need to write code for removing 
urls after some time.

### Cluster replication
Free Redis doesnt have cluster replication functionality, but we can simply add it
by creating proxy service that will replicate all comands from main cluster 
to read clusters. This is simple task because of simplicity of redis protocol.

It can be both Sync and Async task depends on speed and number of GEO clusters.

### Views statistics
For robust stats we will definitelly will need additional service group, 
but to save just views not.

But saving views during high-load also not very straightforward task.
We cant just run increment on redis, because even if it fast, millions of reads,
will create huge load for DB even in-memory one(Always protect DBs from HL).

So what we will do? On each redirect instance will be a thread that will save 
stats once a time/memory limits. Before that all views will besaved only in 
instance memory/disk. This will minimize number of calls to DB.

There is also another probabalistic hack that was popular 15 years ago.
We set a write probability, for example 20%. And once it will shoot, 
we will increment our views not by one, but by 5 (100/20). 
As you understand it not precise on small numbers, but on big one preaty good.

But we will use first one aproach, as the second not guruntee that randomnes 
will not shoot in our leg.

### Now we are ready to build detailed architecture


PS: For drawing I used this service: https://excalidraw.com/
