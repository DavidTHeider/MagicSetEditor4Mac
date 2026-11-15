// Die Statistik über ein Set.
//
// **Die Zahlen kommen aus Abfragen**, nicht aus einer Schleife über die Karten:
// eine Schleife lüde 3.000 Karten in den Arbeitsspeicher, um vier Zahlen zu
// bilden — bei jedem Blick.
//
// **Und sie ist ohne Farbe lesbar.** Jeder Balken trägt seine Zahl und seinen
// Anteil als Text; die Fläche ist die Zugabe, nicht die Auskunft. Ein Diagramm,
// dessen Aussage in der Farbe steckt, ist für jeden zwölften Mann keine.
//
// **Eine Wahl, ein Bild** — UXR-08
// (#1916),
// nach dem Artboard *Werkzeuge*. Unter den Kennzahlen steht „Worüber"
// (`Statistikwahl`), darunter genau das, was gewählt ist: die Belegung je
// Feld (die Vorgabe, wie vorher oben), die Verteilung eines Felds oder die
// Auszählung nach einer Dimension des Spielpakets. Warum nicht mehr drei
// Abschnitte, steht an der Wahl.
import { useCallback, useState } from 'react'

import { api, auspacken, type StatistikDTO, type WertzahlDTO } from '../api'
import { Auszaehlung } from './Auszaehlung'
import { Gruppenbalken } from './Gruppenbalken'
import { Balkenliste, Statistikbild } from './Statistikbild'
import {
  Statistikwahl,
  alsFeld,
  belegungswahl,
  gewaehlteDimension,
  gewaehltesFeld,
} from './Statistikwahl'
import { useAuszaehlung } from '../haken/useAuszaehlung'
import { useGruppenabgleich } from '../haken/useGruppenabgleich'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { useStatistikfilter } from '../statistikfilter'
import { Ladefeld } from './Ladefeld'

interface StatistikEigenschaften {
  // setID ist **gesetzt** — der Arbeitsplatz rendert diese Lasche nur mit
  // offenem Set und zeigt sonst seinen eigenen leeren Zustand
  // (#288).
  // Ein Wächter „kein Set gewählt" wäre hier ein Zweig, den kein Aufruf
  // betreten kann (§3).
  setID: string
}

export function Statistik({ setID }: StatistikEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.statistik
  const [wahl, setzeWahl] = useState(belegungswahl)

  // **Ein Wechsel des Sets räumt die Wahl mit** (§7, Regel 1).
  //
  // Ohne das fragte die Seite unmittelbar `Verteilung(neuesSet, altesFeld)` —
  // ein Feld, das das neue Set gar nicht führt, und die leere Antwort sähe aus,
  // als trüge dort niemand einen Wert. Dasselbe gilt für eine Dimension, die
  // das Spiel des neuen Sets nicht erklärt.
  const [vorlage, setzeVorlage] = useState(setID)

  if (vorlage !== setID) {
    setzeVorlage(setID)
    setzeWahl(belegungswahl)
  }

  const feld = gewaehltesFeld(wahl)
  const dimension = gewaehlteDimension(wahl)

  // **Eine andere Wahl beendet die Wahl eines Balkens** (UXR-F6, #2013): ein
  // Balken gehört zu dem Bild, in dem er stand, und in MSE2 leert
  // `GraphControl::setData` die Wahl, sobald die Daten wechseln
  // (`src/gui/control/graph.cpp`, Zeile 1084, Fassung f9d9356 vom 2021-07-06).
  const filter = useStatistikfilter()

  const wahlWechseln = (neue: string) => {
    setzeWahl(neue)
    filter.aufheben()
  }

  const holenZahlen = useCallback(async () => {
    const antwort = await api().Statistik(setID)

    return auspacken(antwort, () => antwort.statistik)
  }, [setID])

  const holenVerteilung = useCallback(async () => {
    if (feld === '') {
      return []
    }

    const antwort = await api().Verteilung(setID, feld)

    return auspacken(antwort, () => antwort.verteilung)
  }, [setID, feld])

  // **Alle hängen an `setID`** (§7, Regel 1): wer das Set wechselt, bekommt
  // die Zahlen des neuen — und nicht die Verteilung des vorigen.
  const zahlen = useLaden<StatistikDTO | undefined>(holenZahlen, [setID])
  const verteilung = useLaden<WertzahlDTO[]>(holenVerteilung, [setID, feld])
  const zaehlung = useAuszaehlung(setID, dimension)

  // **Die Balken, an denen die Wahl hängt** — die des gewählten Bildes, und
  // keine, solange es lädt oder scheitert: ein leeres Bild vor der Antwort
  // sagt nichts über den Balken (UXR-F6, #2013).
  const balken =
    dimension !== ''
      ? zaehlung.laedt || zaehlung.fehler !== undefined
        ? undefined
        : zaehlung.gruppen
      : feld !== '' && !verteilung.laedt && verteilung.fehler === undefined
        ? verteilung.daten
        : undefined

  useGruppenabgleich(wahl, balken)

  // **Das Ganze der Belegung sind alle Karten des Sets** — die Frage lautet
  // „wie viele meiner Karten haben dieses Feld?".
  const alleKarten = zahlen.daten?.karten ?? 0

  // **Das Ganze einer Verteilung sind die Karten, die das Feld tragen.** Die
  // leeren gehören nicht dazu; sie stehen oben als eigene Zahl.
  const mitFeld = (verteilung.daten ?? []).reduce(
    (summe, eintrag) => summe + eintrag.karten,
    0,
  )

  return (
    <>
      {/* **Eine Lasche ist ein Abschnitt, keine zweite Seite**
          (#812).
          Der `Setarbeitsplatz` trägt die erste Ebene mit dem Namen des Sets. */}
      <h2 className="text-2xl font-semibold tracking-tight">{eigene.titel}</h2>

      <Ladefeld laedt={zahlen.laedt} fehler={zahlen.fehler} leer={false}>
        <></>
      </Ladefeld>

      {zahlen.daten && (
        <>
          <dl className="mt-4 flex flex-wrap gap-8">
            <div>
              <dt className="text-leise">{eigene.karten}</dt>
              <dd className="text-2xl tabular-nums">{zahlen.daten.karten}</dd>
            </div>
            <div>
              <dt className="text-leise">{eigene.felder}</dt>
              <dd className="text-2xl tabular-nums">{zahlen.daten.felder}</dd>
            </div>
            <div>
              <dt className="text-leise">{eigene.leere}</dt>
              <dd className="text-2xl tabular-nums">{zahlen.daten.leere}</dd>
            </div>
          </dl>

          <Statistikwahl
            wahl={wahl}
            waehlen={wahlWechseln}
            felder={zahlen.daten.belegung.map((eintrag) => eintrag.name)}
            dimensionen={zaehlung.dimensionen.map((eine) => eine.name)}
          />

          {dimension !== '' ? (
            <Auszaehlung dimension={dimension} zaehlung={zaehlung} />
          ) : feld !== '' ? (
            <Statistikbild
              ueberschrift={einsetzen(eigene.verteilungVon, { feld })}
            >
              <Ladefeld
                laedt={verteilung.laedt}
                fehler={verteilung.fehler}
                leer={verteilung.daten?.length === 0}
              >
                <p className="mt-3 max-w-prose text-leise">
                  {eigene.verteilungLeer}
                </p>
              </Ladefeld>

              {verteilung.daten && verteilung.daten.length > 0 && (
                <Gruppenbalken
                  eintraege={verteilung.daten}
                  ganzes={mitFeld}
                  wahl={alsFeld(feld)}
                  quelle={feld}
                />
              )}
            </Statistikbild>
          ) : (
            <Statistikbild ueberschrift={eigene.belegung}>
              {zahlen.daten.belegung.length === 0 ? (
                <p className="mt-3 max-w-prose text-leise">
                  {eigene.belegungLeer}
                </p>
              ) : (
                <Balkenliste
                  eintraege={zahlen.daten.belegung}
                  ganzes={alleKarten}
                />
              )}
            </Statistikbild>
          )}
        </>
      )}
    </>
  )
}
