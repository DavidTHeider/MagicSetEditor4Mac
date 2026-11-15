// **Die Wörter des Spielkerns im Fenster** —
// #1623.
//
// # Was hier geprüft wird und was nebenan
//
// **Hier steht der Rückfall.** Dass `bibliothek` zu „Bibliothek" wird, sieht
// man in jedem Lauf über den Tisch; dass ein **unbekanntes** Wort mit seiner
// Kennung dasteht, sieht man dort nie — der Kern schickt keine unbekannte
// Zone. Ein Rückfall ohne Lauf ist ein Versprechen, das niemand einlöst
// (§10), und er hat einen Grund: fremde Daten sind der Normalfall (§8).
import { describe, expect, test } from 'vitest'

import {
  handlungswort,
  kurzzonenwort,
  standwort,
  zeigerwort,
  zonenwort,
} from './partiewoerter'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.spielen
const handlungen = kataloge.de.seiten.handlungen

describe('was der Katalog kennt', () => {
  test('eine Zone bekommt ihr Wort', () => {
    expect(zonenwort(eigene, 'bibliothek')).toBe(eigene.zoneBibliothek)
  })

  // **Das Schlachtfeld heißt Schlachtfeld** — so nennt es CR 403 (Fassung
  // 2026-09-25), so der Entwurf (Artboard *Tisch*), und so hieß es im
  // englischen Katalog schon. Bis zum 2026-09-24 hieß die Zone deutsch
  // „Tisch", auf einer Seite, die selbst „Am Tisch" heißt.
  test('das Schlachtfeld heißt in beiden Sprachen, wie das Regelwerk es nennt', () => {
    expect(zonenwort(eigene, 'tisch')).toBe('Schlachtfeld')
    expect(zonenwort(kataloge.en.seiten.spielen, 'tisch')).toBe('Battlefield')
  })

  test('ein Stand ebenso', () => {
    expect(standwort(eigene, 'laeuft')).toBe(eigene.standLaeuft)
  })

  // **Die Ablagen haben eine Kurzform** — über einem Stapel von 63 Einheiten
  // Breite passt „Bibliothek" nicht, und bis zum 2026-09-24 lief der Name in
  // den des Friedhofs daneben.
  test('eine Ablage bekommt ihre Kurzform', () => {
    expect(kurzzonenwort(eigene, 'bibliothek')).toBe(eigene.zonekurzBibliothek)
    expect(kurzzonenwort(eigene, 'friedhof')).toBe(eigene.zonekurzFriedhof)
    expect(kurzzonenwort(eigene, 'exil')).toBe(eigene.zonekurzExil)
    expect(kurzzonenwort(eigene, 'kommando')).toBe(eigene.zonekurzKommando)
  })

  // **Und wo keine steht, das ganze Wort** — die Hand ist breit genug.
  test('eine andere Zone behält ihr Wort', () => {
    expect(kurzzonenwort(eigene, 'hand')).toBe(eigene.zoneHand)
    expect(kurzzonenwort(eigene, 'keller')).toBe('keller')
  })

  test('und eine Befehlsart', () => {
    expect(handlungswort(handlungen.handlungsname, 'objekt.tappen')).toBe(
      handlungen.handlungsname['objekt.tappen'],
    )
  })
})

describe('was er nicht kennt', () => {
  test('eine unbekannte Zone steht mit ihrer Kennung da', () => {
    expect(zonenwort(eigene, 'keller')).toBe('keller')
  })

  test('ein unbekannter Stand ebenso', () => {
    expect(standwort(eigene, 'schwebend')).toBe('schwebend')
  })

  test('eine unbekannte Präsenzart ebenso', () => {
    expect(zeigerwort(eigene, 'winkt')).toBe('winkt')
  })

  test('und ein unbekannter Nutzlastschlüssel ebenso', () => {
    // **Der Fall, den der Wächter nicht hält**:
    // `internal/werkzeuge/befehlsformen_test.go` prüft die **Arten** gegen
    // den Kern, nicht die Feldnamen. Ein neues Feld ohne Katalogeintrag
    // stünde sonst mit einer leeren Beschriftung da.
    expect(handlungswort(handlungen.handlungsfeld, 'erfunden')).toBe('erfunden')
  })
})
