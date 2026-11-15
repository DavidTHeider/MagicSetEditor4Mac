// **Wer am Zug ist, steht am Tisch** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Wofür es sie gibt
//
// **Weil G1 bis G8 einen Tisch gebaut haben und kein Spiel.** Jeder Sitz
// durfte jederzeit alles; auf die Frage, wer am Zug ist, gab dieses Programm
// keine Antwort.
//
// # Und warum sie verschwindet, wenn es keinen Zug gibt
//
// **Weil eine Partie ohne Zugbegriff keinen Takt hat.** Die Partien aus G1
// bis G8 sind Sandkästen und bleiben es (§15, Punkt 4); eine Leiste, die
// dort „Zug 0" zeigte, wäre eine Auskunft, die es nicht gibt.
import { useId } from 'react'

import { Kampfliste } from './Kampfliste'
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'
import { zugfortschritt } from '../spiel/zugfortschritt'

import type { Katalog } from '../sprache/kataloge'
import type { Partie } from '../apipartie'

interface Zugeigenschaften {
  partie: Partie

  // erlaubt ist, was der Kern gerade annimmt. **Mitten im Beginn einer
  // Partie** hat der Startspieler die Priorität des ersten Zugs schon und
  // darf trotzdem nicht passen (CR 103.8, Befund 2 des Zweitblicks über
  // Phase 6.9) — ein Knopf, den der Kern abwiese, stünde sonst da.
  erlaubt: readonly string[]

  // laeuft sagt, ob gerade ein Befehl unterwegs ist — §7, Regel 4.
  laeuft: boolean

  passen: () => void

  // abkuerzen reicht durch, bis etwas geschieht — CR 732 (M1, #1603).
  abkuerzen: () => void

  // zugende beendet den Zug vorzeitig — CR 724 (M1, #1603).
  zugende: () => void
}

export function Zugleiste({
  partie,
  erlaubt,
  laeuft,
  passen,
  abkuerzen,
  zugende,
}: Zugeigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen
  const abschnitt = useId()

  // **Ohne Zug keine Leiste** — der Grund steht im Kopf dieser Datei. **Und
  // nach dem Ende auch nicht** (P4, #2372): erlaubt ist dann nichts, und
  // „dran, aber nicht passen dürfen“ las sich als „der Beginn läuft“. Das
  // Ergebnis steht oben (`Partieausgang`).
  if (partie.zug.nummer === 0 || partie.stand === 'beendet') {
    return undefined
  }

  const dran = partie.zug.prioritaet === partie.fuer
  const kann = (art: string) => dran && erlaubt.includes(art)

  // **Dran und trotzdem nicht passen** heißt heute: der Beginn läuft
  // (`service.beginnHaelt`). Gefragt wird der Kern und nicht der Stand der
  // Sitze — eine zweite Fassung der Regel liefe irgendwann auseinander.
  const wartetAufBeginn = dran && !kann('zug.passen')
  const abwerfen = partie.zug.abwurf > 0 && partie.zug.aktiv === partie.fuer
  const stand = zugfortschritt(partie.zug.phase, partie.zug.schritt)

  return (
    // **Ein Kasten in der Spalte neben dem Tisch** — Artboard *Tisch*, seit
    // dem 2026-09-24 (Befund F09). Die Zeile des Zuges oben, darunter wer
    // dran ist, die Knöpfe unten; als Leiste über die volle Breite nahm sie
    // der Leinwand eine Zeile.
    <section
      className="flex flex-col items-start gap-2 rounded-sm border border-kante bg-flaeche px-3 py-2"
      aria-label={eigene.zugLeiste}
    >
      <p className="flex flex-wrap items-baseline gap-x-2">
        <span className="text-base font-semibold">
          {eigene.zug} {partie.zug.nummer}
        </span>

        <span id={abschnitt} className="text-sm text-leise">
          {phasenwort(partie.zug.phase, partie.zug.schritt, eigene)}
        </span>
      </p>

      {/* **Wie weit der Zug ist** — Artboard *Tisch*, ein Balken unter der
          Zeile. Sein Name ist der Abschnitt darüber; ein zweiter Name wäre
          eine zweite Aussage über dieselbe Sache (dieselbe Regel wie in
          `Fortschritt`). */}
      {stand !== undefined && (
        <div
          role="progressbar"
          aria-labelledby={abschnitt}
          aria-valuenow={stand.stelle}
          aria-valuemin={1}
          aria-valuemax={stand.von}
          className="h-1.5 w-full overflow-hidden rounded-sm bg-betont"
        >
          <div
            className="h-full bg-marke"
            style={{
              width: `${String(Math.round((stand.stelle / stand.von) * 100))}%`,
            }}
          />
        </div>
      )}

      <span className="text-sm text-leise">
        {eigene.zugAktiv}: {namenwort(partie, partie.zug.aktiv)}
      </span>

      {/* **Wer dran ist, steht da — und wer nicht, erfährt es auch.** §6
          verlangt die Erklärung dort, wo sie gebraucht wird: ein Knopf, der
          grau ist und schweigt, lässt raten. */}
      <span role="status" className="text-sm">
        {wartetAufBeginn
          ? eigene.zugWartetAufBeginn
          : dran
            ? eigene.zugDuBistDran
            : wartewort(partie, eigene)}
      </span>

      <Kampfliste partie={partie} />

      {abwerfen && (
        <span role="status" className="text-sm text-fehler">
          {eigene.zugAbwurf}: {partie.zug.abwurf}
        </span>
      )}

      {/* **Drei Knöpfe mit Rand und keiner laut** — das Artboard setzt „Zug
          beenden" als Hauptknopf. Hier beendet er den Zug **vorzeitig**
          (CR 724.1, siehe unten): was auf dem Stapel liegt, geht ins Exil.
          Der lauteste Knopf einer Leiste ist der, den man ohne Nachdenken
          drückt, und das soll dieser nicht sein. */}
      <div className="flex flex-wrap gap-2">
        <Knopf angehalten={laeuft || !kann('zug.passen')} aufKlick={passen}>
          {eigene.zugPassen}
        </Knopf>

        {/* **Zehn Abschnitte vergeben Priorität** (CR 502.4 und CR 514.3 nicht),
          bei zwei Sitzen sind das zwanzig Passen je Zug. CR 732.1 sagt, dass
          am Tisch niemand so spielt — dieser Knopf ist die Abkürzung, und er
          hält an, sobald etwas geschieht. */}
        <Knopf
          angehalten={laeuft || !kann('zug.abkuerzen')}
          aufKlick={abkuerzen}
        >
          {eigene.zugAbkuerzen}
        </Knopf>

        {/* **Den Zug vorzeitig beenden** — CR 724.1: alles auf dem Stapel wird
          exiliert, und das Spiel springt in den Aufräumschritt. Im Regelwerk
          tut das eine Karte; hier sagt es ein Mensch, wie beim Lebensstand. */}
        <Knopf angehalten={laeuft || !kann('zug.beenden')} aufKlick={zugende}>
          {eigene.zugBeenden}
        </Knopf>
      </div>
    </section>
  )
}

// phasenwort nennt Phase und Schritt in Worten.
//
// **Aus dem Katalog und nicht aus der Kennung**: `hauptVorKampf` ist ein
// Bezeichner und kein Satz, und eine englische Sitzung bekäme ihn sonst
// unübersetzt (§9).
function phasenwort(
  phase: string,
  schritt: string,
  eigene: Katalog['seiten']['spielen'],
): string {
  const phasen: Record<string, string> = {
    anfang: eigene.phaseAnfang,
    hauptVorKampf: eigene.phaseHauptVorKampf,
    kampf: eigene.phaseKampf,
    hauptNachKampf: eigene.phaseHauptNachKampf,
    ende: eigene.phaseEnde,
  }

  const schritte: Record<string, string> = {
    enttappen: eigene.schrittEnttappen,
    versorgung: eigene.schrittVersorgung,
    ziehen: eigene.schrittZiehen,
    kampfbeginn: eigene.schrittKampfbeginn,
    angreifer: eigene.schrittAngreifer,
    blocker: eigene.schrittBlocker,
    kampfschaden: eigene.schrittKampfschaden,
    kampfende: eigene.schrittKampfende,
    endeschritt: eigene.schrittEnde,
    aufraeumen: eigene.schrittAufraeumen,
  }

  const benannt = phasen[phase] ?? phase

  // **Eine Hauptphase hat keinen Schritt** — CR 500.1 zerlegt nur Anfang,
  // Kampf und Ende.
  return schritt === ''
    ? benannt
    : `${benannt} · ${schritte[schritt] ?? schritt}`
}

// namenwort gibt den Namen eines Sitzes — oder seine Kennung.
//
// **Die Kennung als Rückfall**: ein Sitz, den die Sicht nicht führt, ist kein
// Grund für eine leere Stelle (§8).
function namenwort(partie: Partie, kennung: string): string {
  const sitz = partie.sitze.find((einer) => einer.kennung === kennung)

  return sitz === undefined || sitz.name === '' ? kennung : sitz.name
}

// wartewort sagt, auf wen gewartet wird.
//
// **Und „niemand" ist eine eigene Auskunft**: im Enttappen und im Aufräumen
// bekommt niemand Priorität (CR 502.4, CR 514.3) — wer dort „du wartest auf
// …" läse, suchte jemanden, den es nicht gibt.
function wartewort(
  partie: Partie,
  eigene: Katalog['seiten']['spielen'],
): string {
  return partie.zug.prioritaet === ''
    ? eigene.zugNiemandDran
    : `${eigene.zugWartetAuf} ${namenwort(partie, partie.zug.prioritaet)}`
}
