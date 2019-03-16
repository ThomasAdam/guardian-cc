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
cryptic_upper_id = 39480

prize_lower_id = 21622
prize_upper_id = 50000

# CWD for where to place downloaded files, relative to here.
cwd = os.getcwd()

# Tuple for range getting crosswords; this is recalculated if we fetched from
# a known last id.
CRYPTIC = (cryptic_lower_id, cryptic_upper_id, "cryptic")
PRIZE = (prize_lower_id, prize_upper_id, "prize")

# The last read entry is stored externally.  If present, use it to go from
# there, to upper_id.

cryptic_lrid_file = cwd + "/tools/cryptic_last_read_id"
if os.path.isfile(cryptic_lrid_file):
    with open(cryptic_lrid_file, "r") as clrfile:
        lower_id = clrfile.read().replace('\n', '')
        print("Current cryptic id is: {}".format(lower_id))

        # The next crossword...
        cryptic_lower_id = int(lower_id) + 1
        CRYPTIC = (cryptic_lower_id, cryptic_upper_id, "cryptic")
        clrfile.close()

prize_lrid_file = cwd + "/tools/prize_last_read_id"
if os.path.isfile(prize_lrid_file):
    with open(prize_lrid_file, "r") as clrfile:
        lower_id = clrfile.read().replace('\n', '')
        print("Current prize id is: {}".format(lower_id))

        # The next crossword...
        prize_lower_id = cryptic_lower_id - 2
        PRIZE = (prize_lower_id, prize_upper_id, "prize")
        clrfile.close()

# Set the last_id_fetched to the lowest -- as we iterate over the range of
# IDs, this will get overwritten appropriately.
cryptic_last_id_fetched = cryptic_lower_id - 1
prize_last_id_fetched = prize_lower_id + 1

# Kick off!
CROSSWORD_TYPE = [CRYPTIC, PRIZE]
for crossword in CROSSWORD_TYPE:
    crossword_type = crossword[2]

    # If at first you don't succeed...
    tries = 3
    goes = 0

    if crossword_type == "prize":
        lrid_file = prize_lrid_file
        last_id_fetched = prize_last_id_fetched
    else:
        lrid_file = cryptic_lrid_file
        last_id_fetched = cryptic_last_id_fetched

    for num in xrange(*crossword[:2]):
        try:
            url = "https://www.theguardian.com/crosswords/" + crossword_type + "/" + str(num)
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
            if not os.path.exists(cwd + "/crosswords/" + crossword_type + "/setter/" + \
                    clues_json["creator"]["name"]):
                os.makedirs(cwd + "/crosswords/" + crossword_type + "/setter/" + \
                        clues_json["creator"]["name"]);

            save_name = cwd + "/crosswords/" + crossword_type + "/setter/" + \
                clues_json["creator"]["name"] + "/" + str(num) + ".JSON"

            with open(save_name, "wb") as file:
                json.dump(clues_json, file, indent = 4)

            # Reset the counter; we want three consecutive misses to be a failure.
            goes -= 1
            if goes < 0:
                goes = 0

            print("Written [{}]: {}/{} to disk...".format(crossword_type, clues_json["creator"]["name"], num))
            last_id_fetched = num
            os.system("./tools/import.pl {}".format(save_name));
        except:
            goes += 1
            print("Couldn't find crossword: [{}]: {}".format(crossword_type, num))
            if goes >= tries:
                with open(lrid_file, "wb") as lrfile:
                    lrfile.write(str(last_id_fetched))
                    lrfile.close()
                break;
