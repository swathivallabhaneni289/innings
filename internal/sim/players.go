package sim

import (
	"strings"

	"github.com/swathivallabhaneni289/innings/internal/domain"
)

type sq struct{ name, role string }

// squads are playing XIs in batting order, for every team that appears in a fixture.
var squads = map[string][]sq{
	"ind": {{"Abhishek Sharma", "BAT"}, {"Shubman Gill", "BAT"}, {"Suryakumar Yadav", "BAT"}, {"Tilak Varma", "BAT"}, {"Sanju Samson", "WK"}, {"Hardik Pandya", "AR"}, {"Shivam Dube", "AR"}, {"Axar Patel", "AR"}, {"Arshdeep Singh", "BOWL"}, {"Jasprit Bumrah", "BOWL"}, {"Varun Chakaravarthy", "BOWL"}},
	"aus": {{"Travis Head", "BAT"}, {"Mitchell Marsh", "AR"}, {"Josh Inglis", "WK"}, {"Cameron Green", "AR"}, {"Tim David", "BAT"}, {"Glenn Maxwell", "AR"}, {"Marcus Stoinis", "AR"}, {"Xavier Bartlett", "BOWL"}, {"Nathan Ellis", "BOWL"}, {"Adam Zampa", "BOWL"}, {"Josh Hazlewood", "BOWL"}},
	"eng": {{"Zak Crawley", "BAT"}, {"Ben Duckett", "BAT"}, {"Ollie Pope", "BAT"}, {"Joe Root", "BAT"}, {"Harry Brook", "BAT"}, {"Ben Stokes", "AR"}, {"Jamie Smith", "WK"}, {"Gus Atkinson", "BOWL"}, {"Brydon Carse", "BOWL"}, {"Jofra Archer", "BOWL"}, {"Shoaib Bashir", "BOWL"}},
	"sl":  {{"Pathum Nissanka", "BAT"}, {"Oshada Fernando", "BAT"}, {"Dinesh Chandimal", "BAT"}, {"Kamindu Mendis", "AR"}, {"Dhananjaya de Silva", "AR"}, {"Kusal Mendis", "WK"}, {"Ramesh Mendis", "AR"}, {"Prabath Jayasuriya", "BOWL"}, {"Asitha Fernando", "BOWL"}, {"Vishwa Fernando", "BOWL"}, {"Lahiru Kumara", "BOWL"}},

	"ind-w": {{"Smriti Mandhana", "BAT"}, {"Shafali Verma", "BAT"}, {"Jemimah Rodrigues", "BAT"}, {"Harmanpreet Kaur", "BAT"}, {"Richa Ghosh", "WK"}, {"Deepti Sharma", "AR"}, {"Amanjot Kaur", "AR"}, {"Sneh Rana", "AR"}, {"Kranti Goud", "BOWL"}, {"Renuka Singh Thakur", "BOWL"}, {"Shree Charani", "BOWL"}},
	"aus-w": {{"Phoebe Litchfield", "BAT"}, {"Beth Mooney", "WK"}, {"Georgia Voll", "BAT"}, {"Ellyse Perry", "AR"}, {"Annabel Sutherland", "AR"}, {"Ashleigh Gardner", "AR"}, {"Tahlia McGrath", "AR"}, {"Sophie Molineux", "AR"}, {"Kim Garth", "BOWL"}, {"Alana King", "BOWL"}, {"Megan Schutt", "BOWL"}},
	"eng-w": {{"Danni Wyatt-Hodge", "BAT"}, {"Sophia Dunkley", "BAT"}, {"Nat Sciver-Brunt", "AR"}, {"Heather Knight", "BAT"}, {"Amy Jones", "WK"}, {"Alice Capsey", "AR"}, {"Charlie Dean", "BOWL"}, {"Sophie Ecclestone", "BOWL"}, {"Linsey Smith", "BOWL"}, {"Lauren Bell", "BOWL"}, {"Lauren Filer", "BOWL"}},
	"sa-w":  {{"Laura Wolvaardt", "BAT"}, {"Tazmin Brits", "BAT"}, {"Anneke Bosch", "AR"}, {"Marizanne Kapp", "AR"}, {"Sune Luus", "AR"}, {"Chloe Tryon", "AR"}, {"Annerie Dercksen", "AR"}, {"Sinalo Jafta", "WK"}, {"Nadine de Klerk", "AR"}, {"Nonkululeko Mlaba", "BOWL"}, {"Ayabonga Khaka", "BOWL"}},
	"pak-w": {{"Muneeba Ali", "BAT"}, {"Sidra Amin", "BAT"}, {"Omaima Sohail", "AR"}, {"Aliya Riaz", "AR"}, {"Nida Dar", "AR"}, {"Fatima Sana", "AR"}, {"Sidra Nawaz", "WK"}, {"Diana Baig", "BOWL"}, {"Sadia Iqbal", "BOWL"}, {"Nashra Sandhu", "BOWL"}, {"Rameen Shamim", "BOWL"}},
	"nz-w":  {{"Suzie Bates", "BAT"}, {"Georgia Plimmer", "BAT"}, {"Amelia Kerr", "AR"}, {"Sophie Devine", "AR"}, {"Brooke Halliday", "AR"}, {"Maddy Green", "BAT"}, {"Izzy Gaze", "WK"}, {"Jess Kerr", "BOWL"}, {"Lea Tahuhu", "BOWL"}, {"Rosemary Mair", "BOWL"}, {"Eden Carson", "BOWL"}},
	"wi-w":  {{"Hayley Matthews", "AR"}, {"Qiana Joseph", "BAT"}, {"Stafanie Taylor", "AR"}, {"Deandra Dottin", "AR"}, {"Chinelle Henry", "AR"}, {"Shemaine Campbelle", "WK"}, {"Aaliyah Alleyne", "BOWL"}, {"Zaida James", "BOWL"}, {"Afy Fletcher", "BOWL"}, {"Karishma Ramharack", "BOWL"}, {"Shamilia Connell", "BOWL"}},
	"sl-w":  {{"Chamari Athapaththu", "BAT"}, {"Vishmi Gunaratne", "BAT"}, {"Harshitha Samarawickrama", "BAT"}, {"Nilakshika Silva", "BAT"}, {"Kavisha Dilhari", "AR"}, {"Anushka Sanjeewani", "WK"}, {"Hasini Perera", "BAT"}, {"Inoka Ranaweera", "BOWL"}, {"Sugandika Kumari", "BOWL"}, {"Udeshika Prabodhani", "BOWL"}, {"Achini Kulasuriya", "BOWL"}},

	"csk":  {{"Ruturaj Gaikwad", "BAT"}, {"Sanju Samson", "WK"}, {"Ayush Mhatre", "BAT"}, {"Dewald Brevis", "BAT"}, {"Shivam Dube", "AR"}, {"MS Dhoni", "WK"}, {"Jamie Overton", "AR"}, {"Noor Ahmad", "BOWL"}, {"Khaleel Ahmed", "BOWL"}, {"Anshul Kamboj", "BOWL"}, {"Nathan Ellis", "BOWL"}},
	"mi":   {{"Rohit Sharma", "BAT"}, {"Ryan Rickelton", "WK"}, {"Suryakumar Yadav", "BAT"}, {"Tilak Varma", "BAT"}, {"Hardik Pandya", "AR"}, {"Naman Dhir", "AR"}, {"Will Jacks", "AR"}, {"Mitchell Santner", "AR"}, {"Deepak Chahar", "BOWL"}, {"Trent Boult", "BOWL"}, {"Jasprit Bumrah", "BOWL"}},
	"rcb":  {{"Phil Salt", "WK"}, {"Virat Kohli", "BAT"}, {"Devdutt Padikkal", "BAT"}, {"Rajat Patidar", "BAT"}, {"Jitesh Sharma", "WK"}, {"Tim David", "BAT"}, {"Krunal Pandya", "AR"}, {"Romario Shepherd", "AR"}, {"Bhuvneshwar Kumar", "BOWL"}, {"Josh Hazlewood", "BOWL"}, {"Suyash Sharma", "BOWL"}},
	"kkr":  {{"Sunil Narine", "AR"}, {"Finn Allen", "BAT"}, {"Ajinkya Rahane", "BAT"}, {"Cameron Green", "AR"}, {"Rinku Singh", "BAT"}, {"Tim Seifert", "WK"}, {"Ramandeep Singh", "AR"}, {"Harshit Rana", "BOWL"}, {"Varun Chakaravarthy", "BOWL"}, {"Vaibhav Arora", "BOWL"}, {"Matheesha Pathirana", "BOWL"}},
	"srh":  {{"Travis Head", "BAT"}, {"Abhishek Sharma", "BAT"}, {"Ishan Kishan", "WK"}, {"Heinrich Klaasen", "WK"}, {"Nitish Kumar Reddy", "AR"}, {"Liam Livingstone", "AR"}, {"Aniket Verma", "BAT"}, {"Pat Cummins", "BOWL"}, {"Harshal Patel", "BOWL"}, {"Jaydev Unadkat", "BOWL"}, {"Eshan Malinga", "BOWL"}},
	"rr":   {{"Yashasvi Jaiswal", "BAT"}, {"Vaibhav Suryavanshi", "BAT"}, {"Riyan Parag", "AR"}, {"Dhruv Jurel", "WK"}, {"Shimron Hetmyer", "BAT"}, {"Ravindra Jadeja", "AR"}, {"Sam Curran", "AR"}, {"Jofra Archer", "BOWL"}, {"Sandeep Sharma", "BOWL"}, {"Tushar Deshpande", "BOWL"}, {"Ravi Bishnoi", "BOWL"}},
	"dc":   {{"KL Rahul", "WK"}, {"Abishek Porel", "BAT"}, {"Sameer Rizvi", "BAT"}, {"Tristan Stubbs", "BAT"}, {"Axar Patel", "AR"}, {"Ashutosh Sharma", "BAT"}, {"Vipraj Nigam", "AR"}, {"Mitchell Starc", "BOWL"}, {"Kuldeep Yadav", "BOWL"}, {"T Natarajan", "BOWL"}, {"Mukesh Kumar", "BOWL"}},
	"pbks": {{"Prabhsimran Singh", "WK"}, {"Priyansh Arya", "BAT"}, {"Shreyas Iyer", "BAT"}, {"Nehal Wadhera", "BAT"}, {"Shashank Singh", "BAT"}, {"Marcus Stoinis", "AR"}, {"Marco Jansen", "AR"}, {"Azmatullah Omarzai", "AR"}, {"Harpreet Brar", "BOWL"}, {"Arshdeep Singh", "BOWL"}, {"Yuzvendra Chahal", "BOWL"}},

	"mi-w":  {{"Hayley Matthews", "AR"}, {"G Kamalini", "WK"}, {"Nat Sciver-Brunt", "AR"}, {"Harmanpreet Kaur", "BAT"}, {"Amelia Kerr", "AR"}, {"Amanjot Kaur", "AR"}, {"Sajeevan Sajana", "AR"}, {"Shabnim Ismail", "BOWL"}, {"Saika Ishaque", "BOWL"}, {"Jintimani Kalita", "BOWL"}, {"Parunika Sisodia", "BOWL"}},
	"rcb-w": {{"Smriti Mandhana", "BAT"}, {"Grace Harris", "BAT"}, {"Ellyse Perry", "AR"}, {"Georgia Voll", "BAT"}, {"Richa Ghosh", "WK"}, {"Nadine de Klerk", "AR"}, {"Radha Yadav", "BOWL"}, {"Shreyanka Patil", "BOWL"}, {"Arundhati Reddy", "BOWL"}, {"Pooja Vastrakar", "AR"}, {"Lauren Bell", "BOWL"}},
	"dc-w":  {{"Shafali Verma", "BAT"}, {"Laura Wolvaardt", "BAT"}, {"Jemimah Rodrigues", "BAT"}, {"Lizelle Lee", "WK"}, {"Marizanne Kapp", "AR"}, {"Annabel Sutherland", "AR"}, {"Niki Prasad", "BAT"}, {"Sneh Rana", "AR"}, {"Shree Charani", "BOWL"}, {"Minnu Mani", "BOWL"}, {"Titas Sadhu", "BOWL"}},
	"gg-w":  {{"Beth Mooney", "WK"}, {"Sophie Devine", "AR"}, {"Danni Wyatt-Hodge", "BAT"}, {"Ashleigh Gardner", "AR"}, {"Bharti Fulmali", "BAT"}, {"Kashvee Gautam", "AR"}, {"Georgia Wareham", "AR"}, {"Tanuja Kanwar", "BOWL"}, {"Renuka Singh Thakur", "BOWL"}, {"Rajeshwari Gayakwad", "BOWL"}, {"Shabnam Shakil", "BOWL"}},

	"sixers":    {{"Josh Philippe", "WK"}, {"Babar Azam", "BAT"}, {"Steve Smith", "BAT"}, {"Moises Henriques", "AR"}, {"Jordan Silk", "BAT"}, {"Jack Edwards", "AR"}, {"Hayden Kerr", "AR"}, {"Sean Abbott", "BOWL"}, {"Ben Dwarshuis", "BOWL"}, {"Ben Manenti", "BOWL"}, {"Todd Murphy", "BOWL"}},
	"scorchers": {{"Finn Allen", "BAT"}, {"Cooper Connolly", "AR"}, {"Josh Inglis", "WK"}, {"Ashton Turner", "BAT"}, {"Aaron Hardie", "AR"}, {"Laurie Evans", "BAT"}, {"Ashton Agar", "AR"}, {"Andrew Tye", "BOWL"}, {"Jhye Richardson", "BOWL"}, {"Jason Behrendorff", "BOWL"}, {"Lance Morris", "BOWL"}},
	"thunder":   {{"David Warner", "BAT"}, {"Sam Konstas", "BAT"}, {"Ollie Davies", "BAT"}, {"Sam Billings", "WK"}, {"Chris Green", "AR"}, {"Daniel Sams", "AR"}, {"Shadab Khan", "AR"}, {"Nathan McAndrew", "BOWL"}, {"Tanveer Sangha", "BOWL"}, {"Lockie Ferguson", "BOWL"}, {"Wes Agar", "BOWL"}},
	"strikers":  {{"Matt Short", "AR"}, {"Chris Lynn", "BAT"}, {"Jason Sangha", "BAT"}, {"Alex Ross", "BAT"}, {"Harry Nielsen", "WK"}, {"Thomas Kelly", "AR"}, {"Liam Scott", "AR"}, {"Cameron Boyce", "BOWL"}, {"Henry Thornton", "BOWL"}, {"Lloyd Pope", "BOWL"}, {"Jordan Buckingham", "BOWL"}},

	"mict": {{"Ryan Rickelton", "WK"}, {"Reeza Hendricks", "BAT"}, {"Rassie van der Dussen", "BAT"}, {"Nicholas Pooran", "BAT"}, {"George Linde", "AR"}, {"Corbin Bosch", "AR"}, {"Delano Potgieter", "AR"}, {"Rashid Khan", "BOWL"}, {"Kagiso Rabada", "BOWL"}, {"Trent Boult", "BOWL"}, {"Kwena Maphaka", "BOWL"}},
	"jsk":  {{"Faf du Plessis", "BAT"}, {"Devon Conway", "BAT"}, {"James Vince", "BAT"}, {"Leus du Plooy", "BAT"}, {"Donovan Ferreira", "WK"}, {"Evan Jones", "AR"}, {"Gerald Coetzee", "BOWL"}, {"Lizaad Williams", "BOWL"}, {"Hardus Viljoen", "BOWL"}, {"Maheesh Theekshana", "BOWL"}, {"Imran Tahir", "BOWL"}},
	"pc":   {{"Will Jacks", "AR"}, {"Dewald Brevis", "BAT"}, {"Rilee Rossouw", "BAT"}, {"Sherfane Rutherford", "BAT"}, {"Kyle Verreynne", "WK"}, {"Senuran Muthusamy", "AR"}, {"Wayne Parnell", "AR"}, {"Keshav Maharaj", "BOWL"}, {"Anrich Nortje", "BOWL"}, {"Lungi Ngidi", "BOWL"}, {"Eathan Bosch", "BOWL"}},
	"sec":  {{"Quinton de Kock", "WK"}, {"Aiden Markram", "BAT"}, {"Jordan Hermann", "BAT"}, {"Tristan Stubbs", "BAT"}, {"Jonny Bairstow", "WK"}, {"Marco Jansen", "AR"}, {"Roelof van der Merwe", "AR"}, {"Liam Dawson", "AR"}, {"Ottniel Baartman", "BOWL"}, {"Richard Gleeson", "BOWL"}, {"Simon Harmer", "BOWL"}},

	"rockets": {{"Tom Banton", "WK"}, {"Alex Hales", "BAT"}, {"Joe Root", "BAT"}, {"Marcus Stoinis", "AR"}, {"Tom Moores", "BAT"}, {"Imad Wasim", "AR"}, {"Lewis Gregory", "AR"}, {"Rehan Ahmed", "AR"}, {"Sam Cook", "BOWL"}, {"John Turner", "BOWL"}, {"Calvin Harrison", "BOWL"}},
	"spirit":  {{"Jamie Smith", "WK"}, {"Kane Williamson", "BAT"}, {"Ollie Pope", "BAT"}, {"Kieron Pollard", "BAT"}, {"Liam Dawson", "AR"}, {"Andre Russell", "AR"}, {"Daniel Worrall", "BOWL"}, {"Olly Stone", "BOWL"}, {"Mohammad Amir", "BOWL"}, {"Jafer Chohan", "BOWL"}, {"Luke Wood", "BOWL"}},
}

// nicknames are how fans refer to a player in a review (default: last name).
var nicknames = map[string]string{
	"Suryakumar Yadav": "SKY", "Kuldeep Yadav": "Kuldeep", "Radha Yadav": "Radha", "Hardik Pandya": "Hardik",
	"Krunal Pandya": "Krunal", "Abhishek Sharma": "Abhishek", "Rohit Sharma": "Rohit", "Jitesh Sharma": "Jitesh",
	"Suyash Sharma": "Suyash", "Ashutosh Sharma": "Ashutosh", "Sandeep Sharma": "Sandeep", "Deepti Sharma": "Deepti",
	"Shafali Verma": "Shafali", "Aniket Verma": "Aniket", "Smriti Mandhana": "Smriti", "Harmanpreet Kaur": "Harman",
	"Amanjot Kaur": "Amanjot", "Arshdeep Singh": "Arshdeep", "Rinku Singh": "Rinku", "Ramandeep Singh": "Ramandeep",
	"Shashank Singh": "Shashank", "Prabhsimran Singh": "Prabhsimran", "Renuka Singh Thakur": "Renuka",
	"Tilak Varma": "Tilak", "Varun Chakaravarthy": "Varun", "Axar Patel": "Axar", "Harshal Patel": "Harshal",
	"Harshit Rana": "Harshit", "Nitish Kumar Reddy": "Nitish", "Ishan Kishan": "Ishan", "Richa Ghosh": "Richa",
	"Jemimah Rodrigues": "Jemimah", "Sneh Rana": "Sneh", "Kranti Goud": "Kranti", "Shree Charani": "Charani",
	"Amelia Kerr": "Melie Kerr", "Jess Kerr": "Jess Kerr", "Hayden Kerr": "Kerr", "Kamindu Mendis": "Kamindu",
	"Kusal Mendis": "Kusal", "Ramesh Mendis": "Ramesh", "Oshada Fernando": "Oshada", "Asitha Fernando": "Asitha",
	"Vishwa Fernando": "Vishwa", "Jamie Smith": "Jamie Smith", "Linsey Smith": "Linsey Smith", "MS Dhoni": "Dhoni",
	"KL Rahul": "KL", "Virat Kohli": "Kohli", "Shubman Gill": "Gill", "Sanju Samson": "Samson",
	"Chamari Athapaththu": "Chamari", "Tom Banton": "Banton", "Tom Moores": "Moores", "Chris Green": "Chris Green",
	"Cameron Green": "Green", "Maddy Green": "Maddy Green", "Sam Billings": "Billings", "G Kamalini": "Kamalini",
	"Sajeevan Sajana": "Sajana", "T Natarajan": "Natarajan", "Babar Azam": "Babar", "Rashid Khan": "Rashid",
	"Noor Ahmad": "Noor", "Khaleel Ahmed": "Khaleel", "Rehan Ahmed": "Rehan", "Sidra Amin": "Sidra Amin",
	"Sidra Nawaz": "Sidra Nawaz", "Eathan Bosch": "Eathan Bosch", "Corbin Bosch": "Corbin Bosch",
}

func slugify(s string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		default:
			if !dash && b.Len() > 0 {
				b.WriteByte('-')
				dash = true
			}
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

func playerID(teamID, name string) string { return teamID + "-" + slugify(name) }

func nick(name string) string {
	if n, ok := nicknames[name]; ok {
		return n
	}
	parts := strings.Fields(name)
	for i, p := range parts[:len(parts)-1] {
		if i > 0 && (p == "de" || p == "du" || p == "van") {
			return strings.Join(parts[i:], " ")
		}
	}
	return parts[len(parts)-1]
}

// playerTable is every squad player keyed by id.
var playerTable = buildPlayers()

func buildPlayers() map[string]domain.Player {
	m := map[string]domain.Player{}
	for team, xi := range squads {
		for _, p := range xi {
			id := playerID(team, p.name)
			m[id] = domain.Player{ID: id, Name: p.name, TeamID: team, Role: p.role}
		}
	}
	return m
}

// xiIDs returns a team's XI ids in batting order (always non-nil).
func xiIDs(team string) []string {
	out := []string{}
	for _, p := range squads[team] {
		out = append(out, playerID(team, p.name))
	}
	return out
}

func playerNick(id string) string {
	if p, ok := playerTable[id]; ok {
		return nick(p.Name)
	}
	return ""
}

func playerRole(id string) string { return playerTable[id].Role }
