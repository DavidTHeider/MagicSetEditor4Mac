// **Wie die Leitung steht** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Was hier steht und was nicht
//
// **Route, Gegenüber, Standzeit, Warteschlange, Stand im Strom, letzte
// Runde** — und die **Bildzeit**, die als einzige hier entsteht statt aus Go
// zu kommen. Was **nicht** darinsteht, ist jede Zahl über Karten: weder der
// Abdruck des Decks noch der des Stroms. Beide sind Hashes über Karten, und
// gegen einen Hash lässt sich raten (#1462: „Geteilter Eventhash ermöglicht
// keine simple Wörterbuchsuche auf einzelne geheime Karten").
//
// **Fehlende Messgrößen erscheinen als „unbekannt"** und nicht als Null. Das
// steht so in der Abnahme des Issues, und es ist derselbe Satz wie in §8: ein
// Wert, der zwei Dinge bedeutet, wird irgendwo addiert.
import { useId } from 'react'

import { usePartiediagnose } from '../haken/usePartiediagnose'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { Leitungszeile, PartiediagnoseAntwort } from '../apinebenbei'
import { Abschnittstitel } from './Abschnittstitel'

export function Tischdiagnose() {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const lauf = usePartiediagnose()
  const ueberschrift = useId()

  return (
    <section className="flex flex-col gap-2" aria-labelledby={ueberschrift}>
      <div className="flex items-center gap-3">
        <Abschnittstitel id={ueberschrift} className="text-base font-semibold">
          {eigene.diagnose}
        </Abschnittstitel>

        <Knopf gewicht="leise" angehalten={lauf.laeuft} aufKlick={lauf.ansehen}>
          {eigene.diagnoseAnsehen}
        </Knopf>
      </div>

      <Erklaerungsabsatz
        text={eigene.diagnoseErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {lauf.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {lauf.fehler}
        </p>
      )}

      {lauf.stand !== undefined && <Diagnosestand stand={lauf.stand} />}

      <p className="text-sm">
        {eigene.diagnoseBildzeit}:{' '}
        {lauf.bildmillis === undefined
          ? eigene.diagnoseUnbekannt
          : `${String(Math.round(lauf.bildmillis))} ms`}
      </p>
    </section>
  )
}

// Diagnosestand zeigt, was die Go-Seite gemeldet hat.
function Diagnosestand({ stand }: { stand: PartiediagnoseAntwort }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  if (!stand.laeuft) {
    return <p className="text-sm text-leise">{eigene.diagnoseStill}</p>
  }

  return (
    <>
      <p className="text-sm">
        {stand.rolle === 'wirt'
          ? eigene.diagnoseAlsWirt
          : eigene.diagnoseAlsGast}
      </p>

      {stand.pausiert && (
        <p role="status" className="text-sm text-leise">
          {eigene.diagnosePausiert}
        </p>
      )}

      {stand.standBekannt && (
        <p className="text-sm">
          {eigene.diagnoseStand}: {stand.stand}
        </p>
      )}

      <p className="text-sm">
        {eigene.diagnoseRunde}:{' '}
        {stand.rueckkehrBekannt
          ? `${String(stand.rueckkehrMillis)} ms`
          : eigene.diagnoseUnbekannt}
      </p>

      <Leitungsliste leitungen={stand.leitungen} />
    </>
  )
}

// Leitungsliste zeigt eine Zeile je Leitung.
function Leitungsliste({ leitungen }: { leitungen: Leitungszeile[] }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  if (leitungen.length === 0) {
    return <p className="text-sm text-leise">{eigene.diagnoseOhneLeitung}</p>
  }

  return (
    <ul className="flex flex-col gap-1 text-sm">
      {leitungen.map((zeile) => (
        <li key={`${zeile.sitz}-${zeile.gegenueber}`}>
          <Leitungsangaben zeile={zeile} />
        </li>
      ))}
    </ul>
  )
}

// Leitungsangaben nennt die Werte einer Leitung.
//
// **Der Sitz steht nur da, wenn es einen gibt**: als Gast ist das andere Ende
// der Wirt, und „Sitz: (leer)" wäre eine Angabe über nichts.
function Leitungsangaben({ zeile }: { zeile: Leitungszeile }) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  return (
    <span className="flex flex-wrap gap-3">
      {zeile.sitz !== '' && <span>{zeile.sitz}</span>}

      <span>
        {eigene.diagnoseWeg}: {zeile.weg}
      </span>

      {zeile.gegenueber !== '' && (
        <span className="font-mono text-xs">
          {eigene.diagnoseGegenueber}: {zeile.gegenueber}
        </span>
      )}

      <span>
        {eigene.diagnoseSeit}: {zeile.seitSekunden} s
      </span>

      <span>
        {eigene.diagnoseWartend}: {zeile.wartend}
      </span>

      <span>
        {eigene.diagnoseVerworfen}: {zeile.verworfen}
      </span>
    </span>
  )
}
