// **Das Ablageziel sagt, was geschieht, bevor man loslässt** — AP-4
// (#2361),
// Befund B-24.
//
// Wails setzt beim Ziehen aus dem Finder die Klasse
// `file-drop-target-active` an das nächste Element mit
// `data-file-drop-target` (`@wailsio/runtime`, `window.js`,
// `handleDragOver`), und kein Stil griff sie auf: ein Bild, das neben die
// Karte fiel, verschwand wortlos. Jetzt steht ein Satz da — über einem
// Bildfeld, in welches Feld das Bild geht; sonst, dass ein Set geöffnet wird
// und ein Bild auf ein Bildfeld gehört.
import { act, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Ablagehinweis } from './Ablagehinweis'

const worte = kataloge.de.bausteine.ablage

function aufbauen() {
  const wurzel = document.createElement('div')
  wurzel.setAttribute('data-file-drop-target', '')
  const feld = document.createElement('div')
  feld.setAttribute('data-file-drop-target', '')
  feld.setAttribute('data-bildfeld', 'image')
  wurzel.appendChild(feld)
  document.body.appendChild(wurzel)

  return { wurzel, feld }
}

afterEach(() => {
  document.body.innerHTML = ''
})

describe('der Hinweis beim Ziehen', () => {
  it('schweigt, solange nichts gezogen wird', () => {
    aufbauen()
    render(
      <Sprachraum sprache="de">
        <Ablagehinweis />
      </Sprachraum>,
    )

    expect(screen.getByRole('status')).toBeEmptyDOMElement()
  })

  it('sagt über dem Fenster, dass ein Set geöffnet wird', async () => {
    const { wurzel } = aufbauen()
    render(
      <Sprachraum sprache="de">
        <Ablagehinweis />
      </Sprachraum>,
    )

    await act(async () => {
      wurzel.classList.add('file-drop-target-active')
      await Promise.resolve()
    })

    expect(screen.getByRole('status')).toHaveTextContent(worte.fenster)
  })

  it('sagt über einem Bildfeld, in welches Feld das Bild geht', async () => {
    const { feld } = aufbauen()
    render(
      <Sprachraum sprache="de">
        <Ablagehinweis />
      </Sprachraum>,
    )

    await act(async () => {
      feld.classList.add('file-drop-target-active')
      await Promise.resolve()
    })

    expect(screen.getByRole('status')).toHaveTextContent(
      worte.bildfeld.replace('{feld}', 'image'),
    )

    await act(async () => {
      feld.classList.remove('file-drop-target-active')
      await Promise.resolve()
    })

    expect(screen.getByRole('status')).toBeEmptyDOMElement()
  })
})
