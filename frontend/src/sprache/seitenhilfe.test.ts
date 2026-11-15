// **Jede Seite hat eine Hilfe** (#806).
//
// Der Wächter hält den Hilfekatalog gegen `navigation.ts` — **in beide
// Richtungen**. Eine zu kurze Aufzählung ist sonst grün, und niemand sieht es;
// derselbe Grund, aus dem die axe-Prüfung ihre Seitenliste von dort nimmt und
// nicht aus einer Kopie daneben (§3).
import { describe, expect, it } from 'vitest'

import { seitenschluessel } from '../navigation'
import { kataloge } from './kataloge'

describe('die Hilfe je Seite', () => {
  for (const [sprache, katalog] of Object.entries(kataloge)) {
    describe(sprache, () => {
      const eintraege = katalog.hilfe.seiten as Record<
        string,
        { ueberschrift: string; text: string }
      >

      it('führt jeden Seitenschlüssel', () => {
        expect(Object.keys(eintraege).sort()).toEqual(
          [...seitenschluessel].sort(),
        )
      })

      // **Und die Einträge sind gefüllt.** Ein leerer Satz erfüllt die
      // Aufzählung und hilft niemandem — der Wächter oben zählte ihn mit.
      it.each([...seitenschluessel])('sagt für %s etwas', (schluessel) => {
        const eintrag = eintraege[schluessel]

        expect(eintrag?.ueberschrift.trim().length).toBeGreaterThan(0)
        expect(eintrag?.text.trim().length).toBeGreaterThan(20)
      })
    })
  }
})
