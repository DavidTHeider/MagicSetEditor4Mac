// **Welches Ablageziel Wails gerade markiert** — AP-4
// (#2361),
// Befund B-24.
//
// Beim Ziehen aus dem Finder setzt die Laufzeit von Wails die Klasse
// `file-drop-target-active` an das nächste Element mit
// `data-file-drop-target` unter dem Zeiger und nimmt sie beim Verlassen und
// nach dem Ablegen wieder weg (`@wailsio/runtime`, `window.js`,
// `handleDragOver` und `cleanupNativeDrag`, v3.0.0-beta.27). **Die Seite
// erfährt den Zug nur so**: Dateinamen gibt es erst beim Loslassen.
//
// # Was er nicht weiß
//
// **Was gezogen wird.** Ob es ein Set oder ein Bild ist, steht erst beim
// Ablegen fest; der Hinweis nennt deshalb beides.
import { useSyncExternalStore } from 'react'

const aktiv = 'file-drop-target-active'

// Ablageziel ist, was gerade markiert ist — undefined, wenn nichts gezogen
// wird; `bildfeld` ist der Name des Felds, wenn das Ziel eines ist.
export interface Ablageziel {
  bildfeld: string | undefined
}

function abonnieren(melden: () => void): () => void {
  const beobachter = new MutationObserver(melden)
  beobachter.observe(document.body, {
    attributes: true,
    attributeFilter: ['class'],
    subtree: true,
  })

  return () => {
    beobachter.disconnect()
  }
}

// lesen gibt einen festen Wert je Zustand, damit `useSyncExternalStore` bei
// gleichem Ziel nicht neu zeichnet.
let zuletzt: Ablageziel | undefined

function lesen(): Ablageziel | undefined {
  const ziel = document.querySelector(`.${aktiv}`)
  if (ziel === null) {
    zuletzt = undefined

    return undefined
  }

  const bildfeld = ziel.getAttribute('data-bildfeld') ?? undefined
  if (zuletzt?.bildfeld !== bildfeld || zuletzt === undefined) {
    zuletzt = { bildfeld }
  }

  return zuletzt
}

// useAblageziel gibt das Ziel, über dem gerade eine Datei steht.
export function useAblageziel(): Ablageziel | undefined {
  return useSyncExternalStore(abonnieren, lesen)
}
