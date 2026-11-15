// **Der Ausgabeassistent: seine Seiten und seine Ziele** — UXR-09
// (#1917),
// nach dem Artboard *Ausgeben* des Entwurfs.
//
// # Drei Seiten, ein Assistent
//
// **Die Seiten bleiben Seiten** — wie bei den Schubladen (`schubladen.ts`):
// `ausgeben`, `bilder` und `druckbogen` behalten ihren Schlüssel, ihr Kürzel
// (⌥⌘9, ⌥⌘1, ⌥⌘0), ihren Eintrag in Palette und Menü. Was sich ändert, ist
// der Ort: alle drei schlagen **denselben** Assistenten auf, und `bilder` und
// `druckbogen` wählen darin ihr Ziel vor (`docs/PROTOTYP-0.6.0.md`,
// Inventurdifferenz: „Ausgeben — Ziel Bilder", „Ausgeben — Ziel Druckbogen").
//
// # Fünf Ziele
//
// **Aus dem Entwurf, und der aus dem, was schon da war**: die fünf Wege, die
// seit #677 auf der Lasche *Ausgeben* standen — Setdatei, Bilder, Druckbogen,
// Exportvorlage, MTG JSON —, in derselben Reihenfolge. Das Artboard zeichnet
// dieselben fünf.
import type { Seitenschluessel, Setschluessel } from './navigation'

export const ausgabeseiten = [
  'ausgeben',
  'bilder',
  'druckbogen',
] as const satisfies readonly Setschluessel[]

export type Ausgabeseite = (typeof ausgabeseiten)[number]

export const ausgabeziele = [
  'setdatei',
  'bilder',
  'druckbogen',
  'vorlage',
  'mtgjson',
] as const

export type Ausgabeziel = (typeof ausgabeziele)[number]

// verlangtesZiel ist das Ziel, das eine Seite vorwählt — `ausgeben` wählt
// keines: dort steht, was zuletzt gewählt war.
export function verlangtesZiel(seite: Ausgabeseite): Ausgabeziel | undefined {
  return seite === 'ausgeben' ? undefined : seite
}

// istAusgabeseite sagt, ob eine Seite den Ausgabeassistenten aufschlägt.
export function istAusgabeseite(
  seite: Seitenschluessel,
): seite is Ausgabeseite {
  return (ausgabeseiten as readonly Seitenschluessel[]).includes(seite)
}

// ausgabeplatz gibt den Platz, an dem eine Setseite im Arbeitsplatz steht —
// **die drei Seiten des Assistenten teilen sich einen**. Sonst stünden drei
// Assistenten verborgen nebeneinander, jeder mit seiner eigenen Wahl, und
// welcher dastünde, entschiede die Taste, mit der man kam.
export function ausgabeplatz(seite: Setschluessel): Setschluessel {
  return istAusgabeseite(seite) ? 'ausgeben' : seite
}
