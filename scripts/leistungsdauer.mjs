// **Dauerbetrieb und Lastgrenze** — QA-04.AK-03 und AK-04
// (#1735).
//
// Beide Teile haben ein **eigenes Budget in Sekunden** (`LANGLAUF`, `LAST`)
// und laufen erst, wenn die gepaarten Aufgaben gemessen sind.

import { Agent, request } from 'node:http'
import { rmSync } from 'node:fs'
import { join } from 'node:path'
import { performance } from 'node:perf_hooks'

import { verteilung } from './leistungsstatistik.mjs'

/**
 * langlauf zeichnet in **einer** Sitzung je Programm, so lange das Budget
 * reicht, reihum die Karten der Auswahl aus Multicolor. Alle zwei Sekunden
 * wird der residente Stand gelesen; verglichen wird die Dauer im ersten
 * Fünftel der Schritte mit der im letzten.
 *
 * **Was mitwächst, ohne ein Leck zu sein**: MSE2 merkt sich jeden
 * geschriebenen Dateinamen (`ExportInfo::exported_images`,
 * `src/script/functions/export.cpp`, Zeile 425, Fassung f9d9356 vom
 * 2021-07-06) — ein paar Byte je Schritt.
 */
export async function langlauf(c) {
  const karten = c.auswahl.filter((k) => k.satz === 'Multicolor')
  const aus = c.ausgabe('langlauf')
  const ergebnis = {}
  for (const t of c.teilnehmer.filter((x) => c.k.langlaufTeilnehmer.includes(x.name))) {
    const gewartetS = await c.warte()
    // **Das Guthaben ist abgeschaltet**, wo es nicht Gegenstand ist — einmal
    // je Lauf, vom Läufer (`leistungskontingent.mjs`).
    const s = await (t.name === 'mse2' ? t.sitzung(c.satz('Multicolor')) : t.sitzung({ ...c.satz('Multicolor'), id: t.ids.Multicolor }, t.vorrat))
    const pid = s.lauf.pid()
    const anfang = performance.now()
    const speicher = []
    const messen = () => speicher.push({ sekunde: Math.round((performance.now() - anfang) / 100) / 10, byte: c.residentByte(pid) })
    messen()
    const takt = setInterval(messen, 2000)
    const { versuche, fehlertexte } = await zeichneReihum(s, karten, aus, t.name, {
      genug: () => performance.now() - anfang >= c.k.langlauf * 1000,
    })
    clearInterval(takt)
    messen()
    const ende = await s.lauf.schliessen()
    const zahlen = zaehleVersuche(versuche)
    // **Das erste und das letzte Fünftel der gelungenen Bilder** — eine
    // Ablehnung ist schnell und verfälschte den Vergleich.
    const gelungeneDauern = versuche.filter((v) => v.ok && v.dauerMs !== null).map((v) => v.dauerMs)
    const fuenftel = Math.max(1, Math.floor(gelungeneDauern.length / 5))
    ergebnis[t.name] = {
      sekunden: Math.round(performance.now() - anfang) / 1000, ...zahlen,
      absturz: ende.absturz, rueckgabe: ende.rueckgabe,
      anfang: verteilung(gelungeneDauern.slice(0, fuenftel)), schluss: verteilung(gelungeneDauern.slice(-fuenftel)),
      speicher, cpuS: ende.cpuS, rssByte: ende.rssByte, gewartetS, pipelineDanach: c.pipelineLaeuft(),
      fehlertexte,
    }
    console.log(`  langlauf ${t.name}: ${zahlen.versuche} Versuche, ${zahlen.gelungen} gelungen, ${zahlen.fehler} Fehler, ` +
      `Median ${Math.round(zahlen.dauer.median ?? 0)} ms, RSS ${speicher.at(0)?.byte} → ${speicher.at(-1)?.byte}`)
  }
  c.extra.langlauf = ergebnis
}

/**
 * zaehleVersuche zählt die Versuche einer Sitzung — **jeden einmal**.
 *
 * Ein Versuch ist ein Ruf zum Zeichnen; er gelingt oder scheitert. Die Summe
 * aus gelungenen und gescheiterten ist die Zahl der Versuche und nichts
 * anderes: hier stand `schritte = Dauern + Fehler`, und ein gescheiterter Ruf
 * von MSE2 trägt eine Dauer — er stand in beiden, 2663 gegen 2537 Versuche.
 *
 * **Zwei Verteilungen**: die Dauer der Bilder, die entstanden sind, und die
 * der Ablehnungen. Eine Ablehnung ist schnell, und in einer gemeinsamen
 * Verteilung zöge sie Median und Fünftel zu ihren Gunsten.
 */
export function zaehleVersuche(versuche) {
  const gelungen = versuche.filter((v) => v.ok)
  const gescheitert = versuche.filter((v) => !v.ok)
  const dauern = (liste) => liste.map((v) => v.dauerMs).filter((d) => d !== null && d !== undefined)

  return {
    versuche: versuche.length, gelungen: gelungen.length, fehler: gescheitert.length,
    dauer: verteilung(dauern(gelungen)), ablehnung: verteilung(dauern(gescheitert)),
  }
}

/** pause wartet — ersetzbar, damit eine Prüfung nicht wartet. */
const pause = (ms) => new Promise((weiter) => setTimeout(weiter, ms))

/**
 * zeichneReihum zeichnet reihum die Karten der Auswahl, bis `genug` es sagt,
 * und gibt jeden Versuch mit seiner Dauer. **Ein Versuch ohne eigene Dauer**
 * — `mse-mcp` wirft bei einer Abweisung — bekommt die gemessene.
 */
export async function zeichneReihum(s, karten, aus, name, {
  genug, jetzt = () => performance.now(), schlafe = pause, loesche = (pfad) => rmSync(pfad, { force: true }),
}) {
  const versuche = []
  const fehlertexte = new Set()
  for (let n = 0; !genug(); n++) {
    const k = karten[n % karten.length]
    const ziel = join(aus, `${name}-${n}.png`)
    const anfang = jetzt()
    const r = await s.zeichne(k.stelle, ziel)
    const gemessen = jetzt() - anfang
    loesche(ziel)
    versuche.push({ ok: r.ok, dauerMs: r.dauerMs ?? gemessen })
    if (!r.ok) {
      if (fehlertexte.size < 3) fehlertexte.add(`${k.stelle}: ${r.text.slice(-300)}`)
      // **Eine Abweisung wird abgewartet**, nicht übergangen — sonst zählt
      // die Schleife in derselben Sekunde Hunderttausende davon (#2005).
      const warten = wartezeitAus(r.text)
      if (warten > 0) await schlafe(warten * 1000)
    }
  }
  return { versuche, fehlertexte: [...fehlertexte] }
}

/**
 * wartezeitAus liest aus der Abweisung des Werkzeugservers, wie lange zu
 * warten ist — „zu viele Rufe — in N s wieder“ (`internal/mcp/server.go`).
 * Jeder andere Fehler gibt 0.
 */
export function wartezeitAus(text) {
  const treffer = /zu viele Rufe — in (\d+) s wieder/.exec(text)
  return treffer ? Number(treffer[1]) : 0
}

/** holen schickt eine Anfrage und gibt Status und Dauer. */
function holen(agent, adresse, pfad, marke) {
  return new Promise((fertig) => {
    const anfang = performance.now()
    const anfrage = request(`${adresse}${pfad}`, { agent, headers: { Authorization: `Bearer ${marke}` } }, (antwort) => {
      antwort.resume()
      antwort.on('end', () => fertig({
        status: antwort.statusCode, dauerMs: performance.now() - anfang,
        warten: Number(antwort.headers['retry-after'] ?? 0),
      }))
    })
    anfrage.on('error', (fehler) => fertig({ status: 0, dauerMs: performance.now() - anfang, fehler: fehler.message }))
    anfrage.end()
  })
}

/** adresseAus wartet, bis `mse-rest` Adresse und Marke angesagt hat. */
function adresseAus(kind) {
  return new Promise((fertig, scheitern) => {
    let text = ''
    kind.stdout.setEncoding('utf8')
    kind.stdout.on('data', (stueck) => {
      text += stueck
      const adresse = text.match(/Dienst auf (http:\/\/127\.0\.0\.1:\d+)/)
      const marke = text.match(/Marke ([0-9a-f]+)/)
      if (adresse && marke) fertig({ adresse: adresse[1], marke: marke[1] })
    })
    kind.on('close', () => scheitern(new Error(`mse-rest endete ohne Adresse: ${text}`)))
  })
}

/**
 * last setzt den Dienst unter gleichzeitige Leseanfragen — 1, 4 und 16 auf
 * einmal, je Stufe `LAST` Sekunden und **je Stufe ein frischer Dienst**, denn
 * das Guthaben je Principal (`mse grenzen`) gilt über die Stufen hinweg.
 * **Ein Aufrufer, der 429 bekommt, wartet, wie `Retry-After` es sagt** — so
 * misst die Stufe, was ein höflicher Klient erreicht, statt wie schnell der
 * Dienst ablehnt. **Nur MSE4Mac**: MSE2 hat keinen Dienst. Die Marke steht nur
 * im Speicher dieses Laufs und nicht im Beleg.
 */
export async function last(c) {
  const t = c.mse4macs.find((x) => x.name === 'heute')
  const gewartetS = await c.warte()
  const pfad = `/sets/${t.ids.Archenemy}/karten`
  const stufen = []
  for (const parallel of [1, 4, 16]) {
    const lauf = c.starte(join(t.verzeichnis, 'mse-rest'), ['--port', '0'], { umgebung: t.umgebung(t.vorrat) })
    const { adresse, marke } = await adresseAus(lauf.kind)
    const pid = lauf.pid()
    const agent = new Agent({ keepAlive: true, maxSockets: parallel })
    const anfang = performance.now()
    const ende = anfang + c.k.last * 1000
    const ergebnisse = []
    await Promise.all(Array.from({ length: parallel }, async () => {
      while (performance.now() < ende) {
        const e = await holen(agent, adresse, pfad, marke)
        ergebnisse.push({ ...e, nachMs: performance.now() - anfang })
        if (e.status === 429) await new Promise((weiter) => setTimeout(weiter, Math.max(e.warten, 0.2) * 1000))
      }
    }))
    agent.destroy()
    const rss = c.residentByte(pid)
    // **Das Programm bekommt das Signal, nicht `time`** — sonst bliebe der
    // Dienst als Waise zurück.
    process.kill(pid, 'SIGTERM')
    const schluss = await lauf.schliessen()
    const gut = ergebnisse.filter((e) => e.status === 200)
    const abgelehnt = ergebnisse.filter((e) => e.status === 429)
    stufen.push({
      parallel, sekunden: c.k.last, anfragen: ergebnisse.length,
      fehler: ergebnisse.length - gut.length - abgelehnt.length, abgelehnt: abgelehnt.length,
      status: Object.fromEntries([...new Set(ergebnisse.map((e) => e.status))].map((s) => [s, ergebnisse.filter((e) => e.status === s).length])),
      ersteAblehnungNachMs: abgelehnt.length > 0 ? Math.round(Math.min(...abgelehnt.map((e) => e.nachMs))) : null,
      durchsatz: gut.length / c.k.last, dauer: verteilung(gut.map((e) => e.dauerMs)),
      ablehnung: verteilung(abgelehnt.map((e) => e.dauerMs)), rssByte: rss, cpuS: schluss.cpuS, absturz: schluss.absturz,
    })
    console.log(`  last ${parallel}: ${gut.length} angenommen, ${abgelehnt.length} mit 429, ` +
      `${stufen.at(-1).fehler} Fehler, Median ${Math.round(stufen.at(-1).dauer.median ?? 0)} ms`)
  }
  c.extra.last = { programm: t.name, pfad: '/sets/{Archenemy}/karten', stufen, gewartetS, pipelineDanach: c.pipelineLaeuft() }
}
