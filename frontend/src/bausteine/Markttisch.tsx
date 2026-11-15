// **Der Markttisch** — P12
// (#2380),
// Befunde C-14 und C-18, Entwurf `Markttisch09`: links das eigene Fach,
// rechts das freigegebene Fach des Gegenübers, dazwischen die zwei Hälften
// „Von dir" und „Vom Gegenüber" mit Bildern, Zustand und Bilanz (P10).
//
// **Was gezogen wird, wird eine Zeile** (`markttisch.ts`); wie gezogen wird,
// sagt der Gestenvertrag (`useMarktgriff`). **Der bisherige Assistent bleibt
// als Liste erreichbar** (`Handelsvorgang`).
import { Handelspositionen } from './Handelspositionen'
import { Fachliste } from './Marktfach'
import { Knopf } from './Knopf'
import { Ladefehler } from './Ladehinweise'
import { Laschen } from './Laschen'
import { useMarktfaecher } from '../haken/useMarktfaecher'
import { useMarktgriff } from '../haken/useMarktgriff'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Handelsposition } from '../apihandelsleitung'

interface Eigenschaften {
  ich: string
  gegenueber: string
  name: string
  positionen: Handelsposition[]
  setzePositionen: (positionen: Handelsposition[]) => void
}

export function Markttisch({
  ich,
  gegenueber,
  name,
  positionen,
  setzePositionen,
}: Eigenschaften) {
  const tisch = useTexte().seiten.handeln.markttisch
  const faecher = useMarktfaecher(gegenueber)
  const griff = useMarktgriff({
    ich,
    gegenueber,
    karten: faecher.alle,
    positionen,
    setzePositionen,
  })

  const meine = faecher.meine.daten
  const seinFach = einsetzen(tisch.seinFach, { name })

  return (
    // **Drei Spalten erst, wenn der eigene Platz reicht** — gemessen am Tisch
    // (`@container`) und nicht am Fenster (QA01, #2382): bei 1024 Punkten
    // galt `lg:`, und die drei Spalten standen in 343 Punkten neben der Liste
    // der Handel. Die Mitte schrumpft nicht unter ihren Inhalt, und die Seite
    // rollte waagerecht.
    <div className="@container flex flex-col gap-2">
      <p className="text-sm text-leise">{tisch.hinweis}</p>

      <div className="grid gap-3 @2xl:grid-cols-[minmax(0,1fr)_minmax(0,2fr)_minmax(0,1fr)]">
        <section
          aria-label={tisch.meinFach}
          className="flex min-w-0 flex-col gap-2 rounded-sm border border-kante p-3"
        >
          <p className="text-sm font-medium">{tisch.meinFach}</p>
          {/* **Ein Fehler steht da, mit dem Weg zurück** — Gesamtreview über
              0.9.0, Befund 3: das Fach las nur die Daten und zeigte nach
              einem Fehler weiter, dass es lädt. */}
          {faecher.meine.fehler !== undefined ? (
            <Ladefehler
              fehler={faecher.meine.fehler}
              wiederholen={faecher.meine.nachladen}
            />
          ) : (
            <Laschen
              bezeichnung={tisch.meinFach}
              gebiete={[
                {
                  schluessel: 'sammlung',
                  name: tisch.sammlung,
                  inhalt: (
                    <Fachliste
                      karten={meine?.filter((karte) => karte.quelle === 'mein')}
                      leer={tisch.meinLeer}
                      griff={griff}
                    />
                  ),
                },
                {
                  schluessel: 'wunschliste',
                  name: tisch.wunschliste,
                  inhalt: (
                    <Fachliste
                      karten={meine?.filter(
                        (karte) => karte.quelle === 'wunsch',
                      )}
                      leer={tisch.wunschLeer}
                      griff={griff}
                    />
                  ),
                },
              ]}
            />
          )}
        </section>

        <Handelspositionen
          ich={ich}
          gegenueber={gegenueber}
          positionen={positionen}
          setzePositionen={setzePositionen}
          bearbeitbar
          ablagen
          namen={
            new Map(faecher.alle.map((karte) => [karte.printing, karte.name]))
          }
          geholt={faecher.geholt}
        />

        <section
          aria-label={seinFach}
          className="flex min-w-0 flex-col gap-2 rounded-sm border border-kante p-3"
        >
          <p className="text-sm font-medium">{seinFach}</p>
          <p className="text-xs text-leise">
            {einsetzen(tisch.freigegeben, { name })}
          </p>
          {faecher.seine.fehler !== undefined ? (
            <p role="alert" className="text-sm text-fehler">
              {faecher.seine.fehler}
            </p>
          ) : (
            <Fachliste
              karten={faecher.seine.daten}
              leer={tisch.seinLeer}
              griff={griff}
            />
          )}
          <div>
            <Knopf
              gewicht="leise"
              angehalten={faecher.seine.laedt}
              aufKlick={faecher.seine.nachladen}
            >
              {tisch.neuFragen}
            </Knopf>
          </div>
        </section>
      </div>

      <p aria-live="polite" className="sr-only">
        {griff.ansage}
      </p>
    </div>
  )
}
