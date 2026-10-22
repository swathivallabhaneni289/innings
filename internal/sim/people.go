package sim

import "github.com/swathivallabhaneni289/innings/internal/domain"

type personRow struct{ id, name, handle, colour, team, bio string }

var friendRows = []personRow{
	{"f-aditi", "Aditi Rao", "aditi", "#1C2333", "rcb", "Ee sala cup namde. Every year. No notes."},
	{"f-kabir", "Kabir Mehta", "kabir", "#5A3E36", "csk", "Dhoni finishes, I cry"},
	{"f-meera", "Meera Iyer", "meera", "#6B4E71", "ind-w", "Smriti's cover drive is my love language"},
	{"f-arjun", "Arjun Nair", "arjun", "#2F4F4F", "mi", "Bumrah is a cheat code and I won't hear otherwise"},
	{"f-tom", "Tom Hughes", "tomh", "#3E5C76", "eng", "Bazball apologist. Will defend a four-day defeat."},
	{"f-chloe", "Chloe Martin", "chloe", "#7A3B2E", "aus-w", "Ellyse Perry stan since dial-up internet"},
	{"f-sipho", "Sipho Dlamini", "sipho", "#4A5D3A", "sa", "Proteas fan, so heartbreak is a lifestyle"},
	{"f-nimal", "Nimal Perera", "nimal", "#594A3C", "sl", "Galle sunsets and Kamindu cover drives"},
}

var publicRows = []personRow{
	{"p-rohan", "Rohan Desai", "rohan.d", "#2E4057", "mi", "Wankhede North Stand, row 14, since 2008"},
	{"p-sana", "Sana Qureshi", "sanaq", "#4B3F72", "pak-w", "Fatima Sana's yorkers live rent-free in my head"},
	{"p-liam", "Liam O'Connor", "liamoc", "#3D5A40", "ire", "Irish cricket believer. Bring tea and patience."},
	{"p-priya", "Priya Menon", "priyam", "#6E3B3B", "kkr", "Korbo Lorbo Jeetbo, then overthink every over"},
	{"p-jack", "Jack Thompson", "jackt", "#35505E", "sixers", "SCG Members, cold pie, louder opinions"},
	{"p-ananya", "Ananya Bose", "ananya", "#5C4033", "ind-w", "Women's cricket evangelist, scorebook in pencil"},
	{"p-dev", "Dev Malhotra", "devm", "#3B4A6B", "dc", "Delhi boy. Kotla heat is a character-building exercise."},
	{"p-hannah", "Hannah Clarke", "hclarke", "#5E4B56", "eng-w", "Ecclestone appreciation society, founding member"},
	{"p-farhan", "Farhan Siddiqui", "farhans", "#2F3E46", "srh", "Orange Army. 250 is the new 180."},
	{"p-kavya", "Kavya Reddy", "kavya.r", "#614A3A", "srh", "Uppal floodlights and too much biryani"},
	{"p-ben", "Ben Walsh", "benw", "#384B5E", "spirit", "Lord's on a Tuesday is my idea of therapy"},
	{"p-thandi", "Thandi Mokoena", "thandi", "#45503B", "jsk", "Wanderers bullring regular. Yellow till I die."},
	{"p-isla", "Isla Brown", "islab", "#4F3B4F", "nz-w", "Melie Kerr leg-spinners are art, actually"},
	{"p-vikram", "Vikram Joshi", "vikj", "#3F4C5A", "rr", "Pink city, pink jersey, blood pressure through the roof"},
	{"p-maya", "Maya Pillai", "mayap", "#5A4A42", "ind", "Writes a match diary in a notebook older than Gill"},
	{"p-ahmed", "Ahmed Raza", "ahmedr", "#3E4A3D", "pak", "Pakistan fan: expect nothing, feel everything"},
}

func toUsers(rows []personRow) []domain.User {
	out := []domain.User{}
	for _, r := range rows {
		out = append(out, domain.User{ID: r.id, Name: r.name, Handle: r.handle, AvatarColour: r.colour, FavouriteTeamID: r.team, Bio: r.bio})
	}
	return out
}

var friendIDs = func() []string {
	out := []string{}
	for _, r := range friendRows {
		out = append(out, r.id)
	}
	return out
}()

var personByID = func() map[string]personRow {
	m := map[string]personRow{}
	for _, r := range append(append([]personRow{}, friendRows...), publicRows...) {
		m[r.id] = r
	}
	return m
}()

func isFriend(id string) bool { return len(id) > 2 && id[:2] == "f-" }
