// **Wer auf einen Ruf wartet, hat noch nicht auf die Antwort gewartet** —
// Phase 4.8 (#1321).
//
// # Die Gestalt, gegen die dieser Wächter steht
//
//	await waitFor(() => {
//	  expect(irgendeinSpion).toHaveBeenCalled()
//	})
//
//	expect(screen.getByText('…')).toBeInTheDocument()
//
// `toHaveBeenCalled` trifft zu, **sobald gerufen wurde**. Die Antwort ist dann
// noch unterwegs, und was sie bewirkt, steht noch nicht da. Der Griff darunter
// geht ins Leere — oder eben nicht, je nachdem, wie beschäftigt der Rechner
// ist. Das ist §3, „eine Zusicherung, die schon vor der Antwort zutrifft".
//
// # Warum es ihn erst seit dem 2026-09-02 gibt
//
// **Weil die Vorrichtung es bis dahin zugedeckt hat.** `test/api.ts`
// antwortete synchron; zwischen Ruf und Antwort lag nichts. Seit sie eine
// Makroaufgabe später antwortet — so wie die IPC im Betrieb —, ist die Lücke
// echt, und ein voller Lauf unter Last hat sie prompt getroffen.
//
// Gefunden hat der Wächter beim ersten Lauf **sieben** Stellen. Eine davon war
// richtig; sie steht unten mit ihrem Grund.
//
// # Was er nicht hält
//
// **Er liest Zeilen, keine Bedeutung.** Wer die Zusicherung anders schreibt —
// über eine Hilfsfunktion, über `expect.poll`, über einen eigenen Namen für
// `waitFor` —, kommt an ihm vorbei. Und er sagt nichts darüber, ob eine
// Zusicherung, die er durchlässt, das Richtige prüft.
//
// **Er sieht nur `screen.get…`.** Ein Griff über `within(…).get…` oder über
// einen Umschlag, der vorher gegriffen wurde, kommt durch — derselbe Griff,
// nur über einen anderen Namen. Bis zum 2026-09-23 sah er sogar nur den in
// einer Zusicherung (#1809).
import { describe, expect, it } from 'vitest'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'

// erlaubt sind die Stellen, an denen der Ruf der richtige Anker ist.
//
// **Eine Aufzählung mit Grund.** Wer hier etwas einträgt, sagt, warum die
// Antwort an dieser Stelle ausdrücklich **nicht** abgewartet werden soll.
const erlaubt: Record<string, string> = {
  'src/seiten/stilvorschau.test.tsx':
    'Die Zusage lautet „das alte Bild steht, während das neue rechnet". Die ' +
    'Antwort hängt dort absichtlich — sie wird erst am Ende des Tests von ' +
    'Hand aufgelöst. Wer sie abwartete, prüfte das Gegenteil.',

  // **Zwei mehr seit #1809**, beide von derselben Art: der Test hält eine
  // Antwort mit Absicht zurück und handelt, während sie aussteht. Gegriffen
  // wird jeweils nach etwas, das von dieser Antwort nicht abhängt. **Ein
  // dritter, `deckloeschen.test.tsx`, ist seit AP-6 (#2363) keiner mehr**:
  // er greift über das Menü „…“ mit `findByRole`.
  'src/seiten/kartennotizenseite.test.tsx':
    'Das Schreiben hängt mit Absicht und scheitert erst, nachdem die Karte ' +
    'gewechselt ist. Der Griff gilt der Kartenliste, die von dieser Antwort ' +
    'nicht abhängt.',
  'src/seiten/turniervorschlagslauf.test.tsx':
    'Der Paarungsvorschlag hängt mit Absicht und scheitert erst nach dem ' +
    'Wechsel zu einem anderen Turnier, an beiden Stellen. Der Griff gilt der ' +
    'Turnierliste, die von dieser Antwort nicht abhängt.',
}

// istWarten sagt, ob eine Zeile auf eine **Antwort** wartet.
//
// **Und nicht, ob sie ein `await` enthält.** Eine Handlung — tippen, klicken,
// ein Ereignis feuern — wartet auf ihren eigenen Abschluss; was der vorige Ruf
// bewirkt hat, steht danach genauso wenig da wie davor.
//
// **Die Aufzählung ist eng und nicht klug.** Sie kennt die vier Formen, in
// denen in diesem Fenster auf eine Antwort gewartet wird; alles andere gilt
// als Handlung. Wer eine fünfte einführt, trägt sie hier ein — und merkt es
// daran, dass der Wächter seine Stelle meldet. Das ist die richtige Richtung:
// im Zweifel melden.
function istWarten(zeile: string): boolean {
  if (/\b(userEvent|fireEvent)\b|\.click\(|\.type\(|\.tab\(/.test(zeile)) {
    return false
  }

  return /await\s+(waitFor\(|screen\.find|antwortAbwarten\(|effekteNachholen\()/.test(
    zeile,
  )
}

// dateien sammelt alle Testdateien unter `src/`.
function dateien(ordner: string): string[] {
  return readdirSync(ordner).flatMap((name) => {
    const pfad = join(ordner, name)

    if (statSync(pfad).isDirectory()) {
      return dateien(pfad)
    }

    return /\.test\.tsx?$/.test(name) ? [pfad] : []
  })
}

// stellen findet die Gestalt in einer Datei und gibt ihre Zeilennummern.
function stellen(inhalt: string): number[] {
  const zeilen = inhalt.split('\n')
  const gefunden: number[] = []

  for (let i = 0; i < zeilen.length; i += 1) {
    if (!zeilen[i]?.includes('await waitFor(')) {
      continue
    }

    // Der Rumpf des `waitFor` bis zu seiner schließenden Klammer.
    //
    // **Höchstens fünfzehn Zeilen weit.** Ein `waitFor` mit eigener Frist
    // endet auf `}, { timeout: … })` und trifft dieses Muster nicht; ohne
    // Grenze liefe die Suche bis zum Dateiende und urteilte über einen
    // Rumpf, den es nicht gibt. Wer so schreibt, kommt an diesem Wächter
    // vorbei — das steht oben unter »was er nicht hält«.
    let ende = i
    const grenze = Math.min(i + 15, zeilen.length)

    while (ende < grenze && !/^\s*\}\)\s*$/.test(zeilen[ende] ?? '')) {
      ende += 1
    }

    if (ende === grenze) {
      continue
    }

    const rumpf = zeilen.slice(i, ende + 1).join('\n')

    if (!rumpf.includes('toHaveBeenCalled')) {
      continue
    }

    // **Was direkt danach kommt, ist die Frage** — Leerzeilen und Kommentare
    // übersprungen.
    const danach = zeilen
      .slice(ende + 1, ende + 6)
      .filter((zeile) => zeile.trim() !== '' && !zeile.trim().startsWith('//'))

    // **Ein Warten dazwischen ist die Entwarnung — eine Handlung nicht.**
    //
    // Wer nach dem Ruf die Antwort abwartet, hat getan, was hier verlangt
    // wird; ohne diese Zeile meldete der Wächter jede behobene Stelle erneut.
    //
    // **Hier stand bis zum 2026-09-16 jedes `await`**, und das war zu
    // großzügig: `await userEvent.click(…)` wartet auf **sich selbst** und
    // nicht auf die Antwort des vorigen Rufs. Damit galt eine Stelle als
    // entwarnt, an der zwischen Ruf, Wirkung und der nächsten Handlung gar
    // keine Reihenfolge feststand — ein Wächter, der die halbe Zusage hält
    // (§10). Angestrichen hat es ein Lauf, der genau so geschrieben war und
    // einmal in seine Frist lief.
    const beimWarten = danach.findIndex(istWarten)
    const bisZumWarten = danach
      .slice(0, beimWarten === -1 ? 3 : beimWarten + 1)
      .join('\n')

    // **Jeder Griff, nicht nur der in einer Zusicherung** (#1809). Hier
    // stand `expect(screen.get…`, und `await userEvent.click(screen.get…)`
    // kam durch: derselbe Griff ins DOM, nur als Ziel einer Handlung. Rot
    // wurde er am 2026-09-23 in `make ci` — die Seite lud nach, und der Knopf
    // stand einen Augenblick nicht da.
    if (bisZumWarten.includes('screen.get')) {
      gefunden.push(i + 1)
    }
  }

  return gefunden
}

describe('Warten auf den Ruf statt auf die Antwort', () => {
  it('kommt in keiner Testdatei vor', () => {
    const befunde: string[] = []

    for (const pfad of dateien('src')) {
      const relativ = pfad.replace(/\\/g, '/')

      // **Nicht sich selbst.** Diese Datei zeigt die Gestalt im Kommentar,
      // damit man sie erkennt; ein Instrument, das sich selbst misst, ist
      // heute schon zweimal danebengegangen (§10).
      if (relativ === 'src/wartemuster.test.ts' || relativ in erlaubt) {
        continue
      }

      for (const zeile of stellen(readFileSync(pfad, 'utf8'))) {
        befunde.push(
          `${relativ}:${String(zeile)} — nach dem Warten auf einen Ruf wird ` +
            'sofort ins DOM gegriffen; die Antwort ist dann noch unterwegs. ' +
            'Warte auf die Wirkung (findBy…, waitFor über die Zusicherung) ' +
            'oder auf die Antwort (antwortAbwarten).',
        )
      }
    }

    expect(befunde).toEqual([])
  })

  // **Die Gegenprobe zur Aufzählung**: sie darf nicht ins Leere zeigen, und
  // jeder Eintrag braucht einen Grund.
  it('führt nur Ausnahmen, die es gibt und die begründet sind', () => {
    for (const [pfad, grund] of Object.entries(erlaubt)) {
      expect(stellen(readFileSync(pfad, 'utf8')).length).toBeGreaterThan(0)
      expect(grund.length).toBeGreaterThan(60)
    }
  })
})
