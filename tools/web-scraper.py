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
lower_id = 21620
upper_id = 39480

# CWD for where to place downloaded files, relative to here.
cwd = os.getcwd()

# Tuple for range getting crosswords; this is recalculated if we fetched from
# a known last id.
CRYPTIC = (lower_id, upper_id)

# The last read entry is stored externally.  If present, use it to go from
# there, to upper_id.a

lrid_file = cwd + "/tools/last_read_id"
print("looking at {}".format(lrid_file))
if os.path.isfile(lrid_file):
    with open(lrid_file, "r") as lrfile:
        lower_id = lrfile.read().replace('\n', '')

        # The next crossword...
        lower_id = int(lower_id) + 1
        CRYPTIC = (lower_id, upper_id)
        lrfile.close()

# Set the last_id_fetched to the lowest -- as we iterate over the range of
# IDs, this will get overwritten appropriately.
last_id_fetched = lower_id - 1

# If at first you don't succeed...
tries = 3
goes = 0

# Kick off!
for num in xrange(*CRYPTIC[:2]):
    try:
        url = "https://www.theguardian.com/crosswords/cryptic/" + str(num)
        result = requests.get(url)
        clues = ""
        clues1 = ""

        c = result.content
        soup = BeautifulSoup(c, "html5lib")
        # The web site seemingly has two different classes for storing the
        # crossword information -- try both, if they fail, we're doomed
        # anyway.
        try:
            clues1 = soup.find("div", {"class":"js-crossword "})
            clues = clues1["data-crossword-data"]
        except:
            clues1 = soup.find("div", {"class":"js-crossword has-grouped-clues"})
            clues = clues1["data-crossword-data"]

        # Serialise the JSON
        clues_json = json.loads(clues)

        # Create the directory if necessary!
        if not os.path.exists(cwd + "/crosswords/cryptic/setter/" + \
                clues_json["creator"]["name"]):
            os.makedirs(cwd + "/crosswords/cryptic/setter/" + \
                    clues_json["creator"]["name"]);

        save_name = cwd + "/crosswords/cryptic/setter/" + \
            clues_json["creator"]["name"] + "/" + str(num) + ".JSON"

        with open(save_name, "wb") as file:
            json.dump(clues_json, file, indent = 4)

        # Reset the counter; we want three consecutive misses to be a failure.
        goes -= 1
        if goes < 0:
            goes = 0

        print("Written {}/{} to disk...".format(clues_json["creator"]["name"], num))
        last_id_fetched = num
    except:
        goes += 1
        print("Couldn't find crossword: {}".format(num))
        if goes >= tries:
            with open(lrid_file, "wb") as lrfile:
                lrfile.write(str(last_id_fetched))
                lrfile.close()
            break
        else:
            continue
