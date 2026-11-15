// **Nach einer neueren Fassung sehen, und sie holen** —
// #1100.
//
// # Was hier steht und was nicht
//
// **Ob überhaupt gefragt wird, entscheidet Go.** Die Einstellung `aus`
// bedeutet: keine Anfrage, und die Antwort sagt `gefragt: false`. Das Fenster
// trifft die Entscheidung nicht, es zeigt sie (§2). Dasselbe gilt für die
// Grenze zwischen `melden` und `holen`: ein Ladeversuch bei `melden` kommt mit
// einem Fehler zurück, statt hier abgefangen zu werden.
//
// # Einmal je Fassung
//
// Werkstück 3 des Issues verlangt es, und der Ort dafür ist der
// **Browserspeicher** und nicht der Bestand: es ist eine Eigenschaft dieses
// Fensters, keine des Kartenbestands. Wer eine Sicherung zurückspielt, will
// nicht, dass ein weggeklicktes Band wiederkommt.
import { useCallback, useEffect, useState } from 'react'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { ereignisLadefortschritt, laufzeit } from '../fenster'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

// weggelegtSchluessel ist der Ort der zuletzt weggeklickten Fassung.
const weggelegtSchluessel = 'mse.fassung.weggelegt'

export interface Aktualisierungsstand {
  // gefragt sagt, ob die Frage überhaupt hinausging.
  gefragt: boolean

  // fassung ist die neuere Fassung — leer, wenn es keine zu zeigen gibt.
  fassung: string
  seite: string

  // holen sagt, ob geladen werden darf — die Antwort aus Go, nicht eine
  // Entscheidung hier. Unter `melden` gibt es keinen Ladeknopf.
  holen: boolean

  // laeuft sagt, ob gerade geladen wird.
  laeuft: boolean

  // gelesen ist die Zahl der geladenen Bytes.
  gelesen: number

  // anteil ist der Anteil zwischen 0 und 1 — `undefined`, wenn die
  // Gegenstelle keine Länge nennt.
  //
  // **Kein erfundener Anteil** (§10): `ganz` ist dann −1, und ein Balken, der
  // daraus etwas rechnet, zeigt eine Zahl, für die es keine Deckung gibt.
  anteil: number | undefined

  // pfad ist der Ort der geprüften Datei — leer, solange keine da ist.
  pfad: string
  fehler: string

  laden: () => Promise<void>
  weglegen: () => void
}

export function useAktualisierung(): Aktualisierungsstand {
  const texte = useTexte()

  const [gefragt, setzeGefragt] = useState(false)
  const [fassung, setzeFassung] = useState('')
  const [seite, setzeSeite] = useState('')
  const [holen, setzeHolen] = useState(false)
  const [laeuft, setzeLaeuft] = useState(false)
  const [gelesen, setzeGelesen] = useState(0)
  const [anteil, setzeAnteil] = useState<number | undefined>(undefined)
  const [pfad, setzePfad] = useState('')
  const [fehler, setzeFehler] = useState('')

  // **Ein Vorgang je Zeit** (§7, Regel 4). Ein zweiter Knopfdruck, während der
  // erste läuft, holte 188 MB ein zweites Mal — und beide schrieben in
  // dieselbe Nachbardatei.
  const vorgang = useVorgang()

  useEffect(() => {
    let gilt = true

    // **In einer Zeile**, damit `schreibwege_test.go` den Namen lesen kann:
    // der Wächter liest den Methodennamen unmittelbar hinter der Fassade und
    // weist eine umgebrochene Kette ab, damit `docs/SCHREIBWEGE.md`
    // vollständig bleibt.
    const gefragt = api().NachNeuerFassungSehen()

    gefragt
      .then((antwort) => {
        if (!gilt) {
          return
        }

        setzeGefragt(true)

        // **Ein Fehler beim Fragen bleibt still.** Ohne Netz arbeitet die
        // Anwendung weiter (§16, Satz 5) — und ein Band, das beim Start
        // „GitHub antwortete mit 503" sagt, ist keine Hilfe, sondern Lärm.
        if (
          !antwort.neuer ||
          antwort.fassung === localStorage.getItem(weggelegtSchluessel)
        ) {
          return
        }

        setzeFassung(antwort.fassung)
        setzeSeite(antwort.seite)
        setzeHolen(antwort.holen)
      })
      .catch(() => {
        if (gilt) {
          setzeGefragt(true)
        }
      })

    return () => {
      gilt = false
    }
  }, [])

  // **Der Fortschritt hängt am Fenster und nicht am Ladevorgang.** Ein Horcher,
  // der erst beim Knopfdruck angemeldet würde, verpasste die ersten Blöcke —
  // `EventsOn` meldet nicht nach.
  useEffect(() => {
    const abmelden = laufzeit()?.EventsOn?.(
      ereignisLadefortschritt,
      (...daten) => {
        const [gelesenRoh, ganzRoh] = daten
        const geleseneZahl = Number(gelesenRoh)
        const ganze = Number(ganzRoh)

        setzeGelesen(geleseneZahl)
        setzeAnteil(ganze > 0 ? geleseneZahl / ganze : undefined)
      },
    )

    return abmelden
  }, [])

  const laden = useCallback(async () => {
    await vorgang.starte('fassung', async () => {
      setzeLaeuft(true)
      setzeFehler('')
      setzePfad('')

      // **Und der Fortschritt fängt wieder bei null an.**
      //
      // Ohne diese beiden Zeilen zeigte ein zweiter Versuch nach einem
      // Fehlschlag bis zum ersten neuen Ereignis die **alte** Menge — eine
      // Zahl aus einem Lauf, den es nicht mehr gibt (§10). Angestrichen vom
      // Gesamtreview vor 0.2.0.
      setzeGelesen(0)
      setzeAnteil(undefined)

      try {
        const antwort = await api().NeueFassungLaden()

        // **Über `auspacken` und nicht über `antwort.fehler` unmittelbar**:
        // `meldungAus` erwartet einen `Bindungsfehler` oder einen `Error`, und
        // ein rohes Fehlerobjekt fiele bei ihm auf `String(…)` durch.
        auspacken(antwort, () => undefined)

        setzePfad(antwort.pfad)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      } finally {
        setzeLaeuft(false)
      }
    })
  }, [texte, vorgang])

  const weglegen = useCallback(() => {
    if (fassung !== '') {
      localStorage.setItem(weggelegtSchluessel, fassung)
    }

    setzeFassung('')
  }, [fassung])

  return {
    gefragt,
    fassung,
    seite,
    holen,
    laeuft,
    gelesen,
    anteil,
    pfad,
    fehler,
    laden,
    weglegen,
  }
}
