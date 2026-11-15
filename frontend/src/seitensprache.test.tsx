// **Jede Seite spricht wirklich Englisch, wenn die Sitzung es tut** (§9).
//
// # Was das über die anderen Wächter hinaus prüft
//
// `keintextimjsx.test.ts` weist einen Text **im Quelltext** ab, und
// `sprache.test.tsx` hält jeden Katalogschlüssel in beiden Sprachen. Beide
// ließen eine Seite durch, die `kataloge.de` unmittelbar liest statt
// `useTexte()` — kein Literal, kein fehlender Schlüssel, und trotzdem deutscher
// Text in einer englischen Sitzung.
//
// # Warum es diese Datei gibt
//
// Das Sammelissue zu Phase O führte den Kasten „die englische Sitzung
// (`englischesitzung.test.tsx`) deckt jede Seite ab". **Nachgezählt am
// 2026-08-22: die Datei hat einen einzigen Test, und der prüft eine Meldung aus
// Go.** Die Zusage war größer als der Beleg
// (#820).
//
// # Warum der Titel und nicht die Hilfe
//
// Den Titel zeigt jede Seite selbst; die Seitenhilfe steht gesammelt auf der
// Hilfeseite. Wo beide Sprachen dasselbe Wort haben — „Start", „Sets" —, sagt
// der Vergleich nichts, und der Fall wird **übersprungen statt geraten**.
import { afterEach, describe, expect, it } from 'vitest'

import { seiten } from './navigation'
import { kataloge } from './sprache/kataloge'
import { raeumeApi } from './test/api'
import { belegeProbeset } from './test/seitenvorrichtung'
import { geladenesDokumentDer } from './test/geladeneseite'

afterEach(() => {
  raeumeApi()
})

// deutscheBlaetter sind die deutschen Texte, die in einer englischen Sitzung
// nirgends stehen dürfen.
//
// **Aus allen fünf Katalogen, nicht aus einem Feld.** Der erste Anlauf prüfte
// nur den Seitentitel — und ließ deutschen Fließtext, deutsche Knöpfe und
// deutsche Meldungen auf derselben Seite durch. Angestrichen vom Zweitblick
// über Phase O
// (#834).
//
// **Ab zwölf Zeichen und nur, wo sich beide Sprachen unterscheiden.** Kurze
// Blätter sind die Falle: „Name", „Format", „Start" heißen in beiden Sprachen
// gleich oder stecken als Teilkette in einem englischen Wort. Was übrig
// bleibt, sind Sätze — und genau die verraten eine Seite, die am
// Sprachkontext vorbeiliest.
function deutscheBlaetter(): string[] {
  const gefunden: string[] = []

  const gehe = (deutsch: unknown, englisch: unknown) => {
    if (typeof deutsch === 'string') {
      if (
        typeof englisch === 'string' &&
        deutsch !== englisch &&
        deutsch.length >= 12
      ) {
        gefunden.push(deutsch)
      }

      return
    }

    if (typeof deutsch !== 'object' || deutsch === null) {
      return
    }

    for (const [schluessel, wert] of Object.entries(deutsch)) {
      gehe(wert, (englisch as Record<string, unknown> | null)?.[schluessel])
    }
  }

  gehe(kataloge.de, kataloge.en)

  return gefunden
}

describe('Eine englische Sitzung', () => {
  it.each(seiten)('spricht auf $schluessel Englisch', async (seite) => {
    belegeProbeset()

    // **Die geladene Seite, samt Schublade** (G02) — vor den Antworten
    // stand dort „Loading …" und keine Zeile, die sprechen könnte.
    const dokument = await geladenesDokumentDer(seite.schluessel, 'en')
    const gezeigt = dokument.textContent

    // **Der Titel zuerst** — er sagt, ob die Seite überhaupt aus dem Katalog
    // liest, und seine Meldung ist die verständlichste.
    //
    // **Beide Richtungen, und die Verneinung ist die wirksame.** `toContain`
    // allein erfüllt sich selbst: der Baum enthält die Seitenleiste, und die
    // trägt den englischen Namen ohnehin. Nachgemessen an einer Seite, die
    // absichtlich `kataloge.de` liest — sie blieb grün, bis die Verneinung
    // zurück war
    // (#834).
    const deutsch = kataloge.de.seiten[seite.schluessel].titel
    const englisch = kataloge.en.seiten[seite.schluessel].titel

    if (deutsch !== englisch) {
      expect(gezeigt).toContain(englisch)
      expect(gezeigt).not.toContain(deutsch)
    }

    // **Und dann jedes deutsche Blatt.** Steht eines davon auf der Seite,
    // liest sie am Sprachkontext vorbei.
    expect(deutscheBlaetter().filter((satz) => gezeigt.includes(satz))).toEqual(
      [],
    )
  })
})
