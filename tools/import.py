#!/usr/bin/env python3

from tinydb import TinyDB, Query
import sys, json, os.path, glob

# Initialise the DB
db = TinyDB('new.db')

#my_files = glob.glob('crosswords/cryptic/setter/**/*.JSON', recursive=True)
#
#for f in my_files:
#    with open(f, 'r') as jfile:
#        setter_json = json.load(jfile)
#        db.insert(setter_json)
#        print("Added {}...".format(jfile))

a = db.search(Query().entries.any(Query().solution == 'NUDGE'))
b = db.search(Query().creator.name.exists())

s = set()
for k in b:
    s.add(k['creator']['name'])

for k in sorted(s):
    print(k)
