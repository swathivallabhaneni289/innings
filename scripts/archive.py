#!/usr/bin/env python3
"""Builds internal/sim/archive.json: Innings' own record of past tournament matches.

Wikipedia's tournament pages are the reference. Every match box ({{Single-innings cricket match}} and the older
templates) gives the date, both teams, scores, result, ground and the ESPNcricinfo scorecard link, whose number becomes
the match id ("ar-1384403"; a game without one is "ar-<day>-<team>-<team>"). Warm-up games are skipped, and so is a
match the Vault already has, as is a walkover nobody played.

It checks what it reads rather than trusting it. A tournament whose match count differs from the count on its own page
is left out and reported, never half-added. The result puts the side that batted first first ("won by 7 wickets": the
winner batted second), and a match whose margin doesn't fit its scores is left out. Every ground must be in a host
country. A rebuild that would lose a match already in archive.json stops, as a fan may have logged it.

    python3 scripts/archive.py                      # every tournament below
    python3 scripts/archive.py --only "2023 Cricket World Cup" --check   # parse and report, write nothing
    python3 scripts/archive.py --force              # write even though matches already in archive.json would go

Pages are read once through Wikipedia's API, ten to a request and a request every three seconds, and kept in
INNINGS_WIKI_CACHE (default: a folder in the system temp dir) so a rerun asks only for pages it hasn't seen. Standard
library only.

The script talks to en.wikipedia.org over HTTPS and nothing else. It refuses a redirect to any other site, never
opens a link inside a page (a scorecard link is only read for its number), and keeps nothing but the JSON text the
API answers with, which it reads as data and never runs.
"""

import argparse
import calendar
import datetime as dt
import json
import os
import re
import sys
import tempfile
import time
import urllib.error
import urllib.parse
import urllib.request

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
OUT = os.path.join(ROOT, "internal", "sim", "archive.json")
HOST = "en.wikipedia.org"
API = f"https://{HOST}/w/api.php"
AGENT = "InningsArchive/0.1 (personal cricket diary app; one-time fact check)"
MAX_ANSWER = 8 << 20  # a tournament page's text is a few hundred KB; an answer this big isn't one
NEW_TEAM_INK = "#3A4250"  # the feed's colour for a team it has none for (internal/feed/mapping.go)


def T(page, competition, name, fmt, gender, label="", hosts=()):
    """One tournament: its Wikipedia page, the app's competition id, its full name, format and gender. label prefixes
    the stage when the competition id alone doesn't name it ("Asia Cup · Final"); hosts overrides the host countries
    the page's infobox lists."""
    return dict(page=page, competition=competition, name=name, format=fmt, gender=gender, label=label, hosts=hosts)


def world_cups():
    out = []
    for y in (1975, 1979, 1983, 1987, 1992, 1996, 1999, 2003, 2007, 2011, 2015, 2019, 2023):
        out.append(T(f"{y} Cricket World Cup", "world-cup", f"ICC Cricket World Cup {y}", "ODI", "men"))
    for y in (2007, 2009, 2010, 2012, 2014, 2016):
        out.append(T(f"{y} ICC World Twenty20", "world-cup", f"ICC World Twenty20 {y}", "T20", "men"))
    for y in (2021, 2022, 2024, 2026):
        out.append(T(f"{y} ICC Men's T20 World Cup", "world-cup", f"ICC Men's T20 World Cup {y}", "T20", "men"))
    for y in (1973, 1978, 1982, 1988, 1993, 1997, 2000, 2005, 2009, 2013, 2017, 2022, 2025):
        out.append(T(f"{y} Women's Cricket World Cup", "world-cup", f"ICC Women's Cricket World Cup {y}", "ODI", "women"))
    for y in (2009, 2010, 2012, 2014, 2016, 2018):
        out.append(T(f"{y} ICC Women's World Twenty20", "world-cup", f"ICC Women's World Twenty20 {y}", "T20", "women"))
    for y in (2020, 2023, 2024, 2026):
        out.append(T(f"{y} ICC Women's T20 World Cup", "world-cup", f"ICC Women's T20 World Cup {y}", "T20", "women"))
    return out


def others():
    out = []
    for y in (1998, 2000):
        out.append(T(f"{y} ICC KnockOut Trophy", "international", f"ICC KnockOut Trophy {y}", "ODI", "men", "KnockOut Trophy"))
    for y in (2002, 2004, 2006, 2009, 2013, 2017, 2025):
        out.append(T(f"{y} ICC Champions Trophy", "international", f"ICC Champions Trophy {y}", "ODI", "men", "Champions Trophy"))
    asia = {1984: "ODI", 1986: "ODI", 1988: "ODI", "1990–91": "ODI", 1995: "ODI", 1997: "ODI", 2000: "ODI", 2004: "ODI",
            2008: "ODI", 2010: "ODI", 2012: "ODI", 2014: "ODI", 2016: "T20", 2018: "ODI", 2022: "T20", 2023: "ODI", 2025: "T20"}
    for y, f in asia.items():
        out.append(T(f"{y} Asia Cup", "international", f"Asia Cup {y}", f, "men", "Asia Cup"))
    away = {2009: ("South Africa",), 2014: ("India", "UAE"), 2020: ("UAE",), 2021: ("India", "UAE")}
    for y in range(2008, 2027):
        out.append(T(f"{y} Indian Premier League", "ipl", f"Indian Premier League {y}", "T20", "men", hosts=away.get(y, ("India",))))
    return out


TOURNAMENTS = world_cups() + others()

# Wikipedia's flag codes ({{cr|IND}}, {{crw|IND}}) and the team each stands for.
NATIONS = {
    "IND": "India", "AUS": "Australia", "ENG": "England", "SA": "South Africa", "RSA": "South Africa",
    "NZ": "New Zealand", "PAK": "Pakistan", "SL": "Sri Lanka", "WI": "West Indies", "WIN": "West Indies",
    "BAN": "Bangladesh", "AFG": "Afghanistan", "ZIM": "Zimbabwe", "IRE": "Ireland", "NED": "Netherlands",
    "SCO": "Scotland", "USA": "United States", "KEN": "Kenya", "NAM": "Namibia", "UAE": "United Arab Emirates",
    "CAN": "Canada", "NEP": "Nepal", "OMA": "Oman", "HK": "Hong Kong", "HKG": "Hong Kong", "PNG": "Papua New Guinea",
    "UGA": "Uganda", "EAF": "East Africa", "BER": "Bermuda", "ITA": "Italy", "THA": "Thailand", "DEN": "Denmark",
    "JAM": "Jamaica", "TTO": "Trinidad and Tobago", "JPN": "Japan", "NZL": "New Zealand", "SRI": "Sri Lanka",
    "LKA": "Sri Lanka", "SAF": "South Africa", "ZAF": "South Africa",
    # Women's sides that played only the first World Cups
    "YEN": "Young England", "IXI": "International XI",
}

# IPL sides by every name they've played under; renamed ones keep the app's id.
FRANCHISES = {
    "Chennai Super Kings": "csk", "Mumbai Indians": "mi", "Royal Challengers Bangalore": "rcb",
    "Royal Challengers Bengaluru": "rcb", "Kolkata Knight Riders": "kkr", "Sunrisers Hyderabad": "srh",
    "Rajasthan Royals": "rr", "Delhi Capitals": "dc", "Delhi Daredevils": "dc", "Punjab Kings": "pbks",
    "Kings XI Punjab": "pbks", "Gujarat Titans": "gt", "Lucknow Super Giants": "lsg",
}
# Sides that no longer play, with a short code and kit colour of their own.
DEFUNCT = {
    "Deccan Chargers": ("DCH", "#1C3F7A"), "Kochi Tuskers Kerala": ("KTK", "#5A2D82"),
    "Pune Warriors India": ("PWI", "#2CA9E1"), "Rising Pune Supergiant": ("RPS", "#6F2DA8"),
    "Rising Pune Supergiants": ("RPS", "#6F2DA8"), "Gujarat Lions": ("GL", "#E8611A"),
}

# Grounds' countries by city, named as the Vault names them ("UAE", "Barbados"). Hyderabad is India's; the World Cup
# games at Pakistan's Hyderabad link to "Hyderabad, Sindh", and a linked city is looked up before the shown one.
CITY_COUNTRY = {c.strip().lower(): country for country, cities in {
    "India": "Ahmedabad, Bengaluru, Bangalore, Chennai, Madras, Delhi, New Delhi, Dharamsala, Dharamshala, Guwahati, "
             "Hyderabad, Indore, Jaipur, Kanpur, Kolkata, Calcutta, Lucknow, Mohali, Mullanpur, New Chandigarh, Mumbai, "
             "Bombay, Navi Mumbai, Nagpur, Pune, Rajkot, Ranchi, Visakhapatnam, Cuttack, Raipur, Thiruvananthapuram, "
             "Gwalior, Jamshedpur, Vadodara, Margao, Kochi, Chandigarh, Jalandhar, Faridabad, Patna",
    "Pakistan": "Lahore, Karachi, Rawalpindi, Multan, Faisalabad, Peshawar, Gujranwala, Sialkot, Sheikhupura",
    "Sri Lanka": "Colombo, Kandy, Pallekele, Galle, Dambulla, Hambantota, Sooriyawewa, Moratuwa",
    "Bangladesh": "Dhaka, Mirpur, Chattogram, Chittagong, Fatullah, Sylhet, Khulna, Bogra",
    "UAE": "Dubai, Abu Dhabi, Sharjah",
    "England": "London, Birmingham, Manchester, Leeds, Nottingham, Southampton, Bristol, Taunton, Chester-le-Street, "
               "Hove, Chelmsford, Derby, Leicester, Worcester, Northampton, Canterbury, Tunbridge Wells, "
               "Royal Tunbridge Wells, Scarborough, "
               "Beckenham, Arundel",
    "Wales": "Cardiff, Swansea",
    "Australia": "Melbourne, Sydney, Adelaide, Perth, Brisbane, Hobart, Canberra, Geelong, Ballarat, Mackay, Albury",
    "New Zealand": "Auckland, Wellington, Christchurch, Hamilton, Dunedin, Napier, Nelson, Mount Maunganui, Queenstown, "
                   "New Plymouth, Whangarei, Lincoln, Tauranga",
    "South Africa": "Johannesburg, Cape Town, Durban, Centurion, Port Elizabeth, Gqeberha, Bloemfontein, Paarl, "
                    "East London, Potchefstroom, Kimberley, Benoni, Pietermaritzburg",
    "Barbados": "Bridgetown, Barbados",
    "Jamaica": "Kingston, Jamaica",
    "Trinidad and Tobago": "Port of Spain, Tarouba, San Fernando, Trinidad",
    "St Lucia": "Gros Islet, St Lucia, Saint Lucia",
    "Guyana": "Providence, Georgetown, Guyana",
    "Antigua": "North Sound, St John's, St. John's, Antigua",
    "St Kitts": "Basseterre, St Kitts, Saint Kitts",
    "Dominica": "Roseau, Dominica",
    "St Vincent": "Arnos Vale, Kingstown, St Vincent, Saint Vincent",
    "Grenada": "St. George's, St George's, Grenada",
    "Zimbabwe": "Harare, Bulawayo, Kwekwe",
    "Kenya": "Nairobi",
    "USA": "New York, Lauderhill, Dallas, Grand Prairie, East Meadow",
    "Ireland": "Dublin, Belfast, Malahide, Clontarf",
    "Scotland": "Edinburgh, Aberdeen, Glasgow",
    "Netherlands": "Amstelveen, Rotterdam, Deventer",
    "Oman": "Muscat, Al Amerat",
    "Malaysia": "Kuala Lumpur",
    "Canada": "Toronto, King City",
}.items() for c in cities.split(", ")}
CITY_COUNTRY["hyderabad, sindh"] = "Pakistan"
# Country names a page may write, as the Vault names them; the islands host as the West Indies.
COUNTRIES = {c: c for c in set(CITY_COUNTRY.values())} | {
    "United Arab Emirates": "UAE", "United States": "USA", "Saint Lucia": "St Lucia", "Antigua and Barbuda": "Antigua",
    "Saint Kitts and Nevis": "St Kitts", "Saint Vincent and the Grenadines": "St Vincent", "Trinidad": "Trinidad and Tobago",
    "Sri Lanka": "Sri Lanka", "West Indies": "West Indies", "Northern Ireland": "Ireland", "Namibia": "Namibia"}
ISLANDS = {"Barbados", "Jamaica", "Trinidad and Tobago", "St Lucia", "Guyana", "Antigua", "St Kitts", "Dominica",
           "St Vincent", "Grenada"}


def region(country):
    """The host a ground's country counts as: the islands host as the West Indies, Wales with England."""
    return "West Indies" if country in ISLANDS else "England" if country == "Wales" else country


def slug(s):
    return re.sub(r"[^a-z0-9]+", "-", s.lower()).strip("-")


# ---- reading Wikipedia ----

def cache_dir():
    d = os.environ.get("INNINGS_WIKI_CACHE") or os.path.join(tempfile.gettempdir(), "innings-wiki")
    os.makedirs(d, exist_ok=True)
    return d


class StayOnWikipedia(urllib.request.HTTPRedirectHandler):
    """Follows a redirect only within Wikipedia, so no other site is ever contacted."""

    def redirect_request(self, req, fp, code, msg, headers, newurl):
        u = urllib.parse.urlsplit(newurl)
        if u.scheme != "https" or u.hostname != HOST:
            raise urllib.error.HTTPError(newurl, code, f"refused a redirect away from {HOST}", headers, fp)
        return super().redirect_request(req, fp, code, msg, headers, newurl)


OPEN = urllib.request.build_opener(StayOnWikipedia).open
BATCH, GAP = 10, 3.0  # pages per request and seconds between requests: one page a second got "429 Too Many Requests"
_last = [0.0]


def cache_path(title):
    return os.path.join(cache_dir(), title.replace(" ", "_").replace("/", "_") + ".json")


def ask(titles):
    """One API request for up to BATCH pages, each kept in the cache on its own. A 429 or 503 means slow down: wait as
    long as Wikipedia's Retry-After says (a minute if it doesn't say), three times at most."""
    q = urllib.parse.urlencode({"action": "query", "prop": "revisions", "rvprop": "content", "rvslots": "main",
                                "format": "json", "formatversion": "2", "redirects": "1", "titles": "|".join(titles)})
    req = urllib.request.Request(f"{API}?{q}", headers={"User-Agent": AGENT})
    for attempt in range(4):
        time.sleep(max(0.0, GAP - (time.time() - _last[0])))
        try:
            with OPEN(req, timeout=60) as res:
                body = res.read(MAX_ANSWER + 1)
            break
        except urllib.error.HTTPError as e:
            if e.code not in (429, 503) or attempt == 3:
                raise
            after = e.headers.get("Retry-After", "")
            pause = int(after) if after.isdigit() else 60
            if pause > 900:
                sys.exit(f"Wikipedia asked us to wait {pause // 60} minutes; stopping. Pages read so far are kept.")
            print(f"     Wikipedia asked us to slow down; waiting {pause}s", file=sys.stderr)
            time.sleep(pause)
        finally:
            _last[0] = time.time()
    if len(body) > MAX_ANSWER:
        sys.exit(f"Wikipedia's answer for {titles[0]!r}… is over {MAX_ANSWER >> 20} MB, which no {BATCH} pages are; stopping")
    q = json.loads(body).get("query", {})  # anything but the API's JSON stops here, before it's kept
    moves = {r["from"]: r["to"] for step in ("normalized", "redirects") for r in q.get(step, [])}
    pages = {p["title"]: p for p in q.get("pages", [])}
    for t in titles:
        name, hops = t, 0
        while name in moves and hops < 5:
            name, hops = moves[name], hops + 1
        page = pages.get(name)
        if page and (page.get("missing") or page.get("invalid") or page.get("revisions")):  # else it didn't fit; asked alone later
            with open(cache_path(t), "w") as f:
                json.dump({"query": {"pages": [page]}}, f)


def fetch(titles):
    """Reads whichever of these pages aren't in the cache yet, BATCH to a request."""
    todo = [t for t in dict.fromkeys(titles) if not os.path.exists(cache_path(t))]
    for i in range(0, len(todo), BATCH):
        print(f"     reading {len(todo[i:i + BATCH])} pages from Wikipedia ({i + len(todo[i:i + BATCH])}/{len(todo)})", file=sys.stderr)
        ask(todo[i:i + BATCH])


def wikitext(title):
    """The page's wikitext (following redirects), or None if Wikipedia has no such page."""
    fetch([title])
    if not os.path.exists(cache_path(title)):
        sys.exit(f"{title}: Wikipedia's answer didn't include the page; stopping")
    with open(cache_path(title)) as f:
        page = json.load(f)["query"]["pages"][0]
    if page.get("missing") or page.get("invalid"):
        return None
    return page["revisions"][0]["slots"]["main"]["content"]


# ---- wikitext ----

BOX = re.compile(r"\{\{\s*(?:Single-innings cricket match(?: collapsible)?|Limited overs matches|Limited overs international|"
                 r"Cricket match summary)\s*[|\n]", re.I)
HEADING = re.compile(r"^(={2,6})\s*(.*?)\s*\1\s*$", re.M)
# "Road to the final" repeats a finalist's earlier games, often with dates out of step with the season's own page.
SKIP_SECTION = re.compile(r"warm[- ]?up|practice|qualification(?! play)|tour match|exhibition|road to the final", re.I)


def template_at(text, start):
    """The whole {{…}} starting at start, nested templates and links included."""
    depth, i = 0, start
    while i < len(text):
        two = text[i:i + 2]
        if two == "{{":
            depth += 1
            i += 2
        elif two == "}}":
            depth -= 1
            i += 2
            if depth == 0:
                return text[start:i]
        else:
            i += 1
    return text[start:]


def params(template):
    """name → raw value for a template's top-level |name = value pairs."""
    body = template[2:-2]
    parts, depth, cur, i = [], 0, [], 0
    while i < len(body):
        two = body[i:i + 2]
        if two in ("{{", "[["):
            depth += 1
            cur.append(two)
            i += 2
        elif two in ("}}", "]]"):
            depth -= 1
            cur.append(two)
            i += 2
        elif body[i] == "|" and depth == 0:
            parts.append("".join(cur))
            cur = []
            i += 1
        else:
            cur.append(body[i])
            i += 1
    parts.append("".join(cur))
    out = {}
    for p in parts[1:]:
        if "=" in p:
            k, v = p.split("=", 1)
            out[k.strip().lower()] = v.strip()
    return out


NOTES = re.compile(r"\{\{\s*(?:efn|refn|sfn|cn|citation needed|ref label|note)[^{}]*\}\}", re.I)


def clean(v):
    """Wikitext to plain text: notes and references go, links and wrappers ({{nowrap|(DLS method)}}) keep their words."""
    v = re.sub(r"<!--.*?-->", "", v, flags=re.S)
    v = re.sub(r"<ref[^>]*/>", "", v)
    v = re.sub(r"<ref[^>]*>.*?</ref>", "", v, flags=re.S)
    v = NOTES.sub("", v)
    v = re.sub(r"\[\[(?:[^|\]]*\|)?([^\]]*)\]\]", r"\1", v)
    v = re.sub(r"\[https?://\S+\s+([^\]]*)\]", r"\1", v)
    v = re.sub(r"\{\{\s*(?:nowrap|nobr|small|smaller)\s*\|([^{}]*)\}\}", r"\1", v, flags=re.I)
    v = re.sub(r"\{\{\s*(?:abbr|tooltip|abbrlink)\s*\|([^|{}]*)\|[^{}]*\}\}", r"\1", v, flags=re.I)
    v = re.sub(r"\{\{\s*(DLS|D/L|DL|VJD)\s*\}\}", r"\1", v)
    for _ in range(3):
        v = re.sub(r"\{\{[^{}]*\}\}", "", v)
    v = re.sub(r"'{2,}", "", v)
    v = re.sub(r"<[^>]+>", " ", v)
    v = v.replace("&nbsp;", " ")
    return " ".join(v.split())


def sections(text):
    """(heading path, body) for each stretch of text between headings."""
    path, out, last = [], [], 0
    for m in HEADING.finditer(text):
        out.append((list(path), text[last:m.start()]))
        level = len(m.group(1))
        path = [p for p in path if p[0] < level] + [(level, clean(m.group(2)))]
        last = m.end()
    out.append((list(path), text[last:]))
    return out


# ---- one match ----

class Teams:
    """The app's teams (sim's table) plus any the archive adds, named the way the live feed would name them."""

    def __init__(self):
        self.byid, self.byname, self.added = {}, {}, {}
        src = open(os.path.join(ROOT, "internal", "sim", "teams.go")).read()
        for tid, name, short, colour in re.findall(r'\{"([^"]+)", "([^"]+)", "([^"]+)", "(#[0-9A-Fa-f]{6})"\}', src):
            gender = "women" if tid.endswith("-w") else "men"
            self.byid[tid] = dict(id=tid, name=name, short=short, colour=colour, gender=gender)
            self.byname[name.lower()] = tid

    def nation(self, code, gender):
        """A national side by flag code ("NZL") or name ("West Indies", "1=Trinidad & Tobago")."""
        code = re.sub(r"^\s*1\s*=", "", code).replace("&", "and").strip()
        name = NATIONS.get(code.upper()) or next((n for n in NATIONS.values() if n.lower() == code.lower()), None)
        if not name:
            raise ValueError(f"unknown flag code {code!r}")
        code = next(c for c, n in NATIONS.items() if n == name)
        tid = self.byname.get(name.lower()) or self._add(slug(name), name, code.upper(), NEW_TEAM_INK, "men", "international")
        if gender == "men":
            return tid
        wname = name + " Women"
        if wname.lower() in self.byname:
            return self.byname[wname.lower()]
        base = self.byid[tid]
        return self._add(tid + "-w", wname, base["short"] + "-W", base["colour"], "women", "international")

    def franchise(self, name):
        if name in FRANCHISES:
            return FRANCHISES[name]
        if name in DEFUNCT:
            short, colour = DEFUNCT[name]
            canonical = "Rising Pune Supergiant" if name.startswith("Rising Pune") else name
            return self.byname.get(canonical.lower()) or self._add(slug(canonical), canonical, short, colour, "men", "franchise")
        raise ValueError(f"unknown franchise {name!r}")

    def _add(self, tid, name, short, colour, gender, kind):
        t = dict(id=tid, name=name, short=short, colour=colour, gender=gender, kind=kind)
        self.byid[tid] = self.added[tid] = t
        self.byname[name.lower()] = tid
        return tid

    def names(self, tid):
        """Every name a result line might use for the team: "Royal Challengers Bangalore" is rcb too."""
        out = {self.byid[tid]["name"].lower()}
        out |= {n.lower() for n, i in FRANCHISES.items() if i == tid}
        out |= {n.lower() for n in DEFUNCT if slug(n).startswith(tid)}  # "Rising Pune Supergiants" with an s
        return out | {n.removesuffix(" women") for n in out}


FLAG = re.compile(r"\{\{\s*cr(w)?(?:-rt)?\s*\|\s*([^|{}]+)", re.I)
SCORE = re.compile(r"^(\d+)(?:\s*/\s*(\d+))?\s*(?:d(?:ec)?\.?|all\s*out)?\s*\(\s*(\d+(?:\.\d)?)\s*ov", re.I)  # "overs" or "ov"
CRICINFO = re.compile(r"https?://[^\s\]|]*cricinfo\.com/[^\s\]|]*")  # espncricinfo.com, and content.cricinfo.com before it
MONTHS = {m.lower(): i for i, m in enumerate(calendar.month_name) if m} | {m.lower(): i for i, m in enumerate(calendar.month_abbr) if m}


def team_of(raw, gender, teams):
    """The side named in a box. The tournament says men's or women's: a men's final's page may use the women's flag."""
    m = FLAG.search(raw)
    if m:
        return teams.nation(m.group(2), gender)
    name = re.sub(r"\(H\)", "", clean(raw)).strip()
    if gender == "women" and name.lower().endswith(" women"):
        name = name[: -len(" women")]
    if name.upper() in NATIONS:
        return teams.nation(name, gender)
    for code, n in NATIONS.items():
        if n.lower() == name.lower():
            return teams.nation(code, gender)
    return teams.franchise(name)


def date_of(raw, page):
    """"9 June 1983", "June 9, 1983", "13th October 2006", "13–15 June 1979" (its first day), {{Start date|2024|06|01}},
    or "16 February" with the year taken from the page ("1990–91 Asia Cup": January is 1991's)."""
    m = re.search(r"\{\{\s*start date\s*\|([^{}]*)\}\}", raw, re.I)
    if m:
        y, mo, d = [int(x) for x in m.group(1).split("|") if x.strip().isdigit()][:3]
        return dt.date(y, mo, d)
    s = re.sub(r"\(.*?\)", "", clean(raw))
    # "13–15 June 1979", "June 13–15": a game that ran into its reserve days is dated by the day it began
    s = re.sub(r"\b(\d{1,2})(?:st|nd|rd|th)?\s*(?:[–-]|&ndash;|to)\s*\d{1,2}(?:st|nd|rd|th)?\b", r"\1", s)
    if iso := re.search(r"\b(\d{4})-(\d{2})-(\d{2})\b", s):
        return dt.date(*map(int, iso.groups()))
    m = re.search(r"\b(\d{1,2})(?:st|nd|rd|th)?\s+([A-Za-z]+)\.?(?:,?\s+(\d{4}))?\b", s)
    day, mon, year = (m.group(1), m.group(2), m.group(3)) if m and m.group(2).lower() in MONTHS else (None, None, None)
    if day is None and (m := re.search(r"\b([A-Za-z]+)\.?\s+(\d{1,2})(?:st|nd|rd|th)?(?:,?\s+(\d{4}))?\b", s)):
        if m.group(1).lower() in MONTHS:
            day, mon, year = m.group(2), m.group(1), m.group(3)
    if day is None:
        raise ValueError(f"unreadable date {raw!r}")
    month = MONTHS[mon.lower()]
    if year is None:
        year = int(page[:4]) + (1 if page[4] in "–-" and month <= 6 else 0)
    return dt.date(int(year), month, int(day))


def innings_of(raw, tid):
    s = clean(raw)
    m = SCORE.match(s)
    if not m:
        return None
    runs, wkts, overs = int(m.group(1)), m.group(2), m.group(3)
    return dict(teamId=tid, runs=runs, wickets=int(wkts) if wkts is not None else 10, overs=overs)


DEMONYMS = {"australian": "australia", "indian": "india", "english": "england", "pakistani": "pakistan",
            "sri lankan": "sri lanka", "south african": "south africa", "west indian": "west indies",
            "bangladeshi": "bangladesh", "irish": "ireland", "dutch": "netherlands", "scottish": "scotland"}


def winner_of(result, a, b, teams):
    """The side the result line says won, by its full name, part of it or a demonym ("Kolkata won by 140 runs",
    "Australian Women won by 114 runs", "Match tied (South Africa won the second Super Over)"); None when nobody won.
    A winner who isn't one of the two sides is an error, never a draw."""
    r = result.lower().replace("&", "and")
    said = re.findall(r"(?:^|[(,;:]\s*)([a-z][a-z .'-]*?)\s+won\b", r)
    if not said:
        if re.search(r"\bwon\b", r):
            raise ValueError(f"can't tell who won: {result!r}")
        return None
    who = said[-1].strip()
    for d, n in DEMONYMS.items():
        who = re.sub(rf"\b{d}\b", n, who)
    fits = ([tid for tid in (a, b) if who in teams.names(tid)]  # "England" beats "Young England", which only contains it
            or [tid for tid in (a, b) if any(set(who.split()) <= set(n.split()) for n in teams.names(tid))])
    if len(fits) != 1:
        raise ValueError(f"can't tell who won: {result!r}")
    return fits[0]


def venue_of(raw, hosts):
    """The ground, its city and country. The country comes from a linked city ("[[Hyderabad, Sindh|Hyderabad]]"),
    then the city shown, then a country the page writes after it, then the only host; it must be one of the hosts."""
    raw = re.split(r"<br\s*/?>", raw, flags=re.I)[0]  # a final's crowd follows on the next line
    linked = [x.strip().lower() for x in re.findall(r"\[\[([^|\]]+)", raw)]
    s = clean(raw).strip(" .,")
    s = re.sub(r"\.\s+(?=(?:" + "|".join(map(re.escape, COUNTRIES)) + r")$)", ", ", s)  # "Lord's, London. England"
    parts = [x.strip() for x in s.split(",") if x.strip()]
    if not parts:
        raise ValueError("no ground")
    written = None
    while len(parts) > 2 and parts[-1] in COUNTRIES:
        written = written or COUNTRIES[parts.pop()]
    name, city = parts[0], ", ".join(parts[1:])
    in_name = next((c for c in CITY_COUNTRY if not city and re.search(rf"\b{re.escape(c)}\b", name.lower())), None)
    if in_name:  # "Melbourne Cricket Ground": the city is in the name
        city = re.search(rf"\b{re.escape(in_name)}\b", name, re.I).group(0)
    country = (next((CITY_COUNTRY[x] for x in linked if x in CITY_COUNTRY), None) or CITY_COUNTRY.get(city.lower())
               or next((CITY_COUNTRY[x.lower()] for x in parts[1:] if x.lower() in CITY_COUNTRY), None) or written
               or CITY_COUNTRY.get(in_name)  # "Sharjah Cricket Stadium", with no city after it
               or (next(iter(hosts)) if len(hosts) == 1 and "West Indies" not in hosts else None))  # an island, not "West Indies"
    if not country:
        raise ValueError(f"no country for {s!r}")
    if hosts and region(country) not in {region(h) for h in hosts}:
        raise ValueError(f"{s!r} is in {country}, not a host ({', '.join(sorted(hosts))})")
    return dict(id=slug(name), name=name, city=city, country=country)


MARGIN = re.compile(r"\bwon by (\d+) (run|wicket)s?\b", re.I)
REVISED = re.compile(r"D/L|DLS|Duckworth|VJD|revised|reduced|target|rain", re.I)


def shortened(inn):
    """The chase had fewer overs than the first innings and wasn't bowled out: its target was cut, so the runs margin
    isn't the difference in scores (the 1987 and 1992 rain rules as much as DLS)."""
    def balls(o):
        whole, _, part = o.partition(".")
        return int(whole) * 6 + int(part or 0)
    return inn[1]["wickets"] < 10 and bool(inn[0]["overs"] and inn[1]["overs"]) and balls(inn[1]["overs"]) < balls(inn[0]["overs"])


def settle(m):
    """Puts the side that batted first first, as the result has it ("won by 7 wickets": the winner batted second),
    and checks the margin against the scores. Returns what's wrong, or None."""
    a, b = m["teamIds"]
    hit = MARGIN.search(m["resultText"])
    if not m["winnerId"] or not hit:
        return None
    n, kind = int(hit.group(1)), hit.group(2).lower()
    first = m["winnerId"] if kind == "run" else (b if m["winnerId"] == a else a)
    if first != a:
        m["teamIds"], m["innings"] = [b, a], m["innings"][::-1]
    inn = m["innings"]
    if len(inn) != 2 or inn[0]["teamId"] != first:
        return f"{m['id']} ({m['date']}): scores {inn} don't fit {m['resultText']!r}"
    if kind == "wicket" and inn[1]["wickets"] != 10 - n:
        return f"{m['id']} ({m['date']}): won by {n} wickets, but {inn[1]['wickets']} fell"
    if kind == "run" and not REVISED.search(m["resultText"]) and not shortened(inn) and inn[0]["runs"] - inn[1]["runs"] != n:
        return f"{m['id']} ({m['date']}): won by {n} runs, but the scores are {inn[0]['runs']} and {inn[1]['runs']}"
    return None


GENERIC = re.compile(r"^(?:matches|fixtures|results|match results|fixtures and results|results and fixtures|"
                     r"fixtures & results|match summary|summary|scorecards?|match details|schedule|match|\d{1,2} [a-z]+)$", re.I)


def stage_of(path, p, page_stage):
    if clean(p.get("round", "")) and not re.search(r"\bN/?A\b", clean(p["round"])):  # "Match N/A": a voided game
        return clean(p["round"])
    for _, h in reversed(path):
        if re.search(r"\bvs?\.?\b", h) or GENERIC.match(h):
            continue
        return h
    return page_stage


def named(raw):
    """Flags written into a line as the sides they stand for: "{{crw|AUS}} won by 1 run" is "Australia won by 1 run"."""
    def name(m):
        code = re.sub(r"^\s*1\s*=", "", m.group(1)).replace("&", "and").strip()
        return NATIONS.get(code.upper()) or code
    return re.sub(r"\{\{\s*cr(?:w|icon|iconw)?(?:-rt)?\s*\|\s*([^|{}]+)[^{}]*\}\}", name, raw, flags=re.I)


def own_id(day, a, b):
    """The id of a match with no scorecard number to go by: its day and both teams ("ar-2000-12-23-aus-w-nz-w")."""
    a, b = sorted((a, b))
    return f"ar-{day}-{a}-{b}"


def game(m):
    """A match's day and its two teams, in either order: the same game on any page."""
    return m["date"], frozenset(m["teamIds"])


def match_of(p, t, path, page_stage, teams, hosts):
    a, b = team_of(p["team1"], t["gender"], teams), team_of(p["team2"], t["gender"], teams)
    day = date_of(p.get("date", ""), t["page"])
    result = clean(named(p.get("result", "") or p.get("outcome", "")))
    result = re.sub(r"\s*\((?:\s*[A-Z][a-z]{1,3}\s*,)+\s*[A-Z][a-z]{1,3}\s*\)", "", result)  # "(Ind , Pak )": a bowl-out's marks, gone
    if not result:
        raise ValueError(f"no result line ({day}, {a} v {b})")
    innings = [i for i in (innings_of(p.get("score1", ""), a), innings_of(p.get("score2", ""), b)) if i]
    # The scorecard link's last number is the match's ("…-1367856/pakistan-vs-india-12th-match-1384403/full-scorecard").
    # Some pages cite it in the result's footnote instead. Only a scorecard link counts, not a news story, and an archived
    # copy's 14-digit timestamp isn't a match number.
    urls = CRICINFO.findall(" ".join(p.get(k, "") for k in ("report", "scorecard", "result")))
    url = next((u for u in urls if re.search(r"match|scorecard", u, re.I) and "/story/" not in u), "")
    nums = re.findall(r"(?<!\d)\d{5,8}(?!\d)", url)
    mid = f"ar-{nums[-1]}" if nums else own_id(day.isoformat(), a, b)
    stage = stage_of(path, p, page_stage)
    if t["label"]:
        stage = f"{t['label']} · {stage}"
    return dict(id=mid, date=day.isoformat(), competitionId=t["competition"], competitionName=t["name"], stage=stage,
                format=t["format"], gender=t["gender"], teamIds=[a, b], venue=venue_of(p.get("venue", ""), hosts),
                winnerId=winner_of(result, a, b, teams), resultText=result, innings=innings)


# ---- one tournament ----

SUBPAGE = re.compile(r"group stage|pool [a-d]\b|group [a-d]\b|super (?:6|six|8|eight|10|12|four|4|sixes)|knockout|first round|second round|"
                     r"preliminary|playoffs?|league stage|final|round", re.I)
NOT_SUBPAGE = re.compile(r"squads|officials|statistics|qualif|warm|broadcast|controvers|records|personnel|auction", re.I)


def plain(title):
    """A title without the words a sub-page's title may drop: "2025 Champions Trophy final" is 2025 ICC Champions Trophy's."""
    return re.sub(r"\b(?:ICC|Men's) ", "", title)


def pages_of(t, main):
    titles, base = [t["page"]], plain(t["page"])
    for links in re.findall(r"\{\{\s*(?:main|further|see also)\s*\|([^}]*)\}\}", main, re.I):
        for title in links.split("|"):
            title = title.split("#")[0].strip()
            if (plain(title).startswith(base) and title != t["page"] and SUBPAGE.search(plain(title)[len(base):])
                    and not NOT_SUBPAGE.search(title) and title not in titles):
                titles.append(title)
    return titles


def hosts_of(t, main):
    """The host countries: the tournament's own list, else the ones its infobox names ("{{ubl|England|Wales}}")."""
    if t["hosts"]:
        return set(t["hosts"])
    m = re.search(r"^\s*\|\s*host(?:s|_countries|_country)?\s*=(.*?)(?=^\s*\||^\}\})", main, re.M | re.S | re.I)
    if not m:
        return set()
    v = re.sub(r"\{\{\s*(?:flag\w*|cr\w*)\s*\|\s*([^|{}]+)[^{}]*\}\}", r"\1", NOTES.sub("", m.group(1)), flags=re.I)
    v = re.sub(r"\{\{\s*(?:ubl|unbulleted list|plainlist|flatlist|hlist|nowrap)\s*\|([^{}]*)\}\}",
               lambda x: x.group(1).replace("|", ", "), v, flags=re.I)
    v = clean(v)
    found = {c for name, c in COUNTRIES.items() if re.search(rf"\b{re.escape(name)}\b", v)}
    return found | {COUNTRIES.get(NATIONS[c], NATIONS[c]) for c in re.findall(r"\b[A-Z]{2,3}\b", v) if c in NATIONS}


WALKOVER = re.compile(r"walkover|forfeit|by default|awarded|conceded", re.I)

# Where a tournament really had more matches than its infobox counts, checked game by game; anything else that differs
# from the infobox leaves the tournament out.
COUNTS = {
    "2012 ICC Women's World Twenty20": (17, "the two play-offs for fifth and seventh aren't in the infobox's 15"),
    "1997 Asia Cup": (8, "India v Pakistan was rained off on 20 July and abandoned again on 21 July"),
    "2025 Indian Premier League": (75, "Punjab v Delhi, stopped at Dharamsala on 8 May, was voided and played again"),
}


def expected_count(t, main):
    if t["page"] in COUNTS:
        return COUNTS[t["page"]][0]
    m = re.search(r"\|\s*matches\s*=\s*([^\n|]*)", main)
    if not m:
        return None
    n = re.match(r"\s*(\d+)", clean(m.group(1)))
    return int(n.group(1)) if n else None


def tournament(t, teams, vault):
    """The tournament's matches, minus the Vault's, and what's wrong.

    Sub-pages are read first: their boxes carry detail ("(D/L method)") that the main page's copies drop, and the same
    game on another page (same day, same teams) is a copy. A copy whose scores and result agree beats one whose don't.
    A scorecard number that two different games share was pasted by mistake, so neither keeps it. A box that won't read
    only matters when the count comes up short, as the copy that reads is enough. A match whose scores and result still
    disagree is left out on its own; more than two in one tournament means the reading is wrong, and it's left out."""
    main = wikitext(t["page"])
    if main is None:
        return None, "no such page"
    hosts = hosts_of(t, main)
    boxes, unread, problems = [], [], [] if hosts else ["the page names no host country"]
    for title in pages_of(t, main)[1:] + [t["page"]]:
        text = main if title == t["page"] else wikitext(title)
        if text is None:
            problems.append(f"no page {title!r}")
            continue
        stage = plain(title)[len(plain(t["page"])):].strip()
        page_stage = stage[:1].upper() + stage[1:] or "Match"
        for path, body in sections(text):
            if any(SKIP_SECTION.search(h) for _, h in path):
                continue
            for box in BOX.finditer(body):
                p = params(template_at(body, box.start()))
                try:
                    boxes.append(match_of(p, t, path, page_stage, teams, hosts))
                except (ValueError, KeyError) as e:
                    unread.append(f"{title}: {e}")
    # The same scorecard and teams on two pages is one game even when a page has another date: a game that ran into
    # its reserve day is dated by the day it began, as scorecards date it.
    first, clash, began, notes = {}, set(), {}, []
    for m in boxes:
        if set(first.setdefault(m["id"], m)["teamIds"]) != set(m["teamIds"]):
            clash.add(m["id"])
        began[m["id"]] = min(began.get(m["id"], m["date"]), m["date"])
    for m in boxes:
        if m["id"] not in clash and m["date"] != began[m["id"]]:
            notes.append(f"{m['id']} is dated {m['date']} on one page and {began[m['id']]} on another; kept {began[m['id']]}")
            m["date"] = began[m["id"]]
    found, flagged = {}, {}
    for m in boxes:
        if m["id"] in clash:
            m["id"] = own_id(m["date"], *m["teamIds"])
        g, wrong = game(m), settle(m)
        if g not in found or (g in flagged and not wrong):
            if g in found and not found[g]["id"].startswith(f"ar-{m['date']}") and m["id"].startswith(f"ar-{m['date']}"):
                m["id"] = found[g]["id"]  # keep the scorecard number the first copy had
            found[g] = m
            flagged.pop(g, None)
            if wrong:
                flagged[g] = wrong
        elif found[g]["id"].startswith(f"ar-{m['date']}") and not m["id"].startswith(f"ar-{m['date']}"):
            found[g]["id"] = m["id"]  # a later copy has the scorecard number
    matches = sorted(found.values(), key=lambda m: (m["date"], m["id"]))
    want = expected_count(t, main)
    if want is None or want != len(matches):
        problems += [f"page says {want} matches, found {len(matches)}"] + unread
    if len(flagged) > 2:
        problems += list(flagged.values())
    # A walkover counts in the infobox but was never played, so nobody can have been there.
    unplayed = {g for g, m in found.items() if WALKOVER.search(m["resultText"]) and not m["innings"]}
    kept = [m for g, m in sorted(found.items(), key=lambda x: (x[1]["date"], x[1]["id"]))
            if not in_vault(m, vault) and g not in flagged and g not in unplayed]
    return dict(matches=kept, found=len(matches), want=want, problems=problems, unread=unread, flagged=flagged,
                notes=sorted(set(notes)), unplayed=len(unplayed)), None


def in_vault(m, vault):
    """The Vault has this game: same teams, and the same day or one either side (a reserve day dated differently)."""
    day = dt.date.fromisoformat(m["date"])
    return any(((day + dt.timedelta(days=n)).isoformat(), frozenset(m["teamIds"])) in vault for n in (-1, 0, 1))


def vault_keys():
    seeds = json.load(open(os.path.join(ROOT, "internal", "sim", "classics.json")))
    return {(s["date"], frozenset(s["teamIds"])) for s in seeds}


def main():
    ap = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    ap.add_argument("--only", action="append", help="a tournament's page title (repeatable)")
    ap.add_argument("--check", action="store_true", help="report only; don't write archive.json")
    ap.add_argument("--force", action="store_true", help="write even if matches already in archive.json would go")
    args = ap.parse_args()
    todo = [t for t in TOURNAMENTS if not args.only or t["page"] in args.only]
    fetch([t["page"] for t in todo])  # every main page first, ten to a request, then the sub-pages they point to
    fetch([s for t in todo if (text := wikitext(t["page"])) for s in pages_of(t, text)])
    teams, vault = Teams(), vault_keys()
    out, left_out, dropped = [], [], []
    for t in todo:
        res, err = tournament(t, teams, vault)
        if err:
            left_out.append((t["page"], err))
            print(f"  -  {t['page']}: {err}")
            continue
        status = "ok" if not res["problems"] else "LEFT OUT"
        print(f"{'ok' if status == 'ok' else '!!'} {t['page']}: {res['found']} matches (page says {res['want']}), "
              f"{res['found'] - len(res['matches']) - len(res['flagged']) - res['unplayed']} already in the Vault"
              + (f", {res['unplayed']} walkovers never played" if res["unplayed"] else "")
              + (f", {len(res['unread'])} repeated boxes unread" if res["unread"] and status == "ok" else ""))
        for p in res["problems"][:8]:
            print(f"     {p}")
        if status == "ok":
            out.extend(res["matches"])
            for p in res["flagged"].values():
                print(f"     left out, scores and result disagree: {p}")
                dropped.append(p)
            for n in res["notes"]:
                print(f"     note: {n}")
        else:
            left_out.append((t["page"], "; ".join(res["problems"][:3])))
    ids = [m["id"] for m in out]
    dupes = {i for i in ids if ids.count(i) > 1}
    if dupes:
        sys.exit(f"duplicate ids: {sorted(dupes)[:10]}")
    used = {tid for m in out for tid in m["teamIds"]}
    new_teams = sorted((t for tid, t in teams.added.items() if tid in used), key=lambda t: t["id"])
    print(f"\n{len(out)} matches from {len(todo) - len(left_out)} tournaments; {len(new_teams)} new teams; "
          f"{len(left_out)} tournaments left out; {len(dropped)} matches left out where Wikipedia's scores and result disagree")
    if args.check:
        return
    out.sort(key=lambda m: (m["date"], m["id"]))
    # A fan's log points at a match by id, so a rebuild (or an --only run) that would lose one stops here.
    if os.path.exists(OUT) and not args.force:
        gone = sorted({m["id"] for m in json.load(open(OUT))["matches"]} - {m["id"] for m in out})
        if gone:
            sys.exit(f"{len(gone)} matches in archive.json would go ({', '.join(gone[:5])}…), and a fan may have logged one. "
                     f"Check them, then rerun with --force.")
    with open(OUT, "w") as f:
        json.dump(dict(teams=new_teams, matches=out), f, ensure_ascii=False, indent=None, separators=(",", ":"))
        f.write("\n")
    print(f"wrote {os.path.relpath(OUT, ROOT)}")


if __name__ == "__main__":
    main()
