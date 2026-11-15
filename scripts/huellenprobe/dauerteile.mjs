// **Welcher Teil des Dauerbetriebs wächst?** — B04
// (#2223).
//
// Der Zwei-Stunden-Lauf (`dauerbetrieb`) fand einen stetigen Anstieg: der
// Webprozess wuchs um etwa 32 MB, die Anwendung um etwa 13 MB je Minute.
// Er fährt alle Teile reihum und kann nicht sagen, welcher wächst. Dieses
// Drehbuch fährt sie **nacheinander in einer Sitzung**: erst Leerlauf als
// Grundlinie, dann jeden Teil für sich `MSE_DAUER_PHASE_MIN` Minuten
// (Vorgabe 6), und rechnet je Phase die Steigung von Anwendungs- und
// Webspeicher — eine Ausgleichsgerade über die Proben ohne die erste halbe
// Minute. **Mit `+` werden Teile zu einer Phase**: `MSE_DAUER_TEILE=
// bearbeiten+navigation+sammlung+tisch+ausgabe` fährt die gemischte Runde des
// Zwei-Stunden-Laufs, nur mit Heap-Probe.
//
// **Was es nicht belegt**: dass ein Teil allein wächst. Speicher, den eine
// Phase belegt und den die Speicherbereinigung erst später freigibt, kann in
// der nächsten auftauchen; ein Befund wird deshalb mit einem Lauf **nur**
// dieses Teils bestätigt (`MSE_DAUER_TEILE`).
import { fahreDauer } from './dauerbetrieb.mjs'

// **Nur, wer es nennt** — eine Diagnose und kein Beleg einer Zusage.
export const nurEinzeln = true

// phasenAus liest Phasen und Heap-Probe aus der Umgebung.
//
// **`MSE_DAUER_TEILE=` (leer) heißt: nur der Leerlauf.** **`MSE_DAUER_HEAP=aus`
// schaltet die Heap-Probe ab** — sie vergrößert den Puffer der
// Ressourcenzeiten auf eine Million Einträge und liest ihn bei jeder Probe
// ganz, und jeder IPC-Ruf der Hülle ist ein `fetch`. Ein Lauf mit ihr misst
// sich mit (E06, #2278); der Zwei-Stunden-Lauf (`dauerbetrieb`) hat sie nie an.
export function phasenAus(umgebung) {
  const phasenMin = Number(umgebung.MSE_DAUER_PHASE_MIN ?? '6')
  const leerMin = Number(umgebung.MSE_DAUER_LEER_MIN ?? String(phasenMin))
  const liste = umgebung.MSE_DAUER_TEILE ?? 'bearbeiten,navigation,sammlung,tisch,ausgabe'
  const gewaehlt = liste === '' ? [] : liste.split(',')

  return {
    phasen: [
      { name: 'Leerlauf', minuten: leerMin, teile: [] },
      ...gewaehlt.map((teil) => ({ name: teil, minuten: phasenMin, teile: teil.split('+') })),
    ],
    heap: umgebung.MSE_DAUER_HEAP !== 'aus',
  }
}

// steigung ist die Ausgleichsgerade in MB je Minute.
export function steigung(punkte) {
  if (punkte.length < 3) return null
  const n = punkte.length
  const mx = punkte.reduce((s, [x]) => s + x, 0) / n
  const my = punkte.reduce((s, [, y]) => s + y, 0) / n
  const zaehler = punkte.reduce((s, [x, y]) => s + (x - mx) * (y - my), 0)
  const nenner = punkte.reduce((s, [x]) => s + (x - mx) ** 2, 0)

  return nenner === 0 ? null : zaehler / nenner
}

// **Die Einschwingzeit steht vor dem Lauf fest** — Befund 9 des Zweitblicks
// über E/F. Der Leerlauf beginnt mit dem Start, und der Webprozess gibt nach
// knapp zwei Minuten über 100 MB zurück; ein erster Griff in einen Teil lädt
// einmalig nach. Gezählt wird die Steigung **ab Minute 3** jeder Phase, der
// Gesamtfit ab der halben Minute steht daneben. Wer die Zahl nach einem Lauf
// wählte, suchte sich die Steigung aus.
export const EINSCHWINGEN_MIN = 3

// phasensteigung rechnet die Steigungen einer Phase — nach dem Einschwingen
// und über alles.
export function phasensteigung(eigene, haufen, anfang) {
  const ab = (grenze) => {
    const spaeter = eigene.filter((x) => x.minute >= grenze)

    return {
      web: steigung(spaeter.map((x) => [x.minute, x.web])),
      app: steigung(spaeter.map((x) => [x.minute, x.app])),
      goHeap: steigung(haufen.filter((x) => x.minute >= grenze).map((x) => [x.minute, x.heap])),
      proben: spaeter.length,
    }
  }

  return { eingeschwungen: ab(anfang + EINSCHWINGEN_MIN), gesamt: ab(anfang + 0.5) }
}

export default async function dauerteile(p) {
  const { phasen, heap } = phasenAus(process.env)
  p.notiz(heap ? 'mit Heap-Probe' : 'ohne Heap-Probe (MSE_DAUER_HEAP=aus)')
  const { proben, haufen, fehler } = await fahreDauer(p, phasen, 15000, heap)

  p.notiz(`Steigung je Phase in MB je Minute — ab Minute ${String(EINSCHWINGEN_MIN)} der Phase, in Klammern über alles ab der halben Minute`)
  const zahl = (wert) => wert?.toFixed(1) ?? '–'
  for (const phase of phasen) {
    const eigene = proben.filter((x) => x.aufgabe.startsWith(`${phase.name} — `) || (phase.name === 'Leerlauf' && x.aufgabe === 'Start'))
    const { eingeschwungen: e, gesamt: g } = phasensteigung(eigene,
      haufen.filter((x) => x.phase === phase.name), eigene[0]?.minute ?? 0)
    p.pruefe(`${phase.name}: Web ${zahl(e.web)} (${zahl(g.web)}) MB/min, Anwendung ${zahl(e.app)} (${zahl(g.app)}) MB/min, davon Go-Heap ${zahl(e.goHeap)} (${zahl(g.goHeap)}) MB/min`,
      e.web !== null && e.app !== null, `${String(e.proben)} Proben`)
  }

  p.pruefe('jede Runde gelang fachlich', fehler.length === 0, `${String(fehler.length)}: ${fehler.slice(0, 5).join('; ')}`)
  p.zeilen.push('', `Proben (${String(proben.length)}, alle 15 s): ${JSON.stringify(proben)}`, '', `Go-Heap je Runde: ${JSON.stringify(haufen)}`)
  p.pruefe('⌘Q beendet', await p.beende())
}
