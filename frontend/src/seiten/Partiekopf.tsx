// **Was über dem Tisch steht** — herausgelöst am 2026-09-21, als
// `Partietisch.tsx` mit der Zuschauersicht (PLAY-05.AK-02,
// #1713)
// über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Nebenan steht, **was auf dem Tisch liegt** — Zonen, Karten, Befehle. Hier
// steht, **was über die Partie zu wissen ist, bevor man hinsieht**: ihr
// Stand, wessen Sicht man gerade hat, und die zwei Wege hinaus.
//
// # Eine Zeile und kein Block
//
// **Der Tisch steht nach dem Anfangen im Blick** — Befund F09 der nativen
// Prüfung von 0.5.0 (UX-03, UX-08; Artboard *Tisch*). Die Erklärungen zum
// Wirt und zum Aufgeben stehen deshalb hinter einem Fragezeichen bei dem, was
// sie erklären, und der Schalter der Darstellung steht hier, wo der Entwurf
// ihn zeichnet.
import { useEffect, useRef, useState } from 'react'

import { Entscheidung } from '../bausteine/Entscheidung'
import { Feldhilfe } from '../bausteine/Feldhilfe'
import { Knopf } from '../bausteine/Knopf'
import { sitzname, standwort } from './partiewoerter'
import { sitzwort } from './sitzwort'
import { useTexte } from '../sprache/kontext'

import type { Partie } from '../apipartie'
import type { Partielauf } from '../haken/usePartie'

export function Partiekopf({
  partie,
  lauf,
  ueberschrift,
  profilnamen,
  setzeGewaehlt,
  alsListe,
  umschalten,
  offeneHaende,
  setzeOffeneHaende,
  ebene = 2,
  fremd = false,
}: {
  partie: Partie
  lauf: Partielauf

  // fremd heißt: der Tisch eines Wirts, an dem dieses Fenster Gast ist.
  fremd?: boolean

  // ueberschrift ist die Kennung, auf die der Abschnitt daneben zeigt.
  ueberschrift: string

  profilnamen: Map<string, string>

  // setzeGewaehlt räumt die offene Wahl, wenn die Sicht wechselt (§7,
  // Regel 1): eine Karte, die in der neuen Sicht geheim ist, bliebe sonst
  // als Auswahl stehen.
  setzeGewaehlt: (kennung: string | undefined) => void

  // alsListe sagt, welche Darstellung gilt; umschalten wechselt sie.
  alsListe: boolean
  umschalten: () => void

  // offeneHaende heißt: ohne Vorhang am gemeinsamen Gerät (P7, #2375).
  offeneHaende: boolean
  setzeOffeneHaende: (offen: boolean) => void

  // ebene ist die Stufe der Überschrift. **Eins, wo der Tisch die Seite
  // ist** — dann bekommt er auch den Fokus; zwei, wo er in einer Seite
  // steht, etwa als Tisch eines Gastes unter dem eigenen Anfang.
  ebene?: 1 | 2
}) {
  const eigene = useTexte().seiten.spielen
  const Ueberschrift = ebene === 1 ? 'h1' : 'h2'
  const kopf = useRef<HTMLHeadingElement>(null)
  const [fragtAufgeben, setzeFragtAufgeben] = useState(false)

  // **Der Fokus geht an den neuen Arbeitsplatz** — F09 verlangt es
  // wörtlich: „Fokus an neuem Arbeitskontext". Wer eben „Partie anfangen"
  // gedrückt hat, stand mit dem Fokus auf einem Knopf, den es nicht mehr
  // gibt; eine Vorlesehilfe sagte dann nichts, und das Fenster blieb dort
  // gescrollt, wo der Knopf war.
  //
  // **Erst nach oben, dann der Fokus ohne Rollen.** Die Leinwand ist so hoch,
  // wie das Fenster unter dem Kopf ist, wenn die Seite oben steht
  // (`useFensterrest`); rollte der Fokus nur so weit, bis die Überschrift
  // zu sehen ist, stünde die Steuerleiste darunter halb außerhalb.
  useEffect(() => {
    if (ebene !== 1) return

    kopf.current?.ownerDocument.body.scrollIntoView({ block: 'start' })
    kopf.current?.focus({ preventScroll: true })
  }, [ebene])

  return (
    <header className="flex flex-wrap items-center gap-x-6 gap-y-2">
      <div className="flex min-w-0 flex-col gap-1">
        <Ueberschrift
          id={ueberschrift}
          ref={kopf}
          tabIndex={-1}
          className={
            ebene === 1
              ? 'text-2xl font-semibold tracking-tight focus:outline-none'
              : 'text-lg font-semibold focus:outline-none'
          }
        >
          {eigene.tischTitel}
        </Ueberschrift>

        {/* **Der Wirt wird genannt, und was das heißt, steht beim Namen**
            — das Issue verlangt es vor dem Mitspielen (G1, #1458). Ein
            Absatz darunter kostete jedem Tisch zwei Zeilen Leinwand. */}
        <p className="flex items-center gap-1 text-sm text-leise">
          <span>
            {eigene.stand}: {standwort(eigene, partie.stand)} · {eigene.fassung}{' '}
            {partie.fassung} · {eigene.wirt}: {sitzname(partie, partie.wirt)}
          </span>
          <Feldhilfe zu={eigene.wirt} text={eigene.wirtHinweis} />
        </p>
      </div>

      {/* **Die Sichtwahl bleibt, und das Profil wählt darin** (G8,
          #1465):
          ein Sitz mit Profil trägt seinen Namen dabei, einer ohne bleibt
          über dieselbe Liste erreichbar. **Die Sicht bleibt ein Parameter**
          — abgeleitet wird sie nicht, sonst bekäme ein Zuschauer die Sicht
          eines Sitzes, sobald beide dasselbe Profil benutzen. */}
      <div className="ml-auto flex flex-wrap items-center gap-3">
        {/* **Am fremden Tisch keine Sichtwahl** (S12.NAT-04): ein Gast sitzt
            an seinem Sitz, und `ansehen` tut dort nichts. */}
        {!fremd && (
          <label className="flex items-center gap-2 text-sm">
            <span className="font-medium">{eigene.sicht}</span>
            <select
              value={partie.fuer}
              onChange={(ereignis) => {
                setzeGewaehlt(undefined)
                lauf.ansehen(partie.id, ereignis.target.value)
              }}
              className="rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              {/* **Der Zuschauer steht oben und trägt die leere
              Sitzkennung** — PLAY-05.AK-02
              (#1713).

              Die Domäne kennt ihn seit G1: `Partie.Sicht("")` gibt die
              Sicht eines Betrachters **ohne Sitz**. Ansehen konnte sie
              niemand — diese Liste führte nur Sitze.

              **Oben, weil sie die engste ist**: wer sie wählt, sieht
              sicher nichts Fremdes, und das ist der Stand, in dem man
              einen Tisch zeigt. */}
              <option value="">{eigene.sichtZuschauer}</option>

              {partie.sitze.map((sitz) => (
                <option key={sitz.kennung} value={sitz.kennung}>
                  {sitzwort(sitz, profilnamen)}
                </option>
              ))}
            </select>
          </label>
        )}

        {/* **Offene Hände sind eine ausdrückliche Wahl** — P7 (#2375): ohne
            sie geht zwischen zwei Sitzen am selben Gerät der Vorhang zu. Am
            fremden Tisch gibt es keinen. */}
        {!fremd && (
          <span className="flex items-center gap-1">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={offeneHaende}
                onChange={(ereignis) => {
                  setzeOffeneHaende(ereignis.target.checked)
                }}
                className="size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              />
              <span className="font-medium">{eigene.offeneHaende}</span>
            </label>
            <Feldhilfe
              zu={eigene.offeneHaende}
              text={eigene.offeneHaendeHinweis}
            />
          </span>
        )}

        {/* **Aufgeben geht jederzeit** — CR 104.3a (M17, #1618), und es ist
          etwas anderes als „Partie schließen": das schließt das Fenster,
          dies beendet das Spiel. Deshalb stehen beide da und heißen
          verschieden. */}
        <Knopf aufKlick={umschalten}>
          {alsListe ? eigene.tischGrafisch : eigene.tischListe}
        </Knopf>

        <span className="flex items-center gap-1">
          {/* **Erst eine Rückfrage** — P1b (#2370), C-06: ein Einzelklick
              beendete das Spiel, sofort und endgültig. An einem beendeten
              Spiel gibt es nichts mehr aufzugeben. */}
          <Knopf
            gewicht="leise"
            angehalten={fragtAufgeben || partie.stand === 'beendet'}
            aufKlick={() => {
              setzeFragtAufgeben(true)
            }}
          >
            {eigene.aufgeben}
          </Knopf>

          {/* **Was Aufgeben bedeutet, steht dort, wo es geschieht** (§6) —
            es ist sofort endgültig, und die Karten gehen mit (CR 104.3a,
            CR 800.4a). */}
          <Feldhilfe zu={eigene.aufgeben} text={eigene.aufgebenHinweis} />
        </span>

        <Knopf gewicht="leise" aufKlick={lauf.raeumen}>
          {eigene.verlassen}
        </Knopf>
      </div>

      {/* **Endet die Partie, während die Frage steht, geht sie mit** — „Ja,
          aufgeben“ schickte den Befehl sonst an ein beendetes Spiel. */}
      {fragtAufgeben && partie.stand !== 'beendet' && (
        <Entscheidung
          frage={eigene.aufgebenFrage}
          bestaetigen={eigene.aufgebenJa}
          abbrechen={eigene.aufgebenNein}
          aufBestaetigen={() => {
            setzeFragtAufgeben(false)
            lauf.befehlen('zug.aufgeben', {})
          }}
          aufAbbrechen={() => {
            setzeFragtAufgeben(false)
          }}
        />
      )}
    </header>
  )
}
