// **Nach einem Seitenwechsel steht der Fokus nicht im Leeren** — UX-10.AK-02
// (#1682).
//
// # Was gemessen war
//
// **Nach ⌘2 und ⌘6 bewegte Tab den Fokus gar nicht** — gemessen am
// 2026-09-25 in der WebKit-Probe. Der Tastendruck kam an und wurde nicht
// abgefangen; der Fokus stand auf `body`. Erst ein Element, das den Fokus
// bekam, setzte den Startpunkt neu — wer keine Maus benutzt, hatte keines.
// **Warum WebKit so sucht, ist nicht gemessen** — vermutlich von einem
// Knoten der vorigen Seite aus, den der Wechsel abgeräumt hatte.
//
// # Warum der Inhalt
//
// **`main` ist die Stelle, an der jede Seite anfängt**, und der Sprunglink
// zielt schon dorthin. Von dort führt Tab zum ersten Bedienelement der
// neuen Seite, und ein Vorleser sagt an, wo man ist.
//
// # Warum nur, wenn nichts den Fokus hat
//
// **Wer die Leiste bedient, bleibt dort**: ihr Knopf überlebt den Wechsel.
// Und eine Seite, die selbst einen Fokus setzt — ⌘N ins Namensfeld —,
// behält ihn: ihre Effekte laufen vor diesem, denn sie ist ein Kind.
//
// **„Nichts" heißt auch: auf einer verborgenen Seite.** Eine besuchte Seite
// bleibt aufbewahrt (`Seiteninhalt.tsx`, `<Activity mode="hidden">`); ihr
// Knopf wird nicht entfernt, sondern mit `display: none` versteckt, und der
// Fokus stand noch auf ihm. Tab suchte dann von einer Seite aus weiter, die
// niemand sah — gemessen nach ⌘3, ⌘5, ⌘6 und ⌘' am selben Tag.
//
// # Und nicht beim ersten Zeichnen
//
// **Wer die Anwendung öffnet, beginnt mit Tab beim Sprunglink** — beim
// ersten Zeichnen ist nichts gewechselt.
import { useEffect, useRef } from 'react'

import type { Seitenschluessel } from '../navigation'

export function useFokusNachSeitenwechsel(aktuell: Seitenschluessel): void {
  const vorige = useRef(aktuell)

  useEffect(() => {
    if (vorige.current === aktuell) {
      return
    }

    vorige.current = aktuell

    if (imLeeren(document.activeElement)) {
      ;(offeneSchublade() ?? document.getElementById('inhalt'))?.focus({
        preventScroll: true,
      })
    }
  }, [aktuell])
}

// offeneSchublade gibt das erste Bedienelement einer offenen Schublade —
// **dort bleibt der Fokus**, wenn ein Laschenwechsel ihn aus ihr nimmt.
// Angestrichen vom Zweitblick über Phase 6.3 (M2): mit ⌥⌘4–7 ging er auf den
// Inhalt **hinter** der Schublade, ohne sichtbaren Ring, und Escape schloss
// sie danach nicht mehr — Mantines Falle greift nur am Rand der Schublade.
function offeneSchublade(): HTMLElement | null {
  // **Nur eine sichtbare** — WK03 (#2104): eine Schublade einer aufbewahrten
  // Seite bleibt verborgen im Baum (Activity). Der Fokus ging in sie, wirkte
  // nicht, und blieb auf `body`.
  return (
    [
      ...document.querySelectorAll<HTMLElement>(
        '[role="dialog"][aria-modal="true"]',
      ),
    ]
      .find((schublade) => !imVerborgenen(schublade))
      ?.querySelector<HTMLElement>(
        'button, a[href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
      ) ?? null
  )
}

// imLeeren sagt, ob der Fokus nirgends steht, wo man ihn sieht — auf `body`
// oder in einem Teilbaum, den `Activity` verborgen hat.
//
// **Über das Stilattribut und nicht über `checkVisibility`**: React setzt
// `display: none` als Stil an die Wurzeln des verborgenen Teilbaums, und das
// liest jsdom genauso wie WebKit; eine Frage nach dem Layout beantwortet
// jsdom nicht.
function imLeeren(aktiv: Element | null): boolean {
  return aktiv === null || aktiv === document.body || imVerborgenen(aktiv)
}

// imVerborgenen sagt, ob ein Knoten in einer verborgenen Seite steht — so,
// wie Activity sie verbirgt.
function imVerborgenen(knoten: Element): boolean {
  return knoten.closest('[style*="display: none"]') !== null
}
