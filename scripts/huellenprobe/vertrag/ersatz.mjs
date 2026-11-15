// **Ersatzargumente** für die Bindungen ohne eigenen Eintrag — B01 (#2220).
//
// Aus Name und Typ eines Parameters und dem Kontext, den die Gruppen davor
// angelegt haben. **Das Ziel ist der Ruf, nicht der Erfolg**: wer eine
// Kennung braucht, die es nicht gibt, bekommt eine fachliche Ablehnung — und
// auch die geht durch die Naht, die hier geprüft wird (Rumpf, Marken,
// Fehlerübersetzung). Was gelingen **soll**, hat einen Eintrag in seiner
// Gruppe.
import { join } from 'node:path'

const nachName = {
  setID: (k) => k.set?.id ?? '',
  karteID: (k, typ) => (typ === 'int64' ? 0 : (k.neueKarte ?? k.karten?.[0] ?? '')),
  karteIDs: (k) => (k.karten ?? []).slice(0, 1),
  eineID: (k) => k.karten?.[0] ?? '',
  andereID: (k) => k.karten?.[1] ?? '',
  feld: (k) => k.textfeld ?? 'name',
  felder: (k) => [k.textfeld ?? 'name'],
  deckID: (k, typ) => (typ === 'int64' ? 0 : (k.deck ?? '')),
  nutzerID: (k) => k.nutzer ?? '',
  profilID: (k) => k.profil ?? '',
  spiel: (k) => k.set?.spiel ?? 'magic',
  stil: (k) => k.set?.stil ?? '',
  paket: (k) => k.paket ?? '',
  turnierID: (k) => k.turnier ?? '',
  partieID: (k) => k.partie ?? '',
  pfad: (k) => join(k.ort.ort, 'vertrag', 'ziel'),
  verzeichnis: (k) => join(k.ort.ort, 'vertrag'),
  ordner: (k) => join(k.ort.ort, 'vertrag'),
  datei: () => 'vertrag',
  wert: () => 'Vertragswert',
  name: () => 'Vertrag',
  text: () => 'Vertrag',
  inhalt: () => '',
  vergroesserung: () => 1,
  menge: () => 1,
  sprache: () => 'de',
  waehrung: () => 'EUR',
}

// ersatz gibt den Wert für einen Parameter.
export function ersatz({ name, typ }, k) {
  // **Der Name gilt nur, wo der Typ passt** — `wert` ist meist ein Text,
  // bei `KostenbudgetSetzen` aber ein Betrag (erster Rundgang, 2026-10-09).
  const zahl = ['int', 'int64', 'float64'].includes(typ)
  const zahlname = ['vergroesserung', 'menge'].includes(name)
  const passend = zahl ? zahlname || (name.endsWith('ID') && typ === 'int64') : !zahlname && (typ === 'string' || typ === '[]string')
  if (name in nachName && passend) return nachName[name](k, typ)
  if (typ === 'string') return ''
  if (typ === 'bool') return false
  if (['int', 'int64', 'float64'].includes(typ)) return 0
  if (typ.startsWith('[]')) return []
  if (typ.startsWith('map[')) return {}
  if (typ.startsWith('*')) return null

  return {}
}

// zerstoerend sagt, ob eine Bindung etwas wegnimmt oder beendet — sie läuft
// **am Ende**, damit sie den übrigen nicht den Grund nimmt.
export const zerstoerend = (name) => /Loeschen|Entfernen|Zuruecksetzen|Abmelden|Sperren|Widerrufen|Beenden|Verwerfen|Aufloesen|Vergessen|Archivieren|Loesen|Abbrechen|Rotieren|Wechseln|Umschalten|Zurueckspielen|Leeren/.test(name)

// sitzungsende sagt, ob eine Bindung die Sitzung selbst beendet oder ihren
// Grund wechselt — sie läuft **ganz zuletzt**.
//
// **In dieser Reihenfolge**: `Sperren`, dann `Abmelden` zuletzt — danach weist
// das Tor jeden Ruf ab (erster Rundgang, 2026-10-09: `Sperren` kam
// alphabetisch vor `NutzerWechseln` und sperrte fünf Bindungen aus).
export const sitzungsreihe = ['KennwortWechseln', 'ProfilWechseln', 'NutzerWechseln', 'VerschluesselungUmschalten', 'SicherungZurueckspielen', 'Zuruecksetzen', 'Sperren', 'Abmelden']
export const sitzungsende = (name) => sitzungsreihe.includes(name)
