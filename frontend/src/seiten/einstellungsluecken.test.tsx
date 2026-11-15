// **Jede Einstellung, die es gibt, lässt sich auch ändern** — UX-04.AK-01
// und NET-02.NAT-03
// (#1676,
// #1706).
//
// # Der Anlass
//
// **Am laufenden Fenster gemessen, 2026-09-21: 32 von 51.** Neunzehn
// Einstellungen standen nirgends — sechs `netz.*`, vier `arbeitsplatz.*`,
// sieben `richtlinie.*`, `konto.schau` und `bildkonflikt`. Sie sind
// gebaut, geprüft, haben Ebene, Herkunft, Vorgabe und seit AK-02 einen
// Wirkzeitpunkt — und **kein Mensch konnte sie im Fenster anfassen**.
//
// Das ist die erste Dimension aus §10, und sie ist hier doppelt bitter:
// [Netzweg](../bausteine/Netzweg.tsx) sagt wörtlich „Geändert wird auf der
// Einstellungsseite, wie jede andere Einstellung" — und dort stand nichts.
// Ein Kopf, der mehr verspricht, als sein Rumpf hält.
//
// **Und `docs/PROTOTYP-0.6.0.md` behauptete es ebenfalls**: die Tabelle
// nennt „Arbeitsplatz" unter *Erscheinungsbild*, „Marktkonto,
// Handelsschranken" unter *Dienste* und „Bildkonflikt" unter *Daten*. Sie
// war aus der Zuordnung geschrieben, nicht vom gebauten Bildschirm
// abgelesen.
//
// # Was er prüft
//
// **Jede Einstellung, die die Seite bekommt, hat ein Bedienelement** — in
// irgendeiner Gruppe. Nicht in welcher: das entscheidet
// `einstellungsgruppen.ts`, und der Lauf daneben hält es.
//
// # Was er nicht prüft
//
// **Ob die Go-Seite ihm wirklich alle gibt.** Er sieht, was er bekommt.
// Dass `Anwendungsinhalt` die volle Liste reicht, steht dort und ist an
// einem Lauf über die Seite nicht zu sehen.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import {
  alleDienste,
  eintraege,
  zeigeEinstellungen as zeige,
} from '../test/einstellungsseite'
import { einstellungsgruppen } from './einstellungsgruppen'
import { kataloge } from '../sprache/kataloge'

import type { EinstellungDTO } from '../api'

const eigene = kataloge.de.seiten.einstellungen

// **Je eine aus den vier Gruppen, die ganz gefehlt haben.**
const gefehlt: EinstellungDTO[] = (
  [
    ['netz.proxy.http', 'netz'],
    ['arbeitsplatz.navigator', 'arbeitsplatz'],
    ['richtlinie.kostengrenze', 'richtlinien'],
    ['konto.schau', 'dienste'],
    ['bildkonflikt', 'bildausfuhr'],
  ] as const
).map(([schluessel, bereich]) => ({
  schluessel,
  wert: '',
  wirksam: '',
  werte: [],
  bereich,
  ebene: 'profil',
  herkunft: 'vorgabe',
  vorgabe: '',
  wirkung: 'sofort',
}))

describe('die Einstellungsseite', () => {
  it('zeigt zu jeder Einstellung ein Bedienelement', async () => {
    const alleEinstellungen = [...eintraege, ...alleDienste, ...gefehlt]

    zeige({ alleEinstellungen, dienste: alleDienste })

    const gesehen = new Set<string>()

    for (const gruppe of einstellungsgruppen) {
      await userEvent.click(
        screen.getByRole('tab', { name: eigene.gruppen[gruppe] }),
      )

      // **Erreichbar heißt seit UXR-19 auch: einen Klick hinter einem
      // Aufklapper** (#1927).
      // Was selten gebraucht wird, steht eingeklappt (R3 aus
      // `docs/UIREVIEW.md`); geöffnet wird hier, was ein Mensch öffnen kann —
      // jeden Knopf, der etwas aufklappt, und nur ihn.
      for (const knopf of within(screen.getByRole('tabpanel')).queryAllByRole(
        'button',
        { expanded: false },
      )) {
        await userEvent.click(knopf)
      }

      for (const element of screen
        .getByRole('tabpanel')
        .querySelectorAll('[data-einstellung]')) {
        gesehen.add(element.getAttribute('data-einstellung') ?? '')
      }
    }

    const ohne = alleEinstellungen
      .map((eintrag) => eintrag.schluessel)
      .filter((schluessel) => !gesehen.has(schluessel))

    expect(ohne).toEqual([])
  })
})
