// **Suchen und Ersetzen in einem Textfeld** — Werkstück 1 von Phase 3.3
// (#1164).
//
// # Warum groß und klein zählen
//
// In MSE-Script sind `card` und `Card` zweierlei. Ein Ersetzen, das darüber
// hinwegsähe, änderte still etwas anderes als gemeint — und still ist die
// schlimmste Art, etwas Falsches zu tun.
//
// # Warum es Text sucht und keine Muster
//
// Ein regulärer Ausdruck ist mächtiger und teurer: wer `card.name` eingibt,
// meint den Punkt als Punkt. MSE2 hält es ebenso — sein `find/replace` sucht
// Text (`src/gui/set/cards_panel.cpp`, Fassung f9d9356 vom 2021-07-06).
//
// # Was es nicht kann
//
// **Rückgängig als eigenen Schritt.** Ein Ersetzen ändert den Wert des Feldes;
// was daraus im Verlauf wird, entscheidet der Aufrufer. Im Vorlagenbereich ist
// das der Speicherstand — und der vorige liegt als `.bak` neben dem Paket.
import { useEffect, useRef, useState } from 'react'

import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface TextsucheEigenschaften {
  // wert ist der Text, in dem gesucht wird.
  wert: string

  setzeWert: (wert: string) => void

  // zeigeTreffer hebt einen Fund im Feld hervor.
  zeigeTreffer: (von: number, bis: number) => void

  schliessen: () => void
}

export function Textsuche({
  wert,
  setzeWert,
  zeigeTreffer,
  schliessen,
}: TextsucheEigenschaften) {
  const texte = useTexte()
  const eigene = texte.bausteine.suche

  const feld = useRef<HTMLInputElement>(null)

  // **Der Fokus geht ins Suchfeld** — wer ⌘F drückt, will tippen.
  //
  // **Über einen Effekt und nicht über `autoFocus`.** Das Attribut springt
  // auch dann, wenn eine Seite von selbst neu gezeichnet wird; `jsx-a11y`
  // weist es deshalb ab. Dieselbe Gestalt wie in `Palette.tsx`.
  useEffect(() => {
    feld.current?.focus()
  }, [])

  // **Escape horcht am Fenster, nicht am Kasten.** An einem `div` wäre es ein
  // Ereignishorcher an einem nicht bedienbaren Element — auch das dieselbe
  // Überlegung wie bei der Befehlspalette.
  useEffect(() => {
    const aufTaste = (ereignis: KeyboardEvent) => {
      if (ereignis.key === 'Escape') {
        // **Verbraucht** — die Konsole steht in der Schublade *Werkzeuge*,
        // und dasselbe Escape schlösse sie sonst mit (UX-08, #1680).
        ereignis.preventDefault()
        schliessen()
      }
    }

    window.addEventListener('keydown', aufTaste)

    return () => {
      window.removeEventListener('keydown', aufTaste)
    }
  }, [schliessen])

  const [gesucht, setzeGesucht] = useState('')
  const [ersatz, setzeErsatz] = useState('')
  const [bei, setzeBei] = useState(0)

  const stellen = fundstellen(wert, gesucht)

  // **Der Zeiger bleibt in der Liste**, auch wenn sie kürzer wird: wer den
  // dritten Treffer angesehen hat und dann einen Buchstaben tippt, hätte sonst
  // einen Zeiger auf etwas, das es nicht mehr gibt.
  //
  // **Ein Zweig und nicht zwei.** `bei % 0` ist `NaN`, und `stellen[NaN]` ist
  // `undefined`; der Rückfall ist damit der Fall „kein Treffer" und kein
  // toter Zweig, den kein Lauf betritt (§3).
  const stelle = stellen[bei % stellen.length] ?? -1

  useEffect(() => {
    if (stelle >= 0) {
      zeigeTreffer(stelle, stelle + gesucht.length)
    }
  }, [stelle, gesucht.length, zeigeTreffer])

  return (
    <div className="order-2 flex flex-wrap items-end gap-2 rounded-sm border border-kante bg-betont p-2">
      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">{eigene.suchen}</span>
        <input
          ref={feld}
          type="search"
          value={gesucht}
          onChange={(ereignis) => {
            setzeGesucht(ereignis.target.value)
            setzeBei(0)
          }}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
      </label>

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">{eigene.ersetzen}</span>
        <input
          type="text"
          value={ersatz}
          onChange={(ereignis) => {
            setzeErsatz(ereignis.target.value)
          }}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
      </label>

      {/* **Als `status`** (UX-10.AK-02, #1682): wer tippt, hört die Zahl
          der Treffer, ohne den Blick vom Feld zu nehmen. */}
      <p role="status" className="text-sm text-leise">
        {stellen.length === 0
          ? eigene.keinTreffer
          : einsetzen(eigene.treffer, {
              bei: String((bei % stellen.length) + 1),
              ganz: String(stellen.length),
            })}
      </p>

      <Knopf
        bezeichnung={eigene.zurueck}
        aufKlick={() => {
          setzeBei((bisher) => bisher + Math.max(stellen.length - 1, 0))
        }}
      >
        ↑
      </Knopf>

      <Knopf
        bezeichnung={eigene.weiter}
        aufKlick={() => {
          setzeBei((bisher) => bisher + 1)
        }}
      >
        ↓
      </Knopf>

      <Knopf
        bezeichnung={eigene.einsetzen}
        aufKlick={() => {
          if (stelle >= 0) {
            setzeWert(
              wert.slice(0, stelle) +
                ersatz +
                wert.slice(stelle + gesucht.length),
            )
          }
        }}
      >
        {eigene.einsetzen}
      </Knopf>

      <Knopf
        bezeichnung={eigene.alleEinsetzen}
        aufKlick={() => {
          if (stellen.length > 0) {
            setzeWert(wert.split(gesucht).join(ersatz))
          }
        }}
      >
        {eigene.alleEinsetzen}
      </Knopf>

      <Knopf bezeichnung={eigene.schliessen} aufKlick={schliessen}>
        ✕
      </Knopf>

      <p className="w-full text-sm text-leise">{eigene.grossKlein}</p>
    </div>
  )
}

// fundstellen gibt die Anfänge jedes Fundes.
//
// **Ohne Überlappung**: `aa` in `aaa` ist einmal da und nicht zweimal — sonst
// änderte „alle ersetzen" mehr Zeichen, als es gefunden hat.
function fundstellen(wert: string, gesucht: string): number[] {
  if (gesucht === '') {
    return []
  }

  const stellen: number[] = []

  let bei = wert.indexOf(gesucht)

  while (bei >= 0) {
    stellen.push(bei)
    bei = wert.indexOf(gesucht, bei + gesucht.length)
  }

  return stellen
}
