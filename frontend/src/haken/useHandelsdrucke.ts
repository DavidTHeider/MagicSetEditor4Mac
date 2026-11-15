// **Welche Karte eine Handelszeile meint, über das Nachfragen hinweg** — P10
// (#2378).
//
// # Gemerkt je Druck und nicht je Stelle
//
// **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7, Regel 2):
// fragt der Anhalt nach einer neuen Zeile neu, stimmt die alte Antwort nur
// noch für die alten Zeilen — und wer sie nach der Stelle läse, schriebe
// nach dem Entfernen einer Zeile den Namen der entfernten an ihre Nachbarin.
// Gemerkt wird deshalb je Druck; was dabei steht, ist während des Fragens
// derselbe Name wie davor, statt zu flackern.
//
// # Und nach einer Absage
//
// **Bleibt kein Platzhalter stehen**: eine Zeile ohne Antwort gilt als
// unbekannt und nennt ihre Kennung (`Handelsdruckzeile`).
import { useState } from 'react'

import type {
  Handelsdruck,
  Handelsposition,
  Preisanhalt,
} from '../apihandelsleitung'

// unbekannt ist ein Druck, über den nichts bekannt ist.
const unbekannt: Handelsdruck = {
  name: '',
  setcode: '',
  nummer: '',
  bild: '',
  eigen: false,
}

// druckschluessel sagt, wann zwei Zeilen denselben Druck meinen — **die
// Druckzeile des Gegenübers nie**: sie ist eine seines Bestands.
function druckschluessel(position: Handelsposition, ich: string): string {
  if (position.printing !== '') {
    return `kennung:${position.printing}`
  }

  return position.von === ich
    ? `zeile:${String(position.druckId)}`
    : `fremd:${position.von}`
}

// useHandelsdrucke merkt sich je Druck, was ein Preisanhalt über ihn sagte.
//
// **Wie lange und wie viel** (Gesamtreview über 0.9.0, Befund 7): er lebt
// mit dem Handel — `Handelsraum.tsx` baut den Vorgang je Handel mit eigenem
// `key` — und hält höchstens die Drucke, die in diesem Handel vorkamen, je
// vier kurze Texte. Eine entfernte Zeile bleibt gemerkt, damit sie beim
// Wiederhinzufügen gleich ihren Namen trägt; eine Grenze braucht er erst,
// wenn ein Handel Tausende Drucke durchliefe.
export function useHandelsdrucke(
  positionen: Handelsposition[],
  anhalt: { daten: Preisanhalt | undefined; laedt: boolean },
  ich: string,
): (position: Handelsposition, stelle: number) => Handelsdruck | undefined {
  const [bekannt, setzeBekannt] = useState<ReadonlyMap<string, Handelsdruck>>(
    () => new Map(),
  )

  // **Die frische Antwort gilt sofort** — gemerkt wird für das nächste
  // Fragen, **beim Zeichnen** und nicht in einem Effekt: sonst stünde für
  // einen Rahmen der alte Stand da.
  const frisch = anhalt.laedt ? undefined : anhalt.daten
  const [gemerktAus, setzeGemerktAus] = useState<Preisanhalt>()
  if (frisch !== undefined && frisch !== gemerktAus) {
    setzeGemerktAus(frisch)
    setzeBekannt((vorher) => {
      const naechste = new Map(vorher)
      positionen.forEach((position, stelle) => {
        const druck = frisch.zeilen[stelle]?.druck
        if (druck !== undefined) {
          naechste.set(druckschluessel(position, ich), druck)
        }
      })

      return naechste
    })
  }

  return (position, stelle) => {
    const gemerkt = bekannt.get(druckschluessel(position, ich))
    if (anhalt.laedt) {
      return gemerkt
    }

    return frisch?.zeilen[stelle]?.druck ?? gemerkt ?? unbekannt
  }
}
