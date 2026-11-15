// **Die Rechnung des Leerlaufs** — PERF01
// (#2351),
// Befund A-26 des Audits 0.9.0.
//
// **CPU im Mittel je Prozess** heißt hier: verbrauchte CPU-Zeit zwischen zwei
// Abfragen von `ps`, geteilt durch die Wanduhrzeit dazwischen, in Prozent
// eines Kerns. Je Art — Anwendung, Webinhalt, GPU, Netz — summiert, denn
// WebKit startet einen Webinhalt neu, und dann zählt er unter neuer Kennung.
//
// # Was es nicht misst
//
// **Weckungen und Energie** — dafür braucht `powermetrics` Verwalterrechte.
// Ein Prozess, der zwischen den Abfragen verschwand, fehlt in der Rechnung,
// statt mit null zu zählen.

// cpuSekunden liest die Spalte `time` von `ps`: `[[TT-]HH:]MM:SS.ss`.
export function cpuSekunden(zeit) {
  const [tage, rest] = zeit.includes('-') ? zeit.split('-') : ['0', zeit]
  const teile = rest.split(':').map(Number)
  const sekunden = teile.reduce((summe, teil) => summe * 60 + teil, 0)

  return Number(tage) * 86400 + sekunden
}

// zeitenAus liest `ps -o pid=,time=` je Kennung.
export function zeitenAus(ausgabe) {
  const zeiten = new Map()
  for (const zeile of ausgabe.split('\n')) {
    const [pid, zeit] = zeile.trim().split(/\s+/)
    if (pid && zeit) zeiten.set(pid, cpuSekunden(zeit))
  }

  return zeiten
}

// anteile gibt je Art den Anteil eines Kerns in Prozent, über `sekunden`
// Wanduhrzeit — auf vier Stellen.
export function anteile(arten, vorher, nachher, sekunden) {
  const summe = {}
  for (const [pid, art] of arten) {
    if (!vorher.has(pid) || !nachher.has(pid)) continue
    summe[art] = (summe[art] ?? 0) + (nachher.get(pid) - vorher.get(pid))
  }

  return Object.fromEntries(Object.entries(summe).map(([art, cpu]) => [art, Math.round((cpu / sekunden) * 100 * 10000) / 10000]))
}
