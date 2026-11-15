// **Was zwei Sammlungen einander zu bieten haben** — Werkstück T4 von Phase
// 5.6 (#1469).
//
// # Warum der fremde Stand eingesetzt wird
//
// **Weil es die Leitung noch nicht gibt.** Was ein Gegenüber zeigt, gibt sein
// Programm frei (T3); wie es hierherkommt, entscheidet der Weg — heute von
// Hand, mit `T5` über eine Verbindung. Ein Feld dafür ist ehrlicher als ein
// Knopf, der nach einem Dienst greift, den es nicht gibt.
//
// # Warum eine Zeile nur nennt, was von der Vorgabe abweicht
//
// **Weil bei einem genauen Wunsch Angebot und Wunsch dieselbe Zeile sind.**
// Der erste Anlauf schrieb beide hin, und am Fenster stand am 2026-09-18:
// „1 × Shivan Dragon · LEA #232 · Shivan Dragon · LEA #232 · normal · genau".
// Ein DOM-Lauf fängt das nicht — „steht Shivan Dragon da" trifft auch
// zweimal zu; gefunden hat es die Sichtprüfung.
//
// **Was die Vorgabe ist, sagt nichts.** `normal` und `genau` stehen an jeder
// zweiten Zeile und verstecken damit die, an denen sie etwas sagen.
//
// # Warum der Kasten hier steht und nicht im Bereich „Handeln"
//
// **Weil das Gegenüber hier steht.** Der Bereich kommt mit `T9`
// (#1474);
// bis dahin gehört der Abgleich an die Stelle, an der man das Gegenüber
// ansieht — und nicht in eine Seite, die es noch nicht gibt.
import { useId } from 'react'

import { Knopf } from './Knopf'
import { derNeutraleRang, dieEngeArt } from '../apitauschabgleich'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type {
  Tauschabgleichantwort,
  Tauschkandidat,
} from '../apitauschabgleich'

interface Tauschabgleicheigenschaften {
  stand: string
  abgleich: Tauschabgleichantwort | undefined
  laeuft: boolean
  setzeStand: (stand: string) => void
  abgleichen: () => void

  // ausUmschlag liest den fremden Stand aus einem portablen Umschlag
  // (T7, #1472).
  ausUmschlag: () => void
}

export function Tauschabgleich({
  stand,
  abgleich,
  laeuft,
  setzeStand,
  abgleichen,
  ausUmschlag,
}: Tauschabgleicheigenschaften) {
  const eigene = useTexte().seiten.handeln
  const ueberschrift = useId()
  const standfeld = useId()

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-2">
      <h3 id={ueberschrift} className="text-sm font-medium">
        {eigene.tauschTitel}
      </h3>

      <Erklaerungsabsatz
        text={eigene.tauschErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <label className="flex flex-col gap-1 text-sm" htmlFor={standfeld}>
        {eigene.tauschStand}
        <textarea
          id={standfeld}
          rows={3}
          className={standfeldstil}
          value={stand}
          onChange={(ereignis) => {
            setzeStand(ereignis.target.value)
          }}
        />
      </label>

      <div className="flex flex-wrap gap-2">
        <Knopf gewicht="leise" angehalten={laeuft} aufKlick={abgleichen}>
          {eigene.tauschAbgleichen}
        </Knopf>

        {/* **Ein Umschlag ist der Weg von drüben** (T7, #1472) — gelesen
            und nicht gebucht. */}
        <Knopf gewicht="leise" angehalten={laeuft} aufKlick={ausUmschlag}>
          {eigene.tauschAusUmschlag}
        </Knopf>
      </div>

      {abgleich !== undefined && (
        <>
          <Tauschseite
            titel={eigene.tauschGebe}
            zeilen={abgleich.gebe}
            leer={eigene.tauschLeer}
          />

          <Tauschseite
            titel={eigene.tauschBekomme}
            zeilen={abgleich.bekomme}
            leer={eigene.tauschLeer}
          />

          <p className="text-sm text-leise">
            {abgleich.hatVorschlag
              ? `${eigene.tauschVorschlag}: ${String(abgleich.vorschlag.gebe.stueck)} · ${String(abgleich.vorschlag.bekomme.stueck)}`
              : eigene.tauschOhneVorschlag}
          </p>

          {abgleich.befunde.length > 0 && (
            <ul
              aria-label={eigene.tauschBefunde}
              className="flex flex-col gap-1"
            >
              {abgleich.befunde.map((befund) => (
                <li key={befund} className="text-sm text-fehler">
                  {befund}
                </li>
              ))}
            </ul>
          )}
        </>
      )}
    </section>
  )
}

interface Tauschseiteneigenschaften {
  titel: string
  zeilen: Tauschkandidat[]
  leer: string
}

// Tauschseite zeigt eine Richtung samt ihrer Erklärung.
//
// **Die Erklärung steht an jeder Zeile.** „3 × Shivan Dragon für deinen
// Wunsch, Rang hoch, genau dieser Druck" sagt, warum diese Zeile dasteht;
// eine Liste ohne Grund wäre eine Behauptung.
function Tauschseite({ titel, zeilen, leer }: Tauschseiteneigenschaften) {
  return (
    <div className="flex flex-col gap-1">
      <h4 className="text-sm font-medium">{titel}</h4>

      {zeilen.length === 0 && <p className="text-sm text-leise">{leer}</p>}

      {zeilen.length > 0 && (
        <ul aria-label={titel} className="flex flex-col gap-1">
          {zeilen.map((zeile) => (
            <li
              key={`${zeile.angebot}/${zeile.wunsch}`}
              className="text-sm tabular-nums"
            >
              {zeile.menge} × {zeile.angebot}
              {zeile.wunsch !== zeile.angebot && <> · {zeile.wunsch}</>}
              {zeile.rang !== derNeutraleRang && <> · {zeile.rang}</>}
              {zeile.art !== dieEngeArt && <> · {zeile.art}</>}
              {zeile.fehlend > 0 && <> · −{zeile.fehlend}</>}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

const standfeldstil =
  'rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-xs ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'
