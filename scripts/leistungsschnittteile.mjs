// **Die Teile des Messvertrags** — PERF-01
// (#2168).
//
// Was `leistungsschnitte.mjs` liest und entscheidet, ohne etwas zu starten:
// die Ausgabe von `go test -bench`, die Last fremder Prozesse und die Regel,
// wann ein Schnitt wartet und wann er als nicht gemessen gilt.

// benchkopf ist der Anfang einer Ergebniszeile von `go test -bench`: Name und
// Durchläufe. Der Vorsatz `-10` ist GOMAXPROCS und gehört nicht zum Namen.
// Danach folgen Paare aus Wert und Einheit — `ns/op`, **eigene Kennzahlen**
// wie `cpu-ms/op` (`b.ReportMetric`), dann `B/op` und `allocs/op`.
const benchkopf = /^(Benchmark\S+?)(?:-\d+)?\s+(\d+)\s+(.*)$/
const kennzahl = /([\d.]+)\s+(\S+)/g

/** benchLesen gibt je Benchmark die Werte aller Wiederholungen. */
export function benchLesen(text) {
  const gelesen = new Map()
  for (const zeile of text.split('\n')) {
    const treffer = benchkopf.exec(zeile.trim())
    if (!treffer) continue
    const [, name, durchlaeufe, rest] = treffer
    const werte = Object.fromEntries([...rest.matchAll(kennzahl)].map(([, wert, einheit]) => [einheit, Number(wert)]))
    if (werte['ns/op'] === undefined) continue
    if (!gelesen.has(name)) gelesen.set(name, { ns: [], bytes: [], zuteilungen: [], durchlaeufe: [] })
    const eintrag = gelesen.get(name)
    eintrag.ns.push(werte['ns/op'])
    eintrag.durchlaeufe.push(Number(durchlaeufe))
    if (werte['B/op'] !== undefined) eintrag.bytes.push(werte['B/op'])
    if (werte['allocs/op'] !== undefined) eintrag.zuteilungen.push(werte['allocs/op'])
  }

  return gelesen
}

/**
 * fremdLast summiert die CPU der Prozesse außerhalb des eigenen Baums — aus
 * `ps -A -o pid=,ppid=,pcpu=,comm=` — und nennt die fünf größten.
 *
 * **Der eigene Baum zählt nicht**: was der Läufer startet, ist die Messung
 * und nicht ihre Umgebung.
 */
export function fremdLast(ps, wurzel) {
  const prozesse = ps.split('\n').map((zeile) => zeile.trim()).filter(Boolean).map((zeile) => {
    const [pid, ppid, cpu, ...name] = zeile.split(/\s+/)
    return { pid: Number(pid), ppid: Number(ppid), cpu: Number(cpu), name: name.join(' ') }
  })
  const eigene = new Set([wurzel])
  for (let gewachsen = true; gewachsen;) {
    gewachsen = false
    for (const p of prozesse) {
      if (!eigene.has(p.pid) && eigene.has(p.ppid)) {
        eigene.add(p.pid)
        gewachsen = true
      }
    }
  }
  const fremde = prozesse.filter((p) => !eigene.has(p.pid)).sort((a, b) => b.cpu - a.cpu)
  const summe = Math.round(fremde.reduce((s, p) => s + p.cpu, 0) * 10) / 10

  return { summe, groesste: fremde.filter((p) => p.cpu > 0).slice(0, 5) }
}

/** zuViel sagt, warum eine Umgebung zu laut ist — oder `null`. */
export function zuViel(umgebung, grenzen) {
  if (umgebung.last[0] > grenzen.last) return `Last ${String(umgebung.last[0])} über ${String(grenzen.last)}`
  if (umgebung.fremd.summe > grenzen.fremd) {
    return `fremde Prozesse ${String(umgebung.fremd.summe)} % über ${String(grenzen.fremd)} %`
  }

  return null
}

/**
 * warteAufRuhe misst die Umgebung und wartet, solange sie zu laut ist —
 * höchstens `frist` Sekunden, je `takt` Sekunden ein Blick.
 *
 * **Fremde Prozesse werden nicht angehalten** (PERF-01.AK-03): gewartet wird
 * auf sie, und reicht die Frist nicht, gilt der Schnitt als nicht gemessen.
 */
export async function warteAufRuhe({ messen, schlafe, grenzen, frist, takt }) {
  let gewartetS = 0
  for (;;) {
    const umgebung = await messen()
    const grund = zuViel(umgebung, grenzen)
    if (grund === null) return { ruhig: true, gewartetS, umgebung }
    if (gewartetS >= frist) return { ruhig: false, gewartetS, umgebung, grund }
    await schlafe(takt)
    gewartetS += takt
  }
}

/**
 * gestoert sagt, ob fremde Last **während** eines Schnitts über der Grenze
 * lag — oder `null`. Befund 4 des Zweitblicks über Phase 0.7.0-I: bis dahin
 * galt die Grenze nur bis zum Anfang, und was danach kam, stand im Beleg,
 * ohne etwas zu entscheiden.
 *
 * **Nur die fremde Last**: `loadavg` enthält die eigene Arbeit, und ein
 * Schnitt, der die Kerne belegt, höbe sie selbst über jede Grenze. Die Last
 * zählt deshalb nur vorher (`zuViel`).
 */
export function gestoert(proben, grenzen, fehlgeschlagen = 0) {
  if (fehlgeschlagen > 0) return `${String(fehlgeschlagen)} Proben ohne Antwort — dass die Grenze während des Schnitts hielt, ist nicht belegt`
  if (proben.length === 0) return 'keine Probe während des Schnitts'
  const hoechste = proben.reduce((a, b) => (b.fremd.summe > a.fremd.summe ? b : a))
  if (hoechste.fremd.summe <= grenzen.fremd) return null

  const wer = hoechste.fremd.groesste.map((p) => `${p.name} ${String(p.cpu)}`).join(', ')

  return `fremde Prozesse während des Schnitts bis ${String(hoechste.fremd.summe)} % über ${String(grenzen.fremd)} % (${wer})`
}

/**
 * beobachte nimmt alle `takt` Millisekunden eine Probe, bis `ende` gerufen
 * wird — dann eine letzte, und gibt alle zurück, **samt der Zahl der
 * verpassten**: ein Blick, der scheitert, ist kein Beleg dafür, dass die
 * Grenze hielt (Befund 3 des Nachlaufs über Phase I), und `gestoert` macht
 * den Schnitt dann ungültig.
 */
export function beobachte(messen, takt) {
  const proben = []
  let fehlgeschlagen = 0
  const nimm = () => {
    try {
      proben.push(messen())
    } catch {
      fehlgeschlagen += 1
    }
  }
  const uhr = setInterval(nimm, takt)

  return {
    ende: () => {
      clearInterval(uhr)
      nimm()

      return { proben, fehlgeschlagen }
    },
  }
}

/**
 * auswaehlen gibt die gewünschten Schnitte — alle, wenn keiner genannt ist.
 *
 * **Ein unbekannter Name bricht ab**: sonst misst der Lauf nichts und heißt
 * trotzdem vollständig (gefunden am 2026-10-07 an einem ungeteilten `$TEIL`).
 */
export function auswaehlen(schnitte, gewuenscht) {
  for (const name of gewuenscht) {
    if (!schnitte.some((s) => s.name === name)) throw new Error(`unbekannter Schnitt »${name}«`)
  }

  return schnitte.filter((s) => gewuenscht.length === 0 || gewuenscht.includes(s.name))
}
