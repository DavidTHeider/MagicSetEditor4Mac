// **Der Austausch der Einstellungen** — ausgelagert, damit der Katalog
// unter der Zeilengrenze aus §2 bleibt; er kommt als **einer** an (§9).
// **The exchange** — Werkstück O3 of Phase 5.2
// (#1427).
export const austausch = {
  ueberschrift: 'What this file would change',
  erklaerung:
    'Settings travel: as a JSON file holding **only the chosen** values. Whatever belongs to this machine stays behind — no access secret, no folder of this disk, no client id. Reading one in takes two steps: the preview shows line by line what would change, and only then is the rest taken over.',
  mitArbeitsplatz: 'Take the window layout along (it belongs to a screen)',
  sichern: 'Save settings',
  einlesen: 'Read a file',
  uebernehmen: 'Take over the selected',
  verwerfen: 'Discard',
  keineAenderung: 'This file changes nothing in your settings.',
  unveraendert: '{zahl} lines already read like this.',
  uebernommen: '{zahl} lines taken over.',
  befunde: 'What was skipped',
  spalteUebernehmen: 'Take over',
  spalteSchluessel: 'Setting',
  spalteVorher: 'Now',
  spalteNachher: 'Afterwards',
  spalteHerkunft: 'So far chosen by',
  zeileWaehlen: 'Take over the line {schluessel}',
  herkunft: {
    profil: 'this profile',
    nutzer: 'you',
    vorgabe: 'nobody — the default',
  },
}
