// Prüfungen von scripts/lastlaufwirkung.mjs — Z3 des Zweitblicks über Phase 6.11.
//
// **Ein 2xx ist nicht die Wirkung.** Der Lastlauf zählte jeden Ruf mit Status
// 200 bis 299 als gelungen: antwortete der Dienst 201 und schrieb kein Bild,
// stand der Ruf im Durchsatz. Hier steht, was ein Profil zusagt und wie die
// Zusage an der Stichprobe nachgeprüft wird.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { createServer } from 'node:http'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { auswerten, istGelungen } from './lastlaufbericht.mjs'
import { profile } from './lastlaufprofile.mjs'
import { hoeflich, pruefeWirkungen } from './lastlaufwirkung.mjs'
import { png } from './pngbau.mjs'
import { ruf } from './lastlaufnetz.mjs'

const ort = mkdtempSync(join(tmpdir(), 'lastwirkung-'))
const datei = (name, inhalt) => { const pfad = join(ort, name); writeFileSync(pfad, inhalt); return pfad }
const gut = (wirkung, zusatz = {}) => ({ status: 201, ms: 1, text: '', wirkung, ...zusatz })
const antwort = (daten, status = 200) => async () => ({ status, text: JSON.stringify(daten), warten: 0 })

test('ein Bild am Ziel, das vollständig ist, bestätigt den Ruf', async () => {
  const proben = [gut({ art: 'bild', ziel: datei('a.png', png()) })]
  const z = await pruefeWirkungen({ REST: proben }, {})
  assert.deepEqual([z.REST.geprueft, z.REST.fehlend], [1, 0])
  assert.equal(istGelungen(proben[0]), true)
})

test('kein Bild am Ziel: der Ruf zählt nicht mehr als gelungen — und die Zahl nennt es', async () => {
  const proben = [gut({ art: 'bild', ziel: join(ort, 'fehlt.png') })]
  const z = await pruefeWirkungen({ REST: proben }, {})
  assert.deepEqual([z.REST.geprueft, z.REST.fehlend], [1, 1])
  assert.equal(istGelungen(proben[0]), false)
  assert.match(proben[0].wirkungFehlt, /fehlt\.png.*nicht da/)
  assert.match(z.REST.beispiele[0], /fehlt\.png/)
  assert.equal(auswerten(proben, 1).fehler, 1, 'es ist ein Fehler und keine Drosselung')
  assert.match(auswerten(proben, 1).beispielFehler[0], /nicht da/)
})

test('ein abgeschnittenes Bild am Ziel bestätigt nichts', async () => {
  const ganz = png({ breite: 100, hoehe: 100 })
  const proben = [gut({ art: 'bild', ziel: datei('halb.png', ganz.subarray(0, 80)) })]
  await pruefeWirkungen({ REST: proben }, {})
  assert.match(proben[0].wirkungFehlt, /kein vollständiges PNG/)
})

test('ein Bogen ist ein PDF, das anfängt und endet wie eines', async () => {
  const proben = [
    gut({ art: 'bogen', ziel: datei('ganz.pdf', '%PDF-1.4\n…\n%%EOF\n') }),
    gut({ art: 'bogen', ziel: datei('halb.pdf', '%PDF-1.4\n…') }),
    gut({ art: 'bogen', ziel: datei('keins.pdf', 'GIF89a') }),
    gut({ art: 'bogen', ziel: join(ort, 'fehlt.pdf') }),
  ]
  const z = await pruefeWirkungen({ Aufträge: proben }, {})
  assert.deepEqual(proben.map((p) => p.wirkungFehlt === undefined), [true, false, false, false])
  assert.equal(z.Aufträge.fehlend, 3)
})

test('ein Ruf, der schon gescheitert oder gedrosselt war, hat keine Wirkung zu belegen', async () => {
  const proben = [
    gut({ art: 'bild', ziel: join(ort, 'x.png') }, { status: 429 }),
    gut({ art: 'bild', ziel: join(ort, 'y.png') }, { status: 500 }),
    { status: 201, ms: 1, text: '' },
  ]
  const z = await pruefeWirkungen({ REST: proben }, {})
  assert.deepEqual([z.REST.geprueft, z.REST.fehlend], [0, 0])
})

// **Der Endstand belegt einen Ruf und nicht alle** — N3 des Nachlaufs über die
// QA-Arbeit: hier galten alle Schreiber eines Feldes als wirksam, sobald der
// Bestand einen ihrer Werte trug; ein Dienst, der von zehn Änderungen eine
// speicherte, bestand. Belegt ist der Ruf, dessen Wert dasteht; die anderen
// sind **verdeckt** — weder Fehler noch Erfolg, und der Bericht zählt sie
// eigens.
test('ein Feld belegt den Ruf, dessen Wert dasteht — die anderen Schreiber sind verdeckt', async () => {
  const feld = (karte, wert) => gut({ art: 'feld', setID: '1', karte, feld: 'name', wert }, { status: 200 })
  const proben = [feld('10', 'Last 0-0'), feld('10', 'Last 1-0'), feld('11', 'Last 0-1'), feld('12', 'Last 0-2')]
  const abfrage = antwort({ karten: [
    { id: '10', werte: { name: 'Last 0-0' } }, // der erste steht da, der zweite ist verdeckt
    { id: '11', werte: { name: 'Ursprung' } }, // nie geschrieben
    { id: '12', werte: { name: 'Last 0-2' } },
  ] })
  const z = await pruefeWirkungen({ REST: proben }, { abfrage })
  assert.deepEqual(proben.map((p) => p.wirkungFehlt === undefined), [true, true, false, true])
  assert.deepEqual(proben.map((p) => p.wirkungVerdeckt !== undefined), [false, true, false, false])
  assert.deepEqual([z.REST.geprueft, z.REST.fehlend, z.REST.verdeckt], [4, 1, 1])
  assert.match(proben[2].wirkungFehlt, /Karte 11.*Ursprung/)
  assert.match(proben[1].wirkungVerdeckt, /Karte 10.*Last 0-0/)

  // **Verdeckt ist weder gelungen noch Fehler.**
  assert.equal(istGelungen(proben[1]), false)
  const bericht = auswerten(proben, 1)
  assert.deepEqual([bericht.gelungen, bericht.unbelegt, bericht.fehler], [2, 1, 1])
})

test('ein Bestand, den die Auskunft nicht gibt, bestätigt die Schreibrufe nicht', async () => {
  const proben = [gut({ art: 'feld', setID: '1', karte: '10', feld: 'name', wert: 'x' }, { status: 200 })]
  await pruefeWirkungen({ REST: proben }, { abfrage: async () => ({ status: 500, text: '', warten: 0 }), schlafe: async () => {} })
  assert.match(proben[0].wirkungFehlt, /Bestand nicht lesbar \(500\)/)
})

test('eine Drosselung der Auskunft wird abgewartet, nicht als Befund gezählt', async () => {
  const antworten = [{ status: 429, warten: 2, text: '' }, { status: 429, warten: 0, text: '' }, { status: 200, warten: 0, text: '{"karten":[]}' }]
  const gewartet = []
  const a = await hoeflich(async () => antworten.shift(), async (ms) => { gewartet.push(ms) })
  assert.equal(a.status, 200)
  assert.deepEqual(gewartet, [2000, 1000], 'mindestens eine Sekunde')
  const mcp = [{ status: 200, gedrosselt: true, warten: 3 }, { status: 200, gedrosselt: false, text: 'x' }]
  const mcpGewartet = []
  const b = await hoeflich(async () => mcp.shift(), async (ms) => { mcpGewartet.push(ms) })
  assert.equal(b.text, 'x', 'über MCP ist die Drosselung ein Satz im Ergebnis, kein 429')
  assert.deepEqual(mcpGewartet, [3000])
  const dauerhaft = await hoeflich(async () => ({ status: 429, warten: 1, text: '' }), async () => {})
  assert.equal(dauerhaft.status, 429, 'nach acht Versuchen gibt sie auf')
})

test('eine Einfuhr ist bestätigt, wenn ein neues Set mit dem Namen und den Karten dasteht', async () => {
  const w = { art: 'einfuhr', vorher: ['1'], name: 'Probeset', karten: 2 }
  const proben = [gut(w), gut(w)]
  const abfrage = async (weg) => (weg === '/sets'
    ? { status: 200, text: JSON.stringify({ sets: [{ id: '1', name: 'Probeset' }, { id: '5', name: 'Probeset' }] }), warten: 0 }
    : { status: 200, text: JSON.stringify({ karten: [{}, {}] }), warten: 0 })
  const z = await pruefeWirkungen({ REST: proben }, { abfrage })
  assert.deepEqual([z.REST.geprueft, z.REST.fehlend], [2, 0])
})

test('eine Einfuhr ohne neues Set — oder mit zu wenigen Karten — bestätigt keinen Ruf', async () => {
  const w = { art: 'einfuhr', vorher: ['1'], name: 'Probeset', karten: 2 }
  const ohne = [gut(w)]
  await pruefeWirkungen({ REST: ohne }, { abfrage: antwort({ sets: [{ id: '1', name: 'Probeset' }] }) })
  assert.match(ohne[0].wirkungFehlt, /kein neues Set „Probeset“/)

  const leer = [gut(w)]
  const abfrage = async (weg) => (weg === '/sets'
    ? { status: 200, text: JSON.stringify({ sets: [{ id: '5', name: 'Probeset' }] }), warten: 0 }
    : { status: 200, text: JSON.stringify({ karten: [{}] }), warten: 0 })
  await pruefeWirkungen({ REST: leer }, { abfrage })
  assert.match(leer[0].wirkungFehlt, /1 von 2 Karten/)
})

test('die Kartenliste trägt, was der Bestand an Karten hat', async () => {
  const gutes = gut({ art: 'karten', anzahl: 2 }, { text: '{"karten":[{"id":"1"},{"id":"2"}]}' })
  const kurz = gut({ art: 'karten', anzahl: 2 }, { text: '{"karten":[{"id":"1"}]}' })
  const leerer = gut({ art: 'karten', anzahl: 2 }, { text: '' })
  await pruefeWirkungen({ REST: [gutes, kurz, leerer] }, {})
  assert.equal(gutes.wirkungFehlt, undefined)
  assert.match(kurz.wirkungFehlt, /1 statt 2 Karten/)
  assert.match(leerer.wirkungFehlt, /keine Kartenliste/)
})

test('die Setliste ist eine Liste — antwortet der Dienst anders, zählt der Ruf nicht', async () => {
  const gutes = gut({ art: 'sets' }, { text: '{"sets":[{"id":"3"}]}' })
  const leer = gut({ art: 'sets' }, { text: '{"sets":[]}' })
  const falsch = gut({ art: 'sets' }, { text: '{}' })
  await pruefeWirkungen({ REST: [gutes, leer, falsch] }, {})
  assert.deepEqual([gutes.wirkungFehlt, leer.wirkungFehlt], [undefined, undefined])
  assert.match(falsch.wirkungFehlt, /keine Setliste/)
})

test('über MCP gelesene Karten werden aus dem Text des Ergebnisses gezählt — als Liste oder als Hülle', async () => {
  const liste = gut({ art: 'mcp-karten', anzahl: 2 }, { inhalt: '[{"id":1},{"id":2}]' })
  const huelle = gut({ art: 'mcp-karten', anzahl: 2 }, { inhalt: '{"karten":[{"id":1},{"id":2}]}' })
  const leer = gut({ art: 'mcp-karten', anzahl: 2 }, { inhalt: 'ERROR' })
  await pruefeWirkungen({ MCP: [liste, huelle, leer] }, {})
  assert.deepEqual([liste.wirkungFehlt, huelle.wirkungFehlt], [undefined, undefined])
  assert.match(leer.wirkungFehlt, /keine Kartenliste/)
})

test('eine Auskunft, die JSON sein soll, ist es — oder der Ruf zählt nicht', async () => {
  const gutes = gut({ art: 'json' }, { text: '{"kosten":{}}' })
  const kaputt = gut({ art: 'json' }, { text: '<html>' })
  await pruefeWirkungen({ REST: [gutes, kaputt] }, {})
  assert.equal(gutes.wirkungFehlt, undefined)
  assert.match(kaputt.wirkungFehlt, /kein JSON/)
})

test('eine Art, die niemand prüft, ist ein Fehler im Profil und wird nicht stillschweigend bestätigt', async () => {
  await assert.rejects(pruefeWirkungen({ REST: [gut({ art: 'erfunden' })] }, {}), /unbekannte Art.*erfunden/)
})

test('geprüft wird eine Stichprobe, gleichmäßig verteilt, mit dem ersten und dem letzten Ruf', async () => {
  const proben = Array.from({ length: 1000 }, (_, i) => gut({ art: 'bild', ziel: join(ort, `nie-${i}.png`) }))
  const z = await pruefeWirkungen({ REST: proben }, { stichprobe: 20 })
  assert.equal(z.REST.geprueft, 20)
  assert.equal(z.REST.fehlend, 20)
  const geflaggt = proben.map((p, i) => [p, i]).filter(([p]) => p.wirkungFehlt).map(([, i]) => i)
  assert.equal(geflaggt[0], 0)
  assert.equal(geflaggt.at(-1), 999)
  assert.equal(auswerten(proben, 1).fehler, 20, 'gezählt werden die geprüften Rufe, nicht ein Hochrechnen')
})

// **Der Dienst, der 200 sagt und nichts tut.** Ein Double auf der Schleife:
// es antwortet wie `mse-rest` und schreibt nichts — oder, ehrlich, alles.
function dienstDouble({ ehrlich }) {
  const namen = new Map()
  const sets = [{ id: '1', name: 'Saat', spiel: 'magic', stil: '' }]
  const server = createServer((anfrage, antwortStrom) => {
    const stuecke = []
    anfrage.on('data', (s) => stuecke.push(s))
    anfrage.on('end', () => {
      const rumpf = stuecke.length > 0 ? JSON.parse(Buffer.concat(stuecke).toString()) : {}
      const senden = (status, daten) => { antwortStrom.writeHead(status, { 'content-type': 'application/json' }); antwortStrom.end(JSON.stringify(daten)) }
      const pfad = anfrage.url
      if (anfrage.method === 'POST' && /^\/karten\/\w+\/bild$/.test(pfad)) {
        if (ehrlich) writeFileSync(rumpf.ziel, png({ breite: 30, hoehe: 40 }))
        return senden(201, { ziel: rumpf.ziel, breite: 30, hoehe: 40, befunde: [] })
      }
      let treffer
      if (anfrage.method === 'PUT' && (treffer = /^\/sets\/1\/karten\/(\w+)\/felder\/name$/.exec(pfad))) {
        if (ehrlich) namen.set(treffer[1], rumpf.wert)
        return senden(200, { vorher: '', hatte: true })
      }
      if (anfrage.method === 'GET' && pfad === '/sets/1/karten') {
        return senden(200, { karten: ['10', '11'].map((id) => ({ id, nummer: 1, name: namen.get(id) ?? 'Ursprung', werte: { name: namen.get(id) ?? 'Ursprung' } })) })
      }
      if (anfrage.method === 'POST' && pfad === '/mtgjson') {
        if (ehrlich && sets.length === 1) sets.push({ id: '2', name: 'Probeset', spiel: 'magic', stil: '' })
        return senden(201, { angelegt: ['2'], befunde: [] })
      }
      if (anfrage.method === 'GET' && pfad === '/sets') return senden(200, { sets })
      if (anfrage.method === 'GET' && pfad === '/sets/2/karten') return senden(200, { karten: [{ id: '20' }, { id: '21' }] })
      return senden(404, {})
    })
  })
  return new Promise((fertig) => server.listen(0, '127.0.0.1', () => fertig({
    dienst: { adresse: `http://127.0.0.1:${server.address().port}`, marke: 'x' }, schliessen: () => new Promise((z) => server.close(z)),
  })))
}

for (const [name, wie] of [['rest-rendern', 'Bilder'], ['rest-schreiben', 'Felder'], ['rest-einfuhr', 'Sets']]) {
  test(`Profil ${name}: ein Dienst, der ${wie} zusagt und nichts schreibt, zählt als Fehler — ein ehrlicher nicht`, async () => {
    const arbeit = mkdtempSync(join(tmpdir(), 'lastwirkung-'))
    mkdirSync(join(arbeit, 'bilder'), { recursive: true })
    const mtgjson = join(arbeit, 'probe.json')
    writeFileSync(mtgjson, JSON.stringify({ data: { name: 'Probeset', code: 'PRB', cards: [{ name: 'a' }, { name: 'b' }] } }))
    for (const ehrlich of [false, true]) {
      const { dienst, schliessen } = await dienstDouble({ ehrlich })
      try {
        const k = { rest: dienst, karten: ['10', '11'], setID: '1', klienten: 2, dauer: 0.3, arbeit, fall: `t-${ehrlich}`, mtgjson }
        const lauf = await profile[name](k)
        const abfrage = (weg) => ruf(dienst, 'GET', weg)
        if (name === 'rest-einfuhr') assert.match(lauf.befund, /^\d+ Rufe: \d+ legten? ein Set an, \d+ erkannten die Quelle als schon offen, 0 abgewiesen, \d+ gedrosselt$/)
        const z = await pruefeWirkungen(lauf.gruppen, { abfrage })
        const w = auswerten(lauf.gruppen.REST, 1)
        const gruppe = z.REST
        assert.ok(gruppe.geprueft > 0, 'das Profil trägt seine Zusage')
        if (ehrlich) {
          assert.equal(w.fehler, 0, `ehrlich: ${w.beispielFehler}`)
          assert.equal(gruppe.fehlend, 0)
        } else {
          assert.equal(gruppe.fehlend, gruppe.geprueft, 'unehrlich: nichts ist da')
          assert.ok(w.fehler > 0, 'der Durchsatz weist die Rufe als Fehler aus')
          assert.ok(w.gelungen < lauf.gruppen.REST.length)
        }
      } finally {
        await schliessen()
      }
    }
  })
}

// **Eine Variante erbt kein Set der vorigen** — gefunden im Lastlauf vom
// 2026-10-05: beide Varianten lesen dieselbe Quelle in denselben Bestand.
// Die erste legte „Probeset" an, die zweite erkannte es nur noch, und die
// Wirkungsprüfung fand kein **neues** Set — ein Befund der Messung, nicht des
// Dienstes. Das Profil räumt ein gleichnamiges Set vorher weg.
test('Profil rest-einfuhr: ein Set aus einer vorigen Variante wird vorher entfernt', async () => {
  const arbeit = mkdtempSync(join(tmpdir(), 'lastwirkung-'))
  const mtgjson = join(arbeit, 'probe.json')
  writeFileSync(mtgjson, JSON.stringify({ data: { name: 'Probeset', code: 'PRB', cards: [{ name: 'a' }, { name: 'b' }] } }))
  let naechste = 3
  const sets = [{ id: '1', name: 'Saat' }, { id: '2', name: 'Probeset' }]
  const server = createServer((anfrage, antwortStrom) => {
    anfrage.resume()
    anfrage.on('end', () => {
      const senden = (status, daten) => { antwortStrom.writeHead(status, { 'content-type': 'application/json' }); antwortStrom.end(JSON.stringify(daten)) }
      const pfad = anfrage.url
      // **Wie `mse-rest`: gelöscht wird nur mit dem heutigen Namen** als
      // Bestätigung (`?name=…`, `internal/rest/setloeschen.go`).
      if (anfrage.method === 'DELETE' && /^\/sets\/\d+\?/.test(pfad)) {
        const adresse = new URL(pfad, 'http://x')
        const stelle = sets.findIndex((s) => `/sets/${s.id}` === adresse.pathname && s.name === adresse.searchParams.get('name'))
        if (stelle < 0) return senden(409, { fehler: { code: 'konflikt' } })
        sets.splice(stelle, 1)
        return senden(200, { befunde: [] })
      }
      if (anfrage.method === 'POST' && pfad === '/mtgjson') {
        const da = sets.find((s) => s.name === 'Probeset')
        if (da) return senden(201, { angelegt: [], befunde: [], schonOffen: true })
        sets.push({ id: String(naechste), name: 'Probeset' })
        return senden(201, { angelegt: [String(naechste++)], befunde: [] })
      }
      if (anfrage.method === 'GET' && pfad === '/sets') return senden(200, { sets })
      if (anfrage.method === 'GET' && /^\/sets\/\d+\/karten$/.test(pfad)) return senden(200, { karten: [{ id: '20' }, { id: '21' }] })
      return senden(404, {})
    })
  })
  await new Promise((fertig) => server.listen(0, '127.0.0.1', fertig))
  const dienst = { adresse: `http://127.0.0.1:${server.address().port}`, marke: 'x' }
  try {
    const k = { rest: dienst, klienten: 2, dauer: 0.3, arbeit, fall: 't-geerbt', mtgjson }
    const lauf = await profile['rest-einfuhr'](k)
    assert.match(lauf.befund, /^\d+ Rufe: 1 legte ein Set an/)
    const z = await pruefeWirkungen(lauf.gruppen, { abfrage: (weg) => ruf(dienst, 'GET', weg) })
    assert.equal(z.REST.fehlend, 0, 'die Wirkung ist belegt')
  } finally {
    await new Promise((z) => server.close(z))
  }
})

// **Null geprüft ist nicht grün** — QA-LAST01
// (#2092).
// Der Lastlauf für 0.7.0 endete mit Exit 0, obwohl Rendern, Einfuhr und
// Aufträge **keinen** Ruf durchbrachten: geprüft wurden nur gelungene Rufe,
// und ohne sie gab es null Prüfungen und null fehlende Wirkungen.
test('eine Gruppe, die Wirkung zusagt und keinen Ruf durchbringt, ist ungeprüft', async () => {
  const abgewiesen = (wirkung) => ({ status: 400, ms: 1, text: '{"fehler":"verborgen"}', wirkung })
  const z = await pruefeWirkungen({
    REST: [abgewiesen({ art: 'bild', ziel: join(ort, 'x.png') }), abgewiesen({ art: 'bild', ziel: join(ort, 'y.png') })],
    gedrosselt: [{ status: 429, ms: 1, text: '', warten: 1, wirkung: { art: 'bild', ziel: join(ort, 'z.png') } }],
  }, {})
  assert.deepEqual([z.REST.zugesagt, z.REST.geprueft], [2, 0])
  assert.match(z.REST.ungeprueft, /keiner von 2 Rufen mit zugesagter Wirkung gelang — 400 × 2/)
  assert.match(z.gedrosselt.ungeprueft, /429 × 1/)
})

test('eine Gruppe ohne zugesagte Wirkung ist nicht ungeprüft', async () => {
  const z = await pruefeWirkungen({ REST: [{ status: 200, ms: 1, text: '' }] }, {})
  assert.equal(z.REST.ungeprueft, undefined)
})

test('ein geprüfter Ruf hebt »ungeprüft« auf', async () => {
  const z = await pruefeWirkungen({
    REST: [gut({ art: 'bild', ziel: datei('da.png', png()) }), { status: 400, ms: 1, text: '', wirkung: { art: 'bild', ziel: join(ort, 'n.png') } }],
  }, {})
  assert.equal(z.REST.ungeprueft, undefined)
  assert.equal(z.REST.geprueft, 1)
})

// **Nur gedrosselt ist eine eigene Lage** — mit Kontingenten ist die Drosselung
// die Zusage, ohne sie ein Befund. Der Lauf entscheidet je Variante; hier
// steht nur, ob eine ungeprüfte Gruppe ausschließlich gedrosselt wurde.
test('eine ungeprüfte Gruppe sagt, ob sie nur gedrosselt wurde', async () => {
  const gedrosselt = { status: 429, ms: 1, text: '', warten: 1, wirkung: { art: 'bild', ziel: join(ort, 'g.png') } }
  const abgewiesen = { status: 400, ms: 1, text: '', wirkung: { art: 'bild', ziel: join(ort, 'a.png') } }
  const z = await pruefeWirkungen({ nur: [gedrosselt, gedrosselt], gemischt: [gedrosselt, abgewiesen] }, {})
  assert.equal(z.nur.nurGedrosselt, true)
  assert.equal(z.gemischt.nurGedrosselt, false)
})
