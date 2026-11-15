// **Der tatsächliche Netzausgang — Go und WebView** — D06
// (#2237).
//
// `docs/NETZEIGENSCHAFTEN.md` sagt, die Vorschaubilder der Kartensuche lade
// die WebView **an `internal/netz` vorbei**, und welchen Proxy WebKit nehme,
// sei ungemessen. Hier wird es gemessen — **am Prozess und nicht am
// Go-Transport**: `nettop` nennt je Prozess die Verbindungen, und gezählt
// werden die Anwendung und jeder WebKit-Prozess, der mit ihr entstand.
//
// # Der Aufbau
//
// **Scryfall heißt `http://scryfall.test`** — einen Namen, den kein
// Namensdienst kennt. Go löst ihn mit einem Proxy nicht selbst auf, sondern
// schickt die Anfrage dem Proxy; der **protokolliert jede** und reicht sie an
// eine Gegenstelle auf der Schleife. Die nennt als Vorschaubilder zwei
// **echte** Adressen (wie `sec17bilder`), und die WebView holt sie wirklich.
// Sonst geht nichts hinaus.
//
// # Die Schritte
//
//   1. **Erststart** ohne gewollte Netzfunktion: eine Minute Ruhe und ein Gang
//      durch Hilfe und Einstellungen — jeder Ausgang außer der Schleife ist
//      ein Befund.
//   2. **Hinter dem Proxy**: die Suche geht über ihn, die Bilder — gemessen.
//   3. **Dieselbe Suche noch einmal**: was der Zwischenspeicher tut.
//   4. **Proxywechsel** zur Laufzeit über die Kommandozeile: die nächste Suche
//      geht über den neuen.
//   5. **Offline**: der Proxy ist fort — die Meldung nennt den Fall.
//
// **Was es nicht belegt**: einen Systemproxy von macOS — den setzt ein
// Mensch, und ob WebKit ihm folgt, steht dann im Beleg von Hand. **Und weder
// die Frist des Zwischenspeichers noch die Drossel** (Befund 7 des
// Zweitblicks über Phase D): Teil 3 zeigt, dass die zweite Suche aus dem
// Speicher kommt, nicht, wann ein Eintrag verfällt oder ab wann gedrosselt
// wird. Beides hält `internal/netz` mit eingesetzter Uhr, nicht dieser Lauf.
// **`nettop` sieht Schnappschüsse**: eine Verbindung, die zwischen zwei
// Abfragen beginnt und endet, fehlt.
import { execFileSync } from 'node:child_process'
import { lookup } from 'node:dns/promises'
import { createServer, request } from 'node:http'

import { zurSeite } from './uirevue.mjs'

// laufmarke macht die Bildadressen dieses Laufs eindeutig.
const laufmarke = String(Date.now())

const bilder = [
  'https://cards.scryfall.io/small/front/1/1/11bf83bb-c95b-4b4f-9a56-ce7a1816307a.jpg?1783940984',
  'https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=527297&type=card',
]

const treffer = (nummer, bild) => ({
  object: 'card', id: `00000000-0000-4000-8000-00000000006${String(nummer)}`,
  name: `Ausgangsprobe ${String(nummer)}`, lang: 'en', layout: 'normal',
  mana_cost: '{U}', type_line: 'Creature — Testling', oracle_text: '',
  rarity: 'common', set: 'prf', set_name: 'Prüfsatz', collector_number: String(nummer),
  image_uris: { small: bild, normal: bild },
  scryfall_uri: 'https://beispiel.invalid/karte/1',
})

// ausgaengeVon liest die Verbindungen der genannten Prozesse — ohne die
// Schleife. **Je Aufruf ein Schnappschuss**; wer zu kurz lebt, fehlt, darum
// wird während jedes Schritts fortlaufend gesammelt.
//
// **Ob überhaupt gemessen wurde, zählt mit** (Befund 6 des Zweitblicks über
// Phase D): ein `nettop`, das jedes Mal scheitert, gäbe sonst „kein Ausgang"
// als Ergebnis. `gelungen` ist 1 für eine Runde, in der **beide** Arten eine
// Ausgabe in der Gestalt von `nettop -L 1` brachten — eine Kopfzeile, die mit
// einem Komma beginnt (`,state,` bzw. `,`; nachgesehen am 2026-10-10). Eine
// beliebige Ausgabe zählte im Nachlauf noch, und eine gelungene UDP-Messung
// trug eine gescheiterte TCP-Messung mit.
function ausgaengeVon(pids) {
  const gefunden = new Set()
  let arten = 0
  for (const art of ['tcp', 'udp']) {
    let roh = ''
    try {
      roh = execFileSync('nettop', ['-L', '1', '-n', '-m', art, '-J', 'state', ...pids.flatMap((pid) => ['-p', String(pid)])], { encoding: 'utf8', timeout: 10000 })
    } catch { continue }
    if (!roh.startsWith(',')) continue
    arten += 1
    for (const zeile of roh.split('\n')) {
      const treffer = /^(tcp|udp)[46] (\S+)<->(\S+),/.exec(zeile)
      if (!treffer) continue
      const fern = treffer[3]
      if (fern.startsWith('127.') || fern.startsWith('::1') || fern.startsWith('*') || fern.startsWith('[::1]')) continue
      gefunden.add(fern)
    }
  }
  return { gefunden, gelungen: arten === 2 ? 1 : 0 }
}

// webkitProzesse sind die Kennungen aller WebKit-Prozesse dieses Rechners.
function webkitProzesse() {
  const roh = execFileSync('ps', ['-axo', 'pid=,comm='], { encoding: 'utf8' })
  return roh.split('\n').map((z) => z.trim()).filter((z) => z.includes('WebKit.')).map((z) => Number(z.split(/\s+/)[0]))
}

// sammler liest Ausgänge im Halbsekundentakt, bis `stopp` gerufen wird.
function sammler(pids) {
  const alle = new Set()
  const stand = { gelungen: 0 }
  const uhr = setInterval(() => {
    const { gefunden, gelungen } = ausgaengeVon(pids())
    stand.gelungen += gelungen
    for (const a of gefunden) alle.add(a)
  }, 500)
  return { alle, stand, stopp: () => clearInterval(uhr) }
}

// einProxy protokolliert jede Anfrage und reicht `scryfall.test` an die
// Gegenstelle; `CONNECT` reicht er durch und schreibt das Ziel auf.
async function einProxy(p, name, ziel) {
  const buch = []
  const proxy = createServer((anfrage, antwort) => {
    const adresse = new URL(anfrage.url)
    buch.push({ zeit: Date.now(), was: `${anfrage.method} ${adresse.host}${adresse.pathname}` })
    if (adresse.hostname !== 'scryfall.test') {
      antwort.writeHead(502).end()
      return
    }
    const weiter = request({
      host: '127.0.0.1', port: ziel, method: anfrage.method,
      path: `${adresse.pathname}${adresse.search}`, headers: { ...anfrage.headers, host: 'scryfall.test' },
    }, (zurueck) => {
      antwort.writeHead(zurueck.statusCode ?? 502, zurueck.headers)
      zurueck.pipe(antwort)
    })
    weiter.on('error', () => { antwort.writeHead(502).end() })
    anfrage.pipe(weiter)
  })
  proxy.on('connect', (anfrage, sock) => {
    buch.push({ zeit: Date.now(), was: `CONNECT ${anfrage.url}` })
    sock.end('HTTP/1.1 403 Forbidden\r\n\r\n')
  })
  await new Promise((fertig) => { proxy.listen(0, '127.0.0.1', fertig) })
  p.zeige(`${name} horcht auf 127.0.0.1:${String(proxy.address().port)}`)
  return { buch, adresse: `http://127.0.0.1:${String(proxy.address().port)}`, schliessen: () => proxy.close() }
}

async function suche(p, begriff) {
  const feld = "[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && e.labels?.[0]?.innerText.startsWith('Bei Scryfall suchen'))"
  await p.klickeAuf(feld, 'das Suchfeld')
  // **Erst der Fokus, dann die Tasten** — und **kein ⌘A davor**: in zwei
  // nativen Läufen (2026-10-10) ging die erste Taste nach dem Kürzel
  // verloren, aus »Probe« wurde »robe«. Geleert wird über den Wertsetzer,
  // den React liest; getippt wird dann mit echten Tasten.
  await p.warte('das Suchfeld hat den Fokus', () => p.seite(`return document.activeElement === (${feld})`), 10000)
  await p.seite(`const f = ${feld}; Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(f, ''); f.dispatchEvent(new Event('input', { bubbles: true })); return f.value === ''`)
  await p.tippeInsFeld(begriff)
  await p.klickeText('Suchen')
}

const speicher = (p) => p.seite('const a = await window.go.app.Bindungen.Metriken(); return a.speicher')

export default async function d06(p) {
  const gegenstelle = createServer((_anfrage, antwort) => {
    antwort.writeHead(200, { 'content-type': 'application/json' })
    antwort.end(JSON.stringify({
      object: 'list', total_cards: bilder.length, has_more: false,
      // **Je Lauf eine eigene Adresse** — sonst liefert WebKit die Bilder aus
      // seinem Zwischenspeicher, und der Weg ist nicht zu messen: der zweite
      // native Lauf (2026-10-10) sah weder eine direkte Verbindung noch ein
      // CONNECT, und doch waren beide Bilder da.
      data: bilder.map((bild, i) => treffer(i + 1, `${bild}?d06=${laufmarke}`)),
    }))
  })
  await new Promise((fertig) => { gegenstelle.listen(0, '127.0.0.1', fertig) })
  const erster = await einProxy(p, 'der erste Proxy', gegenstelle.address().port)
  const zweiter = await einProxy(p, 'der zweite Proxy', gegenstelle.address().port)

  const bildwirte = {}
  for (const bild of bilder) {
    const wirt = new URL(bild).hostname
    bildwirte[wirt] = (await lookup(wirt, { all: true })).map((a) => a.address)
  }
  p.zeige(`die Bildwirte lösen auf zu ${JSON.stringify(bildwirte)}`)

  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  const vorher = new Set(webkitProzesse())
  try {
    p.notiz('1 — Erststart ohne gewollte Netzfunktion')
    await p.starte({ ort, umgebung: { MSE_SCRYFALL_ADRESSE: 'http://scryfall.test' } })
    const pids = () => [p.pids.at(-1), ...webkitProzesse().filter((pid) => !vorher.has(pid))]
    const ruhe = sammler(pids)
    await new Promise((weiter) => { setTimeout(weiter, 60000) })
    await p.taste(',', 'befehl')
    await new Promise((weiter) => { setTimeout(weiter, 3000) })
    await p.taste('Escape')
    ruhe.stopp()
    p.zeige(`Prozesse: ${pids().join(', ')}`)
    p.pruefe('während des Erststarts wurde gemessen', ruhe.stand.gelungen > 0, `${String(ruhe.stand.gelungen)} Messungen`)
    p.pruefe('der Erststart geht nicht hinaus', ruhe.stand.gelungen > 0 && ruhe.alle.size === 0, [...ruhe.alle].join(', ') || 'keine Verbindung außer der Schleife')
    p.pruefe('die Proxys sahen beim Erststart nichts', erster.buch.length === 0 && zweiter.buch.length === 0)

    p.notiz('2 — die Suche hinter dem ersten Proxy')
    p.mse(ort, 'einstellungen', 'setzen', 'netz.proxy.http', erster.adresse)
    await p.taste('2', 'befehl')
    await p.warte('die Sets stehen', () => p.seite("return !!document.querySelector('main article')"), 20000)
    await p.klickeAuf("[...document.querySelectorAll('main article :is(h2, h3) button, main article :is(h2, h3) a')].find((b) => b.checkVisibility())", 'das erste Set')
    await zurSeite(p, 'Karten')
    await p.klickeAuf("[...document.querySelectorAll('main [aria-haspopup=menu]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Mehr')", 'das Menü »Mehr«')
    await p.klickeText('Aus Scryfall anlegen …')
    const suchweg = sammler(pids)
    await suche(p, 'Ausgangsprobe')
    await p.wartetext('Ausgangsprobe 2', 20000)
    const geladen = await p.warte('die Vorschaubilder sind entschieden', async () => {
      const stand = await p.seite("return [...document.querySelectorAll('main img')].filter((i) => /scryfall\\.io|wizards\\.com/.test(i.src)).map((i) => ({ wirt: new URL(i.src).host, fertig: i.complete, breite: i.naturalWidth }))")
      return stand.length === bilder.length && stand.every((b) => b.fertig) ? stand : null
    }, 30000)
    await new Promise((weiter) => { setTimeout(weiter, 2000) })
    suchweg.stopp()
    p.zeige(`der erste Proxy sah: ${erster.buch.map((e) => e.was).join(' · ')}`)
    p.zeige(`Ausgänge während der Suche: ${[...suchweg.alle].join(', ')}`)
    p.pruefe('die Suche ging über den Proxy', erster.buch.some((e) => e.was.startsWith('GET scryfall.test/cards/search')))
    for (const bild of geladen) p.pruefe(`das Bild von ${bild.wirt} ist geladen`, bild.breite > 0)
    const direkt = Object.entries(bildwirte).filter(([, adressen]) => [...suchweg.alle].some((a) => adressen.some((x) => a.startsWith(x))))
    const ueberProxy = erster.buch.filter((e) => e.was.startsWith('CONNECT'))
    p.zeige(`Bildwirte direkt erreicht: ${direkt.map(([w]) => w).join(', ') || 'keiner'}; CONNECT am Proxy: ${ueberProxy.map((e) => e.was).join(', ') || 'keiner'}`)
    p.pruefe('der Weg der Bilder ist gemessen', direkt.length > 0 || ueberProxy.length > 0, direkt.length > 0 ? 'die WebView geht am Proxy vorbei' : 'die WebView nimmt den Proxy')
    await p.bild('d06-hinter-dem-proxy')

    p.notiz('3 — dieselbe Suche noch einmal')
    const vorZweitem = erster.buch.length
    const speicherVorher = await speicher(p)
    await suche(p, 'Ausgangsprobe')
    // **Gewartet wird auf die zweite Suche**, nicht auf einen Text, der von
    // der ersten noch dasteht (Befund 7 des Zweitblicks über Phase D): die
    // Zählung des Zwischenspeichers bewegt sich, wenn sie durch ist.
    const speicherNachher = await p.warte('die zweite Suche ist durch', async () => {
      const jetzt = await speicher(p)
      return jetzt.treffer + jetzt.geholt + jetzt.veraltet + jetzt.fehlgeschlagen >
        speicherVorher.treffer + speicherVorher.geholt + speicherVorher.veraltet + speicherVorher.fehlgeschlagen ? jetzt : null
    }, 20000)
    p.zeige(`Zwischenspeicher vorher ${JSON.stringify(speicherVorher)}, nachher ${JSON.stringify(speicherNachher)}`)
    p.zeige(`neue Anfragen am Proxy: ${String(erster.buch.length - vorZweitem)}`)
    // **Und die Wirkung wird geprüft, nicht nur gezeigt** (Rest von Befund 7,
    // angestrichen im Nachlauf): ein neuer Treffer, nichts frisch geholt, und
    // keine neue Suche am Proxy. Die Zähler gelten je Prozess — zwischen den
    // zwei Lesungen sucht in dieser Probe nichts anderes.
    p.pruefe('die zweite Suche kam aus dem Zwischenspeicher',
      speicherNachher.treffer > speicherVorher.treffer && speicherNachher.geholt === speicherVorher.geholt,
      `Treffer ${String(speicherVorher.treffer)} → ${String(speicherNachher.treffer)}, geholt ${String(speicherVorher.geholt)} → ${String(speicherNachher.geholt)}`)
    p.pruefe('die zweite Suche fragte den Proxy nicht',
      !erster.buch.slice(vorZweitem).some((e) => e.was.startsWith('GET scryfall.test/cards/search')))

    p.notiz('4 — Proxywechsel zur Laufzeit')
    p.mse(ort, 'einstellungen', 'setzen', 'netz.proxy.http', zweiter.adresse)
    const vorWechsel = erster.buch.length
    await suche(p, 'Wechselprobe')
    await p.warte('der zweite Proxy sah die Suche', () => zweiter.buch.some((e) => e.was.includes('Wechselprobe') || e.was.startsWith('GET scryfall.test')), 20000)
    p.pruefe('nach dem Wechsel geht nichts mehr über den ersten', erster.buch.length === vorWechsel)
    p.zeige(`der zweite Proxy sah: ${zweiter.buch.map((e) => e.was).join(' · ')}`)

    p.notiz('5 — offline: der Proxy ist fort')
    zweiter.schliessen()
    await new Promise((weiter) => { setTimeout(weiter, 500) })
    await suche(p, 'Offlineprobe')
    const satz = await p.warte('die Meldung', async () => {
      const text = await p.seite("return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText).join(' ')")
      return text.length > 0 ? text : null
    }, 20000)
    p.zeige(satz)
    p.pruefe('offline: kein »unerwarteter Fehler«', !satz.includes('unerwarteter Fehler'))
    await p.bild('d06-offline')
    p.pruefe('⌘Q beendet', await p.beende())
  } finally {
    erster.schliessen()
    zweiter.schliessen()
    gegenstelle.close()
  }
}
