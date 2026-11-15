// **Die eine Wahl der Statistik: worüber gezählt wird** — UXR-08
// (#1916),
// nach dem Artboard *Werkzeuge*: ein Feld „Worüber" und darunter ein Bild.
//
// # Warum eine Wahl und nicht drei Abschnitte
//
// **Weil es eine Frage ist** — „wonach zähle ich dieses Set aus?". Bis UXR-08
// standen drei Antworten untereinander, jede mit eigener Überschrift: die
// Belegung immer offen, darunter ein Auswahlfeld „Feld", darunter ein
// Abschnitt mit einem zweiten Auswahlfeld „Dimension" — und die Verteilung
// des gewählten Felds erschien **unter** diesem zweiten Abschnitt, getrennt
// von ihrem eigenen Auswahlfeld. MSE2 kennt ebenfalls nur **eine** Liste von
// Dimensionen (`src/gui/set/stats_panel.cpp`, Fassung f9d9356 vom
// 2021-07-06), in der auch die Felder stehen, nach denen sich zählen lässt.
//
// # Was der Wert einer Wahl ist
//
// Eine Zeichenkette mit Vorsatz — `belegung`, `dimension:<name>` oder
// `feld:<name>`. **Mit Vorsatz**, weil ein Feld und eine Dimension denselben
// Namen tragen können (`rarity` ist beides), und dann wäre unklar, was
// gemeint ist.
//
// # Und warum sie ab fünfzehn Einträgen eine Suche trägt
//
// **R7 aus `docs/UIREVIEW.md`**: eine Auswahl **mit Suche ab fünfzehn**
// Werten. Am Probebestand sind es weniger; an einem echten Set nicht —
// `magic.mse-game` erklärt fünfzig Dimensionen, und `White.mse-set` trägt in
// gut vierzig Feldern einen Wert (beides gezählt am 2026-10-02).
//
// **Bis UXR-F1** (#2007)
// war das ein Suchfeld über einem nativen Auswahlfeld. Jetzt ist es ab
// fünfzehn Einträgen die [`Suchauswahl`](../bausteine/Suchauswahl.tsx): **ein
// Feld, in das man tippt**. Die zwei Gruppen — Dimensionen des Spielpakets und
// Felder des Sets — stehen dort als blasser Zusatz hinter dem Namen und werden
// mitgesucht; sie halten auseinander, was gleich heißt (`rarity` ist beides).
// Darunter bleibt es das native Feld mit seinen `optgroup`.
import { useId } from 'react'

import { Suchauswahl } from '../bausteine/Suchauswahl'
import { suchschwelle, type Suchwahl } from '../bausteine/suchwahl'
import { useTexte } from '../sprache/kontext'

// belegungswahl ist die Vorgabe: die Belegung je Feld, wie vorher oben.
export const belegungswahl = 'belegung'

const feldvorsatz = 'feld:'
const dimensionsvorsatz = 'dimension:'

// gewaehltesFeld gibt das Feld einer Wahl — leer, wenn sie keines nennt.
export function gewaehltesFeld(wahl: string): string {
  return wahl.startsWith(feldvorsatz) ? wahl.slice(feldvorsatz.length) : ''
}

// gewaehlteDimension gibt die Dimension einer Wahl — leer, wenn sie keine
// nennt.
export function gewaehlteDimension(wahl: string): string {
  return wahl.startsWith(dimensionsvorsatz)
    ? wahl.slice(dimensionsvorsatz.length)
    : ''
}

// alsFeld und alsDimension sind die Wahl, die ein Feld oder eine Dimension
// nennt — die Umkehr von `gewaehltesFeld` und `gewaehlteDimension`. **Auch die
// Balken brauchen sie** (UXR-F6, #2013): eine Gruppe gehört zu der Wahl, aus der
// sie stammt.
export function alsFeld(name: string): string {
  return feldvorsatz + name
}

export function alsDimension(name: string): string {
  return dimensionsvorsatz + name
}

const feldklassen =
  'rounded-sm border border-kante bg-transparent px-2 py-1 ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

interface StatistikwahlEigenschaften {
  wahl: string
  waehlen: (wahl: string) => void

  // felder sind die Felder, die im Set einen Wert tragen — aus der Belegung.
  felder: readonly string[]

  // dimensionen sind die Dimensionen, die das Spielpaket erklärt.
  dimensionen: readonly string[]
}

export function Statistikwahl({
  wahl,
  waehlen,
  felder,
  dimensionen,
}: StatistikwahlEigenschaften) {
  const eigene = useTexte().seiten.statistik
  const auswahl = useId()

  const mitSuche = 1 + dimensionen.length + felder.length >= suchschwelle

  // eintraege sind die Zeilen der Suchauswahl: die Belegung zuerst, dann die
  // Dimensionen des Spielpakets, dann die Felder — mit ihrer Gruppe als Zusatz.
  const eintraege: Suchwahl[] = [
    { wert: belegungswahl, name: eigene.belegungWahl },
    ...dimensionen.map((name) => ({
      wert: alsDimension(name),
      name,
      zusatz: eigene.gruppeSpielpaket,
    })),
    ...felder.map((name) => ({
      wert: alsFeld(name),
      name,
      zusatz: eigene.gruppeFelder,
    })),
  ]

  return (
    <div className="mt-6 flex max-w-md flex-col gap-1">
      <label htmlFor={auswahl} className="text-sm font-medium">
        {eigene.worueber}
      </label>

      {mitSuche ? (
        <Suchauswahl
          id={auswahl}
          bezeichnung={eigene.worueber}
          eintraege={eintraege}
          wert={wahl}
          aendern={waehlen}
        />
      ) : (
        <select
          id={auswahl}
          value={wahl}
          onChange={(ereignis) => {
            waehlen(ereignis.target.value)
          }}
          className={feldklassen}
        >
          <option value={belegungswahl}>{eigene.belegungWahl}</option>

          {/* **Eine leere Gruppe steht nicht da** — ein Spiel ohne erklärte
              Dimension und ein Set ohne Feldwert hätten sonst einen Kopf in
              der Liste, unter dem nichts zu wählen ist. */}
          {dimensionen.length > 0 && (
            <optgroup label={eigene.gruppeSpielpaket}>
              {dimensionen.map((name) => (
                <option key={name} value={alsDimension(name)}>
                  {name}
                </option>
              ))}
            </optgroup>
          )}

          {felder.length > 0 && (
            <optgroup label={eigene.gruppeFelder}>
              {felder.map((name) => (
                <option key={name} value={alsFeld(name)}>
                  {name}
                </option>
              ))}
            </optgroup>
          )}
        </select>
      )}
    </div>
  )
}
