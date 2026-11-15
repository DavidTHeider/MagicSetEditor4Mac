// **Ein Satz aus dem Katalog, mit seinen Hervorhebungen** —
// #699.
//
// # Der Befund
//
// Die Hilfeseite zeigte ihre Sätze wörtlich — samt der Sternchen:
//
//	Sagt, **welche Felder** eine Karte hat. Ein `.mse-game` erklärt sie…
//
// Die Auszeichnung war beim Schreiben gemeint und kam nie an. Ein Leser sieht
// stattdessen zwei Zeichen, die niemand erklärt — genau die Sorte Rest aus der
// Werkstatt, die eine Anwendung unfertig aussehen lässt.
//
// # Zwei Formen und keine dritte
//
// `**fett**` und `` `code` ``. Mehr braucht kein Satz dieses Katalogs, und
// mehr wäre eine Auszeichnungssprache, die jemand pflegen müsste. Was nicht
// paart, bleibt stehen, wie es dasteht: ein einzelnes Sternchen ist ein
// Sternchen.
import type { ReactNode } from 'react'

// TEILE trifft `**fett**` und `` `code` `` — **mit Klammern**, damit `split`
// die Treffer behält.
const TEILE = /(\*\*[^*]+\*\*|`[^`]+`)/g

export function Satz({ text }: { text: string }) {
  return <>{stuecke(text)}</>
}

// stuecke zerlegt den Satz in Text und Hervorhebungen.
//
// **Der Schlüssel ist die Stelle, nicht der Text.** Zweimal dasselbe Wort in
// einem Satz gäbe sonst denselben Schlüssel — und React zeichnete das zweite
// nicht.
function stuecke(text: string): ReactNode[] {
  return text.split(TEILE).map((teil, stelle) => {
    if (teil.startsWith('**') && teil.endsWith('**')) {
      return <strong key={stelle}>{teil.slice(2, -2)}</strong>
    }

    if (teil.startsWith('`') && teil.endsWith('`')) {
      return (
        <code
          key={stelle}
          className="rounded-sm bg-betont px-1 font-mono text-[0.9em]"
        >
          {teil.slice(1, -1)}
        </code>
      )
    }

    return teil
  })
}
