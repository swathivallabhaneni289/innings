package sim

import "github.com/swathivallabhaneni289/innings/internal/domain"

type teamRow struct{ id, name, short, colour string }

var menIntl = []teamRow{
	{"ind", "India", "IND", "#1F4FA3"},
	{"aus", "Australia", "AUS", "#FFC72C"},
	{"eng", "England", "ENG", "#1C3F94"},
	{"sa", "South Africa", "SA", "#007A4D"},
	{"nz", "New Zealand", "NZ", "#111111"},
	{"pak", "Pakistan", "PAK", "#01411C"},
	{"sl", "Sri Lanka", "SL", "#0B3D91"},
	{"wi", "West Indies", "WI", "#7B0041"},
	{"ban", "Bangladesh", "BAN", "#006A4E"},
	{"afg", "Afghanistan", "AFG", "#1E5AA8"},
	{"zim", "Zimbabwe", "ZIM", "#D40000"},
	{"ire", "Ireland", "IRE", "#169B62"},
	{"ned", "Netherlands", "NED", "#FF6A00"},
	{"sco", "Scotland", "SCO", "#4B2D7F"},
	{"usa", "United States", "USA", "#0A3161"},
	{"kenya", "Kenya", "KEN", "#006600"},
	{"nam", "Namibia", "NAM", "#003580"},
	{"uae", "United Arab Emirates", "UAE", "#00732F"},
}

var womenIntl = []teamRow{
	{"ind-w", "India Women", "IND-W", "#1F4FA3"},
	{"aus-w", "Australia Women", "AUS-W", "#FFC72C"},
	{"eng-w", "England Women", "ENG-W", "#1C3F94"},
	{"sa-w", "South Africa Women", "SA-W", "#007A4D"},
	{"nz-w", "New Zealand Women", "NZ-W", "#111111"},
	{"pak-w", "Pakistan Women", "PAK-W", "#01411C"},
	{"sl-w", "Sri Lanka Women", "SL-W", "#0B3D91"},
	{"wi-w", "West Indies Women", "WI-W", "#7B0041"},
	{"ban-w", "Bangladesh Women", "BAN-W", "#006A4E"},
	{"ire-w", "Ireland Women", "IRE-W", "#169B62"},
}

var menFranchise = []teamRow{
	// IPL
	{"csk", "Chennai Super Kings", "CSK", "#F9CD05"},
	{"mi", "Mumbai Indians", "MI", "#004BA0"},
	{"rcb", "Royal Challengers Bengaluru", "RCB", "#D11D26"},
	{"kkr", "Kolkata Knight Riders", "KKR", "#3A225D"},
	{"srh", "Sunrisers Hyderabad", "SRH", "#F26522"},
	{"rr", "Rajasthan Royals", "RR", "#EA1A85"},
	{"dc", "Delhi Capitals", "DC", "#17479E"},
	{"pbks", "Punjab Kings", "PBKS", "#DD1F2D"},
	{"gt", "Gujarat Titans", "GT", "#1B2133"},
	{"lsg", "Lucknow Super Giants", "LSG", "#0057E2"},
	// BBL
	{"sixers", "Sydney Sixers", "SIX", "#E2007A"},
	{"scorchers", "Perth Scorchers", "PRS", "#F37021"},
	{"heat", "Brisbane Heat", "HEA", "#00B2A9"},
	{"stars", "Melbourne Stars", "STA", "#00A651"},
	{"renegades", "Melbourne Renegades", "REN", "#E2231A"},
	{"strikers", "Adelaide Strikers", "STR", "#0085CA"},
	{"hurricanes", "Hobart Hurricanes", "HUR", "#5B2C83"},
	{"thunder", "Sydney Thunder", "THU", "#8DC63F"},
	// SA20
	{"mict", "MI Cape Town", "MICT", "#004BA0"},
	{"jsk", "Joburg Super Kings", "JSK", "#F9CD05"},
	{"pc", "Pretoria Capitals", "PC", "#17479E"},
	{"sec", "Sunrisers Eastern Cape", "SEC", "#F26522"},
	{"dsg", "Durban's Super Giants", "DSG", "#00AEEF"},
	{"prc", "Paarl Royals", "PR", "#EA1A85"},
	// The Hundred (men)
	{"invincibles", "Oval Invincibles", "OVI", "#00A19A"},
	{"brave", "Southern Brave", "SOB", "#00A651"},
	{"phoenix", "Birmingham Phoenix", "BPH", "#F15A22"},
	{"spirit", "London Spirit", "LNS", "#1B365D"},
	{"originals", "Manchester Originals", "MNR", "#6B2C91"},
	{"superchargers", "Northern Superchargers", "NOS", "#6CC24A"},
	{"rockets", "Trent Rockets", "TRT", "#F7D417"},
	{"fire", "Welsh Fire", "WEF", "#D50032"},
}

var womenFranchise = []teamRow{
	{"mi-w", "Mumbai Indians Women", "MI-W", "#004BA0"},
	{"rcb-w", "Royal Challengers Bengaluru Women", "RCB-W", "#D11D26"},
	{"dc-w", "Delhi Capitals Women", "DC-W", "#17479E"},
	{"upw", "UP Warriorz", "UPW", "#7B2CBF"},
	{"gg-w", "Gujarat Giants Women", "GG-W", "#E35205"},
}

var teamTable = buildTeams()

func buildTeams() map[string]domain.Team {
	m := map[string]domain.Team{}
	add := func(rows []teamRow, g domain.Gender, kind string) {
		for _, r := range rows {
			m[r.id] = domain.Team{ID: r.id, Name: r.name, Short: r.short, Colour: r.colour, Gender: g, Kind: kind}
		}
	}
	add(menIntl, domain.Men, "international")
	add(womenIntl, domain.Women, "international")
	add(menFranchise, domain.Men, "franchise")
	add(womenFranchise, domain.Women, "franchise")
	for _, t := range archiveTeams { // past matches' sides the table doesn't have
		if _, ok := m[t.ID]; !ok {
			m[t.ID] = t
		}
	}
	return m
}

func teamName(id string) string {
	if t, ok := teamTable[id]; ok {
		return t.Name
	}
	return id
}

func teamShort(id string) string {
	if t, ok := teamTable[id]; ok {
		return t.Short
	}
	return id
}
