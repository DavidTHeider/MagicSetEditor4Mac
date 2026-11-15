// **Die Sammlungsseite für Läufe** — UXR-10
// (#1918).
//
// Ein Bestand in der Menge, die ein Mensch tatsächlich führt, die Seite und
// die Wege in ihre Schubladen. **Die Ausgangslage am Fenster war ein leerer
// Bestand** (`docs/UIREVIEW.md`, Hüllenprobe `uirevue` gegen das Probeset):
// eine Fläche, die dort sauber aussieht, kann mit vierhundert Zeilen, acht
// Marken und einer Auswahl kippen — deshalb zählen die Läufe beides.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi } from 'vitest'

import { keineHerkunft } from '../apiherkunft'
import { Sammlung } from '../seiten/Sammlung'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from './api'

import type { Api, ExemplarDTO } from '../api'

export const texte = kataloge.de.seiten.sammlung

// exemplar baut ein Exemplar, wie die Bindung es gibt.
export function exemplar(teil: Partial<ExemplarDTO> = {}): ExemplarDTO {
  return {
    id: 1,
    fassung: 1,
    printing: 'probe-druck',
    druckID: '',
    vorlaeufig: false,
    menge: 3,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: '',
    setcode: '',
    nummer: '',
    lagerort: '',
    wunsch: false,
    bezeichnung: 'probe-druck',
    herkunft: keineHerkunft,
    ...teil,
  }
}

// bestand gibt `anzahl` Exemplare über drei Sets und fünf Lagerorte.
export function bestand(anzahl: number): ExemplarDTO[] {
  return Array.from({ length: anzahl }, (_, stelle) =>
    exemplar({
      id: stelle + 1,
      printing: `druck-${String(stelle + 1)}`,
      bezeichnung: `Karte ${String(stelle + 1).padStart(4, '0')}`,
      setcode: ['m11', 'khm', 'neo'][stelle % 3] ?? '',
      lagerort: `Binder ${String(stelle % 5)}`,
      menge: (stelle % 4) + 1,
    }),
  )
}

// marken verteilt acht eigene Marken über die ersten Exemplare — so viele,
// wie das Artboard und ein geordneter Bestand tragen.
export function marken(exemplare: readonly ExemplarDTO[]) {
  const namen = [
    'Handelsstapel',
    'Bindermappe 2',
    'verliehen',
    'Commander',
    'doppelt',
    'Foil-Ordner',
    'Wunsch',
    'Kiste 4',
  ]

  return Object.fromEntries(
    exemplare
      .slice(0, 40)
      .map((eines, stelle) => [
        String(eines.id),
        [namen[stelle % namen.length] ?? ''],
      ]),
  )
}

// belegeSammlung setzt die Bindungen der Seite für diesen Bestand.
export function belegeSammlung(
  exemplare: ExemplarDTO[],
  teile: Partial<Api> = {},
): Api {
  return belegeApi({
    Sammlung: () => Promise.resolve({ exemplare, fehler: null }),
    Portfolio: () =>
      Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
    Einstellungen: () => Promise.resolve({ einstellungen: [], fehler: null }),
    MarkenJeObjekt: () => Promise.resolve({ marken: {}, fehler: null }),
    Markenkatalog: () => Promise.resolve({ marken: [], fehler: null }),
    GezeigtesProfil: () => Promise.resolve({ profil: null, fehler: null }),
    Profilvorgaben: () => Promise.resolve({ vorgaben: [], fehler: null }),
    ...teile,
  })
}

// zeigeSammlung zeichnet die Seite in deutscher Sprache; `bestandGeaendert`
// hört, was die Seite dem Fenster als eigenes Schreiben meldet (R3, #2366).
export function zeigeSammlung(bestandGeaendert: () => void = vi.fn()) {
  return render(
    <Sprachraum sprache="de">
      <Sammlung bestandGeaendert={bestandGeaendert} />
    </Sprachraum>,
  )
}

// oeffneAufnehmen öffnet die Schublade mit dem Formular über den Kopf.
export async function oeffneAufnehmen(): Promise<HTMLElement> {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.aufnehmenOeffnen }),
  )

  return screen.findByRole('dialog', { name: texte.neuesTitel })
}

// oeffneEinfuhr öffnet die Schublade mit Listeneinfuhr und Ausgabe.
export async function oeffneEinfuhr(): Promise<HTMLElement> {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.einfuhrOeffnen }),
  )

  return screen.findByRole('dialog', { name: texte.einfuhrTitel })
}

// felder zählt die Felder in einem Bereich wie das Drehbuch `uirevue`:
// Eingaben, Auswahlen, Textfelder, Schalter.
//
// **Ohne die Haken der Liste**: sie sind die Wahl einer Zeile und kein
// Formular — das Drehbuch zählt sie mit, und gezeichnet werden nur die
// Zeilen im Bild. **Und ohne Verborgenes**: was in einem zugeklappten
// Kasten oder einer verborgenen Schublade steht, sieht niemand.
export function felder(bereich: ParentNode = document): number {
  return [
    ...bereich.querySelectorAll<HTMLElement>(
      'input:not([type=hidden]), select, textarea, [role=combobox], [role=switch]',
    ),
  ].filter(
    (feld) =>
      feld.closest('[role=grid]') === null &&
      feld.closest('[hidden]') === null &&
      feld.closest('[inert]') === null &&
      !verborgen(feld),
  ).length
}

// verborgen sagt, ob ein Vorfahr mit `display: none` dasteht — so verbirgt
// `<Activity>` eine Schublade, die zu ist.
function verborgen(element: HTMLElement): boolean {
  for (
    let knoten: HTMLElement | null = element;
    knoten !== null;
    knoten = knoten.parentElement
  ) {
    if (knoten.style.display === 'none') {
      return true
    }
  }

  return false
}

// hauptknoepfeJeRegion gibt die höchste Zahl von Hauptknöpfen in einer
// Region — R1 aus `docs/UIREVIEW.md`, gezählt wie im Drehbuch.
export function hauptknoepfeJeRegion(bereich: ParentNode = document): number {
  const region =
    'section, form, aside, [role=region], [role=tabpanel], [role=dialog]'
  const haupt = [
    ...bereich.querySelectorAll<HTMLElement>('button, [role=button]'),
  ].filter(
    (knopf) =>
      /(^|\s)bg-marke(\s|$)/.test(knopf.className) &&
      knopf.closest('[inert]') === null &&
      !verborgen(knopf),
  )
  const zaehlung = new Map<Element | null, number>()

  for (const knopf of haupt) {
    const ort = knopf.closest(region)
    zaehlung.set(ort, (zaehlung.get(ort) ?? 0) + 1)
  }

  return Math.max(0, ...zaehlung.values())
}
