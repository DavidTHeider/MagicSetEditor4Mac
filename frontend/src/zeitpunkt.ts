// Ein Zeitpunkt, wie ein Mensch ihn liest.
//
// # Warum das hier steht und nicht in Go
//
// **Weil Go die Zone des Menschen nicht kennt.** `internal/scryfall` und
// `internal/mtgjson` reichen jedes Datum **unverändert** durch, und das ist
// richtig: wer dort eine Zone erfände, verschöbe jedes Datum um bis zu einen
// Tag (AGENTS.md §3). Der Browser kennt sie — und damit ist er die Stelle, an
// der aus `2026-08-23T21:01:56.596+00:00` ein Datum wird.
//
// # Woher der Anlass kommt
//
// **Aus der Sichtprüfung am 2026-08-24**, nicht aus einem Test. Im Kasten der
// Sammeldaten stand die rohe Angabe wörtlich. Die Tests prüften, dass der Wert
// **ankommt** — dass er unlesbar ist, sieht nur ein Auge (§3).

// alsZeitpunkt macht aus einer ISO-Angabe ein lesbares Datum.
//
// **Was sich nicht lesen lässt, kommt unverändert zurück.** „Invalid Date"
// wäre die schlechtere Antwort: dann sähe niemand mehr, was ankam.
export function alsZeitpunkt(
  roh: string,
  sprache: string,
  zeitzone?: string,
): string {
  if (roh === '') {
    return ''
  }

  const wann = new Date(roh)

  if (Number.isNaN(wann.getTime())) {
    return roh
  }

  return zeitpunktformat(sprache, zeitzone).format(wann)
}

// zeitpunktformat ist die Gestalt, in der ein Zeitpunkt erscheint.
//
// **Ausgeführt für den Rundlauf** (LOC-03.AK-01,
// #1689):
// wer prüfen will, dass aus dem Gezeigten dieselbe Minute zurückkommt, muss
// **dieselbe** Gestalt fragen können — eine zweite Fassung im Test prüfte
// sich selbst.
//
// **Die Zone ist eine eigene Einstellung** (LOC-03.AK-03): ohne sie gilt die
// des Rechners, wie bisher.
export function zeitpunktformat(
  sprache: string,
  zeitzone?: string,
): Intl.DateTimeFormat {
  return new Intl.DateTimeFormat(sprache, {
    dateStyle: 'medium',
    timeStyle: 'short',
    ...(zeitzone === undefined ? {} : { timeZone: zeitzone }),
  })
}

// alsKalendertag zeigt einen Tag `JJJJ-MM-TT` in der Darstellung — H04
// (#2258).
//
// **In UTC, weil ein Kalendertag keine Zone hat**: über die Zone des Rechners
// gezeigt, wäre der 29. westlich von Greenwich der 28. Was keiner ist, steht
// da, wie es kam.
export function alsKalendertag(roh: string, sprache: string): string {
  const tag = /^(\d{4})-(\d{2})-(\d{2})$/.exec(roh)

  if (tag === null) {
    return roh
  }

  return new Intl.DateTimeFormat(sprache, {
    dateStyle: 'long',
    timeZone: 'UTC',
  }).format(Date.UTC(Number(tag[1]), Number(tag[2]) - 1, Number(tag[3])))
}
