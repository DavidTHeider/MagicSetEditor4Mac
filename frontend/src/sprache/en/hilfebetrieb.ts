// **Operation and trust** — OPS-07
// (#1724).
export const betrieb = {
  ueberschrift: 'Operation and trust',
  einleitung:
    'What is installed, where it lives, what was verified and what was not — read from this application itself.',
  fassungName: 'Version',
  fassungSatz: '{fassung}, built from {stand}, the state of {gebaut}.',
  fassungVeraendert: 'The build contained changes that were not committed.',
  datenName: 'Data',
  datenSatz: '{zahl} migrations applied, the last one {letzte}.',
  datenOhne: 'No store is open.',
  signaturName: 'Signature',
  signaturSatz:
    'Ad-hoc signed, **not notarized**. On first start macOS refuses the application; allow it once under System Settings › Privacy & Security › **Open Anyway**.',
  aktualisierungName: 'Updates',
  aktualisierung: {
    aus: 'Off — this application does not ask for new versions.',
    melden: 'It reports a new version from GitHub and fetches nothing.',
    holen: 'It fetches a new version from GitHub and asks before swapping.',
  },
  orteName: 'Where everything lives',
  orte: {
    bestand: 'Store',
    pakete: 'Template packages',
    bilder: 'Inserted images',
    zwischenspeicher: 'Caches',
    protokoll: 'Log',
    leistungsprofile: 'Performance profiles',
  },
  lizenzenName: 'Licences',
  lizenzenSatz:
    'The source code is under the GNU General Public License 2.0 or later. Third-party parts with their licences are listed in `DRITTANBIETER.md` next to the source; a template package brings its own terms, and an AI model those of its provider.',
  empfaengerName: 'Where data goes',
  empfaengerSatz:
    'Only to the services you switch on — which ones and what leaves is listed under **Help › What this application knows about you**. No usage data is sent.',
  grenzenName: 'Security boundaries',
  grenzenSatz:
    'The service and the tool server listen on `127.0.0.1` only unless set otherwise, and they require a token. Give one an address in the network under **Settings › Diagnostics** and it runs only with TLS, and the page says that it leaves this computer; over stdio the tool server talks over its own streams. The store can be encrypted, under **Settings › Data and backup**.',
  kontaktName: 'When something does not work',
  kontaktSatz:
    'Under **Settings › Diagnostics** you find the diagnosis, the diagnostic bundle and the numbers of this process. Bug reports go to github.com/DavidTHeider/MagicSetEditor4Mac — nothing leaves on its own.',
  pruefstandName: 'What is verified',
  pruefstandSatz:
    'Verified means: a run proved it — a probe whose evidence, naming its commit, lives with the source, or a test that every pipeline runs; only what passed it is released. Present means: built, but never run against the original or the platform.',
  spalteBereich: 'Area',
  spalteGegenstand: 'Subject',
  spalteStand: 'Status',
  spalteBeleg: 'Evidence',
  belegIssue: 'issue #{nummer}',
  bereiche: {
    plattform: 'Platform',
    sprache: 'Languages',
    barrierefreiheit: 'Accessibility',
    regeln: 'Rules',
    integration: 'Outside services',
  },
  staende: {
    geprueft: 'verified',
    vorhanden: 'present',
    beschrieben: 'described',
    blockiert: 'waits for a human',
  },
  gegenstaende: {
    'macOS 15.3 bis 26': 'macOS 15.3 to 26',
    'macOS 15.2 auf Apple Silicon': 'macOS 15.2 on Apple silicon',
    'macOS 27 auf Apple Silicon': 'macOS 27 on Apple silicon',
    'Tastatur und Vorleser im WebKit': 'Keyboard and screen reader in WebKit',
  },
}
