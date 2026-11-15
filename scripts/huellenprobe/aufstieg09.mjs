// **Der Aufstieg von 0.8.0, mit großem Bestand** — M01
// (#2345),
// Befunde A-06 und A-12.
//
// Die Abnahme: 0.8.0 → neue Fassung mit 25.000 Karten in 50 Sets samt Bildern
// — Inhalte und Bilder erhalten; die alte Fassung weist den neuen Stand ab;
// der Rückweg über die Sicherung von vor der Wanderung; frischer Bestand,
// aufgestiegener Bestand und zwei gleichzeitige Starter führen dieselbe
// Buchführung, mit genau einer Wanderung.
//
// # 0.8.0 kommt aus dem Tag
//
// In einem eigenen Arbeitsbaum am Tag `v0.8.0` entstehen seine Kommandozeile
// und sein Abnahmebündel (`scripts/bauen.sh abnahme`) — mit den Modulen und
// der Oberfläche dieses Baums, ohne Netz: seit 0.8.0 hat sich an keiner
// Sperrdatei etwas geändert. **Der Rückweg geht durch sein Fenster**, so wie
// ein Mensch ihn ginge: das alte Bündel öffnet den neuen Stand nicht, zeigt
// die Wiederherstellung samt der Sicherung, die es öffnen kann, und spielt sie
// zurück. Die Kommandozeile von 0.8.0 kann das nicht — sie öffnet den Bestand,
// bevor sie zurückspielt; das steht als Befund im Beleg.
//
// # Der Bestand ist künstlich
//
// 0.8.0 installiert Spiel und Stil der Saat und liest 50 erzeugte Sets mit je
// 500 Karten ein, je zwanzig davon mit einem eigenen Bild. Kein echter
// Bestand, kein Netz.
//
// # Was es nicht belegt
//
// Einen Abbruch **in** der Wanderung, eine volle Platte und einen laufenden
// Dienst — das ist M02 (#2346). Und den Aufstieg durch das neue Fenster: er
// geht über dieselbe Naht (`storage.Migriere`), hier über `mse`.
import { execFileSync, spawn } from 'node:child_process'
import { copyFileSync, existsSync, mkdirSync } from 'node:fs'
import { join } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'
import { bilder, buchfuehrung, fassung080, inhalt, setErzeugen, sicherungen, sqlite, summe, unterschied, zaehlerJeSet, wanderungenNachDerGrundlinie, wurzel } from './aufstiegsteile.mjs'

const SETS = 50
const KARTEN = 500
const BILDER = 20

export default async function aufstieg09(p) {
  const ort = p.ort()
  const arbeit = join(ort.ort, 'arbeit')
  mkdirSync(arbeit)

  p.notiz('0.8.0 aus dem Tag: Kommandozeile und Abnahmebündel')
  const alt = fassung080(arbeit)
  try {
    await aufsteigen(p, ort, arbeit, alt)
  } finally {
    alt.entfernen()
  }
}

async function aufsteigen(p, ort, arbeit, alt) {
  const altRuf = (o, ...a) => execFileSync(alt.mse, a, { cwd: wurzel, env: kindumgebung(o), maxBuffer: 1 << 28, stdio: ['ignore', 'pipe', 'pipe'] }).toString()
  const altVersuch = (...a) => {
    try {
      return { aus: altRuf(ort, ...a), fehler: '' }
    } catch (fehler) {
      return { aus: '', fehler: `${fehler.stderr}`.trim() }
    }
  }
  p.zeige(`0.8.0: Stand ${alt.stand}, Bündelabdruck ${alt.abdruck}, ${altRuf(ort, 'meta').split('\n')[0]}`)
  altRuf(ort, 'paket-installieren', 'saat/probe.mse-game')
  altRuf(ort, 'paket-installieren', 'saat/probe-kunst.mse-style')

  const anfang = Date.now()
  for (let s = 1; s <= SETS; s++) {
    const ordner = join(arbeit, `aufstieg-${s}.mse-set`)
    setErzeugen(ordner, s, { karten: KARTEN, bilder: BILDER })
    altRuf(ort, 'oeffnen', ordner)
  }
  p.zeige(`50 Sets eingelesen in ${Math.round((Date.now() - anfang) / 1000)} s`)

  const saetze = Number(sqlite(ort.bestand, 'SELECT count(*) FROM sets'))
  const karten = Number(sqlite(ort.bestand, 'SELECT count(*) FROM cards'))
  p.pruefe('0.8.0 führt 50 Sets', saetze === SETS, String(saetze))
  p.pruefe('und 25.000 Karten', karten === SETS * KARTEN, String(karten))
  const wanderungenAlt = buchfuehrung(ort.bestand).split('\n')
  p.pruefe('auf der Grundlinie: 97 Wanderungen', wanderungenAlt.length === 97, wanderungenAlt.at(-1))
  const vorher = inhalt(ort.bestand)
  const bilderVorher = bilder(ort.bestand)
  p.pruefe('jeder Bildverweis findet seine Datei', bilderVorher.verweise === SETS * BILDER && bilderVorher.zahl === bilderVorher.verweise, `${bilderVorher.zahl} von ${bilderVorher.verweise}`)
  // **Erst den WAL zurückschreiben**: der Bestand läuft im WAL-Modus, und eine
  // Kopie der Hauptdatei allein verlöre, was noch daneben liegt.
  execFileSync('sqlite3', [ort.bestand, 'PRAGMA wal_checkpoint(TRUNCATE);'], { stdio: 'ignore' })
  const stand080 = join(arbeit, 'stand-0.8.0.db')
  copyFileSync(ort.bestand, stand080)
  const sicherungenVorher = sicherungen(ort)

  p.notiz('Der Aufstieg: das heutige mse öffnet den Bestand')
  const aufstieg = Date.now()
  p.mse(ort, 'sets')
  p.zeige(`Aufstieg in ${Date.now() - aufstieg} ms`)
  const wanderungenNeu = buchfuehrung(ort.bestand).split('\n')
  const neue = wanderungenNachDerGrundlinie()
  p.pruefe(`die ${String(neue.length)} Wanderungen nach der Grundlinie kommen dazu`, wanderungenNeu.length === 97 + neue.length && wanderungenNeu.at(-1).startsWith(neue.at(-1)), wanderungenNeu.at(-1))
  const anders = unterschied(vorher, inhalt(ort.bestand))
  p.pruefe('jede Tafel trägt denselben Inhalt', anders.length === 0, anders.join(', '))
  p.pruefe('jedes Set hat seinen Änderungszähler', zaehlerJeSet(ort.bestand))
  const bilderNachher = bilder(ort.bestand)
  p.pruefe('die Bilder sind dieselben', bilderNachher.abdruck === bilderVorher.abdruck, `${bilderNachher.zahl} von ${bilderNachher.verweise}`)
  p.pruefe('die Tafel der Vermerke ist leer — unbekannt heißt unbekannt', sqlite(ort.bestand, 'SELECT count(*) FROM sicherungsvermerke') === '0')

  // **Am Unterschied und nicht am Namen**: beide Bauten heißen 0.8.0, bis die
  // Freigabe die Fassung hebt, und eine Sicherung überschreibt nie.
  const neu = sicherungen(ort).filter((n) => !sicherungenVorher.includes(n))
  p.pruefe('vor der Wanderung entstand genau eine Sicherung', neu.length === 1, `vorher ${sicherungenVorher.join(', ') || 'keine'}, neu ${neu.join(', ') || 'keine'}`)
  const sicherung = join(ort.ort, neu[0] ?? 'fehlt')
  p.pruefe('sie hält den Stand von 0.8.0', existsSync(sicherung) && unterschied(inhalt(sicherung), vorher).length === 0 && buchfuehrung(sicherung).split('\n').length === 97)
  p.pruefe('und sie ist heil', existsSync(sicherung) && sqlite(sicherung, 'PRAGMA quick_check') === 'ok')

  p.notiz('Die alte Kommandozeile weist den neuen Stand ab')
  const summeVorAbweisung = summe(ort.bestand)
  const abgewiesen = altVersuch('sets')
  p.zeige(abgewiesen.fehler.split('\n')[0])
  p.pruefe('0.8.0 öffnet ihn nicht und sagt, dass er neuer ist', /neueren Fassung/.test(abgewiesen.fehler))
  p.pruefe('und lässt ihn, wie er ist', summe(ort.bestand) === summeVorAbweisung)
  const amTerminal = altVersuch('bestand-zurueckspielen', sicherung, 'egal')
  p.zeige(`**Befund**: die Kommandozeile von 0.8.0 spielt nicht zurück — ${amTerminal.fehler.split('\n')[0]}`)
  p.pruefe('auch dabei bleibt der Bestand, wie er ist', summe(ort.bestand) === summeVorAbweisung)

  p.notiz('Der Rückweg durch das Fenster von 0.8.0')
  await mitAltemFenster(p, () => p.starte({ ort, buendel: alt.buendel }))
  const seite = await p.seite('return document.body.innerText')
  p.pruefe('0.8.0 zeigt die Wiederherstellung', seite.includes('Sicherung zurückspielen'), seite.split('\n').slice(0, 3).join(' ⏎ '))
  p.pruefe('und nennt den Weg zurück über die Sicherung von vor der Wanderung', seite.includes(neu[0] ?? '—'))
  await p.bild('m01-wiederherstellung-080')
  await p.dialogantwort(sicherung)
  await p.klickeText('Sicherung zurückspielen …')
  await p.wartetext('Diese Sicherung zurückspielen?', 20000)
  await p.klickeText('Zurückspielen')
  await p.warte('0.8.0 öffnet den Stand von vorher', () => /Aufstiegsset 50/.test(altVersuch('sets').aus), 60000)
  await p.bild('m01-zurueckgespielt-080')
  p.pruefe('mit allen Karten', Number(sqlite(ort.bestand, 'SELECT count(*) FROM cards')) === SETS * KARTEN)
  p.pruefe('und den 97 Wanderungen von 0.8.0', buchfuehrung(ort.bestand).split('\n').length === 97)
  p.pruefe('und denselben Bildern', bilder(ort.bestand).abdruck === bilderVorher.abdruck)
  p.pruefe('⌘Q beendet 0.8.0', await p.beende())

  p.notiz('Dieselbe Buchführung: frisch, aufgestiegen, zwei zugleich')
  const frisch = p.ort()
  p.mse(frisch, 'sets')
  const zugleich = p.ort()
  copyFileSync(stand080, zugleich.bestand)
  const lauf = () => new Promise((fertig) => {
    const kind = spawn(join(wurzel, '.go/bin/mse'), ['sets'], { cwd: wurzel, env: kindumgebung(zugleich), stdio: ['ignore', 'ignore', 'pipe'] })
    let fehler = ''
    kind.stderr.on('data', (d) => { fehler += d })
    kind.on('exit', (code) => fertig({ code, fehler }))
  })
  const [a, b] = await Promise.all([lauf(), lauf()])
  p.pruefe('beide Starter kamen durch', a.code === 0 && b.code === 0, `${a.code}/${b.code} ${a.fehler}${b.fehler}`.trim())
  const aufgestiegen = buchfuehrung(zugleich.bestand)
  p.pruefe('frisch und aufgestiegen führen dieselbe Buchführung', buchfuehrung(frisch.bestand) === aufgestiegen)
  p.pruefe('jede Wanderung steht genau einmal da', wanderungenNachDerGrundlinie().every((name) => aufgestiegen.split('\n').filter((z) => z.startsWith(name)).length === 1))
  // **Und genau eine Sicherung, mit dem Stand von vorher** — M02 (#2346):
  // bis dahin entschied ein Blick ohne Sperre, ob gesichert wird, beide
  // Starter sicherten, und der zweite legte den Stand **nach** der Wanderung
  // ab. Jetzt stehen Blick, Sicherung und Wanderung unter
  // `bestand.db.wanderung`.
  const angelegt = sicherungen(zugleich)
  p.pruefe('zwei gleichzeitige Starter legen eine Sicherung an', angelegt.length === 1, angelegt.join(', '))
  p.pruefe('und sie hält die 97 Wanderungen von 0.8.0', angelegt.length === 1 && buchfuehrung(join(zugleich.ort, angelegt[0])).split('\n').length === 97)
  p.pruefe('und der Inhalt ist der von 0.8.0', unterschied(inhalt(zugleich.bestand), vorher).length === 0)
}

// mitAltemFenster startet ein Fenster, das die Wiederherstellung noch ohne
// `data-wiederherstellung` zeigt — **0.8.0 erkannte sie am Wort »Sicherung«**
// (`scripts/huellenprobe.mjs` am Tag, `oberflaecheSteht`). Nur für diesen Start.
async function mitAltemFenster(p, starte) {
  const bereit = p.bereit
  p.bereit = async () => {
    await p.warte('die Oberfläche von 0.8.0 steht', () => p.seite("return document.readyState === 'complete' && (document.querySelectorAll('nav button').length > 3 || /Sicherung/.test(document.body.innerText))"), 30000)
  }
  try {
    await starte()
  } finally {
    p.bereit = bereit
  }
}
