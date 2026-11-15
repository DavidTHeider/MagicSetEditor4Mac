// **Die Bewertung des Dauerbetriebs** — B04
// (#2223),
// E06 (#2278).
//
// **Ein eigenes Modul und kein Drehbuch**: was hier steht, rechnet nur über
// Proben, und `scripts/huellenprobe/` hält jede Datei für ein Drehbuch. Das
// Drehbuch `dauerbetrieb` gibt die Namen weiter.

// **Die Regel, festgelegt vor dem Lauf, der sie belegt** (Vermerk 0204,
// #2527): nach einer Stunde Aufwärmen höchstens 10 % Wachstum je Stunde, kein
// Anstieg über drei benachbarte Abschnitte — **ein Abschnitt zählt als Anstieg
// erst ab 1 % über dem vorigen** —, und nach Ruhe und Speicherwarnung liegt
// WebKit unter seinem Median im ersten Abschnitt ([freigegeben]).
//
// **Warum die Stunde**: zweimal gemessen wärmte WebKit genau so lange auf
// (2026-10-12 +109 MB, 2026-10-20 +84 MiB). **Warum 1 %**: das Restwachstum der
// Anwendung aus #2312 liegt bei rund 0,6 % je Abschnitt und brach die Regel bis
// dahin allein. Bis zum 2026-10-20 standen hier 20 Minuten und kein Mindestmaß.
export const AUFWAERMEN_MIN = 60
export const MINDESTANSTIEG = 0.01

// abschnitte teilt die Proben nach der Aufwärmzeit in 20-Minuten-Abschnitte
// und gibt je Abschnitt den Median des Speichers, die Mitte der Zeit und ob er
// **vollständig** ist — seine Proben reichen über neun Zehntel des Fensters.
// Bis zum Zweitblick über Phase B (Befund 7) war jeder belegte Abschnitt ein
// ganzer: der fünf Minuten lange Rest des 125-Minuten-Laufs zählte wie zwanzig.
export function abschnitte(proben, aufwaermenMin = AUFWAERMEN_MIN, laengeMin = 20) {
  const nach = proben.filter((x) => x.minute >= aufwaermenMin)
  const gruppen = new Map()
  for (const x of nach) {
    const nummer = Math.floor((x.minute - aufwaermenMin) / laengeMin)
    gruppen.set(nummer, [...(gruppen.get(nummer) ?? []), x])
  }

  const median = (werte) => [...werte].sort((a, b) => a - b)[Math.floor(werte.length / 2)]

  return [...gruppen.entries()].sort(([a], [b]) => a - b).map(([nummer, xs]) => {
    const minuten = xs.map((x) => x.minute)
    return {
      nummer,
      median: median(xs.map((x) => x.gesamt)),
      mitte: median(minuten),
      vollstaendig: Math.max(...minuten) - Math.min(...minuten) >= 0.9 * laengeMin - 0.5,
    }
  })
}

// bewertung prüft Wachstum je Stunde und stetigen Anstieg — **nur über
// vollständige Abschnitte**, die Rate aus der Zeit zwischen ihren Mitten und
// der Anstieg nur über drei **benachbarte**.
//
// **Ungültig statt still** (Zweitblick über die Folgen vor 0.8.0, Befund 12):
// eine Probe nach der Aufwärmzeit ohne vollständige Messwerte (`gesamt` ist
// `null`) zählte sonst mit 0, und fehlt ein WebKit-Prozess der Anwendung,
// misst die Probe einen anderen als zu Beginn. Fremde zählt sie gar nicht
// mit ([aufteilen]).
// eindeutigEigen sagt, ob die WebKit-Prozesse der ersten Probe die der
// Anwendung sein können: **ein** GPU-, **ein** Netzprozess und ein oder zwei
// Webinhalte, nichts sonst (Nachlauf über die Folgen vor 0.8.0, Befund 4).
// Wer daneben entstand, während das Fenster aufging, wäre sonst den ganzen
// Lauf über mitgezählt worden.
export function eindeutigEigen(arten) {
  const zahl = (art) => arten.filter((x) => x === art).length

  return zahl('GPU') === 1 && zahl('Networking') === 1 &&
    zahl('WebContent') >= 1 && zahl('WebContent') <= 2 &&
    arten.every((x) => ['GPU', 'Networking', 'WebContent'].includes(x))
}

export function bewertung(proben, aufwaermenMin = AUFWAERMEN_MIN) {
  const nach = proben.filter((x) => x.minute >= aufwaermenMin)
  const ohneMessung = nach.filter((x) => x.gesamt === null || x.gesamt === undefined).length
  if (ohneMessung > 0) return { gueltig: false, grund: `${String(ohneMessung)} Probe(n) nach der Aufwärmzeit ohne vollständige Messwerte` }

  const erste = proben.find((x) => x.webArten)
  if (erste && !eindeutigEigen(erste.webArten)) return { gueltig: false, grund: `die WebKit-Prozesse der ersten Probe sind nicht eindeutig die der Anwendung: ${erste.webArten.join(', ')}` }

  const kennungen = nach.filter((x) => x.webPids).map((x) => [...x.webPids].sort().join(','))
  const gewechselt = kennungen.filter((k) => k !== kennungen[0]).length
  if (gewechselt > 0) return { gueltig: false, grund: `die WebKit-Prozesse der Anwendung wechselten in ${String(gewechselt)} Probe(n) — ein neu gestarteter ist ein anderer Prozess` }

  // **Eine Probe je Punkt der Runde** (#2527): der Speicher hängt am Teil,
  // in den eine Probe fällt — im Lauf vom 2026-10-20 lag der Median am Tisch
  // 217 MiB über dem beim Bearbeiten, dreimal so viel wie der Anstieg, an dem
  // die Regel brach. Anfang, Ende und Ruhe tragen keinen Teil und zählen
  // hier nicht.
  const teileDerRunde = new Map()
  for (const x of nach) {
    const teil = /Runde \d+: (.+)$/.exec(x.aufgabe ?? '')?.[1]
    if (teil !== undefined) teileDerRunde.set(teil, (teileDerRunde.get(teil) ?? 0) + 1)
  }
  if (teileDerRunde.size > 1) {
    const zaehlung = [...teileDerRunde.entries()].sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0)).map(([teil, n]) => `${teil} ${String(n)}`).join(', ')
    return { gueltig: false, grund: `Proben aus ${String(teileDerRunde.size)} verschiedenen Teilen der Runde (${zaehlung}) — sie vergleichen Zustände, nicht Zeiten` }
  }

  const alle = abschnitte(proben, aufwaermenMin)
  const teile = alle.filter((t) => t.vollstaendig)
  const unvollstaendig = alle.filter((t) => !t.vollstaendig).map((t) => t.nummer)
  const luecken = []
  for (let i = 0; i + 1 < teile.length; i++) {
    for (let n = teile[i].nummer + 1; n < teile[i + 1].nummer; n++) luecken.push(n)
  }
  if (teile.length < 3) return { gueltig: false, grund: `nur ${String(teile.length)} vollständige Abschnitte nach der Aufwärmzeit`, unvollstaendig, luecken }

  const erster = teile[0]
  const letzter = teile.at(-1)
  const wachstum = (letzter.median / erster.median - 1) / ((letzter.mitte - erster.mitte) / 60)
  let stetig = false
  for (let i = 0; i + 2 < teile.length; i++) {
    const [a, b, c] = teile.slice(i, i + 3)
    const benachbart = b.nummer === a.nummer + 1 && c.nummer === b.nummer + 1
    const steigt = (vorher, nachher) => nachher >= vorher * (1 + MINDESTANSTIEG)
    if (benachbart && steigt(a.median, b.median) && steigt(b.median, c.median)) stetig = true
  }

  return { gueltig: true, teile: teile.map((t) => t.median), unvollstaendig, luecken, wachstumJeStunde: wachstum, stetig, haelt: wachstum <= 0.1 && !stetig }
}

// freigegeben sagt, ob WebKit nach Ruhe und Speicherwarnung unter seinem
// Median im ersten ganzen Abschnitt nach der Aufwärmzeit liegt — **was wächst,
// muss freigebbar sein** (Vermerk 0204). Gesendet ist nicht empfangen:
// reagiert WebKit nicht, hält es nicht. Ohne Probe nach der Warnung oder ohne
// ganzen Abschnitt gibt es kein Urteil.
export function freigegeben(proben, aufwaermenMin = AUFWAERMEN_MIN) {
  const danach = proben.find((x) => x.aufgabe === 'nach der Speicherwarnung')
  const ausserhalb = new Set(['Start', 'Ruhe', 'nach der Speicherwarnung', 'Ende'])
  const web = proben
    .filter((x) => !ausserhalb.has(x.aufgabe) && typeof x.web === 'number')
    .map((x) => ({ minute: x.minute, gesamt: x.web }))
  const erster = abschnitte(web, aufwaermenMin).find((t) => t.vollstaendig)

  if (danach === undefined || typeof danach.web !== 'number') return { gueltig: false, grund: 'keine Probe nach der Speicherwarnung' }
  if (erster === undefined) return { gueltig: false, grund: 'kein ganzer Abschnitt nach der Aufwärmzeit' }

  return { gueltig: true, vorher: erster.median, nachher: danach.web, haelt: danach.web < erster.median }
}
