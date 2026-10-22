package sim

import (
	"strings"
)

// Review templates keyed by match vibe. Placeholders:
// {W} winner, {L} loser, {A}/{Z} the two teams, {T} top scorer, {R} their runs, {B} top wicket-taker,
// {K} their wickets, {P} official POTM, {F} fans' favourite, {V} venue, {M} margin, {H} headline, {Y} year, {N} classic POTM.
var reviewPool = map[string][]string{
	"thriller": {
		"{T}'s {R} kept {L} in it, but {W} held their nerve. My nails did not survive.",
		"Won by {M}. I aged a decade in the last over. {P}, you absolute menace.",
		"Last-over finishes should come with a medical warning. {W} by {M}, and I'm lying on the floor.",
		"Watched the highlights three times and my heart rate still spikes. {P} was ice.",
		"Couldn't sit, couldn't stand, couldn't breathe. {W} by {M}. Cricket is the best and worst thing in my life.",
		"{F} turned it and {P} gets the award? Fine. Whatever. Still a classic.",
		"Paid for the whole seat, used only the edge. {W} sneak it at {V}.",
		"{L} fans, I know that one hurts. Neutral me? Instant classic, no hesitation.",
		"The kind of game you tell your kids about, badly, with too many hand gestures. {B} with {K} was the quiet hero.",
		"Shouted so loud at the final ball the neighbours knocked. Worth it. {W} by {M}.",
	},
	"thrashing": {
		"{W} won by {M} and honestly it felt bigger. {L} never turned up.",
		"Over by the halfway mark. {T}'s {R} was the only thing worth staying up for.",
		"One-way traffic at {V}. Watched the second half with one eye on Netflix.",
		"{L} brought a stick of celery to a gunfight. {P} was unplayable.",
		"Rating is for {T}'s {R}, not the contest. There was no contest.",
		"Good for {W} fans, terrible for my sleep schedule. {M} tells you everything.",
		"{B} took {K} and the life out of the game. Clinical. Slightly boring. Respect.",
		"If you skipped this one, you were right. {W} by {M}, done and dusted.",
		"{L} should send an apology note to everyone who bought tickets at {V}.",
	},
	"collapse": {
		"{L} went from cruising to chaos in about four overs. {B}'s {K} wickets broke them.",
		"Classic {L} collapse. I've seen this film before and I hate the ending.",
		"Wickets falling like dominoes at {V}. {B} smelt blood and never let go.",
		"From 'we've got this' to all out in the time it took to make chai. {W} by {M}.",
		"{L}'s middle order owes me an explanation and a refund. {B}, take a bow.",
		"The collapse was so fast the replay director couldn't keep up. Ruthless from {W}.",
		"Somewhere a {L} coach is staring at a whiteboard. {B} with {K}, brutal.",
		"I texted 'easy win' and then watched the wickets tumble. Never texting again.",
		"{P} lit the fuse, {L} supplied the gunpowder. {W} cruise home.",
	},
	"bigchase": {
		"Chasing that total felt mad until {T} started swinging. {R} of pure audacity.",
		"{W} chased it down like it was a casual net session. {M} to spare. Unreal.",
		"Big total, bigger chase. {T}'s {R} is going in my personal hall of fame.",
		"Every time the asking rate climbed, {W} smashed it back down. Chase of the season.",
		"Didn't think it was on. My dad didn't think it was on. {T} thought it was on.",
		"{L} posted a huge score and still lost. Modern cricket is unhinged and I love it.",
		"Scoreboard pressure? Never heard of her. {W} by {M} at {V}.",
		"Asking rate north of ten and {T} looked completely unbothered. Big-match player.",
		"I had a nap planned for the innings break. Glad I didn't. That chase was everything.",
	},
	"rain": {
		"Rain at {V}. {T} was just getting going too. Pain.",
		"Watched the covers for longer than I watched cricket. {A} v {Z}, we hardly knew you.",
		"No result. Still rating it because {T} hit a couple of absolute beauties before the rain.",
		"Cricket gods said no today. Spent the evening refreshing the radar app.",
		"Half a match, full disappointment. {A} looked set for a big one too.",
		"The groundstaff were the real MVPs at {V}. Shame the sky didn't care.",
		"Packed the snacks, invited mates, got drizzle. Classic.",
		"Not even mad, just damp. {A} v {Z} deserved a proper finish.",
	},
	"century": {
		"{T}, {R}. That's the review. That's the whole review.",
		"Hundreds like {T}'s {R} are why I pay for this streaming subscription.",
		"The noise at {V} when {T} reached three figures. Goosebumps on goosebumps.",
		"{T} batting like the bowlers had personally offended them. {R} and not a single slog.",
		"Rated it for {T}'s {R} alone. {W} winning by {M} was a bonus.",
		"I'd frame that {T} innings if I could. Cover drives straight out of a textbook.",
		"Century, celebration, chaos. {T} has done it again and {L} had no answer.",
		"Some days the best seat is wherever {T} is batting. Masterclass.",
	},
	"routine": {
		"Solid game, {W} were the better side. {P} quietly ran the show.",
		"Not a classic, but {T}'s {R} had some lovely shots. {W} by {M}.",
		"Decent contest at {V}. {B} with {K} was the difference.",
		"Comfortable for {W} in the end, but {L} made them work for a while.",
		"Good, not great. {P} deserved the award. Ordering biryani and going to bed.",
		"{W} by {M}. The kind of professional win you appreciate more on highlights.",
		"Middle-of-the-road match, top-of-the-range {T}. {R} classy runs.",
		"Fine evening of cricket. My chai was hotter than the contest, but still.",
	},
	"level": {
		"{A} v {Z} finished level and my brain still hasn't processed it.",
		"Nobody won, everybody lost sleep. Cricket.",
		"Level at the end. Sport doesn't get more brutal than that.",
	},
	"classic": {
		"Still get chills from this one. {H}. {Y} was a different time.",
		"Rewatched the highlights for the hundredth time. {W} v {L} at {V}, forever.",
		"Where were you in {Y}? I was glued to a tiny TV and I still remember every ball.",
		"{H}, and somehow it's even better on rewatch.",
		"The match that made me a cricket fan. Not exaggerating. {W} forever.",
		"My dad still talks about this one at every family dinner. {Y}, {V}, legends.",
		"If you're new to cricket, start here. {H}.",
		"{L} fans will never forgive {Y}. Neutral me? Pure cinema.",
		"{N}. That's it, that's the review. {Y} belongs to them.",
		"Rewatched just for {N}. Some performances don't age.",
		"{A} v {Z}, {Y}. Nobody blinked. Iconic.",
	},
}

// demoPool is written in the guest's own voice for their starter history.
var demoPool = map[string][]string{
	"fixture": {
		"Logged this the second it finished. {W} by {M}, and {P} was everything.",
		"My first proper stub on here. {T}'s {R} deserved a bigger screen than my phone.",
		"Called it, watched it, rated it. {W} by {M}. Feeling very smug.",
		"{L} broke my heart a little, but {B} with {K} was pure class.",
		"Still thinking about {V}. {P} for president.",
	},
	"classic": {
		"Logging this years late. {H}. I still remember the noise in that room.",
		"{Y}. I didn't sleep that night. {H}.",
		"The match that started it all for me. {H}.",
	},
}

var replyPool = []string{
	"This is the take.",
	"Couldn't agree more.",
	"Half a ball too generous imo.",
	"Still not over it, thanks for the reminder.",
	"You jinxed it by sitting down, admit it.",
	"That last over took years off my life.",
	"Bit harsh on the bowlers, no?",
	"Were you at the ground? I think I saw you on the big screen.",
	"Same. Exactly the same.",
	"Ha, my whole hostel was screaming at that bit.",
	"Respectfully, you're wrong about the player of the match.",
	"Saving this review for when they lose next week.",
	"Need a rewatch party for this one.",
	"Five balls is the only correct answer.",
	"You've ruined me with this review.",
	"Okay but the fielding though.",
	"Why is this so accurate.",
	"Next time I'm watching it with you.",
	"My dad agrees with you, which is rare.",
	"Underrated review.",
	"Pure poetry. Also pure pain.",
	"The chai detail is so real.",
	"Bold rating. I respect it.",
	"I left at the break and I regret everything.",
	"Can't believe I missed this live.",
}

var placeNotes = map[string][]string{
	"stadium":   {"{V}, Stand C", "{V}, upper tier, sunburnt", "{V}, North Stand, row 14", "{V}, cheap seats behind the bowler's arm"},
	"home":      {"Living room sofa", "Parents' place, Dad on commentary duty", "Bed, phone at 4% battery", "Home, volume way too loud"},
	"friends":   {"Rahul's terrace, projector on a bedsheet", "Mates' flat, 14 people, 1 sofa", "Cousin's place in Pune", "Sam's flat in Clapham"},
	"hostel":    {"Hostel room", "Hostel common room, 40 of us", "PG room, laptop balanced on a bucket"},
	"office":    {"Office pantry TV", "Work laptop, stream in a tiny tab", "Meeting on mute, obviously"},
	"cafe":      {"Sports bar on Church Street", "The local pub, big screen", "Chai stall with a radio"},
	"on-the-go": {"Mumbai local, stream buffering", "Back of an Uber, driver also watching", "Airport gate B12"},
}

var classicPlaceNotes = map[string][]string{
	"stadium":   {"{V}, as a kid with my dad", "{V}, bunked school for it"},
	"home":      {"My uncle's place in Hyderabad", "Grandparents' house, the old CRT", "Home, whole family on one sofa"},
	"friends":   {"Neighbour's house, half the colony there", "Best friend's living room"},
	"hostel":    {"Hostel TV room, 60 of us", "College hostel, one tiny TV"},
	"office":    {"First job, office cafeteria", "Office, pretending to work"},
	"cafe":      {"A dhaba with a radio", "Pub in Leeds, packed"},
	"on-the-go": {"Train compartment, radio commentary", "Bus stand, someone's transistor"},
}

// fill replaces placeholders; it fails if any placeholder has no value.
func fill(tpl string, vals map[string]string) (string, bool) {
	var b strings.Builder
	for {
		i := strings.IndexByte(tpl, '{')
		if i < 0 {
			b.WriteString(tpl)
			break
		}
		j := strings.IndexByte(tpl[i:], '}')
		if j < 0 {
			return "", false
		}
		key := tpl[i+1 : i+j]
		v, ok := vals[key]
		if !ok || v == "" {
			return "", false
		}
		b.WriteString(tpl[:i])
		b.WriteString(v)
		tpl = tpl[i+j+1:]
	}
	return b.String(), true
}

// fillFrom picks a template deterministically, skipping any that can't be fully filled.
func fillFrom(r *rng, pool []string, vals map[string]string) string {
	if len(pool) == 0 {
		return ""
	}
	start := r.intn(len(pool))
	for k := range pool {
		if s, ok := fill(pool[(start+k)%len(pool)], vals); ok {
			return s
		}
	}
	return ""
}
