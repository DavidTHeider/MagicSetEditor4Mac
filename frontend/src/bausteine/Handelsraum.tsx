// **Was zwei miteinander tun** — Werkstück T9 von Phase 5.6
// (#1474).
//
// # Warum die drei Kästen beieinanderstehen
//
// **Weil sie ein Weg sind und keine fünf Fragen.** Erst sieht man, was zu
// tauschen wäre, dann sieht man nach, ob etwas unterwegs klemmt, und zuletzt
// wird verbucht, was wirklich den Besitzer gewechselt hat. Sie in eigene
// Laschen zu legen hieße, zwischen ihnen hin und her zu springen.
//
// # Welcher Handel — gewählt und nicht getippt
//
// **Eine Wahl über allen vier Kästen** (UX-08.AK-04,
// #1680).
// `HandelVorgaenge` nennt die Handel mit diesem Gegenüber, und die Kästen
// bekommen den gewählten. **Sie hängen an ihm**: wer einen anderen wählt,
// räumt Stand, Felder und Meldungen mit (§7, Regel 1).

// # Und warum die Kästen eine Trennlinie bekommen haben
//
// **Weil fünf Überschriften in einer Spalte eine Wand sind.** Bei vier ging
// es noch; mit dem Kasten aus
// #1643
// lief die Seite zusammen — gesehen an der laufenden Anwendung, nicht an
// einem Test (§3, „Sichtprüfung").
//
// # Und warum es hier keinen eigenen Vorgangsriegel gibt
//
// **Weil jeder Kasten seinen eigenen führt** (§7, Regel 4), und zwar je
// Gegenüber: `tauschabgleich/<kennung>` und `handelsbuchung/<kennung>`. Ein
// vierter Riegel darüber sperrte Wege gegeneinander, die einander nicht
// stören.
import { useRef, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { Handelsaktionen } from './Handelsaktionen'
import { Handelsausgang } from './Handelsausgang'
import { Handelsbuchung } from './Handelsbuchung'
import { Handelsjournal } from './Handelsjournal'
import { Handelsvorgang } from './Handelsvorgang'
import { Handelswahl } from './Handelswahl'
import { Handelszusagen } from './Handelszusagen'
import { Handelsassistent } from './Handelsassistent'
import { Handelsschritte } from './Handelsschritte'
import { Handelsverlauf } from './Handelsverlauf'
import { Aufklappkasten } from './Aufklappkasten'
import { lagesatzFuer } from './handelslage'
import { useVorgang } from '../haken/useVorgang'
import { useHandelszug } from '../haken/useHandelszug'
import { Zugmeldung } from './Zugmeldung'
import { Kontaktpost } from './Kontaktpost'
import { Kontakttausch } from './Kontakttausch'
import { Markenfilterknopf } from './Markenfilterkasten'
import { Objektkontextmenue } from './Objektkontextmenue'
import { Markenzeile } from './Markenzeile'
import { nachMarken, useMarkenfilter } from '../haken/useMarkenfilter'

import type { Handelsvorgang as Vorgangsstand } from '../apihandelsleitung'

interface Handelsraumeigenschaften {
  kennung: string

  // name ist, wie das Gegenüber heißt — leer, wo keiner bekannt ist. Lage,
  // Stimmen und Gegenbeleg nennen es damit (UXR-15, #1923).
  name?: string

  // vorgang ist der Handel, der aufgeschlagen ankommt — aus der Liste
  // „Was gerade läuft" (TRADE-01, #1715); archiviert sagt, ob er dort im
  // Archiv stand.
  vorgang?: string
  archiviert?: boolean

  // liste meldet eine Änderung an einem Vorgang, damit die Liste nachlädt.
  liste: () => void

  // bestandsstand zählt, was **ohne dieses Fenster** geschrieben wurde — ein
  // anderer Prozess oder die Leitung. Der Raum lädt dann nach wie die Liste.
  bestandsstand?: number
}

export function Handelsraum({
  kennung,
  name = '',
  vorgang = '',
  archiviert = false,
  liste,
  bestandsstand = 0,
}: Handelsraumeigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln.liste
  const assistent = texte.seiten.handeln.assistent
  const [gewaehlt, setzeGewaehltRoh] = useState(vorgang)
  const [kopiert, setzeKopiert] = useState(false)

  // **Die Wahl von jetzt, für Antworten von vorhin** (§7, Regel 2; Befund Z4
  // des Zweitblicks über Phase 6.8). Eine Antwort, die nach einem Wechsel
  // ankommt, stellte die Wahl sonst zurück.
  const wahl = useRef(vorgang)

  // setzeGewaehlt wählt einen Handel — und räumt den Hinweis zur Kopie.
  const setzeGewaehlt = (neu: string) => {
    wahl.current = neu
    setzeKopiert(false)
    setzeGewaehltRoh(neu)
  }

  // **Was ein Kasten zuletzt gehört hat** — nach dem Eröffnen der neue
  // Handel, bevor die Liste ihn nachgeladen hat.
  //
  // **Und nur, bis ohne dieses Fenster etwas kam** (Hüllenprobe
  // `nat04handel`, #1672): sonst stand die eigene Zustimmung vor der des
  // Gegenübers, und Bert wartete auf Anna, die längst zugestimmt hatte.
  const [zuletzt, setzeZuletztRoh] = useState<{
    stand: Vorgangsstand
    bei: number
  }>()
  const setzeZuletzt = (neu: Vorgangsstand | undefined) => {
    setzeZuletztRoh(neu && { stand: neu, bei: bestandsstand })
  }

  // **Behalten**, weil sich nur der Stand bewegt und nicht das Gegenüber:
  // der `key` am Raum räumt beim Wechsel (§7, Regel 1), und ein Handel, der
  // beim Nachladen verschwände, risse den Assistenten mit.
  const vorgaenge = useLaden(
    async () => {
      const antwort = await api().HandelVorgaenge(kennung)

      return auspacken(antwort, () => antwort.vorgaenge)
    },
    [kennung, bestandsstand],
    true,
  )

  // **Marken an Handelsvorgängen** — UX-12 (#1684): die Wahl zeigt nur, was
  // durch den Filter geht; der gewählte Vorgang trägt seine Marken darunter.
  const marken = useMarkenfilter('handel')
  const { getroffen } = marken
  const gefiltert = {
    ...vorgaenge,
    daten:
      vorgaenge.daten &&
      nachMarken(vorgaenge.daten, (einer) => einer.vorgang, getroffen),
  }

  const stand =
    zuletzt?.stand.vorgang === gewaehlt && zuletzt.bei === bestandsstand
      ? zuletzt.stand
      : vorgaenge.daten?.find((einer) => einer.vorgang === gewaehlt)

  // geaendert trägt einen neuen Stand ein und lädt die Liste nach —
  // **Bestenfalls-Arbeit** (§7, Regel 3): scheitert das Nachladen, bleibt der
  // Stand, und die Liste sagt es selbst.
  //
  // **`von` ist der Handel, aus dem geschrieben wurde.** Ist er nicht mehr
  // gewählt, bleibt die Wahl, und nur die Listen laden nach — geschrieben
  // ist geschrieben (Regel 3). Die Antwort sagt, ob sie gewählt hat.
  const geaendert = (neu: Vorgangsstand, von: string): boolean => {
    vorgaenge.nachladen()
    liste()

    if (wahl.current !== von) {
      return false
    }

    setzeZuletzt(neu)
    setzeGewaehlt(neu.vorgang)

    return true
  }

  // gebucht lädt nach einem Schritt der Buchung nach (UXR-15, #1923) — ihre
  // Antwort ist kein ganzer Vorgang, also gilt danach wieder die Liste.
  // **Nur für den Handel, aus dem gebucht wurde** (§7, Regel 2): ist ein
  // anderer gewählt, bleibt dessen Stand, und nur die Listen laden nach.
  //
  // **Und das Journal liest nach** — es steht seit UXR-15 im eigenen Kasten
  // und zählt die Schritte der Buchung mit.
  const [buchungen, setzeBuchungen] = useState(0)
  const gebucht = (von: string) => {
    vorgaenge.nachladen()
    liste()
    setzeBuchungen((bisher) => bisher + 1)

    if (wahl.current === von) {
      setzeZuletzt(undefined)
    }
  }

  // **Ein Riegel für Schritte und Vorschläge** (§7, Regel 4): beide gehen an
  // denselben Handel.
  const lauf = useVorgang()

  // **Ein Zug je Raum** (P13, #2381): seine Meldung bleibt stehen, wenn ein
  // Kasten nach dem Eröffnen neu entsteht, und geht, wenn ein anderer
  // Vorgang gewählt wird.
  const zug = useHandelszug(kennung, lauf)

  return (
    <Handelsassistent
      phase={stand?.phase ?? 'angebot'}
      vorgang={gewaehlt}
      lage={
        stand === undefined
          ? undefined
          : lagesatzFuer(
              stand,
              name === '' ? assistent.deinGegenueber : name,
              texte.seiten.handeln,
            )
      }
      schritte={
        <>
          {stand === undefined ? (
            <p className="text-sm text-leise">{assistent.neu}</p>
          ) : (
            // **Der Rechtsklick auf die Schritte** zeigt die Handlungen des
            // Handels (GESTE, #2358) — dieselben wie „Mit diesem Handel".
            <Objektkontextmenue art="handel" kennung={gewaehlt}>
              <Handelsschritte
                key={`schritte/${gewaehlt}`}
                kennung={kennung}
                stand={stand}
                geaendert={(neu) => {
                  geaendert(neu, gewaehlt)
                }}
                gebucht={() => {
                  gebucht(gewaehlt)
                }}
                zug={zug}
                lauf={lauf}
              />
            </Objektkontextmenue>
          )}

          <Zugmeldung
            ausgang={zug.ausgang}
            name={name === '' ? assistent.deinGegenueber : name}
            angehalten={zug.laeuft}
            nochmal={(bisher, legen) => {
              void zug.nochmal(bisher, legen)
            }}
          />

          {/* **Was am gewählten Handel hängt, steht bei ihm** — kopieren,
              archivieren, verwerfen und seine Marken gelten in jeder Phase
              und gehören nicht hinter „Mit wem". */}
          {gewaehlt !== '' && (
            <div className="flex flex-col gap-2">
              <Handelsaktionen
                key={`aktionen/${gewaehlt}`}
                vorgang={gewaehlt}
                stand={stand}
                archiviert={archiviert && gewaehlt === vorgang}
                hinweis={kopiert ? eigene.kopiert : undefined}
                lauf={lauf}
                kennzeichen={`handelsvorgang/${kennung}`}
                kopiert={(kopie) => {
                  if (geaendert(kopie, gewaehlt)) {
                    setzeKopiert(true)
                  }
                }}
                verworfen={() => {
                  vorgaenge.nachladen()
                  liste()

                  if (wahl.current === gewaehlt) {
                    setzeGewaehlt('')
                  }
                }}
                archiv={liste}
              />
              <Markenzeile
                art="handel"
                kennung={gewaehlt}
                name={gewaehlt}
                marken={marken.je[gewaehlt] ?? []}
                nachladen={marken.nachladen}
              />
            </div>
          )}
        </>
      }
      tafeln={{
        gegenueber: (
          <div className="flex flex-col gap-2">
            <Handelswahl
              vorgaenge={gefiltert}
              gewaehlt={gewaehlt}
              waehlen={(neu) => {
                zug.vergessen()
                setzeGewaehlt(neu)
              }}
            />
            <Markenfilterknopf art="handel" stand={marken} />
          </div>
        ),
        angebot: (
          <>
            <Handelsvorgang
              key={`vorgang/${gewaehlt}`}
              kennung={kennung}
              name={name}
              vorgang={gewaehlt}
              stand={stand}
              geaendert={(neu) => {
                geaendert(neu, gewaehlt)
              }}
              gesehen={setzeZuletzt}
              zug={zug}
              lauf={lauf}
            />
            <Aufklappkasten titel={assistent.abgleich}>
              <Kontakttausch kennung={kennung} />
            </Aufklappkasten>
          </>
        ),
        zusagen: (
          <Handelszusagen
            key={`zusagen/${gewaehlt}`}
            kennung={kennung}
            vorgang={gewaehlt}
            stand={stand}
            lauf={lauf}
          />
        ),
        verbuchen: (
          <>
            <Handelsausgang
              key={`ausgang/${gewaehlt}`}
              kennung={kennung}
              vorgang={gewaehlt}
            />
            <Handelsbuchung
              key={`buchung/${gewaehlt}`}
              kennung={kennung}
              name={name}
              vorgang={gewaehlt}
              lauf={lauf}
              stand={stand}
              gebucht={() => {
                gebucht(gewaehlt)
              }}
              zug={zug}
              nachgeladen={buchungen + bestandsstand}
            />
            {/* **Was sich bewegt hat, neben dem Vorgang und nicht in ihm**
                (UXR-15, #1923): nachsehen ist kein Schritt. */}
            <Aufklappkasten titel={texte.seiten.handeln.buchungJournal}>
              <Handelsjournal vorgang={gewaehlt} stand={buchungen} />
            </Aufklappkasten>
            <Aufklappkasten titel={assistent.unterwegs}>
              <Kontaktpost kennung={kennung} />
            </Aufklappkasten>
            {/* **Wie der Handel dorthin kam** (TRADE-03.AK-03, #1717). */}
            <Aufklappkasten titel={assistent.verlauf}>
              <Handelsverlauf key={`verlauf/${gewaehlt}`} vorgang={gewaehlt} />
            </Aufklappkasten>
          </>
        ),
      }}
    />
  )
}
