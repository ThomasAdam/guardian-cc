#!/usr/bin/env python2.7

from __future__ import print_function
import requests
from bs4 import BeautifulSoup
import sys, traceback, pprint
import json

import os.path

"""
This file is licenced under the GPLv3
"""

"""
Utility to scrape the Guardian cryptic crossword website and extract the JSON
data for the crosswords.

This uses a crude sequence of numbers to derive the URL for the cryptic.
"""

# Guardian website only has data from 1999.  "30000" is very optimistic and
# should be future proof.
cryptic_lower_id = 21620
cryptic_upper_id = 50000

attempts = 0

# CWD for where to place downloaded files, relative to here.
cwd = os.getcwd()

# Tuple for range getting crosswords; this is recalculated if we fetched from
# a known last id.
CRYPTIC = (cryptic_lower_id, cryptic_upper_id, "cryptic")

# The last read entry is stored externally.  If present, use it to go from
# there, to upper_id.

cryptic_lrid_file = cwd + "/tools/cryptic_last_read_id"
if os.path.isfile(cryptic_lrid_file):
    with open(cryptic_lrid_file, "r") as clrfile:
        lower_id = clrfile.read().replace('\n', '')
        last_id_fetched = lower_id
        print("Current cryptic id is: {}".format(lower_id))

        # The next crossword...
        clrfile.close()

def url_fetch(cctype, ccid):
    url = "https://www.theguardian.com/crosswords/" + cctype + "/" + str(ccid)
    result = requests.get(url)
    global attempts
    if result.status_code == 200:
        attempts -= 1

    if result.status_code == 404:
        attempts += 1

    return result

def try_one(crossword_type, num):
            print("{}/{}... ".format(crossword_type, num), end='')
            result = url_fetch(crossword_type, num)
            sc = result.status_code
            if sc == 404:
                    return -1

            clues = ""
            clues1 = ""

            c = result.content
            soup = BeautifulSoup(c, "html5lib")
            # The web site seemingly has two different classes for storing the
            # crossword information -- try both, if they fail, we're doomed
            # anyway.
            clues1 = soup.find("div", {"class":"js-crossword"})
            if clues1 != None:
                clues = clues1["data-crossword-data"]
            else:
                clues1 = soup.find("div", {"class":"js-crossword has-grouped-clues"})
                if clues1 != None:
                    clues = clues1["data-crossword-data"]

            # Serialise the JSON
            try:
                clues_json = json.loads(clues)
            except ValueError:
                print("Invalid JSON... ({})".format(clues))
                return -1

            if not "creator" in clues_json:
                print("Skipping {} as no creator key".format(num))
                with open("/tmp/{}.json".format(num), "wb") as file:
                    json.dump(clues_json, file, indent = 4)
                return -1

            # Create the directory if necessary!
            if not os.path.exists(cwd + "/crosswords/" + crossword_type \
                    + "/setter/" + clues_json["creator"]["name"]):
                        os.makedirs(cwd + "/crosswords/" + crossword_type + \
                                    "/setter/" + clues_json["creator"]["name"])

            save_name = cwd + "/crosswords/" + crossword_type + "/setter/" + \
                clues_json["creator"]["name"] + "/" + str(num) + ".JSON"
            save_name = save_name.encode('utf-8')

            if os.path.exists(save_name):
                print("Skipping {} as it's already saved ({})".format(num, save_name))
            else:
                with open(save_name, "wb") as file:
                    json.dump(clues_json, file, indent = 4)

                print("Written [{}]: {}/{} to disk...".format(crossword_type, clues_json["creator"]["name"], num))
                os.system("./tools/import.pl {}".format(save_name));

            return num
        
# Kick off!
lrid_file = cryptic_lrid_file
# If at first you don't succeed...
tries = 3

for num in range(int(last_id_fetched), cryptic_upper_id):
        sys.stdout.flush()
        ret = try_one("cryptic", num)
        ret = try_one("prize", num)
        sys.stdout.flush()
        last_id_fetched = num
        if attempts >= tries:
            with open(lrid_file, "wb") as lrfile:
                lrfile.write(str(last_id_fetched))
                lrfile.close()
            break
