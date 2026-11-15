// **Der Aktionskatalog hält, was er zusagt** — UX-10.AK-01 und AK-04
// (#1682).
//
// # Drei Zusagen, drei Läufe
//
// 1. **Jede Katalogaktion geht durch den Katalog.** Erkannt wird sie am
//    Schlüssel ihrer Aufschrift — derselbe Griff wie in `loeschwort.test.ts`:
//    ein Knopf, dessen Aufschrift `…Loeschen` heißt, löscht. Ein gewöhnlicher
//    `Knopf` oder `<button>` mit einem solchen Schlüssel ist ein Befund; ein
//    `Aktionsknopf`, ein Menüeintrag oder eine `Entscheidung` mit dem
//    Schlüssel einer **anderen** Aktion auch.
// 2. **Ein angekündigtes Kürzel wirkt.** Wo `mitKuerzel` steht, nennt die
//    Tafel `kuerzelstellen` den Lauf, der das Kürzel drückt — und umgekehrt.
// 3. **Vor jeder löschenden Bindung steht, was ihre Wirkung verlangt**
//    (AK-04). Die Tafel `loeschende` in `aktionstafel.ts` führt jede Bindung
//    aus `bindungsvertrag.json`, die löscht, entfernt, widerruft, vergisst
//    oder zurücksetzt — mit ihrer Bestätigung und dem Beleg dafür.
//
// # Was er nicht hält
//
// **Ob das Wort zur Handlung passt.** Ein Knopf, der Dateien von der Platte
// nimmt, und dessen Schlüssel „entfernen" sagt, ginge hier als Entfernen
// durch; der Fall stand bis zum 2026-09-21 bei den Paketen. Das sieht nur ein
// Mensch — dieselbe Grenze wie bei `loeschwort.test.ts`.
//
// **Und ob die Rückfrage vor dem richtigen Ruf steht.** Geprüft wird, dass es
// sie in der genannten Datei gibt. Wo eine Datei zwei Löschwege hat und nur
// einer fragt, hilft das nicht — deshalb nennt die Tafel die Datei, in der
// die Frage **zu diesem** Ruf steht, und die Tests der Seiten gehen den Weg.
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

import { describe, expect, it } from 'vitest'

import vertrag from './bindungsvertrag.json'
import {
  aktionVon,
  alleStellen,
  beschreiben,
  naheOrte,
  quellen,
} from './test/aktionsstellen'
import {
  kuerzelstellen,
  objektkuerzelstellen,
  loeschende,
  loeschendeMuster,
} from './test/aktionstafel'

describe('Der Aktionskatalog', () => {
  it('geht an jeder Stelle einer Katalogaktion durch den Katalog', () => {
    const stellen = alleStellen()
    const befunde: string[] = []

    // **Eine Gegenprobe, dass der Lauf etwas sieht**: ohne Stellen wäre
    // „kein Befund" keine Aussage (§3).
    expect(
      stellen.filter((stelle) => aktionVon(stelle.schluessel)).length,
    ).toBeGreaterThan(50)

    for (const stelle of stellen) {
      const erwartet = aktionVon(stelle.schluessel)

      if (erwartet === undefined) {
        if (stelle.aktion !== undefined && stelle.art !== 'Aktionsknopf') {
          befunde.push(
            beschreiben(
              stelle,
              `trägt „${stelle.aktion}", sein Schlüssel keine Aktion`,
            ),
          )
        }
        continue
      }

      if (stelle.art === 'Knopf' || stelle.art === 'button') {
        befunde.push(
          beschreiben(stelle, `gewöhnlicher Knopf für „${erwartet}"`),
        )
      } else if (stelle.aktion !== erwartet) {
        befunde.push(
          beschreiben(
            stelle,
            `trägt „${stelle.aktion ?? 'keine Aktion'}", gemeint ist „${erwartet}"`,
          ),
        )
      }
    }

    expect(befunde).toEqual([])
  })

  it('nennt für jedes Kürzel einer Objekthandlung den Lauf, der es drückt', () => {
    const genannt = alleStellen().flatMap((stelle) =>
      stelle.kuerzel === undefined ? [] : [stelle.kuerzel],
    )

    expect(genannt.length).toBeGreaterThan(4)
    expect([...new Set(genannt)].sort()).toEqual(
      Object.keys(objektkuerzelstellen).sort(),
    )
    for (const [kuerzel, laeufe] of Object.entries(objektkuerzelstellen)) {
      for (const [lauf, marke] of laeufe) {
        expect(
          readFileSync(join(quellen, lauf), 'utf8'),
          `${kuerzel} → ${lauf}`,
        ).toContain(marke)
      }
    }
  })

  // **Auch die Listen in `.ts`** — Befund 4 des Zweitblicks über Phase 6:
  // seit AP-5 stehen Einträge in `paketeintraege.ts` und `dateieintraege.ts`,
  // und der Lauf las nur `.tsx`.
  it('liest die ausgelagerten Listen mit', () => {
    const dateien = new Set(alleStellen().map((stelle) => stelle.datei))

    expect(dateien).toContain('seiten/paketeintraege.ts')
    expect(dateien).toContain('seiten/dateieintraege.ts')
  })

  it('kündigt ein Kürzel nur dort an, wo ein Lauf es drückt', () => {
    const angekuendigt = alleStellen().filter((stelle) => stelle.mitKuerzel)

    expect(
      [...new Set(angekuendigt.map((stelle) => stelle.datei))].sort(),
    ).toEqual(Object.keys(kuerzelstellen).sort())

    for (const stelle of angekuendigt) {
      expect(stelle.aktion, stelle.datei).toBe(
        kuerzelstellen[stelle.datei]?.aktion,
      )
    }

    for (const [datei, eintrag] of Object.entries(kuerzelstellen)) {
      for (const [lauf, marke] of eintrag.laeufe) {
        // **Der Lauf drückt die Taste** — nicht nur, dass es ihn gibt.
        expect(
          readFileSync(join(quellen, lauf), 'utf8'),
          `${datei} → ${lauf}`,
        ).toContain(marke)
      }
    }
  })
})

// rueckfrage erkennt die Frage vor einem Löschen: die Bestätigung einer
// `Entscheidung` oder ein Ja in der Zeile, dessen Schlüssel auf eines der
// fünf Wörter endet, an denen `loeschendeMuster` eine Bindung erkennt.
const rueckfrage =
  /bestaetigen=\{[^}]*(loeschen|entfernen|widerrufen|vergessen|zuruecksetzen)(ja)?\}|(loeschen|entfernen|widerrufen|vergessen|zuruecksetzen)ja\}/i

describe('Vor jeder löschenden Bindung', () => {
  const bindungen = Object.keys(vertrag.methoden).filter((name) =>
    loeschendeMuster.test(name),
  )

  it('steht eine Entscheidung in der Tafel', () => {
    expect(bindungen.length).toBeGreaterThan(25)
    expect(Object.keys(loeschende).sort()).toEqual([...bindungen].sort())
  })

  // **Und die Zahlen in `docs/UX-ARCHITEKTUR.md` sind die der Tafel** — eine
  // Zahl in einer Datei, die kein Lauf hält, stimmt nur am Tag, an dem sie
  // geschrieben wurde (§16).
  it('hält die Zahlen der Architekturdatei', () => {
    const datei = readFileSync(
      join(quellen, '../../docs/UX-ARCHITEKTUR.md'),
      'utf8',
    )
    const zahl = (bestaetigung: string) =>
      String(
        Object.values(loeschende).filter(
          (eintrag) => eintrag.bestaetigung === bestaetigung,
        ).length,
      )

    expect(datei).toContain(`**${String(bindungen.length)} Bindungen** löschen`)
    for (const [zeile, bestaetigung] of [
      ['Rückfrage an der Klickstelle', 'rueckfrage'],
      ['getippter Name', 'namen'],
      ['Rücknahme mit ⌘Z', 'ruecknahme'],
      ['keine, mit Grund', 'keine'],
    ] as const) {
      expect(datei).toContain(`| ${zeile} | ${zahl(bestaetigung)} |`)
    }
  })

  it.each(bindungen)('%s hält, was die Tafel sagt', (bindung) => {
    const eintrag = loeschende[bindung]
    expect(eintrag).toBeDefined()
    if (eintrag === undefined) {
      return
    }

    const inhalt = readFileSync(join(quellen, eintrag.ort), 'utf8')

    // **Und der Ruf steht dort** — sonst prüfte die Zeile eine Datei, die
    // mit der Bindung nichts zu tun hat. Ein Haken dazwischen zählt mit: der
    // Ort nennt ihn, der Haken ruft.
    expect(naheOrte(bindung), bindung).toContain(eintrag.ort)

    switch (eintrag.bestaetigung) {
      case 'rueckfrage':
        // **Die Frage steht an der Klickstelle** — als `Entscheidung` oder
        // als Ja und Nein in der Zeile —, **und sie fragt nach dem Löschen**.
        // Eine Frage vor dem Erneuern eines Tokens in derselben Zeile zählt
        // nicht: sie stand dort, als das Löschen daneben nicht fragte.
        expect(inhalt, eintrag.ort).toMatch(rueckfrage)
        break

      case 'namen':
        // **Bestätigt wird mit dem getippten Namen**, und das Feld dafür
        // steht am Ort.
        expect(inhalt, `${eintrag.ort}: ${eintrag.feld}`).toContain(
          eintrag.feld,
        )
        break

      case 'ruecknahme': {
        const [datei, marke] = eintrag.beleg
        expect(existsSync(join(quellen, datei)), datei).toBe(true)
        expect(readFileSync(join(quellen, datei), 'utf8')).toContain(marke)
        break
      }

      case 'keine':
        // **Keine Frage braucht einen Grund**, und er steht in der Tafel.
        expect(eintrag.grund.length).toBeGreaterThan(20)
        break
    }
  })
})
