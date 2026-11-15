// **Die Rechnung des Leistungsvergleichs** — QA-04
// (#1735).
//
// Reine Funktionen: kein Prozess, kein Dateisystem. Der Läufer
// (`scripts/leistungsvergleich.mjs`) misst, hier wird gerechnet — und
// geprüft in `scripts/leistungsstatistik.test.mjs`.
//
// # Quantile wie beim Kanonlauf
//
// **Nächste-Rang-Regel, ohne Interpolation** — dieselbe wie in
// `docs/MESSMETHODIK.md` §5: `Q(p) = x_(k)` mit `k = ⌈p · n⌉`. Eine
// interpolierte Zahl sähe genauer aus, als die Messung ist. **Bei kleinem n
// fällt p99 auf das Maximum**, und das steht so im Beleg: wer aus fünf
// Wiederholungen ein p99 liest, liest das Maximum.

/** quantil gibt das p-Quantil nach der Nächste-Rang-Regel — `null` ohne Werte. */
export function quantil(werte, p) {
  if (werte.length === 0) return null
  const geordnet = [...werte].sort((a, b) => a - b)
  const rang = Math.max(1, Math.ceil(p * geordnet.length))
  return geordnet[Math.min(rang, geordnet.length) - 1]
}

/** verteilung fasst Werte zusammen: n, Ränder, Median, p95, p99 und Mittel. */
export function verteilung(werte) {
  if (werte.length === 0) return { n: 0 }
  const summe = werte.reduce((a, b) => a + b, 0)
  return {
    n: werte.length,
    min: quantil(werte, 0),
    median: quantil(werte, 0.5),
    p95: quantil(werte, 0.95),
    p99: quantil(werte, 0.99),
    max: quantil(werte, 1),
    mittel: summe / werte.length,
  }
}

// # Die Ausgabe von `/usr/bin/time -l`
//
// **macOS schreibt sie ans Ende der Fehlerausgabe** — nach allem, was das
// Programm selbst dort ablegt. Gelesen wird deshalb die **letzte** Zeile mit
// `real … user … sys`; eine frühere könnte Protokoll sein. Zwei Speicherzahlen:
// `maximum resident set size` (der höchste residente Stand, in Byte) und
// `peak memory footprint` (was macOS dem Prozess als Spitze anrechnet,
// komprimierter Speicher eingeschlossen). **Die Zeiten haben zwei
// Nachkommastellen** — für die Wandzeit misst der Läufer deshalb selbst.

const zeitzeile = /^\s*([\d.]+) real\s+([\d.]+) user\s+([\d.]+) sys\s*$/
const rsszeile = /^\s*(\d+)\s+maximum resident set size\s*$/
const spitzenzeile = /^\s*(\d+)\s+peak memory footprint\s*$/

/** zeitLesen liest die Zahlen von `/usr/bin/time -l` — `null`, wenn keine dastehen. */
export function zeitLesen(text) {
  const zeilen = text.split('\n')
  let anfang = -1
  for (let i = zeilen.length - 1; i >= 0; i--) {
    if (zeitzeile.test(zeilen[i])) {
      anfang = i
      break
    }
  }
  if (anfang < 0) return null

  const [, wand, nutzer, system] = zeilen[anfang].match(zeitzeile)
  const ergebnis = {
    wand: Number(wand), nutzer: Number(nutzer), system: Number(system),
    rss: null, spitze: null,
    abgebrochen: zeilen.slice(0, anfang).some((z) => z.includes('Command terminated abnormally')),
  }
  for (const zeile of zeilen.slice(anfang + 1)) {
    const rss = zeile.match(rsszeile)
    if (rss) ergebnis.rss = Number(rss[1])
    const spitze = zeile.match(spitzenzeile)
    if (spitze) ergebnis.spitze = Number(spitze[1])
  }
  return ergebnis
}

// # Reihenfolge
//
// **Je Wiederholung um eine Stelle gedreht** — ein lateinisches Quadrat: bei
// einer Zahl von Wiederholungen, die ein Vielfaches der Programme ist, steht
// jedes gleich oft an jeder Stelle. Wer immer zuerst läuft, bezahlt den
// kalten Seitenspeicher des Betriebssystems allein.

/** reihenfolge gibt die Programme für Wiederholung i, um i gedreht. */
export function reihenfolge(i, namen) {
  const k = i % namen.length
  return [...namen.slice(k), ...namen.slice(0, k)]
}

/**
 * reihenfolgeEffekt vergleicht je Programm den Median der Läufe, die vorn
 * standen, mit dem der übrigen. Ein großer Unterschied heißt: die Reihenfolge
 * misst mit, und der Median über alle ist ein Mittel aus zwei Zuständen.
 */
export function reihenfolgeEffekt(laeufe) {
  const ergebnis = {}
  for (const programm of [...new Set(laeufe.map((l) => l.programm))]) {
    const eigene = laeufe.filter((l) => l.programm === programm)
    const vorn = eigene.filter((l) => l.platz === 0).map((l) => l.wert)
    const dahinter = eigene.filter((l) => l.platz !== 0).map((l) => l.wert)
    ergebnis[programm] = {
      vorn: quantil(vorn, 0.5), dahinter: quantil(dahinter, 0.5),
      nVorn: vorn.length, nDahinter: dahinter.length,
    }
  }
  return ergebnis
}
