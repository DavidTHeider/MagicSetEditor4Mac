// **Diagnosepaket und Unified Log auf eingeschmuggelte Geheimnisse** — I05
// (#2263),
// am nativen Fenster:
//
//   1. eine zufällige **Marke** dorthin, wo Geheimnisse liegen: in den Bund
//      (Scrydex-Schlüssel, CardTrader-Marke — im eigenen Bund der Probe), in
//      die Proxyadresse samt Kennwort (`HTTPS_PROXY`, der einzige Weg, den
//      die Einstellung für ein Kennwort lässt), in den Namen eines Menschen
//      und damit seines Profils und in einen Dateipfad,
//   2. **kontrollierte Fehler** — jeder Abruf geht über den Proxy auf einen
//      geschlossenen Port und scheitert, mit Schlüssel und Kennwort im Weg —
//      und **Erfolg**: Set öffnen, Diagnose zusammenstellen,
//   3. das Paket ziehen, auspacken, das Manifest prüfen und jede Datei nach
//      der Marke durchsuchen — roh, URL-kodiert, Base64,
//   4. den Unified Log seit dem Start: die Einträge der Anwendung ganz, die
//      des Systems nach jeder Form der Marke — getrennt berichtet,
//   5. **die Gegenprobe**: ein absichtlich ungeschwärztes Kontrollstück neben
//      dem Paket wird gefunden, auch als Basic-Auth-Kopf. Ein Scanner, der
//      nichts findet, weil er falsch sucht, sähe sonst aus wie ein sauberes
//      Paket.
//
// **Was er nicht belegt**: den Bund des Rechners — die Probe hat ihren
// eigenen —, und was ein Mensch in einen Freitext tippt, den das Paket
// ausdrücklich mitnimmt.
import { execFileSync } from 'node:child_process'
import { randomBytes } from 'node:crypto'
import { cpSync, existsSync, mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'

import { durchsucheVerzeichnis, ordneLog, pruefeManifest, verkleidungen } from '../geheimnisscan.mjs'

// binde ruft eine Bindung und gibt Code **und Satz** des Fehlers — der Code
// allein sagte im ersten Lauf „unbekannt" und nicht, ob der Schlüssel ins
// Netz ging. Die Marke ist zufällig und kein Geheimnis; steht sie im Satz,
// ist das selbst ein Fund und steht im Beleg.
const binde = (aufruf) => `try { const a = await window.go.app.Bindungen.${aufruf}; return a?.fehler ? a.fehler.code + ' — ' + String(a.fehler.meldung ?? '').slice(0, 200) + (a.fehler.vorgang ? ' [Vorgang ' + a.fehler.vorgang + ']' : '') : 'gut' } catch (e) { return 'abgewiesen: ' + String(e?.message ?? e).slice(0, 160) }`

// grundImProtokoll gibt die Zeilen des Protokolls zu einem Vorgang — dort
// steht der Grund, den der Satz nicht nennt (0075).
function grundImProtokoll(ort, antwort) {
  const vorgang = /\[Vorgang ([0-9a-f]+)\]/.exec(antwort)?.[1]
  if (!vorgang) return ''
  try {
    return readFileSync(join(ort, 'Protokolle', 'mse.log'), 'utf8').split('\n').filter((z) => z.includes(vorgang)).map((z) => z.slice(0, 400)).join(' ⏎ ')
  } catch {
    return 'kein Protokoll'
  }
}

const kasten = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith('Paket sichern'))"

// unifiedLog liest den Unified Log seit `beginn` als JSON.
function unifiedLog(beginn, praedikat) {
  const roh = execFileSync('log', ['show', '--style', 'json', '--start', beginn, '--predicate', praedikat], { encoding: 'utf8', maxBuffer: 256 * 1024 * 1024 })
  const liste = JSON.parse(roh.trim() === '' ? '[]' : roh)

  return Array.isArray(liste) ? liste : []
}

export default async function i05geheimnisse(p) {
  const marke = `MSEGEHEIM${randomBytes(6).toString('hex').toUpperCase()}`
  const ort = p.ort()
  const protokolle = join(ort.ort, 'Protokolle')
  mkdirSync(protokolle)

  // **Der Dateipfad trägt die Marke** — ein Set aus einem Ordner, dessen Name
  // sie enthält; ein Paket, das Herkunftspfade nennt, nennt dann auch sie.
  const ordner = join(ort.ort, `Sets-${marke}`)
  mkdirSync(ordner)
  cpSync(resolve('saat/probe-werkstatt.mse-set'), join(ordner, 'probe-werkstatt.mse-set'), { recursive: true })

  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'einstellungen', 'setzen', 'protokoll.ort', protokolle)
  p.mse(ort, 'einstellungen', 'setzen', 'protokoll.stufe', 'ausfuehrlich')
  p.mse(ort, 'oeffnen', join(ordner, 'probe-werkstatt.mse-set'))
  p.mse(ort, 'nutzer-neu', `Mensch ${marke}`)
  // **Ein Exemplar, damit die Dienste mit Konto überhaupt fragen** — ohne
  // eines antworten sie „nicht gefunden", bevor ein Schlüssel ins Netz geht
  // (erster Lauf am 2026-10-10).
  p.mse(ort, 'exemplar', 'bolt-m10', '1')
  // **Und eingeschaltet** — ein abgeschalteter Dienst fragt nicht, und sein
  // Schlüssel ginge nie auf den Weg (zweiter Lauf: »abgeschaltet«).
  p.mse(ort, 'einstellungen', 'setzen', 'dienst.scrydex', 'an')
  p.mse(ort, 'einstellungen', 'setzen', 'dienst.cardtrader', 'an')

  const beginn = new Date().toLocaleString('sv-SE')
  const proxy = `http://nutzer:${marke}@127.0.0.1:9`
  await p.starte({ ort, umgebung: { HTTP_PROXY: proxy, HTTPS_PROXY: proxy, NO_PROXY: '' } })

  p.notiz('1 — die Marke in den Bund der Probe')
  p.zeige(`Scrydex-Schlüssel: ${await p.seite(binde(`ZugangSetzen('scrydex', 'schluessel', ${JSON.stringify(marke)})`))}`)
  p.zeige(`Scrydex-Team: ${await p.seite(binde(`ZugangSetzen('scrydex', 'team', ${JSON.stringify(`team-${marke}`)})`))}`)
  p.zeige(`CardTrader-Marke: ${await p.seite(binde(`ZugangSetzen('cardtrader', 'marke', ${JSON.stringify(marke)})`))}`)

  p.notiz('2 — kontrollierte Fehler über den Proxy, und Erfolg')
  const abrufe = [
    ['Scryfall-Regeln', await p.seite(binde("ScryfallRegeln('Lightning Bolt')"))],
    ['Scrydex-Verlauf', await p.seite(binde("Fremdpreisverlauf('scrydex', 'bolt-m10', 'nonfoil', [])"))],
    ['CardTrader-Angebote', await p.seite(binde("Marktangebote('cardtrader', 'bolt-m10', 'de', 'nonfoil', [])"))],
  ]
  for (const [was, antwort] of abrufe) {
    p.zeige(`${was}: ${antwort}`)
  }
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('das Set ist offen', () => p.seite("return document.querySelector('main')?.dataset.seite === 'karten'"), 20000)

  p.notiz('3 — das Diagnosepaket')
  await p.taste(',', 'befehl')
  await p.wartetext('Diagnose', 20000)
  await p.klickeText('Diagnose')
  await p.klickeText('Zusammenstellen')
  await p.warte('die Vorschau zeigt die Teile', () => p.seite(`return !!${kasten}`), 20000)
  const paket = join(ort.ort, 'bericht.msediagnose')
  await p.dialogantwort(paket)
  await p.klickeAuf(kasten, '»Paket sichern«')
  await p.warte('das Paket liegt auf der Platte', () => existsSync(paket), 20000)
  const ausgepackt = mkdtempSync(join(tmpdir(), 'i05-paket-'))
  execFileSync('unzip', ['-q', paket, '-d', ausgepackt])
  const teile = execFileSync('unzip', ['-Z1', paket], { encoding: 'utf8' }).split('\n').filter(Boolean)
  p.zeige(`im Paket: ${teile.join(', ')}`)
  // **Der Name aus `internal/app/diagnosepaket.go`** (`Diagnosemanifest`).
  const manifestname = 'diagnose.json'
  const manifest = pruefeManifest(ausgepackt, manifestname)
  p.pruefe('jede Prüfsumme des Manifests gilt, und nichts liegt daneben', manifest.gueltig, manifest.fehler.join(' · ') || manifestname)
  const imPaket = durchsucheVerzeichnis(ausgepackt, [marke])
  p.pruefe('im Paket steht die Marke nirgends — weder im Inhalt noch in einem Namen', imPaket.length === 0, imPaket.map((f) => `${f.ort}:${f.datei}:${f.art}`).join(', ') || 'kein Fund')

  p.notiz('4 — der Unified Log seit dem Start')
  p.pruefe('⌘Q beendet', await p.beende())
  for (const [was, antwort] of abrufe) {
    const grund = grundImProtokoll(ort.ort, antwort)
    if (grund !== '') p.zeige(`Grund im Protokoll, ${was}: ${grund}`)
  }
  const eigen = unifiedLog(beginn, 'process == "MagicSetEditor4Mac"')
  const formen = verkleidungen(marke).map((form) => `eventMessage CONTAINS ${JSON.stringify(form.text)}`).join(' OR ')
  // **Ohne `log` selbst**: es protokolliert sein Prädikat, und darin steht die
  // Marke — der erste Lauf fand genau diesen einen Eintrag.
  const fremd = unifiedLog(beginn, `process != "MagicSetEditor4Mac" AND process != "log" AND (${formen})`)
  const { app, system } = ordneLog([...eigen, ...fremd], 'MagicSetEditor4Mac', [marke])
  p.zeige(`Unified Log: ${String(eigen.length)} Einträge der Anwendung, ${String(fremd.length)} des Systems mit einer Form der Marke`)
  p.pruefe('die Anwendung hat ihre Einträge geschrieben — der Leser greift nicht daneben', eigen.length > 0, String(eigen.length))
  p.pruefe('im Protokoll der Anwendung steht die Marke nirgends', app.length === 0, app.map((f) => f.art).join(', ') || 'kein Fund')
  p.zeige(`beim System: ${system.map((f) => `${f.prozess}:${f.art}`).join(', ') || 'kein Fund'} — kein Leck dieses Programms, wenn es dort steht, aber eine Auskunft`)
  const ausgabe = durchsucheVerzeichnis(ort.ort, [marke]).filter((f) => f.ort === 'inhalt' && (f.datei === 'ausgabe.log' || f.datei.startsWith('Protokolle/')))
  p.pruefe('in der Ausgabe und im Protokoll auf der Platte steht sie nirgends', ausgabe.length === 0, ausgabe.map((f) => `${f.datei}:${f.art}`).join(', ') || 'kein Fund')

  p.notiz('5 — die Gegenprobe: ein ungeschwärztes Kontrollstück')
  const kontrolle = mkdtempSync(join(tmpdir(), 'i05-kontrolle-'))
  writeFileSync(join(kontrolle, 'roh.txt'), `proxy=${proxy}`)
  writeFileSync(join(kontrolle, 'kopf.txt'), `Authorization: Basic ${Buffer.from(`nutzer:${marke}`).toString('base64')}`)
  const gefunden = durchsucheVerzeichnis(kontrolle, [marke]).map((f) => `${f.datei}:${f.art}`).sort()
  p.pruefe('der Scanner findet das Kontrollstück, roh und als Basic-Auth', gefunden.join(',') === 'kopf.txt:base64,roh.txt:roh', gefunden.join(', '))
  const probe = ordneLog([{ processImagePath: '/Applications/X.app/Contents/MacOS/MagicSetEditor4Mac', eventMessage: proxy }], 'MagicSetEditor4Mac', [marke])
  p.pruefe('und ein Eintrag der Anwendung mit der Marke zählt als Fund der Anwendung', probe.app.length === 1 && probe.system.length === 0, JSON.stringify(probe))
}
