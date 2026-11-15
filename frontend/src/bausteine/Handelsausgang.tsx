// **Was zu einem Gegenüber hinausgeht** — Werkstück T5 von Phase 5.6
// (#1470),
// angeschlossen mit
// #1637.
//
// # Warum „schicken" und „übertragen" zwei Knöpfe sind
//
// **Weil „abgeschickt" nicht „angekommen" heißt.** Der Stand wird
// unterschrieben und in den **Ausgang** gelegt; er bleibt dort liegen, bis
// das Gegenüber ihn bestätigt. *Übertragen* baut die Leitung und trägt
// hinaus, was dasteht
// (#1634) —
// **und räumt den Ausgang nicht**: das tut allein die Bestätigung von drüben.
//
// **Bis dahin konnte dieses Fenster einen Handel annehmen und keinen
// hinausbringen.** Ein Angebot lag für immer im Ausgang, und der, dem es
// galt, sah nie etwas.
//
// # Und warum die Klärung zuerst steht
//
// **Weil sie den Rest wertlos macht.** Stehen zwei Verläufe da, liefert
// niemand nach — das lösen zwei Menschen ([domain.Nachzuliefern]).
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

import type { Nachlieferung } from '../apihandelsleitung'
import { Aufklappkasten } from './Aufklappkasten'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { einsetzen } from '../sprache/einsetzen'
import { kennungseingabe } from './kennungseingabe'

const feldstil =
  'rounded-sm border border-kante bg-transparent px-2 py-1 font-mono ' +
  'text-xs focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

interface Handelsausgangeigenschaften {
  kennung: string

  // vorgang ist der gewählte Handel — leer heißt: noch keiner (UX-08.AK-04).
  vorgang: string
}

export function Handelsausgang({
  kennung,
  vorgang,
}: Handelsausgangeigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln
  const lauf = useVorgang()

  const nachrichtenfeld = useId()
  const ueberschrift = useId()

  const [nachricht, setzeNachricht] = useState('')
  const [fehler, setzeFehler] = useState<string>()
  const [gemeldet, setzeGemeldet] = useState<string>()

  // **Ein Satz und keine Kennung** — der Kasten daneben ist `font-mono`, und
  // dort gehört eine Nachrichtenkennung hinein, kein deutscher Satz.
  const [getragen, setzeGetragen] = useState<string>()
  const [offen, setzeOffen] = useState<Nachlieferung[]>()
  const [klaerung, setzeKlaerung] = useState(false)

  const kennzeichen = `handelsausgang/${kennung}`

  // fahre führt einen Ruf und räumt vorher, was er beantwortet.
  const fahre = (
    braucht: string,
    ohne: string,
    ruf: () => Promise<void>,
  ): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)
      setzeGemeldet(undefined)
      setzeGetragen(undefined)

      if (braucht.trim() === '') {
        setzeFehler(ohne)

        return
      }

      try {
        await ruf()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const schicken = (): Promise<void> =>
    fahre(vorgang, eigene.ausgangOhneVorgang, async () => {
      const antwort = await api().HandelSchicken(vorgang, kennung)

      setzeGemeldet(auspacken(antwort, () => antwort.nachrichtId))
    })

  // **Eine weggeräumte Nachricht verschwindet aus der Liste** — sie steht
  // nicht mehr aus (UX-08.AK-04: gewählt aus der Liste statt getippt).
  const bestaetigen = (welche: string): Promise<void> =>
    fahre(welche, eigene.ausgangOhneNachricht, async () => {
      const antwort = await api().HandelPostbestaetigung(welche)

      setzeGemeldet(auspacken(antwort, () => antwort.nachrichtId))
      setzeOffen((bisher) =>
        bisher?.filter((zeile) => zeile.nachrichtId !== welche),
      )
    })

  // **Getragen wird zum Gegenüber dieses Kastens**, nicht zu einem
  // getippten: die Adresse steht am Kontakt, und eine zweite im Feld wäre
  // die Gelegenheit, an eine fremde zu schicken.
  const zustellen = (): Promise<void> =>
    fahre(kennung, eigene.ausgangOhneGegenueber, async () => {
      const antwort = await api().HandelZustellen(kennung)

      const gelesen = auspacken(antwort, () => antwort)

      setzeGetragen(
        eigene.ausgangGetragen.replace('{zahl}', String(gelesen.getragen)),
      )
    })

  // **Der eigene Fortschritt wird gelesen und nicht getippt** — hier steht
  // nur, was das Gegenüber behauptet. Solange es keine Leitung gibt, ist das
  // ein Anfang: Revision 0 heißt „ich habe noch nichts bestätigt", und dann
  // geht alles ab der ersten Folge noch einmal hinaus.
  const wiederaufnehmen = (): Promise<void> =>
    fahre(vorgang, eigene.ausgangOhneVorgang, async () => {
      const antwort = await api().HandelWiederaufnahme(vorgang, kennung, {
        vorgang,
        revision: 0,
        abdruck: '',
        folge: 0,
      })

      const gelesen = auspacken(antwort, () => antwort)

      setzeOffen(gelesen.nachzuliefern)
      setzeKlaerung(gelesen.klaerung)
    })

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-2">
      <h3 id={ueberschrift} className="text-sm font-medium">
        {eigene.ausgangTitel}
      </h3>

      <Erklaerungsabsatz
        text={eigene.ausgangErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      <div className="flex flex-wrap gap-2">
        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft(kennzeichen)}
          aufKlick={() => {
            void schicken()
          }}
        >
          {eigene.ausgangSchicken}
        </Knopf>

        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft(kennzeichen)}
          aufKlick={() => {
            void zustellen()
          }}
        >
          {eigene.ausgangZustellen}
        </Knopf>

        <Knopf
          gewicht="leise"
          angehalten={lauf.laeuft(kennzeichen)}
          aufKlick={() => {
            void wiederaufnehmen()
          }}
        >
          {eigene.ausgangWiederaufnahme}
        </Knopf>
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {gemeldet !== undefined && (
        <p role="status" className="font-mono text-xs text-leise">
          {gemeldet}
        </p>
      )}

      {getragen !== undefined && (
        <p role="status" className="text-sm text-leise">
          {getragen}
        </p>
      )}

      {klaerung && (
        <p role="status" className="max-w-prose text-sm text-fehler">
          {eigene.ausgangKlaerung}
        </p>
      )}

      {offen !== undefined && !klaerung && (
        <div className="flex flex-col gap-1 text-sm text-leise">
          {/* **Angesagt wird der Satz, nicht die Knöpfe** — die Liste steht
              neben der Statuszeile und nicht in ihr. */}
          <p role="status">
            {offen.length === 0
              ? eigene.ausgangLeer
              : eigene.ausgangNachzuliefern}
          </p>

          <ul className="flex flex-col gap-1">
            {offen.map((zeile) => (
              <li
                key={zeile.nachrichtId}
                className="flex flex-wrap items-center gap-2"
              >
                <span className="font-mono text-xs">
                  {zeile.nachrichtId} · {eigene.ausgangFolge} {zeile.folge} ·{' '}
                  {zeile.art}
                </span>
                <Knopf
                  gewicht="leise"
                  bezeichnung={einsetzen(eigene.ausgangWegraeumenName, {
                    nachricht: zeile.nachrichtId,
                  })}
                  angehalten={lauf.laeuft(kennzeichen)}
                  aufKlick={() => {
                    void bestaetigen(zeile.nachrichtId)
                  }}
                >
                  {eigene.ausgangWegraeumen}
                </Knopf>
              </li>
            ))}
          </ul>
        </div>
      )}

      <Aufklappkasten titel={eigene.ausgangFachleute}>
        <div className="flex flex-col gap-2">
          <Erklaerungsabsatz
            text={eigene.ausgangFachleuteErklaerung}
            klasse="max-w-prose text-sm text-leise"
          />

          <div className="flex flex-wrap items-end gap-2">
            <label
              className="flex flex-col gap-1 text-sm"
              htmlFor={nachrichtenfeld}
            >
              {eigene.ausgangNachricht}
              <input
                {...kennungseingabe}
                id={nachrichtenfeld}
                type="text"
                className={feldstil}
                value={nachricht}
                onChange={(ereignis) => {
                  setzeNachricht(ereignis.target.value)
                }}
              />
            </label>

            <Knopf
              gewicht="leise"
              angehalten={lauf.laeuft(kennzeichen)}
              aufKlick={() => {
                void bestaetigen(nachricht)
              }}
            >
              {eigene.ausgangBestaetigen}
            </Knopf>
          </div>
        </div>
      </Aufklappkasten>
    </section>
  )
}
