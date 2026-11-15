import { dienstkatalog } from './einstellungenkatalog'
import { einstellungsbelegung } from './einstellungsbelegung'
import { einstellungsgruppennamen } from './einstellungsgruppennamen'
import {
  schalternamen,
  schalterwerte,
  seitenschalter,
} from './einstellungenschalter'

// **Die Texte der Einstellungsseite** — ausgelagert für Phase 4.8
// (#1321).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als
// einer ankommt** (§9) — dieselbe Gestalt wie `seitenkarten.ts`. Mit den
// Dienstschaltern ging `seiten.ts` über die Zeilengrenze aus §2.
import { ki } from './seitenki'
import { netzbereich, netzhilfe, netznamen } from './einstellungennetz'
import { handelshilfe, handelsnamen, handelswerte } from './einstellungenhandel'
import {
  gestaltungsbeiwerk,
  gestaltungshinweise,
  gestaltungsnamen,
  gestaltungswerte,
} from './einstellungengestaltung'
import { zugaenge } from './einstellungenzugaenge'
import { ansichtsnamen, ansichtswerte } from './einstellungenansicht'
import { vorgabennamen } from './einstellungenvorgaben'
import {
  protokollhinweise,
  protokollnamen,
  protokollwerte,
} from './einstellungenprotokoll'
import { austausch } from './einstellungenaustausch'
import { eigenuebersetzung } from './eigenuebersetzung'
import { kihinweise, kinamen, kiwerte } from './einstellungenki'

export const einstellungen = {
  ...einstellungsgruppennamen,
  ...seitenschalter,

  titel: 'Settings',

  namen: {
    ...schalternamen,
    sprache: 'Language',
    'sprache.eigene': 'Own translation',
    darstellung: 'Regional format',
    zeitzone: 'Time zone',
    'filter.gespeichert': 'Saved filters',
    'kuerzel.ziffern': 'Mapping of ⌘0 to ⌘9',
    erscheinungsbild: 'Theme',
    schriftgroesse: 'Text size',
    aktualisierung: 'Check for new versions',
    ersteschritte: 'Introduction on start',
    seitenleiste: 'Sidebar',
    leistenlage: 'Sidebar position',
    fensterstart: 'Window at startup',
    fensterbreite: 'Last remembered width',
    fensterhoehe: 'Last remembered height',
    schriftart: 'Display typeface',
    feldtiefe: 'Card form',
    'tisch.autoanordnung': 'Arrange cards on the table automatically',

    'barrierefreiheit.bewegung': 'Motion',
    'barrierefreiheit.kontrast': 'Contrast',
    'sicherheit.faktorbeimwechsel': 'Second factor when switching profiles',

    ...vorgabennamen,
    ...protokollnamen,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **A cloud provider's address** — piece 3 of A18 (#1533), labelled with
    // #1540.
    'dienst.anthropic.adresse': 'Anthropic address',
    'dienst.openai.adresse': 'OpenAI address',
    'dienst.xai.adresse': 'xAI address',
    ...netznamen,
    ...handelsnamen,
    ...gestaltungsnamen,
    ...ansichtsnamen,
    'konto.schau': 'Read my own market account',
    ...kinamen,
    // **The client id per provider** — Werkstück 4 of Phase 4.12 (#1351).
    'anbieter.github.kennung': 'Client id at GitHub',
    'anbieter.google.kennung': 'Client id at Google',
    'anbieter.microsoft.kennung': 'Client id at Microsoft',
    'anbieter.x.kennung': 'Client id at X',

    // **The workspace and the two limits** — Werkstück O3, Phase 5.2 (#1427).
    'arbeitsplatz.navigator': 'Left column',
    'arbeitsplatz.inspektor': 'Right column',
    'arbeitsplatz.navigatorbreite': 'Width of the left column',
    'arbeitsplatz.inspektorbreite': 'Width of the right column',
    'richtlinie.nurlokal': 'Work locally only',
    'richtlinie.kostengrenze':
      'Cost limit in the smallest unit (cents for EUR)',
  },

  // Origin, default and search — work item U11 of phase 5.1 (#1422).
  herkunftMarke: 'set by {herkunft}',
  wirksamKurz: 'In effect: “{wert}”.',
  wirksamOffenKurz: 'The computer’s default applies.',
  wirksamGehalten: 'In effect: “{wert}” — a policy holds it.',
  herkunftZeile: 'Set by {herkunft}, level {ebene}.',
  herkunftNamen: {
    profil: 'this profile',
    nutzer: 'you',
    geraet: 'this device',
    vorgabe: 'nobody (default)',
  },
  ebenenNamen: {
    profil: 'profile',
    nutzer: 'user',
    geraet: 'device',
    arbeitsplatz: 'workspace',
  },
  ebenenErklaerung:
    'Every setting applies at one level: device, user, profile or workspace. **Device** describes this machine — a path, a window size — and does not travel with an exchange. **User** applies to you in each of your profiles. **Profile** applies only here and overrides what you set as a user. **Workspace** is how the columns of this window stand. What applies is what the profile sets, otherwise what you set as a user, otherwise the default — for a policy, the strictest value. No setting requires administrator rights.',
  fachleute: 'What experts can also set here',
  fachleuteSatz:
    'Timeouts, reachability over other networks, the way out and the services for other programs. Expanding changes nothing — it only shows more.',
  zuruecksetzen: 'Reset to “{vorgabe}”',
  erbeMitWert: '{erbe} ({wert})',
  zuruecksetzenVon: 'Reset {name} to its default',
  suche: 'Find a setting',
  suchePlatzhalter: 'Name, key or value',
  sucheTreffer: '{zahl} matches',
  sucheKeine: 'No setting matches “{frage}”.',
  zumGebiet: 'Show in the “{gebiet}” area',

  // **Below the appearance settings** — UX-09 (#1681).
  gestaltung: gestaltungsbeiwerk,

  hinweise: {
    ...protokollhinweise,
    feldtiefe:
      'How much of the card form shows. **Normal** is what you fill in; **Advanced** adds what a game-package script computes; **Expert** adds what the game package does not explain at all — `stylesheet` and its kin. The three nest: deeper means more, never something else. Applies per profile.',
    'sprache.eigene':
      'Which translation of your own the window speaks — empty means none. If it is gone, the chosen language applies again.',
    sprache:
      'Applies to the whole window and to the menu bar. What comes from an MSE package — field names, game names — stays as written in the file. German is the source language; all other languages are machine-translated and not yet reviewed by an editor.',
    darstellung:
      'How numbers, amounts and dates look — for example **de-CH** or **en-GB**. Empty means: as the system has it. The language of the texts is not affected.',
    zeitzone:
      'Which zone times are shown in — for example **Europe/Berlin**. “Same as the system” means the computer’s own. It is stored per profile and does not travel with an exchange.',
    'filter.gespeichert':
      'The saved filters of every view as a list in JSON — name, view, marks and whether **and** or **or** applies. They are created in the navigator of their view; they belong to the profile and travel with the settings export.',
    'kuerzel.ziffern':
      'Which page a digit opens with ⌘. Empty means: the areas of the sidebar. It is set at the shortcut table under “General”.',
    erscheinungsbild:
      'Light, dark, or whatever the computer says. The choice applies to this application only.',
    schriftgroesse:
      'Scales the whole interface, not just the text — spacing and borders grow with it. The card preview stays as it is: it shows what would be printed.',
    aktualisierung:
      'The default is **off**: nothing goes out to the network and nothing is asked. With **notify**, the application checks at startup whether a newer version exists and says so — nothing more. With **download** it additionally offers a button that fetches the version into the download folder and compares its checksum; nothing is installed on its own.',
    ersteschritte:
      'The three steps to your first card, as a window on startup. The same switch is inside that window — turn it off there and you find it again here.',
    seitenleiste:
      'The bar on the left with the areas of the application. Folded, its width goes to the workspaces; ⌥⌘S, the button at its edge and the entry under **View** fold and unfold it. It is stored per profile and does not travel with an exchange.',
    leistenlage:
      'On the left it stands upright beside the content, at the top as a row above it. The same entries, shortcuts and arrow keys — only across. It is stored per profile and does not travel with an exchange.',
    fensterstart:
      'The default is **Fill the screen** — the window takes the room it has, and **no full screen**: that would take the menu bar with it. **Normal size** opens at the built-in size, **As last time** at whatever you left it. It is stored per profile and does not travel with an exchange.',
    fensterbreite:
      'What was measured when the window last closed — **only for "As last time"**. Empty means there was no last time yet, and the window opens at its built-in size.',
    fensterhoehe:
      'The same for the height. Both are written when the window closes; a measurement outside 700 to 8,192 points is not remembered — a window that does not fit the next screen could not be grabbed any more.',
    schriftart:
      'The typeface for headings and set names. Body text stays the system font — a decorative face over a paragraph reads badly. All four ship with macOS; nothing is downloaded.',
    'barrierefreiheit.bewegung':
      'The override goes **one way only**: it turns motion off, never on. Anyone who has set **reduced motion** in the system does not get it here anyway — an application overriding that would work against them.',
    'barrierefreiheit.kontrast':
      'Same direction: **high** removes the fibre texture and raises the colours until every text holds 7 : 1. Anyone asking the system for more contrast gets the same.',
    'sicherheit.faktorbeimwechsel':
      'The **only** override that loosens — and only when **switching** profiles. It changes nothing about signing in: there the second factor does what it exists for. On a computer with one person, a one-time code at every switch is friction without benefit.',
    'vorgabe.zustand':
      'What a new row in the collection starts with. **No default** means the field stays empty, as it always was.',
    'vorgabe.waehrung':
      'What the currency field of a new row starts with, for example **EUR**. Nothing is converted.',
    'vorgabe.einzelspiel':
      'Whether a new tournament runs matches of a **single** game. It decides the display, not the maths — game win percentage and opponents’ are kept either way.',
    'vorgabe.hochladepfad':
      'The folder a dialog for opening a file starts in — images, MTG JSON and everything else coming off disk. A full path or one starting with **~/**; empty means wherever the operating system opens it.',
    'vorgabe.ausfuhrpfad':
      'Where a save dialog starts when the action has no place of its own. If the folder is missing when you save — an unplugged drive, say —, the app tells you instead of saving somewhere else.',
    'vorgabe.sicherungspfad':
      'Where the dialog for a backup of your library starts. A missing folder stops the backup with a message.',
    'vorgabe.diagnosepfad':
      'Where diagnostic text and bundles are saved. A missing folder stops with a message.',
    'bildausfuhr.ordner':
      'Where the image export last wrote. The Images page suggests it and records every newly chosen folder here — one for all sets, as in MSE. Empty means the folder dialog opens at the folder for files you save.',
    'ausgabe.ziele':
      'Which target Export last showed for each set. A list in JSON, the latest choice last, at most 200; it is set when you choose under “Where to?”. Empty means nothing is preselected.',
    'dienst.anthropic.adresse':
      'Where requests go. Empty means Anthropic’s own interface. A different address is a **gateway** — a service that forwards the request; its key is stored separately.',
    'dienst.openai.adresse':
      'The same for OpenAI. Empty means `https://api.openai.com`. Anything entered here means you are no longer talking to OpenAI but to whatever answers at that address.',
    'dienst.xai.adresse': 'The same for xAI. Empty means `https://api.x.ai`.',
    ...netzhilfe,
    ...handelshilfe,
    ...gestaltungshinweise,
    'konto.schau':
      'Whether this program may read your **own** account at a marketplace — your listings and your orders. **Off** means: not even a look. Switching the marketplace on is not enough: that one answers what **other** sellers ask, this one what lies **with you**. Nothing is ever written there — no listing, no cart, no purchase.',
    ...kihinweise,
    'anbieter.github.kennung':
      'The `client_id` of the OAuth application you created at GitHub. Empty means: signing in with GitHub is not offered.',
    'anbieter.google.kennung':
      'The `client_id` of the OAuth application you created at Google. It usually ends in `apps.googleusercontent.com`.',
    'anbieter.microsoft.kennung':
      'The application id from the Microsoft Entra portal — a UUID. Microsoft requires **no** secret for this flow.',
    'anbieter.x.kennung':
      'The `client_id` of the application you created in the X developer portal.',
  },

  werte: {
    ...schalterwerte,
    ...protokollwerte,
    ...handelswerte,
    ...gestaltungswerte,
    ...ansichtswerte,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **Die Kontoschau** — T11, #1476. Dieselben Worte wie bei der
    // KI-Laufzeit darüber: zwei Schreibweisen für denselben Zustand
    // lernt niemand gern.
    'konto.schau.aus': 'Off',
    'konto.schau.an': 'On',
    ...kiwerte,
    'sprache.erben': 'As set for the user',
    'erscheinungsbild.erben': 'As set for the user',
    'schriftgroesse.erben': 'As set for the user',
    'schriftart.erben': 'As set for the user',
    'sprache.de': 'Deutsch',
    'sprache.en': 'English',
    'sprache.es': 'Español',
    'sprache.fr': 'Français',
    'sprache.it': 'Italiano',
    'sprache.ja': '日本語',
    'sprache.ru': 'Русский',
    'sprache.zh': '简体中文',
    'erscheinungsbild.system': 'Match the system',
    'erscheinungsbild.hell': 'Light',
    'erscheinungsbild.dunkel': 'Dark',
    'schriftgroesse.klein': 'Small',
    'schriftgroesse.normal': 'Normal',
    'schriftgroesse.gross': 'Large',
    'schriftgroesse.sehrgross': 'Very large',
    'aktualisierung.aus': 'Off',
    'aktualisierung.melden': 'Notify only',
    'aktualisierung.holen': 'Notify and download',
    'ersteschritte.an': 'On',
    'ersteschritte.aus': 'Off',
    'seitenleiste.sichtbar': 'Visible',
    'seitenleiste.verborgen': 'Hidden',
    'leistenlage.senkrecht': 'Left',
    'leistenlage.waagerecht': 'Top',
    'fensterstart.maximiert': 'Fill the screen',
    'fensterstart.normal': 'Normal size',
    'fensterstart.zuletzt': 'As last time',

    'feldtiefe.normal': 'Normal',
    'feldtiefe.erweitert': 'Advanced',
    'feldtiefe.experte': 'Expert',
    'tisch.autoanordnung.an': 'Yes — the table lays them out',
    'tisch.autoanordnung.aus': 'No — what you place stays put',
    'tisch.beschriftung': 'Labels on the table',
    'tisch.beschriftung.aus': 'Off — counts only',
    'tisch.beschriftung.klein': 'Small',
    'tisch.beschriftung.normal': 'Medium',
    'tisch.beschriftung.gross': 'Large',
    'schriftart.pergament': 'Parchment — calm and readable',
    'schriftart.fantasy': 'Fantasy — playful',
    'schriftart.antike': 'Antiquity — Roman capitals',
    'schriftart.kalligrafie': 'Calligraphy — handwritten',
    'barrierefreiheit.bewegung.system': 'Match the system',
    'barrierefreiheit.bewegung.reduziert': 'Always little',
    'barrierefreiheit.kontrast.system': 'Match the system',
    'barrierefreiheit.kontrast.hoch': 'Always high',
    'sicherheit.faktorbeimwechsel.verlangen': 'Also when switching',
    'sicherheit.faktorbeimwechsel.nurAnmelden': 'Only when signing in',
    'vorgabe.zustand.ohne': 'No default',
    'vorgabe.zustand.mint': 'Mint',
    'vorgabe.zustand.nearMint': 'Near Mint',
    'vorgabe.zustand.excellent': 'Excellent',
    'vorgabe.zustand.good': 'Good',
    'vorgabe.zustand.lightPlayed': 'Light Played',
    'vorgabe.zustand.played': 'Played',
    'vorgabe.zustand.poor': 'Poor',
    'vorgabe.einzelspiel.aus': 'Off',
    'vorgabe.einzelspiel.an': 'On',

    // **The workspace and the limit** — Werkstück O3 of Phase 5.2 (#1427).
    'arbeitsplatz.navigator.erben': 'As set for the user',
    'arbeitsplatz.navigatorbreite.erben': 'As set for the user',
    'arbeitsplatz.inspektorbreite.erben': 'As set for the user',
    'arbeitsplatz.navigator.sichtbar': 'Visible',
    'arbeitsplatz.navigator.verborgen': 'Hidden',
    'arbeitsplatz.inspektor.erben': 'As set for the user',
    'arbeitsplatz.inspektor.sichtbar': 'Visible',
    'arbeitsplatz.inspektor.verborgen': 'Hidden',
    'richtlinie.nurlokal.aus': 'Off',
    'richtlinie.nurlokal.an': 'On',
  },

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351).
  // **Where a login comes from in the first place** — follow-up to Werkstück 4
  // of Phase 4.12 (#1351).
  anmeldung: 'Signing in with a provider',
  anmeldungHinweis:
    'This application ships **no credentials**. To sign in with GitHub, Google, Microsoft or X, create an **OAuth application** at the provider yourself and enter its `client_id` here. As the redirect URI, give the address shown in the browser during sign-in — it points at this machine and is valid only while the sign-in runs.',

  anmeldungGeheimnis:
    'GitHub additionally requires a `client_secret` — Google leaves it optional, Microsoft forbids it for programs like this one. It goes **into the keychain**, not here: the library gets shared, the keychain does not. Create it in **Keychain Access** as an **application password**, with the name `MagicSetEditor4Mac-Anmeldung` and the account `github`. On the command line: `security add-generic-password -s MagicSetEditor4Mac-Anmeldung -a github -w`.',

  geheimnis: 'Secret',
  geheimnisPflicht: 'is required and missing',
  geheimnisVerboten:
    'requires no secret and rejects one — the client ID is enough here.',

  anbietername: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  },

  wirktNachNeustart: 'Takes effect after a restart.',

  trefferInGruppe: '{zahl} matches for “{frage}”',
  gesuchtMarke: 'found',
  sucheSprung: 'Jump to the first match',

  vorgaben: 'Defaults',

  // **The pointer to the help** — Werkstück 13 of Phase 4.12 (#1351).

  dienste: 'Third-party card services',
  diensteHinweis:
    'Off means: not a single request goes to this service. The application keeps working without any service — only what comes from outside will not.',

  kiAnbieter: 'Language models',
  kiAnbieterHinweis:
    'Where questions may go. Every provider computes on their machines and sees what you send — so each is switched on its own, and none is on by itself.',
  // See the German catalogue for why these two live here and not under
  // `werte`: they are the same for every service.
  dienstAus: 'Off',
  dienstAn: 'On',
  dienstUngeprueft: 'Never run against the real service',
  dienstUngeprueftHinweis:
    'There is no account for this service here. What is tested is our side of the boundary — the request, the headers, the mapping. Whether the service answers the way we assume is not measured.',
  fragen: 'Who answers what',
  fragenHinweis:
    'When several services can answer the same question, this decides which one is asked. **There is no silent fallback**: if the chosen one stays quiet, you are told — instead of quietly seeing an answer from a different source.',
  'frage.suche': 'Card search',

  // The service catalogue — work item A1 of phase 5.3 (#1440).
  ...dienstkatalog,

  dienstSchluesselFehlt: 'No credentials stored',
  dienstSchluesselHinweis:
    'This service needs credentials. They live in the macOS keychain, not in the collection file — that one gets passed around.',
  zugaenge: 'Credentials',
  zugaengeHinweis:
    'They live in the macOS keychain and **not** in the collection file — that one gets passed around. What you see here is only **whether** something is stored; the value itself never comes back.',
  // See the German catalogue: these name the headers each service asks for
  // (documented in decision 0056).

  kuerzelUeberschrift: 'Keyboard shortcuts',
  ...einstellungsbelegung,
  kuerzelHinweis:
    'They apply throughout the window. Inside a text field, ⌘Z and ⌘F belong to the field.',
  kuerzelZweiteReihe:
    'The eight shortcuts with ⌥ sit on the keys of the top row, from the left. ' +
    'On a keyboard with a different layout — AZERTY, say — a different character ' +
    'is printed there; the key is the same.',
  kuerzelTaste: 'Key',
  kuerzelHandlung: 'Action',

  austausch,
  eigenuebersetzung,
  // **An account's remaining quota at a provider** — A5, #1444.
  kontingentTitel: 'Provider quota',
  kontingentErklaerung:
    'How many credits the plan holds and how many are left. The lookup goes out to the network and costs a credit itself.',
  kontingentHolen: 'Check quota',
  kontingentHolenBei: 'Check quota at {dienst}',
  kontingentLaeuft: 'Checking …',
  kontingentZahlen: '{rest} of {gesamt} credits left.',
  kontingentFrist:
    'The provider may report this figure up to {minuten} minutes late — it is not a cap. What actually limits spending here is this library\u2019s budget.',
  kontingentUeberzug: 'A credit beyond the plan costs {satz}.',
  // **What is known about AI models** — A6, #1445.
  ...ki,

  // **Die Zugangsdaten** (#1534) — eigene Datei, ein Katalog (§9).
  ...zugaenge,

  ...netzbereich,
}
