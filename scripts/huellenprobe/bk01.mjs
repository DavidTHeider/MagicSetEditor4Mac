// **Ein abgebrochener Ruf am nativen Fenster** — BK-01.AK-03
// (#2109).
//
// Ein **fremder Schreiber** hält die Schreibsperre des Bestands — so, wie es
// `mse`, MCP oder REST täten. Das Fenster ruft `EinstellungSetzen` und bricht
// den Ruf nach 300 ms ab (`CancellablePromise.cancel` aus `@wailsio/runtime`,
// über die Gestalt aus `anbindung.ts`). Gemessen wird zweierlei:
//
// - **Das Fenster verlässt das Warten**: die Zusage ist danach abgewiesen und
//   nicht mehr offen.
// - **Go hat aufgegeben**: nachdem der fremde Schreiber die Sperre freigibt,
//   steht der alte Wert noch da. Vor BK-01 wartete der Ruf mit dem Kontext
//   des Fensters bis zu fünf Sekunden (`busy_timeout(5000)`) — und schrieb,
//   sobald die Sperre fiel.
//
// **Was sie nicht belegt**: den Abbruch beim Schließen des Fensters
// (`CancelWindowCalls` in Wails) — derselbe Kontext, aber ein anderer
// Auslöser, und das Fenster ist danach nicht mehr da, um es zu zeigen.
import { DatabaseSync } from 'node:sqlite'

const bindung = (p, ausdruck) => p.seite(`const b = window.go.app.Bindungen; return await ${ausdruck}`)

async function bildkonflikt(p) {
  const antwort = await bindung(p, 'b.Einstellungen()')

  return antwort.einstellungen.find((e) => e.schluessel === 'bildkonflikt')
}

const ruhe = (ms) => new Promise((fertig) => { setTimeout(fertig, ms) })

export default async function bk01(p) {
  const ort = p.ort()
  await p.starte({ ort })

  const vorher = await bildkonflikt(p)
  const neu = vorher.werte.find((w) => w !== vorher.wert)
  p.zeige(`bildkonflikt: ${vorher.wert}, gewünscht wird ${neu}`)

  p.notiz('1 — ohne fremde Sperre schreibt derselbe Weg, und wieder zurück')
  p.pruefe('er schreibt', !(await bindung(p, `b.EinstellungSetzen('bildkonflikt', '${neu}')`)).fehler)
  p.pruefe('der Wert steht', (await bildkonflikt(p)).wert === neu)
  p.pruefe('und zurück', !(await bindung(p, `b.EinstellungSetzen('bildkonflikt', '${vorher.wert}')`)).fehler)

  p.notiz('2 — ein fremder Schreiber hält die Sperre; das Fenster ruft und bricht nach 300 ms ab')
  const fremd = new DatabaseSync(ort.bestand)
  fremd.exec('BEGIN IMMEDIATE')

  let gemessen
  try {
    gemessen = await p.seite(`
      const ruf = window.go.app.Bindungen.EinstellungSetzen('bildkonflikt', '${neu}')
      let ende = 'offen'
      ruf.then(() => { ende = 'erfüllt' }, (grund) => { ende = 'abgewiesen (' + (grund?.name ?? String(grund)) + ')' })
      await new Promise((fertig) => setTimeout(fertig, 300))
      const vorAbbruch = ende
      const abbrechbar = typeof ruf.cancel === 'function'
      if (abbrechbar) ruf.cancel()
      await new Promise((fertig) => setTimeout(fertig, 100))
      return { vorAbbruch, abbrechbar, nachAbbruch: ende }
    `)
    p.zeige(`vor dem Abbruch ${gemessen.vorAbbruch}, abbrechbar ${gemessen.abbrechbar}, danach ${gemessen.nachAbbruch}`)
    p.pruefe('der Ruf wartete auf die Sperre', gemessen.vorAbbruch === 'offen')
    p.pruefe('die Zusage ist abbrechbar', gemessen.abbrechbar)
    p.pruefe('das Fenster wartet nicht mehr', gemessen.nachAbbruch.startsWith('abgewiesen'))

    // **Die Sperre bleibt noch eine Weile** — lange genug, dass ein Ruf
    // alter Art weiter wartete, und kürzer als `busy_timeout`.
    await ruhe(1500)
  } finally {
    fremd.exec('ROLLBACK')
    fremd.close()
  }

  p.notiz('3 — nach der Freigabe: geschrieben hätte nur ein Ruf, der weiterwartete')
  await ruhe(1500)
  const danach = await bildkonflikt(p)
  p.zeige(`bildkonflikt danach: ${danach.wert}`)
  p.pruefe('der abgebrochene Ruf hat nicht geschrieben', danach.wert === vorher.wert)

  p.notiz('4 — und derselbe Weg geht danach wieder')
  p.pruefe('er schreibt', !(await bindung(p, `b.EinstellungSetzen('bildkonflikt', '${neu}')`)).fehler)
  p.pruefe('der Wert steht', (await bildkonflikt(p)).wert === neu)

  p.pruefe('⌘Q beendet', await p.beende())
}
