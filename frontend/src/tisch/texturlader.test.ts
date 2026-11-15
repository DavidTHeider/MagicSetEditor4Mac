// **Welches Bild ein Sprite trägt oder erwartet** — die Entscheidung, die bis
// zum 2026-09-24 in `pixinaht.ts` stand und dort keiner maß.
//
// # Wogegen hier geprüft wird
//
// **Gegen den Tisch vom 2026-09-24**, gesehen in WebKit mit dem Set „Green":
// zwei Karten in Adas Hand blieben Platzhalter, obwohl ihr Bild längst da
// war — die Bindung gab die Adresse in 74 ms, das Bild ließ sich laden und
// dekodieren, und nach einem Wechsel auf die Liste und zurück standen sie
// da. Der Vermerk „schon geladen" hing an **Kennung und Adresse** und
// überlebte, was er beschrieb: ein Sprite, das auf den Platzhalter
// zurückfiel, und ein Sprite, das neu entstand. Fragte danach dieselbe
// Adresse wieder, kehrte der Zeichner zurück, ohne zu laden.
import { describe, expect, it } from 'vitest'

import { neuerTexturlader } from './texturlader'

// sprite ist ein Stellvertreter — der Lader kennt vom Sprite nur seine
// Identität.
const sprite = () => ({})

describe('der Texturlader', () => {
  it('lädt ein Bild für ein Sprite einmal', () => {
    const lader = neuerTexturlader<object>()
    const eines = sprite()

    expect(lader.soll(eines, '/bild/a.png')).toBe(true)
    expect(lader.soll(eines, '/bild/a.png')).toBe(false)
  })

  it('lädt ein anderes Bild neu', () => {
    const lader = neuerTexturlader<object>()
    const eines = sprite()

    lader.soll(eines, '/bild/a.png')

    expect(lader.soll(eines, '/bild/b.png')).toBe(true)
  })

  // **Der Fall vom 2026-09-24**: der Platzhalter kam zurück, dann dieselbe
  // Adresse — und es wurde nicht geladen.
  it('lädt nach dem Platzhalter dasselbe Bild wieder', () => {
    const lader = neuerTexturlader<object>()
    const eines = sprite()

    lader.soll(eines, '/bild/a.png')
    lader.vergiss(eines)

    expect(lader.soll(eines, '/bild/a.png')).toBe(true)
  })

  // **Ein neues Sprite für dieselbe Karte ist ein neues Sprite** — der
  // Vermerk hing an der Kennung und hätte es für geladen gehalten.
  it('lädt für ein neues Sprite, auch wenn ein altes dasselbe Bild trug', () => {
    const lader = neuerTexturlader<object>()

    lader.soll(sprite(), '/bild/a.png')

    expect(lader.soll(sprite(), '/bild/a.png')).toBe(true)
  })

  // **Eine späte Antwort gilt nur für das, was noch erwartet wird** — sonst
  // setzte ein langsames altes Bild sich über ein schnelles neues.
  it('lässt eine überholte Antwort nicht gelten', () => {
    const lader = neuerTexturlader<object>()
    const eines = sprite()

    lader.soll(eines, '/bild/a.png')
    lader.soll(eines, '/bild/b.png')

    expect(lader.gilt(eines, '/bild/a.png')).toBe(false)
    expect(lader.gilt(eines, '/bild/b.png')).toBe(true)

    lader.vergiss(eines)

    expect(lader.gilt(eines, '/bild/b.png')).toBe(false)
  })
})
