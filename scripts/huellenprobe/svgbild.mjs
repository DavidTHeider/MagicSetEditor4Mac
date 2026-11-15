// **Ein SVG über `<img>` im echten WebKit** — Befund 5 des Zweitblicks über
// UX-13 (#1685).
//
// `docs/DATEIANSICHTEN.md` sagt: ein SVG in der Ansicht führt kein Skript aus
// und lädt keine fremde Adresse. Die Primärquelle (W3C SVG Integration,
// Entwurf vom 2014-04-17, §2 und §3.4) sagt das für `<img>`, bindet HTML aber
// ausdrücklich nicht — also wird es hier **an der WKWebView der Anwendung**
// gemessen:
//
// 1. Ein Dienst auf `127.0.0.1` zählt jede Anfrage.
// 2. **Gegenprobe A**: ein gewöhnliches Bild aus der Seite erreicht ihn —
//    sonst hieße „keine Anfrage" nur, dass niemand zuhört (§3).
// 3. Das SVG — Skript, `onload`, `@import`, `<image>` auf den Dienst — als
//    `data:`-Adresse in ein `<img>`, **genau so gebaut wie
//    `internal/app/dateiansicht.go`** (`data:` + Typ + `;base64,` + Bytes).
// 4. **Gegenprobe B**: dasselbe SVG als Dokument in einem `<iframe>` fragt den
//    Dienst an — das SVG *würde* also laden, wenn man es ließe.
//
// **Beobachtend gestartet** (`MSE_CSP_BEOBACHTEN=1`, nur im Abnahmebau) —
// seit SEC-17 weist die Schutzrichtlinie jedes Bild von `127.0.0.1` ab, und
// beide Gegenproben kämen nie an: „keine Anfrage aus dem SVG" hieße dann nur,
// dass die Richtlinie davorsteht, nicht, was WebKit mit einem SVG in `<img>`
// tut. Gefunden im Gesamtlauf der Abnahme von 0.7.0. Was die Richtlinie dabei
// meldet, sind die Anfragen an den eigenen Dienst und das `data:`-Dokument
// der Gegenprobe B — erwartet, und nur sie.

import http from 'node:http'

const ruhe = (ms) => new Promise((fertig) => setTimeout(fertig, ms))

export default async function svgbild(p) {
  const anfragen = []
  const dienst = http.createServer((anfrage, antwort) => {
    anfragen.push(anfrage.url)
    antwort.writeHead(204, { 'Access-Control-Allow-Origin': '*' })
    antwort.end()
  })
  await new Promise((fertig) => dienst.listen(0, '127.0.0.1', fertig))
  const ziel = (weg) => `http://127.0.0.1:${dienst.address().port}/${weg}`
  // Erwartet sind die Anfragen an den eigenen Dienst — und das `<iframe>`
  // der Gegenprobe B selbst, ein `data:`-Dokument, das `frame-src` nicht
  // erlaubt.
  p.erwarteteVerstoesse = new RegExp(`abgewiesen=http://127\\.0\\.0\\.1:${String(dienst.address().port)}( |$)|anweisung=frame-src abgewiesen=data( |$)`)

  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40" onload="fetch('${ziel('onload')}')">`
    + `<style>@import url(${ziel('stil')});</style>`
    + `<script>fetch('${ziel('skript')}')</script>`
    + `<image href="${ziel('bild')}" width="10" height="10"/>`
    + '<rect width="40" height="40" fill="teal"/></svg>'
  const adresse = `data:image/svg+xml;base64,${Buffer.from(svg).toString('base64')}`

  try {
    await p.starte({ umgebung: { MSE_CSP_BEOBACHTEN: '1' } })

    p.notiz('Gegenprobe A — die Seite erreicht den Dienst')
    await p.seite(`const b = new Image(); b.src = ${JSON.stringify(ziel('gegenprobe'))}; document.body.append(b)`)
    await p.warte('die Gegenprobe kommt an', async () => anfragen.includes('/gegenprobe'), 10000)
    p.pruefe('der Dienst hört zu', anfragen.includes('/gegenprobe'), JSON.stringify(anfragen))

    p.notiz('Das SVG über <img>, wie die Dateiansicht es zeigt')
    const breite = await p.seite(`const b = document.createElement('img'); b.id = 'svgprobe'; b.src = ${JSON.stringify(adresse)}; document.body.append(b); await b.decode(); return b.naturalWidth`)
    p.pruefe('das SVG ist gezeichnet', breite === 40, `Breite ${breite}`)
    await ruhe(3000)
    const ausDemBild = anfragen.filter((weg) => weg !== '/gegenprobe')
    p.pruefe('kein Skript, kein onload, kein @import, kein <image> von außen', ausDemBild.length === 0, JSON.stringify(ausDemBild))

    p.notiz('Gegenprobe B — dasselbe SVG als Dokument lädt')
    await p.seite(`const r = document.createElement('iframe'); r.id = 'svgdokument'; r.src = ${JSON.stringify(adresse)}; document.body.append(r)`)
    await p.warte('das Dokument fragt an', async () => anfragen.some((weg) => weg !== '/gegenprobe'), 10000)
    p.pruefe('als Dokument fragt dasselbe SVG an', anfragen.some((weg) => weg !== '/gegenprobe'), JSON.stringify(anfragen))

    await p.seite("for (const id of ['svgprobe', 'svgdokument']) document.getElementById(id)?.remove()")
    p.pruefe('⌘Q beendet', await p.beende())
  } finally {
    dienst.close()
  }
}
