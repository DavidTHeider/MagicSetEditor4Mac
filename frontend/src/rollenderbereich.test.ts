// **Was rollt, erreicht die Tastatur** — angestrichen vom Zweitblick über
// Phase 6.3 (G10) und vom Nachlauf (N10): zwei `<pre>` rollten waagerecht
// und hatten keinen Tabhalt, und kein Wächter sah es. Ein rollender Bereich,
// den die Tastatur nicht erreicht, ist für manche gar nicht da (WCAG 2.1.1).
//
// **Das Muster steht in `Diagnose.tsx`**: `role="region"` mit Namen und
// `tabIndex={0}` — `jsx-a11y` lässt einen Tabhalt nur an einer Region zu.
//
// # Was er nicht hält
//
// **Nur `<pre>`**: ein `<div>` oder `<ul>`, das rollt, sieht er nicht —
// bei ihnen sagt die Klasse allein nicht, ob sie rollen, weil der Inhalt es
// entscheidet. **Und einen `<pre>`, der für Vorleser verborgen ist**
// (`aria-hidden`), lässt er aus: er doppelt ein Feld, das den Fokus hat.
import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

import { describe, expect, it } from 'vitest'

const quelle = import.meta.dirname

function bauteile(ort: string): string[] {
  return readdirSync(ort).flatMap((name) => {
    const pfad = join(ort, name)
    if (statSync(pfad).isDirectory()) {
      return bauteile(pfad)
    }

    return name.endsWith('.tsx') && !name.includes('.test.') ? [pfad] : []
  })
}

// rollendeOhneHalt gibt jeden `<pre>`, der rollt und die Tastatur nicht
// erreicht — das Element bis zu seinem ersten `>`.
function rollendeOhneHalt(text: string): string[] {
  return [...text.matchAll(/<pre\b[^>]*>/g)]
    .map((treffer) => treffer[0])
    .filter(
      (element) =>
        /\boverflow-(?:auto|scroll|x-auto|y-auto|x-scroll|y-scroll)\b/.test(
          element,
        ) &&
        !element.includes('aria-hidden="true"') &&
        !(
          element.includes('role="region"') &&
          element.includes('tabIndex={0}') &&
          element.includes('aria-label=')
        ),
    )
}

describe('Ein rollender Bereich', () => {
  it('erkennt einen ohne Halt und lässt einen mit Halt durch', () => {
    expect(
      rollendeOhneHalt('<pre className="overflow-x-auto">x</pre>'),
    ).toHaveLength(1)
    expect(
      rollendeOhneHalt(
        '<pre role="region" tabIndex={0} aria-label={a} className="overflow-auto">x</pre>',
      ),
    ).toEqual([])
    expect(
      rollendeOhneHalt('<pre aria-hidden="true" className="overflow-auto">'),
    ).toEqual([])
    expect(rollendeOhneHalt('<pre className="font-mono">x</pre>')).toEqual([])
  })

  it('erreicht überall die Tastatur', () => {
    const funde = bauteile(quelle).flatMap((pfad) =>
      rollendeOhneHalt(readFileSync(pfad, 'utf8')).map(() =>
        relative(quelle, pfad),
      ),
    )

    expect(funde).toEqual([])
  })
})
