package sim

import "github.com/swathivallabhaneni289/innings/internal/domain"

type ven struct{ id, name, short, city, country string }

var venueRows = []ven{
	{"wankhede-stadium", "Wankhede Stadium", "Wankhede", "Mumbai", "India"},
	{"m-chinnaswamy-stadium", "M. Chinnaswamy Stadium", "Chinnaswamy", "Bengaluru", "India"},
	{"eden-gardens", "Eden Gardens", "Eden Gardens", "Kolkata", "India"},
	{"arun-jaitley-stadium", "Arun Jaitley Stadium", "Kotla", "Delhi", "India"},
	{"narendra-modi-stadium", "Narendra Modi Stadium", "Motera", "Ahmedabad", "India"},
	{"ma-chidambaram-stadium", "M. A. Chidambaram Stadium", "Chepauk", "Chennai", "India"},
	{"sawai-mansingh-stadium", "Sawai Mansingh Stadium", "SMS", "Jaipur", "India"},
	{"brabourne-stadium", "Brabourne Stadium", "Brabourne", "Mumbai", "India"},
	{"dy-patil-stadium", "DY Patil Stadium", "DY Patil", "Navi Mumbai", "India"},
	{"sydney-cricket-ground", "Sydney Cricket Ground", "SCG", "Sydney", "Australia"},
	{"sydney-showground-stadium", "Sydney Showground Stadium", "the Showground", "Sydney", "Australia"},
	{"galle-international-stadium", "Galle International Stadium", "Galle", "Galle", "Sri Lanka"},
	{"sinhalese-sports-club-ground", "Sinhalese Sports Club Ground", "SSC", "Colombo", "Sri Lanka"},
	{"the-oval", "The Oval", "The Oval", "London", "England"},
	{"headingley", "Headingley", "Headingley", "Leeds", "England"},
	{"old-trafford", "Old Trafford", "Old Trafford", "Manchester", "England"},
	{"edgbaston", "Edgbaston", "Edgbaston", "Birmingham", "England"},
	{"the-rose-bowl", "The Rose Bowl", "the Rose Bowl", "Southampton", "England"},
	{"county-ground-bristol", "County Ground", "Bristol", "Bristol", "England"},
	{"county-ground-taunton", "Cooper Associates County Ground", "Taunton", "Taunton", "England"},
	{"trent-bridge", "Trent Bridge", "Trent Bridge", "Nottingham", "England"},
	{"newlands", "Newlands", "Newlands", "Cape Town", "South Africa"},
	{"supersport-park", "SuperSport Park", "Centurion", "Centurion", "South Africa"},
}

var venueTable = func() map[string]ven {
	m := map[string]ven{}
	for _, v := range venueRows {
		m[v.id] = v
	}
	return m
}()

func venue(id string) domain.Venue {
	v := venueTable[id]
	return domain.Venue{ID: v.id, Name: v.name, City: v.city, Country: v.country}
}

func venueShort(id string) string {
	if v, ok := venueTable[id]; ok {
		return v.short
	}
	return ""
}
