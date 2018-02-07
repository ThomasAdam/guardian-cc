#!/usr/bin/env python2.7

from __future__ import print_function
import requests
from bs4 import BeautifulSoup
import json
import sys, traceback, pprint

# Guardian website only has data from circa 2000, these are
# the crossword numbers
CRYPTIC = (21620,27412, "cryptic")

CROSSWORD_TYPE = [CRYPTIC]

for crossword in CROSSWORD_TYPE:
    crossword_type = crossword[2]
    for crossword_no in xrange(*crossword[:2]):

        try:
            url = "https://www.theguardian.com/crosswords/" + crossword_type + "/" + str(crossword_no)
            result = requests.get(url)
            clues = ""
            clues1 = ""

            c = result.content
            # strip out JSON crossword data from the page
            soup = BeautifulSoup(c, "html5lib")
            try:
                clues1 = soup.find("div", {"class":"js-crossword "})
                clues = clues1["data-crossword-data"]
            except:
                clues1 = soup.find("div", {"class":"js-crossword has-grouped-clues"})
                clues = clues1["data-crossword-data"]


            clues_json = json.loads(clues)

            save_name = clues_json["id"] + ".JSON"

            with open(save_name, "wb") as file:
                json.dump(clues_json, file, indent=4)
        except:
            print("couldn't find crossword no:{}".format(crossword_no))
            continue
