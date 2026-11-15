// **Die Vorrichtung für die Spielenseite** — geteilt von zwei Läufen.
//
// **Eine eigene Datei**, weil `spielen.test.tsx` und
// `spielenregeln.test.tsx` dieselbe Seite zeichnen und dieselbe Fassade
// belegen. Eine Vorrichtung, die zweimal dastünde, liefe beim ersten neuen
// Feld auseinander (§10).
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { vi } from 'vitest'

import { Spielen } from './Spielen'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { belegeApi } from '../test/api'
import { eineListe } from './peerkontaktprobe'
import { zweiDecks } from './partieprobe'

import type { Seitenschluessel } from '../navigation'

// zeige zeichnet die Seite und gibt die Hand.
//
// **`wechseln` führt seit UXR-13 (#1921) zu den Decks**, wenn keines da ist;
// ein Lauf, der davon nichts wissen will, gibt keinen eigenen. Die Vorgabe
// ist ein `vi.fn()` und keine eigene leere Funktion: die zählte die
// Abdeckung als nie gerufen.
export function zeige(wechseln: (seite: Seitenschluessel) => void = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Spielen wechseln={wechseln} />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// steigeEin setzt sich aus der Partienliste an einen Sitz — über das Menü
// „Weiterspielen" der Zeile, wie seit UXR-13
// (#1921).
// Vorher stand je Sitz ein Knopf.
export async function steigeEin(
  hand: ReturnType<typeof nutzerereignis.setup>,
  partie: string,
  sitz: string,
): Promise<void> {
  const eigene = kataloge.de.seiten.spielen

  await hand.click(
    await screen.findByRole('button', {
      name: einsetzen(eigene.partienWeiterspielenIn, { name: partie }),
    }),
  )
  await hand.click(
    screen.getByRole('menuitem', {
      name: einsetzen(eigene.partienEinsteigenAls, { name: sitz }),
    }),
  )
}

// mitDecks belegt die Fassade so, dass der Anfang zwei Decks sieht.
//
// **Und die Regelfreigabe gleich mit** (G4, #1461): der Startassistent fragt
// sie, sobald ein Spiel feststeht. Eine Vorrichtung ohne sie ließe jeden
// Lauf eine **Fehlermeldung** zeigen, die mit seiner Sache nichts zu tun hat
// — gesehen am 2026-09-16, als zwei `alert` in derselben Seite standen.
export function mitDecks(weiteres: Record<string, unknown> = {}) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: zweiDecks(), fehler: null }),
    Formate: () => Promise.resolve({ formate: [], fehler: null }),
    Regelfreigabe: () => Promise.resolve(eineFreigabe()),

    // **Und die eigenen Profile** (G8, #1465): der Startassistent fragt, wem
    // er einen Sitz anbieten darf. **Ohne Profil ist die Vorgabe** — dann
    // steht das Feld nicht da, und ein Lauf, der von ihm nichts wissen will,
    // sieht es auch nicht.
    Nutzer: () => Promise.resolve({ nutzer: [], fehler: null }),
    WerArbeitet: () => Promise.resolve({ nutzerID: '1', fehler: null }),

    // **Und die eigenen Partien** (#1600): die Seite fragt, ob eine offene
    // dasteht, bevor jemand eine anfängt. **Keine ist die Vorgabe** — ein
    // Lauf, der von der Liste nichts wissen will, sieht sie nicht.
    Partien: () => Promise.resolve({ partien: [], fehler: null }),

    // **Und die Kontakte** (UX-08.AK-04, #1680): am fremden Tisch und bei der
    // Leihe steht zur Wahl, wer bestätigt ist. **Niemand ist die Vorgabe** —
    // ein Lauf, der von der Wahl nichts wissen will, sieht nur ihren Satz.
    Peerkontakte: () => Promise.resolve(eineListe([])),
    ...weiteres,
  })
}

// eineFreigabe gibt die Antwort des Grundprofils, wie die Naht sie gibt.
export function eineFreigabe() {
  return {
    freigabe: {
      erlaubt: true,
      regelwerk: 'sandkasten',
      fassung: 1,
      wirksam: '2026-08-19',
      hoechste: 'begleitet',
      matrix: [],
      luecken: [],
      hindernisse: [],

      // **Modus und Baum kamen mit PLAY-02** (#1710) — eine Vorrichtung
      // ohne sie wäre großzügiger als der Betrieb (§3).
      modus: 'anfaenger',
      baum: [
        { kennung: 'ablauf', bereiche: ['zonen'] },
        { kennung: 'mechanik', bereiche: ['marken'] },
      ],
      widersprueche: [],
    },
    fehler: null,
  }
}
