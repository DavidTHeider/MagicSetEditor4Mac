// **Was jeder Spieler am Tisch hat** — Werkstück M3 der Regelmaschine
// (#1605).
//
// # Wofür es sie gibt
//
// **Weil die Zahlen dastanden und niemand sie sah.** `leben` und
// `vergiftung` liegen seit G9 am Modell
// (#1592),
// und seit `M17` (#1618)
// scheidet ein Sitz bei null Leben oder zehn Marken aus. Gemessen am
// 2026-09-17: im ganzen Fenster kam **keine der beiden Zahlen** vor — die
// einzige Stelle mit dem Wort „Leben" war der Grund **nach** dem
// Ausscheiden. Eine Regel greift, und niemand kann sehen, worauf (§10,
// erste Dimension).
//
// **Der Manavorrat kommt mit `M3` dazu** und gehört an dieselbe Stelle: er
// ist Spielerzustand wie die anderen beiden, und CR 106.4a/b verlangen, dass
// ein übriger Vorrat angesagt wird — das geht nur, wenn er dasteht.
//
// # Und warum neben der Zugleiste und nicht darin
//
// **Weil die Zugleiste den Takt zeigt und diese den Stand.** Die eine
// verschwindet, wo es keinen Zug gibt (ein Sandkasten aus G1 bis G8); Leben
// und Mana gibt es dort trotzdem.
import { einsetzen } from '../sprache/einsetzen'
import { Lebensregler } from './Lebensregler'
import { useTexte } from '../sprache/kontext'

import type { Partie, Sitz } from '../apipartie'

// Commandertod ist der Schaden, ab dem die Partie vorbei ist — CR 903.10a
// (Fassung 2026-08-19).
//
// **Eine benannte Zahl und kein `>= 21` im JSX**: sie steht einmal da, mit
// ihrer Regel daneben, statt an jeder Fundstelle neu.
const Commandertod = 21

export interface Spielerleisteneigenschaften {
  partie: Partie

  // aendereLeben ändert das Leben des eigenen Sitzes — P4 (#2372), C-20.
  // Fehlt es, steht der Stand nur da: wo der Kern `spieler.leben` nicht
  // erlaubt, gibt es nichts zu drücken.
  aendereLeben?: (um: number) => void
}

export function Spielerleiste({
  partie,
  aendereLeben,
}: Spielerleisteneigenschaften) {
  const eigene = useTexte().seiten.spielen

  return (
    <section
      className="flex flex-wrap items-center gap-4 rounded-sm border border-kante px-3 py-2"
      aria-label={eigene.standLeiste}
    >
      {partie.sitze.map((sitz) => (
        <Sitzstand
          key={sitz.kennung}
          sitz={sitz}
          eigen={sitz.kennung === partie.fuer}
          benenne={(kennung) => benenneIn(partie, kennung)}
          {...(sitz.kennung === partie.fuer && aendereLeben !== undefined
            ? { aendereLeben }
            : {})}
        />
      ))}
    </section>
  )
}

// benenneIn nennt ein Objekt beim Namen — **soweit es einen zeigt**.
//
// **Drei Fälle und kein `||`** (§8): das Objekt liegt am Tisch und trägt
// einen Namen; es liegt da und zeigt keinen, weil es verdeckt ist oder die
// Sicht ihn nicht hergibt; oder es liegt gar nicht mehr da. In den letzten
// beiden bleibt die Kennung — schlechter als ein Name und besser als eine
// leere Stelle.
function benenneIn(partie: Partie, kennung: string): string {
  const gefunden = partie.objekte.find((eines) => eines.kennung === kennung)

  if (gefunden === undefined || gefunden.name === '') {
    return kennung
  }

  return gefunden.name
}

// Sitzstand gibt einen Sitz mit dem, was an ihm zählt.
//
// **Leben steht immer da, die anderen beiden nur, wo es sie gibt.** Eine
// Null ist keine Auskunft: die meisten Partien haben nie eine
// Vergiftungsmarke, und „Gift: 0" an jedem Sitz wäre Lärm, in dem die eine
// Marke untergeht, auf die es ankommt.
function Sitzstand({
  sitz,
  eigen,
  benenne,
  aendereLeben,
}: {
  sitz: Sitz
  eigen: boolean
  benenne: (kennung: string) => string
  aendereLeben?: (um: number) => void
}) {
  const eigene = useTexte().seiten.spielen

  return (
    <span
      className="flex items-baseline gap-2 text-sm"
      // **Der eigene Sitz ist benannt und nicht nur gefärbt.** Zwei Sitze
      // dürfen denselben Namen tragen — gegen sich selbst zu spielen ist seit
      // G8 der Normalfall —, und eine Farbe allein sagt einem
      // Vorleseprogramm nichts.
      aria-label={eigen ? eigene.standDu : undefined}
    >
      <span className={eigen ? 'font-medium' : 'text-leise'}>{sitz.name}</span>

      {aendereLeben === undefined ? (
        <span aria-label={eigene.standLeben}>{sitz.leben}</span>
      ) : (
        <Lebensregler wert={sitz.leben} aendere={aendereLeben} />
      )}

      {sitz.vergiftung > 0 && (
        <span className="text-fehler">
          {eigene.standGift}: {sitz.vergiftung}
        </span>
      )}

      {sitz.manavorrat !== '' && (
        <span className="text-leise">
          {eigene.standMana}: {sitz.manavorrat}
        </span>
      )}

      {/* **Die fünf Zahlen aus PLAY-03**
          (#1711).
          Alle nach derselben Regel wie das Gift darüber: **eine Null ist
          keine Auskunft**. */}
      {sitz.strahlung > 0 && (
        <span className="text-fehler">
          {eigene.standStrahlung}: {sitz.strahlung}
        </span>
      )}

      {sitz.commandersteuer > 0 && (
        <span className="text-leise">
          {eigene.standSteuer}: {sitz.commandersteuer}
        </span>
      )}

      <Commanderschaden schaden={sitz.commanderschaden} benenne={benenne} />

      {sitz.leergezogen && (
        <span className="text-fehler">{eigene.standLeergezogen}</span>
      )}

      {/* **Die Mulligans** (#1839) — am Tisch erklärt (CR 103.5), und nach
          derselben Regel: eine Null ist keine Auskunft. */}
      {sitz.mulligans > 0 && (
        <span className="text-leise">
          {eigene.standMulligans}: {sitz.mulligans}
        </span>
      )}
    </span>
  )
}

// Commanderschaden gibt den Kampfschaden je Commander — CR 903.10a.
//
// **Je Kennung und nicht als Summe**: zwei mit je zwanzig sind vierzig und
// trotzdem keine einundzwanzig. Wer sie addierte, nähme einen Spieler aus
// einer Partie, in der er noch sitzt.
//
// **Und die Reihenfolge ist die der Tabelle.** Sie zu sortieren hieße, eine
// Rangfolge zu behaupten, die es nicht gibt — gefährlich ist der höchste
// Wert, und der steht ohnehin hervorgehoben da.
function Commanderschaden({
  schaden,
  benenne,
}: {
  schaden: Record<string, number>
  benenne: (kennung: string) => string
}) {
  const eigene = useTexte().seiten.spielen

  return (
    <>
      {Object.entries(schaden).map(([wessen, wieviel]) => (
        <span
          key={wessen}
          className={
            wieviel >= Commandertod ? 'font-medium text-fehler' : 'text-leise'
          }
        >
          {einsetzen(eigene.standCommanderschaden, {
            name: benenne(wessen),
            schaden: String(wieviel),
          })}

          {/* **Der Satz statt nur der Farbe** (§6): eine Hervorhebung allein
              sagt einem Vorleseprogramm nichts, und sie sagt auch einem
              Sehenden nicht, **warum**. */}
          {wieviel >= Commandertod && <> {eigene.standCommanderTod}</>}
        </span>
      ))}
    </>
  )
}
