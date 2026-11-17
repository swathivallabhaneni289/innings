# The prompts behind Innings

I made Innings by prompting Claude Code. This file collects the prompts that shaped it, in the order I gave them, from the first evening on 22 October 2026 to 17 November, the day after the repository went public. I've edited them for grammar and clarity and left out small talk and status questions, but each one still asks for what I asked for at the time. Under each prompt is what came of it.

The plan I started from is in [PLAN.md](PLAN.md), and the product requirements are in [PRD.md](PRD.md).

## How I worked

I started from a written product plan and set the limits before any code was written: Go as the language, one app for both iPhone and Android, and nothing that costs money. Within those limits I let Claude Code make the technical choices, and I tested the builds myself on the simulator and sent back what didn't work. When a build fell short of the plan, I held it to the plan: real matches and the match-day chat were the point of the app, so they were built rather than cut. As the project grew I added working rules. Claude Code had to ask me before using any outside service, downloading anything or pushing to GitHub, and could use only reputable sources. The docs had to be written in plain language, and the whole history had to be checked for secrets before the repository went public.

## Day one: the brief (22 October)

> Here is the product plan for Innings. Before you build it, keep a few things in mind. Treat this as a one-evening build: move fast and take full charge of the technical decisions, using the GSD workflow if it helps. I'll review the finished app at the end, and everything should be pushed to GitHub.

The plan became [PLAN.md](PLAN.md), with a table of build decisions that filled its gaps, and the work ran as six GSD phases, each with its own plan and summary.

> Use Go as the programming language. The app also has to work on Android phones, not just iPhones.

> If using Go means the app won't look good, tell me. I'm not after a list of approvals, though: I want the finished product. Ask me only when you need a confirmation or something is unclear.

Go runs the API, and the app is built with Expo and React Native in TypeScript, because Go can't draw good native phone screens. One codebase covers both phones.

> Keep the GitHub repository neat and well documented, the way GSD structures a project. Take the time quality needs. The project should look professionally built, not generic or AI-made, so use proper design tools and techniques.

> Before I test it tomorrow, make sure everything is pushed, every test has run and everything works.

Version 1.0 was finished that evening: six phases and 64 requirements, with tests for the API and its scoring rules.

## The look, briefly (22 to 29 October)

> The current look feels generic, as if AI made it. Let's go with the matchday direction.

> The welcome screen should show what the app is about at a glance, with some motion, without feeling busy or cutting off any text.

Three directions were sketched and I picked "Matchday Ticket", where every match is printed as a ticket. The welcome screen went through several rounds before it became one match ticket that rolls through real matches like a scoreboard.

## Accounts (25 October)

> Keep the progress record up to date, and add a requirements document to the repository.

[REQUIREMENTS.md](REQUIREMENTS.md) lists every requirement with an acceptance check and where it lives in the code.

> How do people sign in right now? Instead of that, I want a proper sign-in, like the Beli app has.

Phase 7 replaced the guest accounts with real sign-in and an onboarding flow: a fan card with a unique @username, then finding friends from contacts. It started with a phone number and a texted code, and Phase 10 later moved it to email.

> Write the docs like a person would, in plain and natural language.

## Real matches and the match-day chat (29 October)

> Are the matches up to date? If there's a match today, will the app show it? Having that match appear for predictions is the reason we're building Innings in the first place.

> Research it if you need to, and make a better choice. I don't want to pay anything for this app.

Phase 8 brought real fixtures, live scores, squads, playing XIs and results from ESPN's public scores feed, which needs no account and costs nothing. Only the server reads the feed; the app never calls it directly.

> At its heart, this app is about talking a match through and predicting what a player will do or how the match will go, with a chat room for each game. Without today's live match, it misses the point.

> Clarifying questions are fine, but every feature already in the brief needs to be built.

Phase 9 added a chat room for every real match, and calls during the match worth 5 bragging points each.

> The app has to keep itself current when squads and playing XIs are announced, following a reliable source.

The server looks a week ahead, so squads appear as soon as they're named, and checks every five minutes before the toss for the playing XIs.

> Can fans post their own calls, like whether a batter will make 50, instead of only answering the app's?

Fans can make their own calls from the chat, five per match, and a call is refused once it's as good as decided.

## Sign-in, done properly (29 and 30 October)

> If real fans start posting, where is all of that stored? Keeping everything on my laptop doesn't seem like a good idea. Would a hosted database like Supabase be better?

> Sign-in has to be done properly, with a forgot-password flow, because it's the first thing anyone sees. It mustn't cost anything, but there still needs to be a proper way to sign in.

> Ask my permission before you use any outside service.

This became a standing rule for the rest of the build: downloads, web searches, hosting providers and pushes to GitHub all waited for my yes.

> Keep the data on my Mac for now, and build sign-up properly. Take only the information an account really needs, and make sign-up and sign-in work the way people expect from other apps. Then test all of it.

Phase 10 moved sign-in to an email and a password, confirmed by a 6-digit code sent to that address, with forgot password, change password and delete account. Texted codes were dropped because each one costs money. Phone numbers are no longer collected, passwords are hashed with PBKDF2-SHA256, and friends are matched on hashed email addresses. The data stays in SQLite, with SQL written so it could move to Postgres later.

> Let fans pick more than one team when they sign up.

Fans now back up to five teams, main team first.

## Testing it by hand (8 to 13 November)

> Let's do a hands-on check of the whole app so we know how it really works.

> How does a new user receive their sign-up code by email, and is sending it free?

Codes go out through any mail account, such as Gmail with an app password, at no cost. Until one is set up, the server runs in dev mode and shows the code on screen.

> Everything runs smoothly except "I was there too", which takes up too much room. It should be a button.

It became a compact button.

> I can't find every match I want to log. Search should take what fans actually type, like a year and "ind vs aus".

Search now understands queries like "ind vs aus 2023".

## Past matches (13 November)

> I'd like the earlier matches too, and I'm not sure ESPN is the most accurate source for them. Rather than having fans add matches themselves, let's include the past matches in the app.

> Use a well-known site only as a reference and build our own archive. Don't download anything. It should be up to date and go back much further.

> Use only reputable websites. Nothing should put malware on my devices.

Innings now has its own archive: every World Cup (men's and women's, ODI and T20), Champions Trophy, Asia Cup and IPL season from 1973 to 2026, 2,948 matches. A script builds it from Wikipedia's tournament pages, read as text through Wikipedia's own API and no other site. It checks itself: each tournament's match count against its page, each winning margin against the scores, and each ground against the host countries.

## Rating and the Top 5, everywhere (13 November)

> Ranking works, but rating shouldn't live only in the Vault. Wherever a match's details open, I should be able to rate it or put it in my Top 5. Reviews also has a comment section but nowhere to write my own review.

> Add a way to edit the Top 5. People make mistakes and will want to take games out.

Every finished match's page now has a rating panel that saves with one tap, "Pin to Top 5" and a review box above the reviews. The Top 5 has an editor to reorder it or take games out, and nothing changes until the fan taps Save.

## Going public (16 November)

> What else do we need to check before we call this done, and what would make it ready to demo?

A full end-to-end run passed 88 checks against a fresh server, new tests covered taking a review off and clearing the Top 5, and an unused phone-number package left over from Phase 7 was removed.

> Make the repository public, but first make sure no passwords or other secrets are in it. My email on the commits is fine.

Every commit was scanned for keys, tokens, passwords and personal data before the switch, and the scan came back clean. The file that holds real passwords has never been committed.

> Make the README professional, and add a PRD and a file of the prompts behind the build.

That produced this file, [PRD.md](PRD.md) and the current README.

## Ready to demo (17 November)

> I don't want to use my email password, or any other personal password, anywhere in this project. I just want it ready to demo. What else does it need?

The code that confirms a sign-up keeps showing on screen in dev mode, and the docs now say that sending real codes would need a mail account set up just for Innings.

> Should we move the data to Supabase?

> OK, let's keep SQLite and start the demo setup.

`make demo` starts the API on a fresh database of its own, so my everyday data isn't touched. It opens demo@example.com, an account that already has stubs, five friends and a Top 5, and prints a new password for it each run.

> For the demo, show the matches that are on that day.

The demo runs on that day's real matches from ESPN's feed, and once the week's results have loaded, the demo account gets a few made-up predictions on them, with their points. `INNINGS_MATCHES=sim make demo` runs it on the simulated world instead, for a day with nothing live.

> Will the email and password be the same every time I run the demo?

The email is, but the password is new every run, because this repository is public and a fixed password in it would work for anyone. I kept it that way.
