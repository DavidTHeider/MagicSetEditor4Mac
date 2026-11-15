// **Ein deutsches Wort, das zwei Dinge benennt** — Wächter zu
// #1551.
//
// # Wogegen er läuft
//
// Zwei Schlüssel **desselben Verbunds** tragen denselben deutschen Text,
// während ihre englischen sich unterscheiden. Dann sagt der englische
// Katalog, dass es zwei verschiedene Dinge sind, und der deutsche zeigt
// eines.
//
// **Gefunden hat diese Klasse kein Wächter, sondern ein `getByText`, das bei
// zwei Treffern scheitert** — also der Zufall, dass ein Lauf gerade diesen
// Text abfragte. Harmlos ist sie nicht: bis zum 2026-09-13 griff
// `Massenvorgaenge.test.tsx` neunmal mit `texte.aendern` und `texte.loeschen`
// nach den Knöpfen `masseAendern` und `masseLoeschen` — und traf sie, weil
// beide Paare dasselbe deutsche Wort trugen. Im englischen Katalog hätte
// derselbe Test nichts gefunden.
//
// # Der Verbund und nicht der ganze Katalog
//
// Verglichen wird, was **in demselben Objekt** nebeneinander steht — also
// das, was eine Seite oder ein Bereich führt. Über den ganzen Katalog hinweg
// wäre die Regel wertlos: `seiten.sets.loeschen` und `seiten.sammlung.loeschen`
// heißen zu Recht beide „Löschen", denn sie stehen nie nebeneinander.
//
// # Was er nicht prüft
//
// **Er sieht nur die Fälle, in denen das Englische die Unterscheidung schon
// macht.** Ein deutsches Wort, das zwei Dinge benennt und auch im Englischen
// gleich heißt, bleibt ihm unsichtbar — dagegen hilft nur, wer hinsieht.
//
// **Und er vergleicht nur Geschwister.** Zwei Schlüssel, die auf derselben
// Seite landen, aber in verschiedenen Verbünden liegen —
// `seiten.sammlung.zustaende.*` gegen `seiten.sammlung.*` —, sieht er nicht.
//
// **Er sagt auch nicht, welche Seite recht hat.** Dass die beiden Kataloge
// uneins sind, ist eine Messung; ob das deutsche Wort zu schärfen oder das
// englische anzugleichen ist, ist eine Entscheidung je Zeile.
import { describe, expect, it } from 'vitest'

import { kataloge } from './kataloge'

// Verbund ist, was in einem Objekt des Katalogs nebeneinandersteht — je
// Schlüssel sein deutscher und sein englischer Text.
type Verbund = Map<string, { de: string; en: string }>

describe('Ein deutsches Wort je Sache', () => {
  it('benennt in keinem Verbund zwei Dinge mit demselben deutschen Wort', () => {
    expect(befunde(kataloge.de, kataloge.en)).toEqual([])
  })

  // **Die Gegenprobe steht im Wächter und nicht in einer Notiz** (§10): ein
  // Wächter, der nur einmal grün war, hat nichts gezeigt. Der gepflanzte Fall
  // ist der aus der Sammlung, der diesen Wächter veranlasst hat.
  it('findet einen gepflanzten Fall', () => {
    expect(
      befunde(
        {
          seite: { aendern: 'Ändern', masseAendern: 'Ändern', menge: 'Menge' },
        },
        {
          seite: { aendern: 'Save', masseAendern: 'Change', menge: 'Quantity' },
        },
      ),
    ).toEqual([
      'seite: „Ändern" für aendern und masseAendern — englisch Save gegen Change',
    ])
  })

  // **Und er geht an dem vorbei, was zu Recht gleich heißt**: Spaltenname und
  // Gruppierung desselben Feldes sind dieselbe Sache, in beiden Sprachen.
  // Ohne diesen Fall stünden in `seiten.sammlung` allein zehn Treffer, und
  // ein Wächter mit zehn berechtigten Treffern wird abgeschaltet.
  it('geht an zwei Einträgen vorbei, die auch englisch gleich heißen', () => {
    expect(
      befunde(
        { seite: { spalteSet: 'Set', gruppierungSet: 'Set' } },
        { seite: { spalteSet: 'Set', gruppierungSet: 'Set' } },
      ),
    ).toEqual([])
  })
})

// ------------------------------------------------------------------ Hilfsmittel

// befunde nennt jede Stelle, an der ein deutsches Wort für zwei Dinge steht.
function befunde(deutsch: object, englisch: object): string[] {
  const gefunden: string[] = []

  for (const [pfad, verbund] of verbuende(deutsch, englisch)) {
    const nachDeutsch = new Map<string, { schluessel: string; en: string }[]>()

    for (const [schluessel, paar] of verbund) {
      const bisher = nachDeutsch.get(paar.de) ?? []

      nachDeutsch.set(paar.de, [...bisher, { schluessel, en: paar.en }])
    }

    for (const [wort, traeger] of nachDeutsch) {
      // **Mehr als ein englisches Wort heißt: mehr als ein Träger.** Eine
      // zusätzliche Längenprüfung wäre ein Zweig, den kein Lauf betritt (§3).
      const englischeWorte = new Set(traeger.map((einer) => einer.en))

      if (englischeWorte.size > 1) {
        const schluessel = traeger
          .map((einer) => einer.schluessel)
          .join(' und ')

        gefunden.push(
          `${pfad}: „${wort}" für ${schluessel} — englisch ${[...englischeWorte].join(' gegen ')}`,
        )
      }
    }
  }

  return gefunden.sort()
}

// verbuende sammelt je Objekt des Katalogs seine Zeichenketteneinträge.
//
// **Die Schlüsselmengen beider Kataloge sind gleich** — das hält der Typ
// `Katalog` und zusätzlich `sprache.test.tsx`. Deshalb steht hier kein Zweig
// für einen fehlenden englischen Eintrag: er wäre unbetretbar.
function verbuende(deutsch: object, englisch: object): Map<string, Verbund> {
  const gesammelt = new Map<string, Verbund>()

  const gehe = (de: object, en: object, pfad: string): void => {
    const verbund: Verbund = new Map()

    for (const [schluessel, wert] of Object.entries(de)) {
      const gegenstueck = (en as Record<string, unknown>)[schluessel]

      if (typeof wert === 'string') {
        verbund.set(schluessel, { de: wert, en: String(gegenstueck) })
      } else {
        gehe(
          wert as object,
          gegenstueck as object,
          pfad === '' ? schluessel : `${pfad}.${schluessel}`,
        )
      }
    }

    gesammelt.set(pfad === '' ? '(Wurzel)' : pfad, verbund)
  }

  gehe(deutsch, englisch, '')

  return gesammelt
}
