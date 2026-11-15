import { betrieb } from './hilfebetrieb'
import { suche } from './hilfesuche'

export const hilfe = {
  betrieb,
  suche,
  // One help entry per page — the keys come from `navigation.ts`.
  seiten: {
    start: {
      ueberschrift: 'Start',
      text: 'The way in. The numbers of the store at the top — sets, cards, packages, copies — and below them what you worked on last: opened, imported, exported, and what you keep pinned at the top. Then what needs attention — sets with reservations, decks missing cards, packages with script findings — and what is being fetched in the background. The three first steps appear as a window at the first start; the tours in the help show them again.',
    },
    sets: {
      ueberschrift: 'Sets',
      text: 'Every set this application knows. From here one is opened, a new one created, or an `.mse-set` file read in. The open set decides what its tabs and the “Properties” and “Tools” drawers show.',
    },
    pakete: {
      ueberschrift: 'Packages',
      text: 'Games, stylesheets, symbol fonts and export templates. They are fetched here — at a fixed revision and with a checksum — and installed. Without them a card has no fields and no look.',
    },
    vorlagen: {
      ueberschrift: 'Templates',
      text: 'The text files inside a template package — fields, scripts, measurements. They can be changed here without a separate text editor. A change applies to every set that uses this package; the previous state stays as a backup next to the package.',
    },
    sammlung: {
      ueberschrift: 'Collection',
      text: 'What you actually own — one row per printing, condition, language and finish. It has nothing to do with your sets: a designed card is a draft, a copy is a piece of cardboard. Prices are fetched on request and are estimates nobody stands behind.',
    },
    decks: {
      ueberschrift: 'Decks',
      text: 'Assemblies of your own cards and real printings. A deck references cards instead of copying them, and belongs to a game — without one there would be no format. A format is a record here, not a built-in list: it is created, not shipped. The check asks Scryfall for real printings; a card you designed yourself stays unchecked, and that does not mean illegal.',
    },
    turniere: {
      ueberschrift: 'Tournaments',
      text: 'A tournament is a course over time: rounds, pairings, results. The standings are computed from it and stored nowhere — correct a result and the standings follow at once. Pairing follows the Swiss system; the four tiebreakers and their figures come from the Magic Tournament Rules. Registering a deck submits a list: it is copied, and later changes to the deck no longer reach it.',
    },
    spielen: {
      ueberschrift: 'Play',
      text: "A game is a table with frozen decks: what a deck becomes when the game starts does not change if someone edits that deck later. What happens at the table is a stream of events — each hangs on the previous one, and none is overwritten. What you see is always **one** seat's view: a library is face down even for its owner, and searching it leaves you owing a shuffle. The host holds the full state; the page says so, because that is to be known before joining.",
    },
    handeln: {
      ueberschrift: 'Trading',
      text: 'You trade with a peer whose fingerprint you have compared — no account and no middleman. The matching is computed on **this** machine: nothing leaves it, and what a peer sees of you is your decision under “Settings › Connections”. Booking happens only once the cards have really changed hands; your holdings change then, and the journal records what went where. **Booked here is not booked on both sides** — two holdings on two machines are not one shared booking, and the state says so. What your collection is worth stands beside it as an observation, not as a promise.',
    },
    hilfe: {
      ueberschrift: 'Help',
      text: 'The terms that come from Magic Set Editor and that nobody brings along, one sentence each. Plus the tours, the keyboard shortcuts, why some labels stay English, and what this application sends outside.',
    },
    nutzer: {
      ueberschrift: 'Users and profiles',
      text: 'Who owns what is kept here. A user is the person; a profile is one working environment of that person — its own sets, its own collection, its own settings. Anyone who wants to collect separately for a tournament needs a second profile, not a second account.',
    },
    einstellungen: {
      ueberschrift: 'Settings',
      text: 'Language, appearance, and the places where packages are looked for. What stands here applies to the whole program, not to a single set.',
    },
    karten: {
      ueberschrift: 'Cards',
      text: 'The cards of the open set: the list on the left, the preview in the middle, the fields of the opened card on the right. A change to a field shows at once; nothing needs confirming.',
    },
    setangaben: {
      ueberschrift: 'Set details',
      text: 'What applies to the whole set: name, description, set symbol, author. These are available to every card — a stylesheet can draw the symbol onto each one.',
    },
    stil: {
      ueberschrift: 'Style',
      text: "The set's stylesheet and its options. The sheet says how a card looks; the options adjust it — to a different frame, say, or a different printing.",
    },
    suche: {
      ueberschrift: 'Search',
      text: 'Find cards by their field values, across all sets or within one. The search runs on the full text, and a hit leads straight to the card.',
    },
    ersetzen: {
      ueberschrift: 'Replace',
      text: 'Swap a word in a set\u2019s text fields. It shows what would change first; one \u2318Z reverses the whole move.',
    },
    statistik: {
      ueberschrift: 'Statistics',
      text: 'What is in the open set: how many cards, how often a field is filled, how a field value is distributed. For checking whether a run is complete.',
    },
    bilder: {
      ueberschrift: 'Images',
      text: 'Writes the cards of the set out as image files — all at once, with a naming pattern and a rule for clashing names. **Assigning** a card image happens at the image field of the card, not here.',
    },
    druckbogen: {
      ueberschrift: 'Print sheet',
      text: 'The cards side by side on paper, as a PDF with cut marks. Meant for the printer, not for passing on — export is for that.',
    },
    schluesselwoerter: {
      ueberschrift: 'Keywords',
      text: 'A keyword recognises its text on every card and appends the reminder text. Those from the game are fixed and shown for reference; your own apply to this set only and can be added, edited and deleted.',
    },
    pack: {
      ueberschrift: 'Random pack',
      text: 'A booster from the cards of this set, following the rules of the game package — which pack types exist is stated there. The same seed gives the same draw; to share a result, pass on the number.',
    },
    symbol: {
      ueberschrift: 'Symbol',
      text: "This set's symbol — drawn, not loaded. A basic shape is the start; every point can be moved afterwards. Tab moves between points, the arrow keys move them. What is drawn here appears in the preview at once and goes into the file when the set is saved.",
    },
    konsole: {
      ueberschrift: 'Console',
      text: 'An MSE script expression in the context of the open card — the same environment as when drawing. It changes nothing; it computes and returns. The result comes as source text so a string can be told apart from a name.',
    },
    ausgeben: {
      ueberschrift: 'Export',
      text: 'Cards as an image, or as text through an export template. The template comes from a package and decides what comes out: a list for a forum, a table, a file for another program.',
    },
  },

  // **Über** (#1168).
  ueber: {
    ueberschrift: 'About MagicSetEditor4Mac',
    fassung: 'Version',
    lizenzName: 'Licence',
    lizenzSatz:
      'This application\u2019s source code is licensed under the GNU General Public License, version 2 or later.',
    rechteName: 'Third-party content',
    rechteSatz:
      'The licence covers the source code, **not** card frames, mana symbols, fonts, rules text, names and trademarks. Those belong to their rights holders \u2014 Wizards of the Coast among them \u2014 and not to this project.',
    quelleName: 'Source code',
    quelleSatz: 'github.com/DavidTHeider/MagicSetEditor4Mac',
  },

  weiterlesen: {
    ueberschrift: 'Further reading',
    einleitung:
      "Two documents belong to this application's source — in the project directory under `docs/`, not in the disk image you installed. This application cannot open them itself: it has no way out, and a link that goes nowhere would be worse than none.",
    handbuchName: 'Manual',
    handbuchSatz:
      'What stands here in keywords, in full sentences: the first launch, ' +
      'building a card, exporting, the shortcuts — and what to do when a card ' +
      'stays empty. `docs/HANDBUCH.md`',
    administrationName: 'Data and backups',
    administrationSatz:
      'Where the data lives, how to back it up and restore it, what the log ' +
      'contains and what happens on a version change — including the sentence ' +
      'that a backup **without a password** is plain text. `docs/ADMINISTRATION.md`',
  },

  erststart: {
    ueberschrift: 'First launch',
    text:
      'This application is not notarised. macOS blocks it once; ' +
      'you allow it in System Settings under “Privacy & Security”.',
  },

  // Tours, and the areas that have no page (#1422).
  bereiche: {
    ueberschrift: 'And one area that has no page in the sidebar',
    einleitung:
      'Not everything this application can do lives in the sidebar. What comes in from outside is not a place — and therefore has no page.',
    kiName: 'AI',
    kiSatz:
      'A language model reads and changes cards through the **MCP server** — a separate process on the same store, which you register in your AI tool. What leaves the machine that way and what does not is under **Interfaces**.',
    kiKnopf: 'Go to the interfaces',
  },
  fuehrungen: {
    ueberschrift: 'Tours',
    einleitung:
      'A tour walks you through the application step by step and can be repeated at any time. Next to it, one sentence each on how the application is laid out.',
    startName: 'The first steps',
    startSatz:
      'Three steps to the first card: open a set, edit cards, fetch templates. They appear by themselves at the first start; whether they return at the next start is a setting.',
    startZeigen: 'Show the first steps now',
    arbeitsplatzName: 'The workplace',
    arbeitsplatzSatz:
      'Cards, templates, collection and decks share one layout: navigator on the left, inspector on the right, the centre between them. **⌥⌘[** folds the navigator, **⌥⌘]** the inspector — and what you fold stays folded per profile. **⌥⌘S** folds and unfolds the sidebar so its width goes to the workspaces — that state belongs to this device and is stored per profile.',
    paletteName: 'The command palette',
    paletteSatz:
      '**⌘K** opens it above every page: pages, sets and commands, found with a few letters.',
    leisteName: 'The context bar',
    leisteSatz:
      'Under the application name it says who works where as what — person, profile, role. Switching happens on the page **Users and profiles**; the bar leads there.',
  },

  begriffe: {
    set: {
      name: 'Set',
      satz: 'A collection of cards together with its own data — name, description, symbol. It lives in a file ending in `.mse-set`.',
    },
    karte: {
      name: 'Card',
      satz: 'The thing this is all about. A card carries field values — name, cost, rules text — and is drawn according to a stylesheet.',
    },
    spiel: {
      name: 'Game',
      satz: 'Says **which fields** a card has. An `.mse-game` declares them: `magic` knows casting cost and rarity, another game knows others.',
    },
    stil: {
      name: 'Stylesheet',
      satz: 'Says **how** a card looks: which layer sits where, which image gives the frame. An `.mse-style` always belongs to a game.',
    },
    stiloption: {
      name: 'Styling option',
      satz: 'A stylesheet setting that applies to the whole set or to one card — the frame of a draft card, say, or a different printing.',
    },
    symbolschrift: {
      name: 'Symbol font',
      satz: 'Draws mana symbols and the like. Not a typeface but a collection of small images with rules about which text shows which one.',
    },
    paket: {
      name: 'Package',
      satz: 'The umbrella term for game, stylesheet, symbol font and export template. Packages are fetched and installed; without them a card stays empty.',
    },
    exportvorlage: {
      name: 'Export template',
      satz: 'A small program that turns the cards of a set into **text** — a list for a forum, a table, a file for another program.',
    },
    druckbogen: {
      name: 'Print sheet',
      satz: 'A PDF with the cards side by side on paper, with cut marks. For printing and cutting out — not for passing on as a file.',
    },
    vorbehalt: {
      name: 'Unread line',
      satz: 'A line from an opened file that this program did not interpret. It is not lost: it travels along unchanged and is written back on save.',
    },
    befund: {
      name: 'Finding',
      satz: 'A note about something that did not work out while drawing — a missing layer, a character without a symbol. The card is drawn regardless.',
    },
  },

  sprache: {
    ueberschrift: 'Why some texts stay English',
    einleitung:
      "A card field's caption and the sentence below it come from the **game package**, not from this application. They read the way the package author wrote them — usually English.",
    schluesselName: 'The name is the key',
    schluesselSatz:
      'Your `.mse-set` file stores the value under it. An invented translation would look nicer and make the file unreadable for anyone searching it — so this application invents none.',
    uebersetztName: 'What is translated gets translated',
    uebersetztSatz:
      'Where a package ships a translation, it is used. In the fetched `M15-Magic-Pack` that was 25 card fields on 2026-08-21. For the other 224 it shipped none, and then the word is the same one Magic Set Editor 2 shows.',
    fensterName: 'The window itself speaks eight languages',
    fensterSatz:
      'Everything this application says on its own — buttons, messages, this page — follows the language from the settings. German is the source language; the other seven are machine-translated and not yet editorially reviewed — the language choice says so as well.',
    leisteName: 'The menu bar too',
    leisteSatz:
      'The application menu and the menus for editing and for the window use the words macOS uses for them in this language. **Cut**, **Copy** and **Paste** keep their tie to the operating system.',
  },

  // **What this application knows about you** — Werkstück 13 of Phase 4.12
  // (#1351).
  datenschutz: {
    ueberschrift: 'What this application knows about you',
    einleitung:
      'Since there are users and profiles, this application records **personal data**. Here is what that is, where it lives, what leaves this machine — and what expressly does not.',

    wasName: 'What is recorded',
    wasSatz:
      'All of it is **optional** except the display name. About the person: **form of address**, **name affix**, **first name**, **further first names**, **surname**, **tagline** and two **biographies**, a short and a long one. About the account: **address**, **second address**, **phone number** and, if you set one, a **password**. Plus a **profile picture** and one **badge** per profile.',

    woName: 'Where it lives',
    woSatz:
      'In your library — a SQLite file in your home folder — and in images next to it. **Three things live elsewhere**, namely in the macOS **keychain**: the secret of your second factor, the secret of a sign-in at a provider, and credentials for third-party card services. The library gets shared, the keychain does not — hence the split.',

    formName: 'In what form',
    formSatz:
      'A **password** is never stored in the clear: it is guarded with **Argon2id**, and the guarded form carries its own parameters. An **access token** is stored as a checksum; the plain text exists exactly once, right after it is created. Everything else is stored as you typed it — it is your file. **And the file itself can be encrypted** — under **Settings › Data and backup › Database**. Nothing is then readable on disk; the key goes into this Mac’s keychain and **cannot be recovered**. A **backup** gets its own **password** — it is then encrypted and still opens on any computer; without a password an encrypted database hands it out only when explicitly asked.',

    ausName: 'What leaves this machine',
    ausSatz:
      'On a **sign-in**, your choice goes to the provider you picked: `github.com` and `api.github.com`, `accounts.google.com` and `oauth2.googleapis.com`, `login.microsoftonline.com`, or `twitter.com` and `api.twitter.com`. All that is requested there is your **name and your id** — **not** your e-mail address, not your contacts, not your posts. Three things come back and are stored: the provider, your id there, and your display name there.',

    nichtName: 'What does not go out',
    nichtSatz:
      'Your **name parts, biographies, phone number and address** never leave this machine. They live in your library and are sent nowhere. There is **no telemetry**, no crash reports, no usage counts and no account with us — there is no “us”.',

    dienstName: 'And the card services',
    dienstSatz:
      'Searching for cards asks `api.scryfall.com` and — when enabled — `api.magicthegathering.io`, `graphql.mtgjson.com` or `api.scrydex.com`. Only the **search query** goes out: no user name, no id, nothing from your profile. **The thumbnails of the results** are fetched by the window itself — from `cards.scryfall.io` and, for magicthegathering.io, from `gatherer.wizards.com` and `gatherer-static.wizards.com`; only the image address goes there. Prices for a whole collection come from `mtgjson.com` when enabled, marketplace offers from `api.cardtrader.com`; what goes out is the printing id and, for CardTrader, your access token — nothing else. Template packages come via `codeload.github.com`, release notices via `api.github.com`. What each service can do here and on which terms is under **Settings › Services › Third-party card services**; the links there lead to `scryfall.com`, `docs.magicthegathering.io`, `mtgjson.com`, `scrydex.com` and `www.cardtrader.com` — the application itself never calls those pages, they open in your browser only when you click. **And the market links on a card** (T8): they lead to `www.cardmarket.com`, `www.cardtrader.com`, `www.tcgplayer.com`, `www.cardkingdom.com`, `gatherer.wizards.com` and `www.mtggoldfish.com` — **built, not called**: the address is assembled on this machine from the card name or a provider identifier, and it only goes out when you click, in the browser of your choice. No key is needed, and none is in these addresses.',

    kiName: 'And AI',
    kiSatz:
      'A **local** model runs on this machine and never leaves it: the application talks to `127.0.0.1` and to nothing else. A **cloud model** is a different matter — switch Anthropic, OpenAI or xAI on and `api.anthropic.com`, `api.openai.com` or `api.x.ai` receives the **task and the fields of the card** in question, together with your key from the keychain. **With OpenAI, nothing stays there**: the interface stores a response by default, and this application switches that off explicitly. **The same with xAI**: the endpoint that keeps answers for 30 days is never used. Nothing from your profile, no second card, no set. Both are **off** until you switch them on, and under **Settings › Services › Work locally only** every cloud model stays off even when its own switch says otherwise. What comes back is **shown, not adopted** — a human confirms what goes into a field. What Anthropic can do here and on what terms is listed under **Settings › Services › Third-party card services**; the links there point to `www.anthropic.com`, `platform.claude.com`, `openai.com`, `developers.openai.com`, `x.ai` and `docs.x.ai` — the application never calls those pages itself, they open in your browser on your click.',

    auskunftName: 'Access and deletion',
    auskunftSatz:
      'Under **Users and profiles › My account** you export your data as a file — everything listed here, in one go. And under **Users and profiles** you delete a user — with their profiles, their picture and the secret of their second factor. **What lies inside their profiles stays**: sets, decks, collection and tournaments then belong to no profile and are no longer reachable from the interface. Neither needs a network nor anyone’s consent.',
  },

  schnittstellen: {
    ueberschrift: 'What leaves and what comes in',
    einleitung:
      'This application works offline. **Out onto the network** it goes only when someone triggers it: fetching a template, a card search and a sign-in all need a button press; it looks for new releases only if a setting allows it, and that one is **off** by default. Which addresses are contacted and what goes out is described under **What this application knows about you**. Beside that there are two paths onto this computer that nobody takes by themselves either — and one that does run by itself: if the start fails, it shows the reason in a window.',
    bezugName: 'Fetching templates',
    bezugSatz:
      'At the press of a button the application fetches a template package over HTTP, at a fixed revision and with a checksum. Without the button press nothing happens.',
    scryfallName: 'Card data from Scryfall',
    scryfallSatz:
      'At the press of a button the application asks `api.scryfall.com` for a card, a set or a card image. What comes back is **shown, not adopted** — what goes into a field is a human decision. Without a network the last good state remains, and it says how old it is.',
    fassungName: 'Checking for new versions',
    fassungSatz:
      'The only place that can ask without a button press — and **only** if it is switched on under **Settings**. The default is **off**, and then not a single request goes out. Switched on, it asks `api.github.com` for the latest release. Downloading happens only at the press of a button, and then **signature and checksum are verified before any file is placed**; nothing is installed on its own.',
    mcpName: 'The MCP server',
    mcpSatz:
      'A separate program called `mse-mcp` that lets a language model work on a set: read cards and fields, set fields, create, copy, delete and reorder cards, import a set, save a set and render a card. The measure is not quantity but reversibility: no tool overwrites anything that already exists — where a file is already in place, it refuses — and every change says how to undo it. When a client such as Claude Desktop starts it, it speaks over standard input and output: no port, no address. **Over HTTP** it listens only once you start it under **Settings › Connections**, and only where **Where the services listen** says.',
    mcpNichtAn:
      'It does not run alongside, but it **is** part of this bundle: under `MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp`, together with `mse`, the command line. **By itself this application places nothing outside its own bundle.** Under **Settings › Connections › Tools without a window** you set up, at the press of a button, what you want — the link to `mse` in `~/.local/bin` with a `PATH` entry in `~/.zprofile`, the entry for Codex, Claude Desktop or Claude Code — and you take each entry back there too; anything outside its own block it never touches. What the tools change, this window notices by itself: it asks for the state of the database every two seconds and reloads the page concerned.',
    dienstName: 'The service for scripts',
    dienstSatz:
      'A separate program called `mse-rest` that gives a script the same capabilities as `mse` — read sets and cards, set a field, import or save a set, render a card. Every request needs a token. **Where it listens is your choice** under **Where the services listen**: the default is this Mac only (`127.0.0.1`); if you choose an address that leaves this Mac, it requires TLS and an access token.',
    dienstNichtAn:
      'It is part of this bundle, under `MagicSetEditor4Mac.app/Contents/MacOS/mse-rest`, and does not start by itself: whoever wants no service has none. You start it under **Settings › Connections** — or as a LaunchAgent under **Tools without a window**; then macOS starts it at every login until you take it back there. Whoever starts it is told the address and the token. Requests from a browser are refused before the token is even checked.',
    startfensterName: 'The message when the start fails',
    startfensterSatz:
      'The only path the application takes **by itself**. If the database cannot be opened, it calls macOS\u2019 `osascript` to show the reason in a window — otherwise it would sit on a channel nobody sees after a double-click. The text goes in as a value, not as a script.',
    dateienName: 'Files on this computer',
    dateienSatz:
      'Sets, template packages, written images and PDFs. They sit where you put them; the store sits in your user folder.',
    sonstSatz:
      'Nothing else. **No telemetry**, no crash reports, no usage counts. There is an account — but with the provider you choose, not with us; what goes out in the process is described under **What this application knows about you**.',
    grenzenName: 'Limits',
    grenzenSatz:
      'Every machine access has a **budget**, and every call costs by its kind; when it runs out, the answer says so with the wait. Heavy work — rendering, import, export, bulk operations — has a fixed number of slots per class that window, command line, MCP and service share. Whoever guesses locks only their own credential.',
    grenzenZahlen:
      'Today: {burst} units at once, {nachfuellen} per minute; a read costs {lesen}, a write {schreiben}, an import {einfuhr}; at most {rendern} renderings at a time; {reserve} units stay with the window.',
    grenzenOhneKontingente:
      '**The rate budgets are switched off** (`grenzen.kontingente`) — for a controlled environment. The slots per kind of work, the limits on request and response bodies, the timeouts, the lock against guessing, access and TLS still apply.',
  },
} as const
