// **The “Users and profiles” page** — phase 4.12
// (#1351).
import { faktor, konto } from './seitenkonto'
import { sitzung } from './seitensitzung'

export const nutzer = {
  titel: 'Users and profiles',

  laschen: {
    allgemein: 'General',
    konto: 'Account',
    profil: 'Profiles',
    persoenlich: 'Personal',
  },

  allgemeinErklaerung:
    'A user is the person. A profile is one working environment of that person — its own sets, its own collection, its own settings. Who is working where right now shows in the sidebar under the application name; switching happens here.',

  kontoErklaerung:
    'What this user signs in with. As long as there is exactly one user with one profile, nothing is asked.',

  profilErklaerung:
    'Anyone who wants to collect separately for a tournament needs a second profile, not a second account.',

  niemand: 'No further user yet. The first one was created on first start.',
  keineProfile: 'This user has no profile yet.',

  neuerNutzer: 'Create a user',
  neuerNutzerName: 'Display name of the new user',
  neuerNutzerNameHinweis:
    'Unique and without spaces at the edges — it tells two users apart',
  neuerNutzerKnopf: 'Create user',
  nutzerAngelegt: 'Created: {name}',

  neuesProfil: 'Create another profile',
  neuesProfilNutzer: 'For which user',
  neuesProfilName: 'Profile name',
  neuesProfilNameHinweis:
    'Unique per user — two people may both keep a “Tournament”',
  neuesProfilKnopf: 'Create profile',
  profilAnlegenOeffnen: 'Create profile …',
  profilAngelegt: 'Created: {name}',

  wahl: 'Which profile you are working in',
  wahlErklaerung:
    'Sets, decks, collection, tournaments and settings belong to the chosen profile. Switching deletes nothing — what is here will be back after switching back.',
  hier: 'You are working here',
  hierhin: 'Switch here',
  hierhinFuer: 'Switch to {nutzer} · {profil}',

  nutzerwahl: 'Who is working',
  nutzerwahlErklaerung:
    'Looking into someone else\u2019s profile does not change who you are \u2014 your own shares decide where you may go. This is not the same as the profile above. To become someone else you prove their second factor; you then work in their first profile.',
  dasBinIch: 'That is you right now',
  werden: 'Become this person',
  werdenFuer: 'Become {nutzer}',

  anmeldung: 'Sign in with a provider',
  anmeldungErklaerung:
    'To sign in with a provider, enter its `client_id` under **Settings › User and profile › Signing in with a provider**. MagicSetEditor4Mac ships no credentials; you create your own OAuth application at the provider, and a secret, if one is required, goes into the keychain and not into the library.',
  anmelden: 'Sign in',
  anbieterstand: '{eingerichtet}, {geprueft}',
  eingerichtet: 'set up',
  nichtEingerichtet: 'not set up',
  geprueft: {
    simuliert: 'tested only against a simulation',
    realGeprueft: 'tested against the real provider',
  },
  anmeldenBei: 'Sign in with {anbieter}',
  anmeldungAbbrechen: 'Cancel',
  anmeldungAbbrechenBei: 'Cancel sign-in with {anbieter}',
  verbunden: 'Connected',
  verbundenAls: 'Connected as {name}',
  loesen: 'Disconnect',
  loesenVon: 'Disconnect from {anbieter}',

  anbieternamen: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  } as Record<string, string>,

  pflege: 'Edit users and profiles',
  pflegeErklaerung:
    'A name can be changed. To delete, the name must stay as it is — it confirms who is meant.',
  bearbeiten: 'Edit {name}',
  bearbeitenKurz: 'Edit',
  neuerNameFuer: 'Name',
  umbenennen: 'Rename',
  loeschen: 'Delete',
  loeschenBestaetigt: 'Delete {name} permanently',
  loeschenFrage: 'Delete {name} permanently?',
  loeschenJa: 'Delete permanently',
  abbrechen: 'Cancel',
  nutzerLoeschenHinweis:
    'A deleted user takes their profiles along. The last user cannot be deleted.',
  profilLoeschenHinweis:
    'The last profile of a user cannot be deleted. What lies in the profile stays in the inventory but can no longer be reached.',

  angaben: 'Personal details',
  angabenErklaerung:
    'Everything except the display name is optional. **Salutation** is a free field: any list would leave somebody out, and an application that forces a choice between wrong entries has addressed the person wrongly. Address, phone number and password are under **Account**.',
  angabenSichern: 'Save details',
  felder: {
    name: 'Display name',
    anrede: 'Salutation',
    namenszusatz: 'Name affix',
    vorname: 'First name',
    weiterenamen: 'Further first names',
    nachname: 'Last name',
    slogan: 'Tagline',
    kurzebio: 'Short biography',
    langebio: 'Long biography',
  },

  // **Die Profil-Einstellungen** — Werkstück 9 von Phase 4.12 (#1351).
  badge: 'Profile badge',
  badgeVon: 'Badge of the profile of {name}',
  badgeWaehlen: 'Choose badge …',
  badgeEntfernen: 'Remove badge',
  profilangaben: 'Profile details',
  profilangabenErklaerung:
    'These belong to **this profile**, not to you — anyone keeping several tells them apart here. The **tagline** goes under your own and does not replace it. The **badge** sits at the bottom right over your profile picture: it shows the role you are currently working in.',
  profilangabenSichern: 'Save profile',
  profilfelder: {
    name: 'Profile name',
    subslogan: 'Tagline',
    kurzebeschreibung: 'Short description',
    langebeschreibung: 'Long description',
  },

  // **Technische Profile und ihre Zugriffstoken** — Werkstück 10 von
  // Phase 4.12 (#1351), Punkte 12 und 13.
  technisch: 'Technical profile — nobody stands behind it',
  technischErklaerung:
    'A technical profile keeps what a service writes on its own apart from what you do yourself. Anyone letting a service work on their behalf does not want it writing in **their** profile: what appears there was not made by their hand. It still belongs to you, because somebody has to be able to switch it off.',
  ablauf: 'Expires on (empty: never)',
  ablaufErklaerung:
    'A point in time after which this profile has expired — **not an order to delete**. Nothing vanishes on its own.',

  token: 'Access tokens',
  tokenErklaerung:
    'This is how an automated caller identifies itself at the REST endpoint — in **this** profile, with the permission set here. The window, the command line and MCP ask for no token: there it is already settled who is working.',
  tokenLeer: 'This profile has no access token yet.',
  tokenBezeichnung: 'What it is for',
  tokenScope: 'Permission',
  tokenLesen: 'Read only',
  tokenSchreiben: 'Read and write',
  tokenAblauf: 'Expires on (empty: never)',
  tokenAnlegen: 'Create token',
  tokenAendern: 'Apply',
  tokenLoeschen: 'Revoke',
  tokenLoeschenFrage: 'Revoke the token “{name}”?',
  tokenLoeschenErklaerung:
    'Whoever uses it is locked out afterwards. A new token has to be entered there again.',
  tokenLoeschenJa: 'Revoke',
  tokenAbgelaufen: 'expired',
  tokenUnbefristet: 'never expires',

  tokenEinmalig:
    'The plain value appears **exactly once**, right after it is created. After that it exists nowhere; whoever loses it creates a new token.',
  tokenKlartext: 'Copy it now — this value will not come back:',

  // **Rotation, bulk revocation, expiry presets and MCP client set-up** —
  // piece O2 of phase 5.2 (#1426).
  tokenPrincipal: 'Who it works for',
  tokenAblaufwahl: 'Expiry',
  tokenAblaufwahlen: {
    tag: 'One day',
    monat: 'Thirty days',
    jahr: 'One year',
    nie: 'Never',
    eigen: 'Custom time',
  },
  tokenRotieren: 'New secret',
  tokenRotierenFrage: 'Generate a new secret?',
  tokenRotierenErklaerung:
    'The old one stays valid for one more hour so a running job can finish. The new plain value appears exactly once.',
  tokenAlleWiderrufen: 'Revoke all',
  tokenAlleWiderrufenFrage: 'Revoke every access token of this profile?',
  tokenAlleWiderrufenErklaerung:
    'Any automated caller using one of them is locked out afterwards — even mid-session. Whatever it created stays.',
  tokenGezogen: '{zahl} tokens revoked.',
  tokenLetzteNutzung: 'last used {zeit}',
  tokenNieBenutzt: 'never used',
  tokenVorgaengerBis: 'old secret valid until {zeit}',
  tokenOnboarding: 'How to set up an MCP client',
  tokenOnboardingSchritt1:
    'Create a token and copy the plain value — it appears only once.',
  tokenOnboardingSchritt2:
    'Write it into a file only you can read (mode 0600). A file readable by others is rejected by “mse-mcp”.',
  tokenOnboardingSchritt3:
    'Start “mse-mcp” with MSE_TOKEN_DATEI set to that file — as an environment variable, never an argument: that shows in the process list.',
  tokenOnboardingSchritt4:
    'The token is checked on every tool call: a revocation, an expiry or a smaller permission takes effect on the next call. Without a token, “mse-mcp” runs as before, unauthenticated in the shown profile.',
  tokenOnboardingBeispiel:
    '{ "mcpServers": { "mse": { "command": "/Applications/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp", "env": { "MSE_TOKEN_DATEI": "/Users/ich/.mse/mcp-token" } } } }',
  tokenOnboardingBeispielTitel: 'Setup example',

  // **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
  // (#1351), Punkte 4 bis 7.
  vorgaben: 'Defaults for this profile',
  vorgabenErklaerung:
    'What a new record starts with in **this profile** — a format, a game, a template from the store. Unlike the settings, defaults point at **rows in the store**; what you can choose depends on what is there.',
  vorgabenamen: {
    format: 'Default format for new decks',
    spiel: 'Default game for new sets',
    vorlage: 'Default template for new sets',
    deck: 'Preferred decks',
    lagerort: 'Assigned storage locations',
  } as Record<string, string>,
  vorgabenOhne: 'no default',
  vorgabenLeer: 'There is nothing for this in the inventory yet.',
  vorgabenVerwaist: 'Points nowhere, because this no longer exists:',

  // **Rechte, Rollen und Sichtbarkeit** — Werkstück 11 von Phase 4.12
  // (#1351).
  freigaben: 'Who may work in this profile',
  freigabenErklaerung:
    'A **reader** only sees that this profile exists — not what is in it. A **collaborator** switches into it and works there. A **manager** grants access themselves. This profile is yours if you created it; then you manage it anyway.',
  freigabenNurLesen: 'Only whoever manages this profile can grant access.',
  freigabenLeer: 'So far nobody but the owner may enter this profile.',
  freigabeNutzer: 'Whom',
  freigabeNiemand: 'nobody',
  freigabeRolle: 'As what',
  freigabeVergeben: 'Grant',
  freigabeZurueckziehen: 'Revoke',
  rollennamen: {
    leser: 'Reader — sees that the profile exists',
    mitarbeiter: 'Collaborator — switches in and works there',
    verwalter: 'Manager — grants access themselves',
  } as Record<string, string>,

  bild: 'Profile picture',
  bildVon: 'Profile picture of {name}',
  bildWaehlen: 'Choose picture …',
  bildEntfernen: 'Remove picture',
  ohneBild: '{name} has no profile picture',

  nochNicht:
    'The data is separated between profiles — as long as nobody opens the store file with a foreign tool. Encrypting the store under “Settings › Data and backup” helps against that.',

  // **Die vier Kontoarten** — Werkstück 6 (#1351), in einer eigenen Datei.
  konto,
  sitzung,

  // **Der zweite Faktor** — Werkstück 7 (#1351).
  faktor,

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351),
  // Punkte 9, 10 und 11.
  //
  // **Am 2026-09-06 aus `seiteneinstellungen` hierher gezogen**, nach dem
  // Zweitblick über die Phase: die Einstellungen einer Person lagen an zwei
  // Orten, und wer sie suchte, suchte an beiden.

  persoenlich: 'My account',

  persoenlichErklaerung:
    'These settings belong to **you**, not to this profile — you take them into every profile of yours. The upper two override the operating system, and **one way only**: they turn on, never off.',

  auskunft: 'My data',

  auskunftErklaerung:
    'What this program has stored about you — details, sign-in ways, the **kinds** of second factors, settings and profile names. **Without password and without secret**: neither is a datum about you, but a means of being you. Sets, decks and collection are not in there — they belong to a profile and leave by their own ways.',

  auskunftSichern: 'Save my data …',

  datenschutz: 'What this application knows about you',

  datenschutzHinweis:
    'Which details are recorded, where they live, what leaves this machine — and what expressly does not. Described under **Help → What this application knows about you**.',
} as const
