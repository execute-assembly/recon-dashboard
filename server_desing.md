# recon dashboard server

# overview
hello claude i am swapping from python too golang for the server

structure so far

```
server/
   cmd/main.go -> just handles checking if db and paths exist, if not creates them and the database then calls server.run()

   internal/
   		server/
   			server.go -> main server with routes, handles taking requests, dispatching to correct code

   		db/
   			db_ops.go -> handles all the reads, writes and updates to database

   		handlers/
   			import.go -> handles import json data from disk
```
 NOTE:
 	**this is what im just handling now, if i mention any features that werent in this file, consult this file again to see if its in here!!!!**



# Database structure

all databases will be stored under
**server2/database/<domain>_db.sql**

DOMAINS TABLE
```
domain_name string
status_code string
open_ports string formatted as port1, port2, port3 ...
title string
tech_stack string
content_type string
server string
IPs string formatted as IP1, IP2
CNAME string


JUICTY HITS TABLE
url string
statuc_code string
size string
severity string
```


# WE WILL HANDLE THOSE DATABASE STUFF FIRST AND THE ROUTES BELOW

ROUTES

```
**/api/hosts** -> reads from the domains table and returns all that data
**/api/hits** -> reads the juicy hits table and returns it
```
**NOTE**
	**THE SERVER WILL RETURN JSON TO THE FRONT END, SO request -> read db -> craft json -> return to user**