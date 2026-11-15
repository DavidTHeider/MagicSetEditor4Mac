// **Leerlauf am nativen Fenster: sichtbar, ausgeblendet, minimiert** —
// PERF01 (#2351),
// Befund A-26 des Audits 0.9.0.
//
// # Der Vertrag, vor dem Lauf festgelegt
//
// - **Ein Set mit aufgeschlagener Karte** (die Saat `probe-werkstatt`), dann
//   eine Minute Einschwingen.
// - **Je Zustand zehn Minuten** (`MSE_LEERLAUF_MIN` verkürzt sie — dann heißt
//   jede Reihe „zu kurz", und der Beleg ist keiner für PERF01:
//   `mindestSekunden` im Vertrag), davor 30 Sekunden, bis der Zustand steht:
//   **sichtbar** (vorn), **ausgeblendet** (⌘H) und **minimiert** (⌘M).
// - **Gemessen**: die CPU-Zeit je Prozess über `ps`, zu Beginn und am Ende
//   jedes Zustands — die Anwendung und **ihre** WebKit-Prozesse (Webinhalt,
//   GPU, Netz), die mit dem Fenster entstanden. Dazu, was die Seite selbst
//   über ihre Sichtbarkeit weiß (`document.visibilityState`).
// - **Bewertet**: höchstens 1 % eines Kerns im Mittel, je Art und Zustand
//   (`scripts/messvertrag.mjs`).
//
// # Warum ausgeblendet und nicht verdeckt
//
// **Ein Fenster, das ein anderes ganz verdeckt, gehört einer anderen
// Anwendung**, und dieses Drehbuch fasst keine an. WebKit liest die
// Sichtbarkeit aus dem Verdeckungsstand des Fensters
// (`NSWindowOcclusionStateVisible`); den setzt ⌘H genauso zurück wie ein
// darübergelegtes Fenster. **Das ist eine Annahme über WebKit**, und belegt
// ist sie nur für diese beiden Griffe: die Spalte `visibilityState` sagt, ob
// die Seite sich verborgen sah.
//
// # Was es nicht misst
//
// Weckungen und Energie — `powermetrics` braucht Verwalterrechte.
import { execFileSync } from 'node:child_process'
import { cpus, loadavg } from 'node:os'

import { MESSVERTRAG, tafel, urteil } from '../messvertrag.mjs'
import { anteile, zeitenAus } from '../leerlauf.mjs'

// **32 Minuten** — nur, wer es nennt.
export const nurEinzeln = true

const minuten = Number(process.env.MSE_LEERLAUF_MIN ?? '10')
const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))
const ARTEN = { WebContent: 'Webinhalt', GPU: 'GPU', Networking: 'Netz' }

const prozesse = () => execFileSync('ps', ['-Ao', 'pid=,comm=']).toString().split('\n')
  .map((z) => z.trim().split(/\s+/)).filter((t) => t.length >= 2).map(([pid, ...name]) => ({ pid, name: name.join(' ') }))

// eigeneArten ordnet die Prozesse der Anwendung ihrer Art zu — die
// WebKit-Prozesse, die nach `vorher` entstanden; `EnhancedSecurity` startet
// das System selbst und zählt nicht.
function eigeneArten(vorher, anwendung) {
  const arten = new Map([[String(anwendung), 'Anwendung']])
  for (const x of prozesse()) {
    const fund = /WebKit\.(WebContent|GPU|Networking)$/.exec(x.name)
    if (fund && !vorher.has(x.pid)) arten.set(x.pid, ARTEN[fund[1]])
  }

  return arten
}

const zeiten = (arten) => {
  try {
    return zeitenAus(execFileSync('ps', ['-o', 'pid=,time=', '-p', [...arten.keys()].join(',')]).toString())
  } catch (fehler) {
    return zeitenAus(fehler.stdout?.toString() ?? '')
  }
}

export default async function leerlauf09(p) {
  p.zeige(`Last vor dem Lauf ${loadavg().map((z) => z.toFixed(2)).join(' / ')} auf ${String(cpus().length)} Kernen; je Zustand ${String(minuten)} Minuten`)
  const vorher = new Set(prozesse().map((x) => x.pid))
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  await p.starte({ ort })
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('die Vorschau steht', () => p.seite("return !!document.querySelector('img[alt=\"Vorschau der Karte\"]')?.checkVisibility()"), 20000)
  await schlafe(60000)

  const zustaende = [
    { name: 'sichtbar', herstellen: async () => { await p.aktiviere() }, steht: (lage) => lage.sichtbar && !lage.verborgen },
    { name: 'ausgeblendet', herstellen: async () => { await p.taste('h', 'befehl') }, steht: (lage) => lage.verborgen },
    { name: 'minimiert', herstellen: async () => { await p.aktiviere(); await schlafe(1000); await p.taste('m', 'befehl') }, steht: (lage) => !lage.sichtbar && !lage.verborgen },
  ]
  const ergebnis = {}
  const dauer = {}
  for (const zustand of zustaende) {
    p.notiz(`${zustand.name}, ${String(minuten)} Minuten`)
    await zustand.herstellen()
    const lage = await p.warte(`das Fenster ist ${zustand.name}`, async () => { const l = await p.fenster(); return zustand.steht(l) ? l : false }, 15000)
    await schlafe(30000)
    const sicht = await p.seite('return document.visibilityState')
    const arten = eigeneArten(vorher, p.pids.at(-1))
    const anfang = zeiten(arten)
    const beginn = Date.now()
    await schlafe(minuten * 60000)
    const ende = zeiten(eigeneArten(vorher, p.pids.at(-1)))
    dauer[zustand.name] = (Date.now() - beginn) / 1000
    ergebnis[zustand.name] = anteile(eigeneArten(vorher, p.pids.at(-1)), anfang, ende, dauer[zustand.name])
    p.zeige(`${zustand.name}: Fenster sichtbar ${String(lage.sichtbar)}, verborgen ${String(lage.verborgen)}, die Seite sieht sich »${String(sicht)}«; ${JSON.stringify(ergebnis[zustand.name])}`)
  }
  await p.aktiviere()
  await p.warte('das Fenster ist wieder vorn', async () => (await p.fenster()).sichtbar, 15000)
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('Ergebnis')
  p.zeilen.push('', '| Reihe | n | min | Median | P95 | max | Grenze | Urteil |', '| --- | --- | --- | --- | --- | --- | --- | --- |')
  for (const e of MESSVERTRAG.filter((x) => x.beleg === 'huellenprobe/leerlauf09.md')) {
    const [, zustand, art] = /^Leerlauf (\S+), (\S+) /.exec(e.reihe)
    const wert = ergebnis[zustand]?.[art]
    const werte = wert === undefined ? [] : [wert]
    // **Die Dauer geht ins Urteil** (Zweitblick über Phase 3, Befund 13): ein
    // Lauf mit `MSE_LEERLAUF_MIN` unter zehn heißt „zu kurz" und belegt nichts.
    const lauf = { sekunden: dauer[zustand] }
    p.zeilen.push(tafel(e, werte, lauf))
    p.pruefe(`${e.reihe}: ≤ ${String(e.grenze)}`, urteil(e, werte, lauf) === 'gehalten', urteil(e, werte, lauf))
  }
}
