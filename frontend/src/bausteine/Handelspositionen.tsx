// **Was bei einem Handel hin und her geht** — Werkstück T1 von Phase 5.6
// (#1466),
// angeschlossen mit
// #1637.
//
// # Warum die Seite eine Spalte ist
//
// **Weil eine Position eine Partei nennt und keine Zeichenkette** — und seit
// TRADE-02b (#1951)
// auch keine Auswahl mehr: zwei Spalten *Von dir* und *Vom Gegenüber*, und
// eine Zeile steht in der Spalte, deren Knopf sie anlegte (AK-01). Dazu je
// Zeile ein Preis, je Seite eine Summe und darunter die Differenz (AK-02) —
// ein Anhalt aus dem eigenen Preisbestand, der nirgends hingeht (AK-03).
//
// # Und warum hier kein Exemplar steht
//
// **Weil ein Angebot einen Druck nennt und keine Besitzzeile.** Welches der
// eigenen Stücke weggeht, wählt ein Mensch später — das ist eine eigene
// Fähigkeit und ein eigenes Issue
// (#1643).
import { useId, useState } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { Handelsdifferenz } from './Handelsdifferenz'
import { Handelspreise } from './Handelspreise'
import { Handelsspalte, type Spaltenzeile } from './Handelsspalte'
import { dieLeereZeile, type Formularzeile } from './Handelspositionsform'
import { Satz } from './Satz'
import { Sichtvorschlag } from './Sichtvorschlag'
import { Tauschwunschaufnahme } from './Tauschwunschaufnahme'
import { useAnschluss, useAufgenommen } from '../anschlussraum'
import { useHandelsanhalt } from '../haken/useHandelsanhalt'
import { useHandelsdrucke } from '../haken/useHandelsdrucke'
import { useTexte } from '../sprache/kontext'

import type {
  Handelsdruck,
  Handelsposition,
  Preisanhalt,
} from '../apihandelsleitung'
import type { Geholt } from '../haken/useSichtvorschlag'

interface Handelspositioneneigenschaften {
  // ich ist die eigene Peerkennung — die Seite »von dir«.
  ich: string

  // gegenueber ist die Kennung der anderen Partei.
  gegenueber: string

  positionen: Handelsposition[]
  setzePositionen: (positionen: Handelsposition[]) => void

  // bearbeitbar heißt: ein Entwurf oder ein Gegenangebot entsteht hier.
  // **Sonst stehen die Zeilen nur da** — nach der Einigung geht kein
  // Vorschlag mehr, und ein offenes Formular verspräche einen.
  bearbeitbar: boolean

  // ablagen macht die zwei Spalten zu Zielen am Markttisch (P12, #2380).
  ablagen?: boolean

  // namen sind, wie die Fächer einen Druck nennen — der Rückfall, wo der
  // Anhalt ihn nicht kennt (ein Druck des Gegenübers, den dieser Rechner nie
  // gesehen hat).
  namen?: ReadonlyMap<string, string>

  // geholt ist die Sicht, die der Markttisch schon geholt hat (#2494).
  geholt?: Geholt
}

// Seite ist eine der zwei Spalten.
type Seite = 'ich' | 'gegenueber'

export function Handelspositionen({
  ich,
  gegenueber,
  positionen,
  setzePositionen,
  bearbeitbar,
  ablagen = false,
  namen,
  geholt,
}: Handelspositioneneigenschaften) {
  const eigene = useTexte().seiten.handeln
  const ueberschrift = useId()

  const anhalt = useHandelsanhalt(positionen)
  const gemerkt = useHandelsdrucke(positionen, anhalt, ich)
  const druckAn = (position: Handelsposition, stelle: number) =>
    mitNamen(gemerkt(position, stelle), position, namen)

  // offen ist die Spalte, in der das Formular steht — höchstens eine.
  const [offen, setzeOffen] = useState<Seite>()
  const [vorlage, setzeVorlage] = useState<Formularzeile>(dieLeereZeile)

  // **Was aus der Sammlung mitgereist ist, steht schon im Formular** —
  // UX-03.AK-01
  // (#1675),
  // und zwar in der eigenen Spalte: was aus der eigenen Sammlung kommt, gibt
  // man her.
  //
  // **Die Güte reist unverändert mit** (nachgelesen am 2026-09-21: keine
  // Naht vergleicht `Handelsposition.Guete` gegen eine Aufzählung), **die
  // Menge bleibt bei eins** — wie viele Stücke weggehen, ist eine
  // Entscheidung und keine Eigenschaft des Exemplars.
  const anschluss = useAnschluss()
  const mitgereist =
    anschluss.offen?.art === 'handel' ? anschluss.offen : undefined

  useAufgenommen(mitgereist, (angekommen) => {
    setzeVorlage({
      ...dieLeereZeile,
      printing: angekommen.printing,
      guete: angekommen.guete,
      sprache: angekommen.sprache,
      finish: angekommen.finish,
    })
    setzeOffen('ich')
  })

  // uebernehmen hängt die Zeile an ihre Seite und schließt das Formular.
  const uebernehmen = (seite: Seite, zeile: Formularzeile) => {
    setzePositionen([
      ...positionen,
      {
        von: seite === 'ich' ? ich : gegenueber,
        printing: zeile.printing,
        druckId: 0,
        // **Gelesen hat sie das Formular** (F03): sie ist eine ganze Zahl.
        menge: Number(zeile.menge),
        guete: zeile.guete,
        sprache: zeile.sprache,
        finish: zeile.finish,
        abbild: '',
        bewertung: 0,
      },
    ])

    setzeOffen(undefined)
    setzeVorlage(dieLeereZeile)

    // **Verbraucht ist verbraucht** (§7, Regel 3): die Zeile steht, und der
    // Satz darüber spräche sonst weiter über eine Handlung von vorhin.
    anschluss.verbrauche()
  }

  const spalte = (seite: Seite, titel: string) => (
    <Handelsspalte
      titel={titel}
      zeilen={zeilenDer(positionen, anhalt.daten, druckAn, (von) =>
        seite === 'ich' ? von === ich : von !== ich,
      )}
      waehrung={anhalt.daten?.hauptwaehrung ?? ''}
      summe={summeDer(anhalt.daten, seite === 'ich' ? ich : gegenueber)}
      bearbeitbar={bearbeitbar}
      ablage={ablagen ? (seite === 'ich' ? 'gebe' : 'bekomme') : undefined}
      bereit={seite === 'gegenueber' || ich !== ''}
      offen={offen === seite}
      vorlage={vorlage}
      oeffnen={() => {
        setzeOffen(seite)
      }}
      schliessen={() => {
        setzeOffen(undefined)
        setzeVorlage(dieLeereZeile)
      }}
      uebernehmen={(zeile) => {
        uebernehmen(seite, zeile)
      }}
      entfernen={(stelle) => {
        setzePositionen(positionen.filter((_, andere) => andere !== stelle))
      }}
    />
  )

  return (
    <fieldset aria-labelledby={ueberschrift} className="flex flex-col gap-2">
      <legend id={ueberschrift} className="text-sm font-medium">
        {eigene.positionTitel}
      </legend>

      {/* **Ein Wunsch aus einem Deck** (UX-03.AK-01, #1675): was fehlt,
          kommt als Positionen **vom Gegenüber** herein — mit Zustand,
          Sprache und Ausführung einer neuen Zeile (#2495): ohne sie ließ
          sich die Zeile anbieten und scheiterte erst beim Verbuchen. */}
      {bearbeitbar && (
        <Tauschwunschaufnahme
          eintragen={(stuecke) => {
            setzePositionen([
              ...positionen,
              ...stuecke.map((stueck) => ({
                von: gegenueber,
                printing: stueck.printing,
                druckId: 0,
                menge: stueck.menge,
                guete: dieLeereZeile.guete,
                sprache: dieLeereZeile.sprache,
                finish: dieLeereZeile.finish,
                abbild: '',
                bewertung: 0,
              })),
            ])
          }}
        />
      )}

      {/* **Was das Gegenüber freigibt, über die Leitung** (P11, #2379). */}
      {bearbeitbar && (
        <Sichtvorschlag
          ich={ich}
          gegenueber={gegenueber}
          positionen={positionen}
          setzePositionen={setzePositionen}
          geholt={geholt}
        />
      )}

      {mitgereist !== undefined && (
        <p className="text-sm text-leise">
          {einsetzen(eigene.anschlussWartet, { name: mitgereist.name })}
        </p>
      )}

      {positionen.length === 0 && (
        <p className="max-w-prose text-sm text-leise">{eigene.positionLeer}</p>
      )}

      {/* **Am Tisch untereinander** (P12, `Markttisch09`): zwischen den
          Fächern ist die Mitte zu schmal für zwei Spalten — gemessen am
          nativen Fenster, wo jede Zeile Wort für Wort umbrach. */}
      <div className={ablagen ? 'grid gap-3' : 'grid gap-3 md:grid-cols-2'}>
        {spalte('ich', eigene.positionVonIch)}
        {spalte('gegenueber', eigene.positionVonGegenueber)}
      </div>

      {positionen.length > 0 && (
        <>
          <Handelsdifferenz anhalt={anhalt.daten} fehler={anhalt.fehler} />
          <Handelspreise positionen={positionen} geholt={anhalt.nachladen} />
        </>
      )}

      {bearbeitbar && (
        <p className="max-w-prose text-sm text-leise">
          <Satz text={eigene.positionVorbereitung} />
        </p>
      )}
    </fieldset>
  )
}

// zeilenDer gibt die Positionen einer Seite samt Stelle, Preis und Druck.
//
// **Die Stelle ist die in der ganzen Liste** — Entfernen und der Anhalt
// zählen dort, nicht in der Spalte.
function zeilenDer(
  positionen: Handelsposition[],
  anhalt: Preisanhalt | undefined,
  druckAn: (
    position: Handelsposition,
    stelle: number,
  ) => Handelsdruck | undefined,
  gehoert: (von: string) => boolean,
): Spaltenzeile[] {
  return positionen.flatMap((position, stelle) =>
    gehoert(position.von)
      ? [
          {
            position,
            stelle,
            preise: anhalt?.zeilen[stelle]?.preise,
            druck: druckAn(position, stelle),
          },
        ]
      : [],
  )
}

// summeDer gibt die Summe einer Seite in der Hauptwährung.
function summeDer(anhalt: Preisanhalt | undefined, von: string) {
  return anhalt?.seiten
    .find((seite) => seite.von === von)
    ?.summen.find((summe) => summe.waehrung === anhalt.hauptwaehrung)
}

// mitNamen nimmt den Namen aus einem Fach, wo der Anhalt keinen kennt —
// **nicht für einen eigenen Druck**: sein Name steht im eigenen Bestand.
function mitNamen(
  druck: Handelsdruck | undefined,
  position: Handelsposition,
  namen: ReadonlyMap<string, string> | undefined,
): Handelsdruck | undefined {
  const name = namen?.get(position.printing)

  return druck?.name === '' && !druck.eigen && name !== undefined
    ? { ...druck, name }
    : druck
}
