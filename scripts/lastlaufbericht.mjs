// **Was der Lastlauf aus seinen Proben macht** — QA-08.AK-04
// (#1739).
//
// Eine eigene Datei aus demselben Grund wie `lastlaufnetz.mjs`. Hier steht
// die **Rechnung** — Durchsatz, Perzentile, Fehlerrate, Anteil der
// Drosselung — und die **Gestalt** des Belegs.
//
// # Zwei Uhren, und beide stehen im Beleg
//
// **Die Uhr des Klienten** misst jeden Ruf genau, samt Leitung und
// Warteschlange vor dem Annehmen. **Die Uhr des Dienstes** steht in
// `GET /metriken` und `lesen.metriken` (OPS-05): sie misst ab dem Behandler
// und kennt nur Eimer — 1, 5, 25, 100, 500, 2500 ms und darüber. Ein p95 von
// dort ist deshalb eine **Obergrenze**: „höchstens 25 ms" und nicht „19 ms".
// Das Issue verlangt, p95/p99 dort abzulesen; der Bericht nennt beide und
// sagt, welche welche ist.
//
// # Was als Fehler zählt
//
// **Eine Drosselung nicht.** 429 über REST und „zu viele Rufe" über MCP sind
// die erwartete Schutzreaktion (AK-04) und stehen in einer eigenen Spalte.
// Ein Fehler ist, was weder gelang noch gedrosselt wurde: eine abgebrochene
// Verbindung, ein 5xx, ein anderer 4xx, ein `isError` aus einem anderen Grund.

// perzentil rechnet über sortierte Werte — nächster Rang, wie üblich.
export function perzentil(sortiert, anteil) {
  if (sortiert.length === 0) return null

  return sortiert[Math.min(sortiert.length - 1, Math.ceil(anteil * sortiert.length) - 1)]
}

// istGedrosselt sagt, ob eine Probe eine Schutzreaktion war.
export function istGedrosselt(probe) {
  return probe.status === 429 || probe.gedrosselt === true
}

// istGelungen sagt, ob eine Probe getan hat, was sie sollte: der Status stimmt,
// MCP meldet keinen Fehler, und — wo die Wirkung geprüft wurde — sie ist da.
// **Ein 2xx allein ist nicht die Wirkung** (`lastlaufwirkung.mjs`).
//
// **Ein verdeckter Schreibruf ist nicht gelungen** (N3 des Nachlaufs über die
// QA-Arbeit): sein Wert ist überschrieben, und gewirkt zu haben, belegt
// nichts — er steht in `unbelegt` und nicht unter den Fehlern.
export function istGelungen(probe) {
  return probe.status >= 200 && probe.status < 300 && probe.mcpFehler !== true &&
    probe.wirkungFehlt === undefined && probe.wirkungVerdeckt === undefined
}

// auswerten fasst die Proben einer Gruppe zusammen.
export function auswerten(proben, sekunden) {
  const zeiten = proben.map((p) => p.ms).sort((a, b) => a - b)
  const gelungen = proben.filter(istGelungen)
  const gedrosselt = proben.filter(istGedrosselt)
  const unbelegt = proben.filter((p) => p.wirkungVerdeckt !== undefined && p.wirkungFehlt === undefined)
  const fehler = proben.filter((p) => !istGelungen(p) && !istGedrosselt(p) && !unbelegt.includes(p))
  const gelungenZeiten = gelungen.map((p) => p.ms).sort((a, b) => a - b)
  const wartezeiten = gedrosselt.map((p) => p.warten).filter((w) => w > 0)

  return {
    rufe: proben.length,
    jeSekunde: runde(proben.length / sekunden),
    gelungenJeSekunde: runde(gelungen.length / sekunden),
    gelungen: gelungen.length,
    unbelegt: unbelegt.length,
    gedrosselt: gedrosselt.length,
    fehler: fehler.length,
    fehlerrate: proben.length === 0 ? 0 : runde(fehler.length / proben.length, 4),
    anteilGedrosselt: proben.length === 0 ? 0 : runde(gedrosselt.length / proben.length, 4),
    msP50: runde(perzentil(zeiten, 0.5)), msP95: runde(perzentil(zeiten, 0.95)), msP99: runde(perzentil(zeiten, 0.99)),
    gelungenP50: runde(perzentil(gelungenZeiten, 0.5)), gelungenP95: runde(perzentil(gelungenZeiten, 0.95)),
    gelungenP99: runde(perzentil(gelungenZeiten, 0.99)),
    // **Über `reduce` und nicht über `Math.min(...)`**: eine Drosselung ist
    // billig, und ein Fall mit Hunderttausenden Proben sprengte den Stapel.
    wartenMin: wartezeiten.length ? wartezeiten.reduce((a, b) => Math.min(a, b)) : null,
    wartenMax: wartezeiten.length ? wartezeiten.reduce((a, b) => Math.max(a, b)) : null,
    beispielFehler: fehler.slice(0, 3).map((p) => `${p.status} ${(p.wirkungFehlt ?? p.fehler ?? p.inhalt ?? p.text ?? '').slice(0, 160)}`),
  }
}

// serverseitig liest die Verteilung einer Fläche aus einer Antwort von
// `GET /metriken` oder `lesen.metriken` — Summe über beide Arten.
export function serverseitig(text, flaeche) {
  let stand

  try {
    stand = JSON.parse(text)
  } catch {
    return null
  }

  const reihen = (stand.laufzeiten ?? []).filter((r) => r.flaeche === flaeche)
  if (reihen.length === 0) return null

  const eimer = stand.eimerMs ?? []
  const summe = new Array(eimer.length + 1).fill(0)
  let anzahl = 0
  let abgewiesen = 0
  let gescheitert = 0

  for (const reihe of reihen) {
    anzahl += reihe.anzahl
    abgewiesen += reihe.abgewiesen
    gescheitert += reihe.gescheitert
    reihe.eimer.forEach((wert, stelle) => { summe[stelle] += wert })
  }

  const obergrenze = (anteil) => {
    let bisher = 0
    for (let stelle = 0; stelle < summe.length; stelle += 1) {
      bisher += summe[stelle]
      if (bisher >= Math.ceil(anteil * anzahl)) return stelle < eimer.length ? `≤ ${eimer[stelle]} ms` : `> ${eimer.at(-1)} ms`
    }

    return null
  }

  return {
    anzahl, abgewiesen, gescheitert, p95: obergrenze(0.95), p99: obergrenze(0.99),
    heapMiB: runde((stand.prozess?.heapBytes ?? 0) / 2 ** 20), goroutinen: stand.prozess?.goroutinen ?? null,
  }
}

function runde(wert, stellen = 1) {
  if (wert === null || wert === undefined || Number.isNaN(wert)) return null

  return Math.round(wert * 10 ** stellen) / 10 ** stellen
}

// markdown schreibt den Beleg als Tafel je Variante.
export function markdown(bericht) {
  const zeilen = [
    '# Lastlauf',
    '',
    `**${bericht.faelle.length} Fälle** — Commit \`${bericht.commit}\`${bericht.veraendert ? ' **mit Änderungen im Baum**' : ''}, ` +
      `${bericht.zeit}, ${bericht.maschine}. Last der Maschine (1/5/15 min) vorher ${bericht.last.vorher.join(' / ')}, ` +
      `nachher ${bericht.last.nachher.join(' / ')}.`,
    `Gefahren mit \`scripts/lastlauf.mjs\`: je Fall **${bericht.eingaben.dauer} s**, **${bericht.eingaben.klienten}** Klienten, ` +
      'gegen einen **frisch gestarteten** Dienst mit eigenem Bestand, eigenem Zuhause und eigenem Schlüsselbund. ' +
      'Was der Lauf misst und was nicht, steht in seinem Kopf und in `docs/LASTMESSUNG.md`.',
    '',
    '**Zeit** ist die des Klienten (p50/p95/p99 in ms über alle Rufe, gelungen oder nicht); **Dienst** ist die Obergrenze ' +
      'aus den Eimern von `/metriken` bzw. `lesen.metriken`. **Gedrosselt** ist die erwartete Schutzreaktion und kein Fehler. ' +
      '**Gelungen** heißt Status 2xx **und**, wo ein Profil eine Wirkung zusagt, dass eine Stichprobe sie belegt (Bild am Ziel, Feld im ' +
      'Bestand, Set nach der Einfuhr — `scripts/lastlaufwirkung.mjs`); ein Ruf ohne Wirkung zählt als Fehler.',
    '',
    // **Erst die Rohpfade** — QA-LAST01 (#2092): ohne sie misst die Last nur,
    // wie schnell der Dienst ablehnt.
    ...(bericht.rohpfade ? [`**Vorher die Rohpfade**, je einmal ohne Last und mit belegter Wirkung: ${bericht.rohpfade.join(', ')}. ` +
      'Eine Gruppe, die Wirkung zusagt und keinen Ruf durchbringt, steht als **ungeprüft** da; mit Kontingenten ist eine nur ' +
      'gedrosselte Gruppe kein Befund, ohne sie macht sie den Lauf rot.', ''] : []),
  ]

  for (const variante of bericht.varianten) {
    zeilen.push(`## Kontingente ${variante === 'mit' ? 'an' : 'aus'}`, '')
    zeilen.push('| Fall | Gruppe | Rufe | je s | gelungen je s | gedrosselt | Fehler | p50 | p95 | p99 | Retry-After | Dienst p95/p99 | RSS max MiB |')
    zeilen.push('| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- | ---: |')

    for (const fall of bericht.faelle.filter((f) => f.variante === variante)) {
      for (const [gruppe, w] of Object.entries(fall.gruppen)) {
        const warten = w.wartenMin === null ? '—' : `${w.wartenMin}–${w.wartenMax} s`
        const dienst = fall.dienst?.[gruppe] ? `${fall.dienst[gruppe].p95} / ${fall.dienst[gruppe].p99}` : '—'
        zeilen.push(`| \`${fall.name}\` | ${gruppe} | ${w.rufe} | ${w.jeSekunde} | ${w.gelungenJeSekunde} | ` +
          `${prozent(w.anteilGedrosselt)} | ${prozent(w.fehlerrate)} | ${w.msP50} | ${w.msP95} | ${w.msP99} | ${warten} | ${dienst} | ` +
          `${fall.speicherMiB ?? '—'} |`)
      }
    }

    zeilen.push('')

    for (const fall of bericht.faelle.filter((f) => f.variante === variante && f.befund)) {
      zeilen.push(`- **\`${fall.name}\`**: ${fall.befund}`)
    }

    zeilen.push('')
  }

  return `${zeilen.join('\n').trimEnd()}\n`
}

function prozent(anteil) {
  return `${Math.round(anteil * 1000) / 10} %`
}
