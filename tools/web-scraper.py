#!/usr/bin/env python3.11

import requests
from bs4 import BeautifulSoup
import sys, traceback, pprint
import json, re, os.path, string, glob, numpy
import os.path, html

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

last_id_fetched = cryptic_lower_id
attempts = 0

# Grid globals
#
# Assign A = 0 -> Z = 25.  One less so that when manipulating arrays, we
# don't need to keep calculating an offset of -1.
alphanums = dict(list(zip(string.ascii_letters, [ord(c) % 32 - 1 for c in string.ascii_letters])))
all_grids = {}

# Set up the regexp
accessible_regexp = re.compile(r":\s*(?P<lights>.*?)\<\/li\>$")

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
        last_id_fetched = int(lower_id) + 1
        print(("Current cryptic id is: {}".format(lower_id)))

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

def accessible_url(cctype, ccid):
    url = "https://www.theguardian.com/crosswords/accessible/" + cctype + "/" + str(ccid)
    result = requests.get(url)
    if result.status_code != 200:
        return (False, None)
    return (True, result.content)

def grid_preamble():
    for f in glob.glob("grids/*.grid"):
        fname = os.path.basename(f)
        if fname.endswith('.grid'):
            # Chop off .grid at the end.
            fname = fname[:-5]
        data = numpy.loadtxt(f, delimiter=',')
        all_grids[fname] = {}
        all_grids[fname]['grid'] = data


def lookup_grid(num, content):
    soup = BeautifulSoup(content, "html5lib")
    clues = soup.find("div", {"class": "crossword__accessible-data"})
    all_clues = soup.find_all("li", {"class": "crossword__accessible-row-data"})

    this_grid = numpy.ones((15, 15))
    matched_grid = None

    row = 0
    for l in all_clues:
        newl = str(l)
        m = accessible_regexp.search(newl)
        if m is None:
            break
        for matches in m.group('lights').split(' '):
            if matches == '':
                continue
            try:
                this_grid[row][alphanums[matches]] = 0
            except IndexError as e:
                print("Got an IndexError")
                return None
        row += 1

    for g in all_grids:
        try:
            numpy.testing.assert_array_almost_equal(all_grids[g]['grid'], this_grid)
        except AssertionError as e:
            continue

        if numpy.array_equiv(all_grids[g]['grid'], this_grid):
            matched_grid = g
            break
    return matched_grid

def update_json(grid_info, json):
   json['_gridType'] = grid_info

def try_one(crossword_type, num):
            result = url_fetch(crossword_type, num)
            sc = result.status_code
            if sc == 404:
                    return -1

            clues = ""

            c = result.content
            soup = BeautifulSoup(c, "html5lib")
            # The web site seemingly has two different classes for storing the
            # crossword information -- try both, if they fail, we're doomed
            # anyway.
            gu_island_tag = soup.find("gu-island", {"name": "CrosswordComponent"})

            if gu_island_tag and gu_island_tag.has_attr("props"):
                props_content = gu_island_tag["props"]

                try:
                    clues = html.unescape(props_content)
                except TypeError:
                    print("Unable to convert HTML json to json")
                    return -1

            # Serialise the JSON
            try:
                clues_json = json.loads(clues)
            except ValueError:
                print(("Invalid JSON... ({})".format(clues)))
                return -1

            if not "creator" in clues_json["data"]:
                print(("Skipping {} as no creator key".format(num)))
                with open("/tmp/{}.json".format(num), "w") as file:
                    json.dump(clues_json, file, indent = 4)
                return -1

            creator = clues_json["data"]["creator"]["name"].rstrip()

            # Create the directory if necessary!
            if not os.path.exists(cwd + "/crosswords/" + crossword_type \
                    + "/setter/" + creator):
                        os.makedirs(cwd + "/crosswords/" + crossword_type + \
                                    "/setter/" + creator)

            save_name = cwd + "/crosswords/" + crossword_type + "/setter/" + \
                creator + "/" + str(num) + ".JSON"
            save_name = save_name.encode('utf-8')

            # Grid detection.
            #(works, content) = accessible_url(crossword_type, num)
            #if works:
            #    my_grid = lookup_grid(num, content)
            #    update_json(my_grid, clues_json)

            with open(save_name, "w") as file:
                json.dump(clues_json["data"], file, indent = 4)

            try:
                print(("{}: {}: {}".format(num, crossword_type, creator)))
                os.system("guardian-cc import {}".format(save_name.decode('utf-8')));
            except UnicodeEncodeError as e:
                print(("{}: OK".format(num)))

            return num
        
# Kick off!
lrid_file = cryptic_lrid_file
# If at first you don't succeed...
tries = 10

# Set up some preamble for grid comparisons
# grid_preamble()

for num in range(int(last_id_fetched), cryptic_upper_id):
        sys.stdout.flush()
        ret = try_one("cryptic", num)
        last_id_fetched = ret
        if ret != -1:
            with open(lrid_file, "w") as lrfile:
                lrfile.write(str(last_id_fetched))
                lrfile.close()
        ret = try_one("prize", num)
        last_id_fetched = ret
        if ret != -1:
            with open(lrid_file, "w") as lrfile:
                lrfile.write(str(last_id_fetched))
                lrfile.close()
        sys.stdout.flush()
        if attempts >= tries:
            break
