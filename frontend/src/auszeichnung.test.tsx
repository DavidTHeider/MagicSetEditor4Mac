// **Keine rohe Auszeichnung im Fenster** — Werkstück 8 von Phase 4.8
// (#1321).
//
// # Warum es diesen Test gibt
//
// Die Sprachkataloge benutzen `**…**` und `` `…` ``; gezeichnet wird das von
// `Satz`. Wer einen Katalogtext **roh** in JSX setzt, bekommt die Sternchen
// wörtlich ins Fenster — und niemand sieht es, weil kein Test den Text so
// liest, wie er dasteht.
//
// **Es ist schon einmal geschehen** (#1100):
// unter „Nach neuen Fassungen sehen" stand wörtlich `**aus**`. Behoben wurde
// damals die **Stelle**. In Phase 4.8 kam es an zwei neuen Stellen wieder —
// `Dienstfragen` und `Zugaenge` —, und gefunden hat es wieder die
// Sichtprüfung am laufenden Fenster.
//
// **Deshalb steht hier die Klasse und nicht der Fall** (§10): jede Seite aus
// `navigation.ts`, in jeder Sprache, ihr ganzer sichtbarer Text.
//
// # Was er nicht hält
//
// **Nur, was eine Seite ohne Zutun zeichnet.** Ein Text, der erst nach einem
// Klick erscheint, steht hier nicht drin — dafür gibt es die Tests der
// jeweiligen Seite. Und er sagt nichts darüber, ob der ausgezeichnete Text
// **richtig** ausgezeichnet ist; er sieht nur, dass die Marken nicht mehr
// dastehen.
//
// # Gewartet wird, bevor gelesen wird
//
// **Der erste Anlauf las synchron** — und sah damit nur die **erste**
// Zeichnung. Der Bereich „Zugangsdaten" holt seinen Stand über eine Bindung
// und steht deshalb erst eine Makroaufgabe später da; er war für diesen
// Wächter unsichtbar, und seine rohe Auszeichnung wäre durchgekommen.
// Aufgefallen ist es an einer Gegenprobe, die **nichts verschob** (§10).
import { describe, expect, it } from 'vitest'

import { belegeProbeset } from './test/seitenvorrichtung'
import { geladenesDokumentDer } from './test/geladeneseite'
import { seiten } from './navigation'
import { sprachen } from './sprache/kataloge'

// marken sind die Auszeichnungen, die `Satz` versteht.
//
// **Der Zeilenumbruch ist ausgenommen**: `\n` ist keine Marke, sondern ein
// Zeichen, das jeder Text tragen darf.
const marken: { name: string; muster: RegExp }[] = [
  { name: 'Fettung `**…**`', muster: /\*\*[^*]+\*\*/ },
  { name: 'Schreibmaschine `` `…` ``', muster: /`[^`]+`/ },
]

describe('Rohe Auszeichnung', () => {
  for (const sprache of sprachen) {
    it.each(seiten)(
      `steht auf $schluessel in ${sprache} nirgends im Text`,
      async (seite) => {
        belegeProbeset()

        // **Erst muss die Seite fertig geladen sein.** Was eine Bindung
        // nachliefert, steht sonst nicht da — siehe oben. **Und das ganze
        // Dokument** (G02): sieben Setseiten stehen in einer Schublade, und
        // die zeichnet Mantine in ein Portal neben `container`.
        const dokument = await geladenesDokumentDer(seite.schluessel, sprache)

        const text = dokument.textContent

        for (const marke of marken) {
          const treffer = marke.muster.exec(text)

          expect(
            treffer,
            treffer === null
              ? ''
              : `auf „${seite.schluessel}" steht ${marke.name} wörtlich im ` +
                  `Text: „${treffer[0]}" — der Katalogtext gehört durch ` +
                  `\`Satz\`, nicht roh ins JSX`,
          ).toBeNull()
        }
      },
    )
  }
})
