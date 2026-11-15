export const fehler = {
  unbekannt:
    'An unexpected error occurred. What exactly is in the log — under “Settings › Diagnostics” you can see its tail and save it.',
  unbekanntMitVorgang:
    'An unexpected error occurred. What exactly is in the log — under “Settings › Diagnostics” you can see its tail and save it. Look there for {vorgang}: every line of this operation carries that identifier.',
  nichtGefunden:
    'The requested entry was not found. Perhaps somebody else deleted it — reload the list.',
  ohnePeeridentitaet:
    'This profile does not have an identity yet. Create one in the “How this profile identifies itself” box — without it, nobody can prove who they are.',
  peerNichtErreichbar:
    'Your peer cannot be reached right now. Is it listening, and is its address right? Otherwise try again later.',
  peerOhneAdresse:
    'No address is set for this peer. Enter it on its contact under “Address and note” — without it, this program can only find it on the same network.',
  ohneNetzerlaubnis:
    'Nothing got out onto the local network. If macOS just asked for permission, allow it and try again. Otherwise the reason can’t be told for sure from here: either the permission is missing — System Settings → Privacy & Security → Local Network — or the network doesn’t let multicast through, for example behind a VPN.',
  ohneNetzroute:
    'There is no route to the local network right now. Is a network interface connected? You can pick a specific one under Settings → Connections.',
  schluesselbundAblage:
    'The key of this profile could not be stored in the keychain. Is the login keychain unlocked? Then try again.',
  einladungZuNeu:
    'This invitation comes from a newer version of this program. Ask the other side for an invitation from their older version — or get the newer one yourself.',
  fernhorchenLaeuft:
    'This profile is already reachable across networks. Stop it first if you want a new invitation — the old one becomes void.',
  anmeldungAbgelaufen:
    'The sign-in did not come back within ten minutes. Start it again once you are ready in the browser.',
  umschlagFassung:
    'This package comes from a different version of this program. Ask the other side for one from their older version — or get the newer one yourself.',
  umschlagUnterschrift:
    'The signature on this package does not hold. It was changed in transit or comes from somebody other than it claims — do not accept it and check with the other side.',
  umschlagInhalt:
    'A piece of this package does not match its manifest. Please ask the other side to send it again — something was damaged or changed in transit.',
  fremderNutzer:
    'Only this user can do that. Switch to this user first — the app will ask for their second factor.',
  nichtFreigegeben:
    'Your counterpart shares nothing with you for trading. Please ask them to share a view with you — or trade via a file.',
  zeileNichtVerbuchbar:
    'Line {zeile} could not be booked on the receiving side — the condition or language is missing, or a detail does not exist. Remove it and add it again with condition, language and finish.',
  zusageUnpassend:
    'This copy fulfils none of the lines you give — condition, language or finish differ. Choose a matching one or add one to your collection.',
  erlaubnis:
    'You do not have the role for this. Whoever manages this profile can give it to you.',
  unbekannteHerkunftsart:
    'The program does not know this origin. Pick “taken over from a profile” or “imported”.',
  // The command centre (#1423).
  befehlBestaetigen:
    'This command changes something that cannot easily be undone — read the preview, then confirm.',
  befehlArgumente:
    'The arguments are not a JSON object. Enter an object in curly braces — one field per parameter.',
  herkunftOhneProfil:
    'An origin from a profile names it; an import names none — give one or leave it out.',
  nurLesend:
    'You can only read in this profile. It belongs to someone else and your role in it is \u201creader\u201d. Switch to your own profile to work \u2014 or ask whoever manages this one for the \u201ccollaborator\u201d role.',
  bestandUnbrauchbar:
    'This file is damaged or not a MagicSetEditor4Mac library — it cannot be opened. Restore a backup; the last one is wherever you saved it.',
  bestandNeuereFassung:
    'This library comes from a newer version of MagicSetEditor4Mac. The file itself is fine — do not replace it. Get the newer version, or restore an earlier backup.',
  bestandAeltereFassung:
    'This library comes from a version before 0.7.0, and this version can no longer upgrade it. The file itself is fine — do not replace it. Open it once with MagicSetEditor4Mac 0.7.0; after that, this version opens it too.',
  bestandNurLesbar:
    'The library can be read but not written — the file or its folder lacks write permission. The file is fine — do not replace it. Give it write permission in the Finder under “Get Info” and try again.',
  bestandEA:
    'The disk reported an error while reading the library. Whether the file is intact is unknown — do not replace it right away. Check the disk, for example with Disk Utility, and try again.',
  kaumPlatz:
    'The library could not be opened or written, and there is hardly any space left on the disk. The file is probably fine — do not replace it. Free up space and try again; if that does not help, check the disk, for example with Disk Utility, and the write permission.',
  startUngeklaert:
    'The library could not be opened, and the reason is unclear — it is shown below. Try again. Only restore a backup once you are sure this file can no longer be opened.',
  sicherungVerworfen:
    'The backup failed its check and was discarded — nothing has been backed up. Try again, or back up to a different location.',
  sicherungUnvermerkt:
    'The backup is written and checked; only its time could not be recorded. Keep the file and try again once the library is free — until then the age shows the previous backup.',
  schluesselFehlt:
    'This library is encrypted and its key is not in this Mac\u2019s keychain. The file itself is fine \u2014 do not replace it. The key was created when encryption was switched on and cannot be recomputed; without it the data is unreachable.',
  belegt:
    'The library is busy right now — another window or command is probably writing. Please try again in a moment.',
  bestandWirdGetauscht:
    'The library is being replaced right now — by restoring a backup, switching encryption or starting over. Wait a moment, then try again.',
  anwendungEndet:
    'MagicSetEditor4Mac is quitting — this step was not carried out. Start the app again to continue.',
  vorgangAbgebrochen:
    'The operation was cancelled before it finished. Start it again if you still need it.',
  konflikt: 'The entry changed in the meantime. Please reload.',
  konfliktgrund:
    'That did not work: {grund}. Check the current state before you try again.',
  doppelt:
    'An entry with these details already exists. Change one of them — or edit the existing entry.',
  doppelterName: 'This name is already taken here — choose another one.',
  doppeltesExemplar:
    'The collection already holds this printing — in the same condition, language and finish. Raise the quantity there or change one of the details.',
  gesperrt:
    'This app may not access this location — it is locked. Choose another one or allow access, in the file’s permissions or under System Settings › Privacy & Security › Files and Folders, and try again.',
  voll: 'There is no space left on the disk. Free up some space and try again.',
  netzName:
    'The name “{wirt}” could not be resolved. Check your internet connection and, if one is set up, the proxy under “Settings › Connections”. What the connection reported is under “Settings › Diagnostics” at {vorgang}.',
  netzNichtErreichbar:
    'Nothing accepts connections at “{wirt}”. Is the service running, and is its address right? Otherwise try again later. What the connection reported is under “Settings › Diagnostics” at {vorgang}.',
  netzKeineAntwort:
    '“{wirt}” did not answer within {sekunden} seconds. Try again later. What the connection reported is under “Settings › Diagnostics” at {vorgang}.',
  netzAntwortStockt:
    'The answer from “{wirt}” stalled for {sekunden} seconds. Check the connection and try again. What the connection reported is under “Settings › Diagnostics” at {vorgang}.',
  netzAbgerissen:
    'The connection to “{wirt}” broke off before the answer was complete. Try again; if it keeps happening, the service or a proxy in between is at fault. What the connection reported is under “Settings › Diagnostics” at {vorgang}.',
  netzProxyAnmeldung:
    'The proxy “{wirt}” requires a sign-in. Set its credentials through HTTPS_PROXY in the environment — under “Settings › Connections” the proxy deliberately takes no password. What the connection reported is under “Settings › Diagnostics” at {vorgang}.',
  gedrosselt:
    'The other side reports too many requests. Wait a moment, then try again. What the connection reported is under “Settings › Diagnostics” at {vorgang}.',
  gedrosseltBis:
    '“{wirt}” is not accepting requests right now. Try again from {uhrzeit} — until then this program will not ask again. What the connection reported is under “Settings › Diagnostics” at {vorgang}.',
  kontingent:
    'Too many calls in a short time. Wait {sekunden} s and then try again.',
  prozessVeraltet:
    'This process comes from an older version than the library and is shutting down — nothing was written. An installed service restarts on its own, with the program it is installed for; restart any other process with the new version, then try again.',
  bestandVomDienstGehalten:
    'The scripting service is holding the library open. Revoke it under Settings › Diagnostics › “Set up the command line and tool server”, then restore again.',
  falschesSpiel:
    'The deck belongs to a different game than this tournament. Pick a deck from the same game.',
  turnierformatSpiel:
    'The format belongs to a different game than this tournament. Pick a format from the same game — or none.',
  paketGeaendert:
    'This package differs from how it was installed. Confirm the replacement — the change is kept as the previous version.',
  symbolZuVieleTeile:
    'This symbol has {anzahl} parts, {grenze} are allowed. Merge parts or choose a simpler drawing.',
  symbolZuVielePunkte:
    'This symbol has {anzahl} points, {grenze} are allowed. Simplify the shapes or choose a simpler drawing.',
  symbolZuTief:
    'The groups of this symbol are nested {anzahl} levels deep, {grenze} are allowed. Choose a flatter drawing or ungroup inner groups.',
  symbolZuGross:
    'This file is larger than {grenze} and therefore not a symbol file. Choose an .mse-symbol file.',
  symbolGespeichertZuGross:
    'Saved, this symbol would be larger than {grenze}, more than a symbol file can hold. Simplify the shapes or remove some points.',
  bildZuGross:
    'This image is too large to unpack. Choose a smaller one or scale it down first.',
  setOhneSymbol:
    'This set has no symbol yet. Draw one and save it — then it can be added.',
  // **Your own translations** — LOC-02 (#1688).
  sprachkennungUngueltig:
    '“{kennung}” is not a BCP 47 language tag. Pick one like pt-BR or zh-Hant.',
  sprachkennungEingebaut:
    '“{kennung}” is a built-in language. Choose a tag of your own such as {kennung}-x-mine.',
  platzhalterAbweichend:
    'This translation names other placeholders than the German sentence. Enter exactly these: {erwartet}.',
  auszeichnungUnsicher:
    'This translation carries something that is not text, or markup that does not close. Remove tags and close ** and `.',
  sprachpaketNeuer:
    'This translation comes from a newer program ({format}). Get the new version.',
  sprachpaketOhneLizenz:
    'No translation goes out without a licence. Enter one, for example CC0-1.0.',
  bereichAbgeschaltet:
    'This area is switched off ({schalter}). Switch it back on under Settings › Areas and functions — nothing has been lost.',
  dienstAus:
    'The service {dienst} is switched off. Switch it on under Settings › Services if you want to ask it.',
  fremdeEigeneKarten:
    'The sender’s own cards did not come along — their identifier is only valid in the sender’s database: {karten}. If you need them, open their set and add them here.',
  nichtGelegtePlaetze:
    'These entries could not be added here: {plaetze}. The rest of the deck was accepted; check them and add them by hand.',
  fremdeKartenUndNichtGelegte:
    'The sender’s own cards did not come along — their identifier is only valid in the sender’s database: {karten}. And these entries could not be added here: {plaetze}. Check both and add them by hand.',
  symbolVerwendet:
    'These sets carry the symbol: {sets}. Confirm the deletion — they keep their copy.',
  symbolOhneVorige:
    'This symbol has no previous version. Save a new one first — then the current one can be restored.',
  turnierAbgeschlossen:
    'This tournament is closed. Choose “Reopen” to change anything.',
  sitzGegangen:
    'This seat has left the game and can no longer act. Choose a seat that is still in play.',
  fremderWert:
    'Someone else changed this field in the meantime. Check the new value and decide whether to keep your draft.',
  vorschlagNurGegenseite:
    'Only the other side can change an open offer. Wait for their answer, or cancel the trade and offer again.',
  turnierNichtFertig:
    'Enter every result first: a tournament can be closed once it has at least one round and every pairing has a result.',
  turniernameLeer: 'A tournament needs a name — enter one.',
  turnierspielLeer:
    'A tournament belongs to a game — otherwise there is no format to check against. Pick one.',
  validierungsgrund:
    'That did not work: {grund}. Correct the input and try again.',

  validierung:
    'The input is not valid. The hint at the field says why — correct it there.',

  // **The rule itself is the remedy** (#1350).
  // **The password of a backup** (OPS-06, #1723).
  sicherungskennwortKurz:
    'A backup password has at least twelve characters. Choose a longer one — a sentence of four words is good and easy to remember.',
  sicherungskennwortFehlt:
    'This backup is encrypted. Enter its password to view or restore it.',
  sicherungskennwortFalsch:
    'The password does not fit this backup. Check it and enter it again.',
  // **Eine MTG-JSON-Datei, die sich nicht lesen lässt** (G01, #2248).
  mtgjsonKeinJSON:
    '{datei} is not readable JSON — at byte {stelle} the file ends or is malformed. Choose another file.',
  mtgjsonKeinSet:
    '{datei} is JSON but contains no set in MTG JSON format (it differs at byte {stelle}). Choose a file with a set — or a whole folder.',
  sicherungImKlartext:
    'The collection is encrypted — without a password the backup would be plain text. Set a password under Settings › Data and backup or explicitly choose “back up without password” there.',

  planVerschluesselt:
    'The collection is encrypted, and the plan backs up without a password — so it is not backed up by the plan. Back it up by hand with a password.',

  paketnameUngueltig:
    'A package name is made of letters, digits, hyphens and underscores — it becomes the directory name. Pick one from those characters.',
  ordnerangabe:
    '“{wert}” is not a full location. Enter one that starts with / or ~/, without .. in between.',
  ordnerFehlt:
    'The folder “{ordner}” cannot be reached. Connect the drive or pick another folder.',
  ordnerIstDatei:
    '“{ordner}” is a file, not a folder. Pick the folder it is in, or another one.',
  ordnerGesperrt:
    'This app may not write to “{ordner}”. Pick another folder or give access to this one.',
  ordnerVoll:
    'Only {frei} left on the drive of “{ordner}”. Free up space or pick another folder.',
  karteInDeck:
    'This card is used in {anzahl} deck(s) — {decks}. Deleting it removes it there too; confirm if that is what you mean.',
  pflichtfeld: 'This field must not be empty — enter something.',
  nichtEditierbar:
    'This field cannot be edited — the template only shows it. Change the value it is derived from.',
  berechnetesFeld:
    'This field is derived from others and cannot be set across many cards. Its parts are listed separately — pick one of those.',
  keinSetMitStil:
    'No open set uses this stylesheet. Open a set that uses it — or change the stylesheet of the open one.',
  vorlageFehlt:
    'No template is installed for this card. Check the Packages page for the set’s game and stylesheet — without them a card has no fields and no areas.',
  sicherungNeuereFassung:
    'This backup comes from a newer version of MagicSetEditor4Mac. It carries {unbekannt}, and this version reaches up to {stand}. Get the newer version — or pick an older backup.',
  sicherungAeltereFassung:
    'This backup comes from a version before 0.7.0 — it lacks schema steps up to {grundlinie}, and this version can no longer upgrade it. Restore it with MagicSetEditor4Mac 0.7.0; after that, this version opens the library.',
  sicherungUnbrauchbar:
    'This file is not a MagicSetEditor4Mac backup. Pick one that this application wrote.',
  // The parser's own sentence is German for now — see issue #1178.
  skriptFehler:
    'The expression cannot be evaluated: {meldung} Correct it at the place named.',

  // **The missing second factor** — piece 7 of phase 4.12 (#1351).
  // **The local session** — ID-02 (#1701).
  kennwortFalsch:
    'This password is not correct. Enter it again — or set a new one with the recovery sheet.',
  kennwortLaenge:
    'A password has at least {mindestens} and at most {hoechstens} characters. Choose one of that length.',
  sperreOhneRueckweg:
    'Without a password and without local unlock, the lock could not be lifted. Set a password first.',
  lokaleEntsperrungFehlt:
    'Local unlock is not set up for this user on this Mac. Enter the password instead.',
  codeFalsch:
    'This code is not correct — or it has been used already. Check the code on the sheet — each works once.',
  zweiterFaktor:
    'This user requires a second factor. Enter the code from your authenticator app or confirm with Touch ID.',
  faktorGesperrt: 'Too many wrong codes. Wait a moment and try again.',
  sicherungFremdeFassung:
    'This backup does not match this version of MagicSetEditor4Mac. It names {dort}, and this application is {hier}. Restore it with the version that wrote it.',

  // **Die Sätze des Zerteilers und des Auswerters** —
  // #1178.
  //
  // **`durchgereicht` ist die Ausnahme, und sie hat einen Grund**: die 48
  // Sätze der Standardbibliothek nennen Funktions- und Argumentnamen. Ein
  // Eintrag je Funktion wäre eine zweite Abschrift der Bibliothek in zwei
  // Sprachen; dort steht der deutsche Satz, wie er kommt.
  skript: {
    fremdesZeichen: 'the character {0} is not part of the language',
    einzelnesGleich: 'a single `=`; assign with `:=`, compare with `==`',
    offeneZeichenkette: 'a string without a closing quotation mark',
    zuTiefGeschachtelt: 'the expression is nested deeper than {0} levels',
    keinSkript: 'there is no script named {0}',
    skriptkreis: 'these fields read each other in a circle: {0}',
    keinVerweis: '`+` cannot concatenate a reference to another script',
    keineZahl: '{0} is not a number',
    formUnbekannt: 'the form {0} cannot be evaluated',
    nameUnbekannt: 'the name {0} is not known here',
    keineFunktion: '{0} is being called here, which only works for a function',
    ohneRumpf: 'this function has no body',
    nichtWahrheitswert: '`not` needs a truth value, here stands {0}',
    vorzeichenZahl: 'a sign needs a number, here stands {0}',
    zaehlschleifeZahl:
      'the {0} of a counting loop needs a number, here stands {1}',
    rechnenLinks: '`{0}` calculates with numbers, on the left {1}',
    rechnenRechts: '`{0}` calculates with numbers, on the right {1}',
    vergleichLinks: '`{0}` compares numbers, on the left {1}',
    vergleichRechts: '`{0}` compares numbers, on the right {1}',
    keineListe: 'sort_list orders a list, here stands {0}',
    zuTief:
      'the evaluation goes more than {0} levels deep — does something call itself here?',
    schleifeSteht:
      'a counting loop makes no progress at {0} — the number is too large',
    funktionVerbinden:
      'a function can only be combined with a function, here stand {0} and {1}',
    fehltNach: 'after {0}, {1} is missing',
    fehltHier: '{0} is missing here',
    klammerOffen: 'this {0} is never closed — {1} is missing',
    ausdruckGehtWeiter: 'the expression ends here, but the text continues',
    zuweisungBrauchtNamen: 'assignment goes to a name',
    durchgereicht: '{0}',
    bedingungWahrheit: 'a condition needs a truth value, here stands {0}',
    fuerJedenListe: '`for each` walks through a list, here stands {0}',
    wahrheitLinks: '`{0}` needs truth values, on the left stands {1}',
    wahrheitRechts: '`{0}` needs truth values, on the right stands {1}',
    bindenFunktion: '`@` binds arguments to a function, here stands {0}',
  },
} as const
