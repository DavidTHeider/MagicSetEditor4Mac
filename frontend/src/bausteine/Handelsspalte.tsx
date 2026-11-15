// **Was eine Seite gibt** — TRADE-02b
// (#1951),
// nach dem Artboard *Handelsassistent*: Titel und Summe oben, je Zeile der
// Preis rechts, darunter der Knopf, der in **diese** Spalte anlegt.
import { useId } from 'react'

import { alsGeld } from '../geld'
import { einsetzen } from '../sprache/einsetzen'
import { Aktionsknopf } from './Aktionsknopf'
import { Handelsdruckzeile } from './Handelsdruckzeile'
import { Knopf } from './Knopf'
import {
  Handelspositionsform,
  type Formularzeile,
} from './Handelspositionsform'
import { useDarstellung, useTexte } from '../sprache/kontext'

import type {
  Anhaltspreis,
  Handelsdruck,
  Handelsposition,
} from '../apihandelsleitung'

// Spaltenzeile ist eine Position samt ihrer Stelle in der ganzen Liste.
export interface Spaltenzeile {
  position: Handelsposition
  stelle: number

  // preise ist, was der Anhalt für diese Zeile nennt — `undefined`, solange
  // er nicht da ist.
  preise: Anhaltspreis[] | undefined

  // druck ist die Karte der Zeile (P10, #2378) — ebenso `undefined`, solange
  // der Anhalt nicht da ist.
  druck: Handelsdruck | undefined
}

interface Eigenschaften {
  titel: string
  zeilen: Spaltenzeile[]
  waehrung: string
  summe: Anhaltspreis | undefined
  bearbeitbar: boolean

  // bereit heißt: der Knopf dieser Spalte geht. **Die eigene wartet auf den
  // Ausweis** — ohne ihn ginge eine Zeile mit leerer Seite hinaus.
  bereit: boolean

  // offen heißt: das Formular steht in dieser Spalte.
  offen: boolean

  // ablage macht die Spalte zum Ziel am Markttisch (P12, #2380).
  ablage?: 'gebe' | 'bekomme' | undefined
  vorlage: Formularzeile
  oeffnen: () => void
  schliessen: () => void
  uebernehmen: (zeile: Formularzeile) => void
  entfernen: (stelle: number) => void
}

export function Handelsspalte({
  titel,
  zeilen,
  waehrung,
  summe,
  bearbeitbar,
  bereit,
  offen,
  ablage,
  vorlage,
  oeffnen,
  schliessen,
  uebernehmen,
  entfernen,
}: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln
  const darstellung = useDarstellung()
  const ueberschrift = useId()

  // merkmale nennt Güte, Sprache und Ausführung **als Wörter** — `nearMint`
  // und `nonfoil` sind Schlüssel, keine Sätze (Sichtprüfung am 2026-09-30).
  // Was leer ist, fehlt, etwa an einer Zeile aus einem Tauschwunsch.
  const sammlung = texte.seiten.sammlung
  const merkmale = (position: Handelsposition) =>
    [
      position.guete in sammlung.zustaende
        ? sammlung.zustaende[position.guete as keyof typeof sammlung.zustaende]
        : position.guete,
      position.sprache,
      position.finish === 'nonfoil' ||
      position.finish === 'foil' ||
      position.finish === 'etched'
        ? sammlung[position.finish]
        : position.finish,
    ]
      .filter((teil) => teil !== '')
      .join(' · ')

  // preisText nennt den Preis einer Zeile in der Hauptwährung — oder sagt,
  // dass es keinen gibt, statt null zu schreiben (AK-02).
  const preisText = (preise: Anhaltspreis[] | undefined) => {
    if (preise === undefined || waehrung === '') {
      return ''
    }

    const preis = preise.find((einer) => einer.waehrung === waehrung)

    return preis === undefined
      ? eigene.assistent.keinPreis
      : alsGeld(preis.wert, waehrung, darstellung)
  }

  return (
    <section
      aria-labelledby={ueberschrift}
      data-ablage={ablage}
      className="flex flex-col gap-2 rounded-sm border border-kante p-3"
    >
      <div className="flex items-baseline justify-between gap-2">
        <p id={ueberschrift} className="text-sm font-medium">
          {titel}
        </p>

        {summe !== undefined && (
          <p className="text-sm font-medium">
            {alsGeld(summe.wert, summe.waehrung, darstellung)}
          </p>
        )}
      </div>

      {zeilen.length > 0 && (
        <ul className="flex flex-col gap-1 text-sm">
          {zeilen.map(({ position, stelle, preise, druck }) => (
            <li
              key={`${position.printing}/${String(stelle)}`}
              className="flex items-center justify-between gap-2"
            >
              <Handelsdruckzeile
                position={position}
                druck={druck}
                merkmale={merkmale(position)}
              />

              <span className="flex shrink-0 items-center gap-2">
                <span className="text-leise">{preisText(preise)}</span>

                {bearbeitbar && (
                  <Aktionsknopf
                    aktion="entfernen"
                    gewicht="leise"
                    bezeichnung={eigene.positionEntfernenName.replace(
                      '{nummer}',
                      String(stelle + 1),
                    )}
                    aufKlick={() => {
                      entfernen(stelle)
                    }}
                  >
                    {eigene.positionEntfernen}
                  </Aktionsknopf>
                )}
              </span>
            </li>
          ))}
        </ul>
      )}

      {bearbeitbar &&
        (offen ? (
          <Handelspositionsform
            vorlage={vorlage}
            uebernehmen={uebernehmen}
            verwerfen={schliessen}
          />
        ) : (
          <Knopf
            gewicht="leise"
            angehalten={!bereit}
            bezeichnung={einsetzen(eigene.assistent.zeileIn, { seite: titel })}
            aufKlick={oeffnen}
          >
            {eigene.assistent.zeileKurz}
          </Knopf>
        ))}
    </section>
  )
}
