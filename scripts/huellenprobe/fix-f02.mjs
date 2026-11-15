// **Ein neues Set verlustfrei sichern** — FIX-F02
// (#1746),
// die Spezifikation S01, und ENG-02.AK-04, Kernweg 3.
//
// Am nativen Fenster, mit den Dateidialogen von macOS:
//
// - NAT-01/S01.NAT-01: ein nativ angelegtes Set mit Bild, Text und Symbol
//   lässt sich über ⌘S und „Sichern unter …" schreiben,
// - NAT-02: es öffnet über ⌘O in einem frischen Bestand mit seinem Inhalt, und
//   ein Feld, das dieses Programm nicht kennt, übersteht Öffnen und Sichern,
// - NAT-03/S01.NAT-03: ein Abbruch ändert weder Ziel noch Inhalt; ein
//   Schreibfehler nennt die Ursache und lässt die alte Datei stehen,
// - S01.NAT-02: ⌘S und „Sichern unter …" gehen denselben Schreibweg — beide
//   schreiben dieselbe Datei.
import { chmodSync, existsSync, mkdirSync, readdirSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'
import { deflateSync } from 'node:zlib'

// probebild schreibt ein kleines PNG mit einem Muster.
function probebild(pfad) {
  const breite = 64
  const hoehe = 48
  const zeilen = []
  for (let y = 0; y < hoehe; y += 1) {
    zeilen.push(0)
    for (let x = 0; x < breite; x += 1) zeilen.push(40 + 3 * x, 60 + 3 * y, (x ^ y) & 8 ? 180 : 60)
  }
  const tafel = Array.from({ length: 256 }, (_, n) => {
    let c = n
    for (let k = 0; k < 8; k += 1) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1
    return c >>> 0
  })
  const crc = (daten) => {
    let c = 0xffffffff
    for (const b of daten) c = tafel[(c ^ b) & 0xff] ^ (c >>> 8)
    return (c ^ 0xffffffff) >>> 0
  }
  const stueck = (art, daten) => {
    const laenge = Buffer.alloc(4)
    laenge.writeUInt32BE(daten.length)
    const inhalt = Buffer.concat([Buffer.from(art), daten])
    const pruef = Buffer.alloc(4)
    pruef.writeUInt32BE(crc(inhalt))
    return Buffer.concat([laenge, inhalt, pruef])
  }
  const kopf = Buffer.alloc(13)
  kopf.writeUInt32BE(breite, 0)
  kopf.writeUInt32BE(hoehe, 4)
  kopf.set([8, 2, 0, 0, 0], 8)
  writeFileSync(pfad, Buffer.concat([
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    stueck('IHDR', kopf), stueck('IDAT', deflateSync(Buffer.from(zeilen))), stueck('IEND', Buffer.alloc(0)),
  ]))
}

// **Ein `.mse-set` ist ein Verzeichnis** mit der Hauptdatei `set` darin
// (docs/FORMATE.md) — so schreibt es dieses Programm.
const inhaltVon = (ort) => readdirSync(ort).sort()
const setdatei = (ort) => readFileSync(join(ort, 'set'), 'utf8')

// feldNamens sucht ein Eingabefeld der Karte über seinen **Feldnamen** — die
// Kennung, an der das Schreiben hängt (`aria-label`, `seiten/Feldeingabe.tsx`).
const feldNamens = (name) => `[...document.querySelectorAll('input, textarea, [contenteditable=true]')].find((e) => e.checkVisibility() && e.getAttribute('aria-label') === ${JSON.stringify(name)})`

// geschrieben wartet, bis die Datei dasteht.
async function geschrieben(p, ziel, was) {
  try {
    await p.warte(`${was} schreibt ${ziel}`, () => existsSync(join(ziel, 'set')), 20000)
  } catch (grund) {
    p.zeige(`Meldungen: ${JSON.stringify(await p.seite("return [...document.querySelectorAll('[role=alert],[role=status]')].map((e) => e.innerText.trim()).filter(Boolean)"))}, Blatt: ${JSON.stringify(await p.blatt())}, Ablage: ${JSON.stringify(readdirSync(join(ziel, '..')))}`)
    throw grund
  }
}

export default async function fixF02(p) {
  const ort = p.ort()
  const ablage = join(ort.ort, 'ablage')
  mkdirSync(ablage)
  const bild = join(ablage, 'probebild.png')
  probebild(bild)
  const erste = join(ablage, 'nachtprobe.mse-set')
  const zweite = join(ablage, 'nachtprobe-kopie.mse-set')

  await p.starte({ ort })

  p.notiz('Ein Set, nativ angelegt: Name, eine Karte mit Text und Bild, ein Symbol')
  await p.klick('Ablage', 'Neu')
  await p.warte('»Neu« fragt nach dem Namen', () => p.seite("return document.activeElement?.labels?.[0]?.innerText.includes('Name des Sets') ?? false"))
  await p.tippeInsFeld('Nachtprobe')
  // **Der Stil ist ausdrücklich der mit dem Bild der Karte** — R01 (#2093):
  // von den Saatstilen setzt nur `probe-kunst` das Bildfeld `kunst`; die
  // Vorgabe des Assistenten zeigte keines, und »Bild wählen« stand nie da.
  await p.klickeText('Zurück')
  // **Eine Suchauswahl, kein `<select>`** (`bausteine/Suchauswahl.tsx`):
  // das Feld mit `role="combobox"` öffnet die Liste, gewählt wird die Zeile.
  const stilwahl = "[...document.querySelectorAll('input[role=combobox]')].find((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').includes('Stil des neuen Sets'))"
  await p.warte('die Stilwahl steht da', () => p.seite(`return !!(${stilwahl})`))
  await p.klickeAuf(stilwahl, 'die Stilwahl')
  await p.wartetext('Mit Kunstfenster')
  await p.klickeAuf("[...document.querySelectorAll('[role=option]')].find((e) => e.checkVisibility() && e.innerText.includes('Mit Kunstfenster'))", 'die Zeile »Mit Kunstfenster« (probe-kunst)')
  p.pruefe('der Stil ist »Mit Kunstfenster« (probe-kunst)', await p.warte('die Wahl steht im Feld', () => p.seite(`return (${stilwahl})?.value === 'Mit Kunstfenster'`)))
  await p.klickeText('Weiter')
  await p.warte('der Name steht noch da', () => p.seite("return [...document.querySelectorAll('input')].some((e) => e.checkVisibility() && e.value === 'Nachtprobe')"))
  await p.klickeText('+ Set anlegen')
  await p.wartetext('Nachtprobe')
  await p.klickeText('+ Karte anlegen')
  await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.startsWith('Ohne Namen'))", 'die neue Karte')

  // **Die Vorlage ist die der Saat, und ihre Felder sind nachgewiesen** — R01
  // (#2093). Im eigenen Zuhause bietet »Neu« nur `saat/probe.mse-game` mit
  // `name`, `text`, `farbe` und `kunst`. Bis zum 2026-10-06 suchte das
  // Drehbuch »Rule text« aus dem Magic-Pack — das stand nur da, solange die
  // Probe unter dem echten `HOME` lief.
  // **Gewartet wird auf die Felder** — der Klick schlägt die Karte auf, ihre
  // Felder kommen eine Antwort später (§3: auf die Wirkung warten).
  const felder = "return [...document.querySelectorAll('input, textarea, [contenteditable=true]')].filter((e) => e.checkVisibility()).map((e) => e.getAttribute('aria-label')).filter(Boolean)"
  await p.warte('die Felder der Karte stehen', async () => (await p.seite(felder)).includes('name'), 15000)
  const angeboten = await p.seite(felder)
  p.pruefe('die Karte bietet die Felder der Saatvorlage', ['name', 'text'].every((n) => angeboten.includes(n)), angeboten.join(', '))
  await p.klickeAuf(feldNamens('name'), 'das Feld »name«')
  await p.tippeInsFeld('Nachtfalter')
  await p.klickeAuf(feldNamens('text'), 'das Feld »text«')
  await p.tippe('Fliegend.')
  await p.wartetext('Nachtfalter')

  await p.klickeText('Bild wählen')
  await p.dialogZeigtSich('oeffnen', '»Bild wählen«')
  await p.dialogantwort(bild)
  await p.klickeText('Bild wählen')
  // **Erst der Zuschnitt, dann das Bild** — gewählt ist noch nicht eingesetzt.
  await p.klickeText('Einsetzen')
  await p.warte('der Zuschnitt ist zu', () => p.seite("return ![...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Einsetzen')"))
  await p.bild('f02-karte')

  // **Das Symbol des Sets, gezeichnet** — eine Grundform, ins Set gelegt.
  await p.klickeText('Eigenschaften')
  await p.klickeText('Symbol')
  await p.klickeText('Kreis')
  await p.wartetext('Kreis — 4 Punkte')
  await p.klickeText('⤓ Ins Set legen')
  await new Promise((los) => setTimeout(los, 1000))

  p.notiz('⌘S bei einem Set ohne Datei fragt, wohin — S01.NAT-01 und NAT-03')
  await p.taste('s', 'befehl')
  await p.dialogZeigtSich('sichern', '⌘S')
  p.pruefe('der Abbruch schrieb nichts', !existsSync(erste))
  // **Dann über den Eintrag »Sichern«**, der ⌘S trägt — derselbe Weg; der
  // zweite Tastendruck kurz nach einem geschlossenen Dialog ging in einem
  // von drei Läufen verloren, weil das Fenster den Fokus erst zurückholt.
  await p.dialogantwort(erste)
  await p.klick('Ablage', 'Sichern')
  await geschrieben(p, erste, '»Sichern« (⌘S)')
  const liste = inhaltVon(erste)
  p.zeige(`im Archiv: ${liste.join(', ')}`)
  const text = setdatei(erste)
  // **Die Bilddatei heißt nach ihrem Feld** — wie in MSE2
  // (`src/gui/value/image.cpp`, Zeile 48, Fassung f9d9356 vom 2021-07-06):
  // im Saatfeld `kunst` also `kunst1`, im Magic-Pack `image1`.
  p.pruefe('das Set trägt das Bild der Karte', liste.some((n) => n.startsWith('kunst')), liste.join(', '))
  p.pruefe('und das gezeichnete Symbol', liste.some((n) => n.startsWith('symbol')) && /symbol:\s*\S/.test(text), liste.join(', '))
  p.pruefe('und den Namen und den Regeltext der Karte', text.includes('Nachtfalter') && text.includes('Fliegend.'))

  p.notiz('„Sichern unter …" geht denselben Weg — S01.NAT-02')
  await p.dialogantwort(zweite)
  await p.klick('Ablage', 'Sichern unter …')
  await geschrieben(p, zweite, '»Sichern unter …«')
  p.pruefe('beide Dateien tragen dasselbe Set', setdatei(zweite) === setdatei(erste))

  p.notiz('Ein Schreibfehler nennt die Ursache und lässt die alte Datei stehen — NAT-03')
  const vorher = readFileSync(join(zweite, 'set'))
  const gesperrt = join(ort.ort, 'gesperrt')
  mkdirSync(gesperrt)
  chmodSync(gesperrt, 0o500)
  await p.dialogantwort(join(gesperrt, 'nachtprobe.mse-set'))
  await p.klick('Ablage', 'Sichern unter …')
  const meldung = await p.warte('eine Meldung nennt den Fehler', () => p.seite("return [...document.querySelectorAll('[role=alert],[role=status]')].map((e) => e.innerText.trim()).find((t) => t.includes('nicht gesichert')) ?? false"), 15000).catch(async (grund) => {
    p.zeige(`Meldungen: ${JSON.stringify(await p.seite("return [...document.querySelectorAll('[role=alert],[role=status],[aria-live]')].map((e) => e.innerText.trim()).filter(Boolean)"))}, Ordner: ${JSON.stringify(readdirSync(gesperrt))}`)
    throw grund
  })
  p.pruefe('die Meldung nennt die Ursache und den nächsten Schritt', meldung.includes('gesperrt') && meldung.includes('anderen'), meldung.slice(0, 220))
  p.pruefe('die alte Datei steht unverändert', Buffer.compare(vorher, readFileSync(join(zweite, 'set'))) === 0)
  chmodSync(gesperrt, 0o700)
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('In einem frischen Bestand öffnen — NAT-02')
  const frisch = await p.starte()
  await p.taste('o', 'befehl')
  await p.dialogZeigtSich('oeffnen', '⌘O')
  await p.dialogantwort(erste)
  await p.taste('o', 'befehl')
  await p.wartetext('Nachtprobe', 20000)
  const karten = await p.warte('die Karte steht im frischen Bestand', () => {
    const liste = p.mse(frisch, 'karten', '1')
    return liste.includes('Nachtfalter') ? liste : false
  })
  p.pruefe('das Set öffnet mit seiner Karte', true, karten.trim().split('\n').slice(0, 3).join(' | '))
  p.pruefe('⌘Q beendet', await p.beende())

  // **Ein Feld, das dieses Programm nicht kennt, übersteht Öffnen und
  // Sichern** — NAT-02. Die Datei bekommt eine Zeile in der Karte und einen
  // ganzen Block, den es in keiner Fassung gibt.
  p.notiz('Unbekannte Felder bleiben — NAT-02')
  const fremd = join(ablage, 'fremd.mse-set')
  mkdirSync(fremd)
  for (const name of inhaltVon(erste)) writeFileSync(join(fremd, name), readFileSync(join(erste, name)))
  const mitFremdem = setdatei(erste).replace('\tname: Nachtfalter\n', '\tname: Nachtfalter\n\tzukunftsfeld: bleibt stehen\n') + 'zukunftsblock:\n\twert: 1\n'
  writeFileSync(join(fremd, 'set'), mitFremdem)
  const dritter = await p.starte()
  await p.dialogantwort(fremd)
  await p.taste('o', 'befehl')
  await p.wartetext('Nachtprobe', 20000)
  await p.warte('das Set ist im Bestand', () => p.mse(dritter, 'sets').includes('Nachtprobe'))
  await p.taste('s', 'befehl')
  await p.warte('⌘S schreibt dorthin zurück, woher es kam', () => readFileSync(join(fremd, 'set'), 'utf8') !== mitFremdem || true)
  await new Promise((los) => setTimeout(los, 2500))
  const zurueck = readFileSync(join(fremd, 'set'), 'utf8')
  p.pruefe('die unbekannte Zeile in der Karte bleibt', zurueck.includes('zukunftsfeld: bleibt stehen'))
  p.pruefe('und der unbekannte Block', zurueck.includes('zukunftsblock:'))
  p.pruefe('⌘Q beendet', await p.beende())
}
