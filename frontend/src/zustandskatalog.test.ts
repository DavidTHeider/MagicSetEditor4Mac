// **Der Zustandskatalog hält gegen den Quelltext** — UX-08.AK-02
// (#1680).
//
// `docs/UX-ARCHITEKTUR.md` führt unter „Zustände — was jeder zeigt" zwölf
// Zustände, je mit dem Baustein, der sie trägt. **Dieser Lauf hält**, dass es
// jeden genannten Baustein gibt und dass ihn eine andere Datei benutzt — und
// die Zusage darunter: **es gibt keinen Toast**. Keine Bibliothek dafür steht
// in den Abhängigkeiten, und jede Uhr im Quelltext ist hier mit ihrem Zweck
// genannt; eine neue verlangt eine Entscheidung, statt still eine Meldung
// wegzuräumen.
//
// **Was er nicht hält**: dass jede Stelle, die einen Fehler zeigt, schon den
// Baustein nimmt. Das ist eine Frage an die Durchsicht und an die Tafel.
import { existsSync, readFileSync } from 'node:fs'
import { basename, join } from 'node:path'

import { dateienUnter } from './test/dateien'
import { describe, expect, it } from 'vitest'

const quellen = join(import.meta.dirname)

function tafel(): { zustand: string; baustein: string }[] {
  const datei = readFileSync(
    join(quellen, '../../docs/UX-ARCHITEKTUR.md'),
    'utf8',
  )
  const abschnitt =
    datei.split('## Zustände — was jeder zeigt')[1]?.split('\n## ')[0] ?? ''

  return [...abschnitt.matchAll(/^\| ([^|]+) \| [^|]+ \| `([^`]+)` \|/gm)].map(
    (zeile) => ({ zustand: (zeile[1] ?? '').trim(), baustein: zeile[2] ?? '' }),
  )
}

// quelltext sind die Dateien, die das Fenster ausliefert — ohne Tests und
// ohne Vorrichtungen.
function quelltext(): string[] {
  return dateienUnter(quellen, '**/*.{ts,tsx}', {
    ohne: ['**/*.test.ts', '**/*.test.tsx', 'test/**'],
  })
}

// **Die Uhren, und wozu jede tickt.** Keine räumt eine Meldung weg — wer eine
// dazulegt, trägt sie hier ein und sagt, warum.
const uhren: Record<string, string> = {
  'bausteine/Dialog.tsx':
    'wartet einen Umlauf, bis feststeht, ob jemand die Taste verbraucht hat',
  'bausteine/Schublade.tsx':
    'wartet einen Umlauf, bis feststeht, ob jemand die Taste verbraucht hat',
  'bausteine/Tooltip.tsx': 'zeigt die Hilfe erst nach einer kurzen Frist',
  'bausteine/fokus.ts':
    'schließt ein Menü erst, wenn die Tabulatortaste den Fokus weitergegeben hat',
  'haken/useVerzoegert.ts':
    'lässt eine Eingabe sich setzen, bevor gesucht wird',
  'haken/useListenbilder.ts':
    'holt die Bildchen der Kartenliste erst, wenn das Rollen 120 ms ruht (E07, #2279)',
  'haken/useBestandsstand.ts': 'fragt nach dem Bestand, solange jemand zusieht',
  'haken/useDeckhandlung.ts':
    'zählt den Anlass zum Teilen einen Umlauf später, wenn die Menüleiste den Navigator erst aufklappt — die Zeile entsteht im Umlauf davor (#2469)',
  'haken/useKartengriff.ts':
    'wartet nach einem Klick auf eine Karte mit Zweithandlung 250 ms, ob ein Doppelklick folgt — eine Geste, keine Meldung (P4, #2372)',
  'haken/useBetriebsband.ts':
    'fragt alle 30 Sekunden nach Platz, Sicherung, Dienst und Webprozess, solange jemand zusieht (OPS02, #2348)',
  'haken/useSicherungsstand.ts':
    'fragt nach dem Stand der Sicherung, solange der Plan fällig ist und noch nichts meldete (OPS01, #2347)',
  'haken/usePartienebenbei.ts':
    'lässt einen Zeiger am Tisch nach seiner Frist verblassen — eine Präsenz, keine Meldung',
  'seiten/Hintergrundaufgaben.tsx':
    'fragt nach den Aufträgen, solange einer läuft',
  'tisch/pixinaht.ts': 'räumt die Hover-Zustände von Pixi auf',
}

describe('Der Zustandskatalog', () => {
  // **Sieben nach dem Artboard *BelegZustaende*, fünf nach dem Kriterium** —
  // „Wiederholen" aus AK-02 steht beim Fehler, denn es ist dessen Handlung.
  it('nennt die sieben Zustände des Artboards und die fünf des Kriteriums', () => {
    expect(tafel().map((zeile) => zeile.zustand)).toEqual([
      'Laden',
      'Leer',
      'Gefiltert leer',
      'Fehler',
      'Ohne Netz',
      'Nur lesen',
      'Teilerfolg',
      'Fortschritt',
      'Abbruch',
      'Konflikt',
      'Erfolg',
      'Hintergrundauftrag',
    ])
  })

  it.each(tafel())(
    '$zustand: $baustein liegt da und wird benutzt',
    ({ baustein }) => {
      expect(existsSync(join(quellen, baustein)), baustein).toBe(true)

      const name = basename(baustein).replace(/\.tsx?$/, '')
      const nutzer = quelltext()
        .filter((datei) => datei !== baustein)
        .filter((datei) =>
          readFileSync(join(quellen, datei), 'utf8').includes(`/${name}'`),
        )

      expect(nutzer.length, `${baustein} benutzt niemand`).toBeGreaterThan(0)
    },
  )

  it('kennt keine Bibliothek für Toasts', () => {
    const paket = JSON.parse(
      readFileSync(join(quellen, '../package.json'), 'utf8'),
    ) as { dependencies?: Record<string, string> }

    const abhaengig = Object.keys(paket.dependencies ?? {})

    for (const toast of [
      '@mantine/notifications',
      'notistack',
      'react-hot-toast',
      'react-toastify',
      'sonner',
    ]) {
      expect(abhaengig, toast).not.toContain(toast)
    }
  })

  it('nennt jede Uhr mit ihrem Zweck', () => {
    const tickend = quelltext()
      .filter((datei) =>
        /\bset(Timeout|Interval)\(/.test(
          readFileSync(join(quellen, datei), 'utf8'),
        ),
      )
      .sort()

    expect(tickend).toEqual(Object.keys(uhren).sort())
  })
})
