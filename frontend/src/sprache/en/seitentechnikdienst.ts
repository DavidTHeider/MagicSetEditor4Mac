// Die Texte der Technikseite über **Dienste und Wege** — welches Zeugnis
// ein Dienst vorzeigt (NET-03, #1707) und wie ein Gegenüber über fremde
// Netze angebunden ist (NET-02, #1706). Aus `seitentechnik.ts` herausgelöst,
// weil die Datei an der Zeilengrenze aus §2 stand.

export const dienstzeugnis = {
  eigenes: 'Own certificate from {aussteller}, valid until {bis}',
  identitaet:
    'Certificate on this profile’s identity — whoever connects compares the fingerprint.',
}

export const fernwege = {
  titel: 'How the other sides are connected',
  keine: 'Nobody has used this invitation yet.',
  direkt: 'Direct, via {adresse}',
  relais: 'Through the relay {adresse} — no direct path yet',
  aufbau: 'Known, but not reachable over any path yet',
  unbestimmt: 'Connected: {anzahl} — the path is not known yet.',
}

// **Command line and tool server with one click** (OPS-09, #1907).
export const schnittstellen = {
  titel: 'Set up the command line and tool server',
  satz: 'One button adds what the shell, Codex, Claude and the service need, a second removes exactly that again. Beforehand you see which file and which lines; entries by others stay as they are.',
  namen: {
    befehle: 'The links to mse and mse-mcp in ~/.local/bin',
    befehlszeile: 'The command line in the login shell’s PATH',
    codex: 'The tool server for Codex',
    claudedesktop: 'The tool server for Claude Desktop',
    claudecode: 'The tool server for Claude Code',
    dienst: 'The scripting service as a LaunchAgent',
  },
  staende: {
    offen: 'not set up',
    eingerichtet: 'set up',
    veraltet: 'out of date',
    fremd: 'someone else’s entry',
    nichtGefunden: 'not on this Mac',
    unnoetig: 'not needed',
  },
  staendeSaetze: {
    veraltet: 'The app is somewhere else now; Set up repairs it.',
    fremd: 'Someone else’s entry is already there — it stays as it is.',
    unnoetig: 'It is already in the PATH.',
  },
  zeilen: 'These lines are added',
  vorschau: {
    json: 'This entry goes into the file',
    befehl: 'This command runs',
    agent: 'This file is created, and this command loads it',
  },
  nurEintrag:
    '{name} had changed since — only the app’s own entry was removed; the rest stays as you left it.',
  einrichten: 'Set up',
  widerrufen: 'Revoke',
  alle: {
    titel: 'Everything at once',
    satz: 'Sets up every target that exists on this Mac and is still missing — or takes back everything that is set up. Foreign entries stay as they are.',
    vorschau: 'What will be written ({anzahl} targets)',
    vorschauLeer: 'Nothing is missing — whatever exists here is set up.',
    einrichten: 'Set up everything',
    widerrufen: 'Take everything back',
    ergebnis: 'What happened',
    ausgaenge: {
      eingerichtet: 'set up',
      widerrufen: 'taken back',
      fremd: 'foreign entry, unchanged',
      unveraendert: 'unchanged',
      gescheitert: 'failed',
      abgebrochen: 'cancelled',
    },
  },
  lauschwahl: {
    titel: 'Where the services listen',
    satz: 'Applies to the service for scripts — started in the window and as a LaunchAgent — and to the tool server over HTTP, from their next start.',
    wahl: 'Reachable through',
    schleife: 'This Mac only (127.0.0.1)',
    auswahl: 'Selected addresses',
    alle: 'All interfaces (0.0.0.0)',
    alleSatz: 'Every interface, including one added later.',
    aussen:
      'A chosen address leaves this Mac: the services then require TLS and an access token, and macOS may ask for local network permission.',
    inaktiv: 'switched off',
    ohneAdresse: 'no address',
    liste: 'Interfaces of this Mac',
    arten: {
      schleife: 'Loopback',
      netz: 'Network',
      tunnel: 'Tunnel (VPN)',
      bruecke: 'Bridge',
      direkt: 'Direct wireless',
    },
    speichern: 'Apply choice',
    gespeichert: 'Applied — takes effect the next time a service starts.',
    keineGewaehlt: 'Choose at least one address.',
    platzhalter: 'as under “Where the services listen”',
    weitere: 'Also listens on',
    ohne: 'Without {adresse}: {grund}',
  },
}
