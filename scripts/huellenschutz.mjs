// **Was die Schutzrichtlinie während eines Drehbuchs abgewiesen hat** —
// SEC-17 (#2164).
//
// Die Anwendung schreibt jeden Verstoß gegen ihre Schutzrichtlinie als
// `CSP-Verstoß` in ihr Protokoll (`internal/vorschauweg/schutzweiche.go`).
// Nach jedem Drehbuch liest `huellenlauf.mjs` alle Orte, die es angelegt hat:
// ein Verstoß, den es nicht ausdrücklich erwartet, ist eine Ressource, die
// die Richtlinie der Anwendung wegnimmt — und damit ein Befund dieses
// Drehbuchs.
//
// **Jedes Protokoll unter dem Ort, in beiden Formen** — angestrichen vom
// Zweitblick über Phase H: ein Drehbuch darf Ort und Format umstellen
// (`ops04`), und das Protokoll dreht sich (`mse.log.1` …). Und **ein Bericht,
// der sich nicht lesen ließ, ist ein Befund**: was er gemeldet hätte, weiß
// niemand.

import { existsSync, readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'

const protokollname = /^mse\.log(\.\d+)?$/
const meldungen = new Set(['CSP-Verstoß', 'CSP-Bericht unlesbar'])

// protokolle gibt jede Protokolldatei unter `ordner`, auch in Unterordnern.
function protokolle(ordner) {
  if (!existsSync(ordner)) return []

  return readdirSync(ordner, { recursive: true, withFileTypes: true })
    .filter((eintrag) => eintrag.isFile() && protokollname.test(eintrag.name))
    .map((eintrag) => join(eintrag.parentPath, eintrag.name))
}

// gelesen gibt eine Meldung aus Text- oder JSON-Zeile in **einer** Gestalt —
// so trifft ein `erwartet` beide Formate.
function gelesen(zeile) {
  if (zeile.startsWith('{')) {
    let satz
    try {
      satz = JSON.parse(zeile)
    } catch {
      return undefined
    }
    if (!meldungen.has(satz.msg)) return undefined
    const { time: _zeit, level: _stufe, msg, ...rest } = satz

    return [`msg=${msg.includes(' ') ? JSON.stringify(msg) : msg}`, ...Object.entries(rest).map(([schluessel, wert]) => `${schluessel}=${String(wert)}`)].join(' ')
  }

  const treffer = /msg=(CSP-Verstoß|"CSP-Bericht unlesbar")( .*)?$/.exec(zeile)

  return treffer ? treffer[0] : undefined
}

// cspVerstoesse gibt die Verstöße aus den Protokollen der Orte — ohne die,
// auf die `erwartet` passt.
export function cspVerstoesse(orte, erwartet) {
  const gefunden = []

  for (const ort of orte) {
    for (const datei of protokolle(ort.ort)) {
      for (const zeile of readFileSync(datei, 'utf8').split('\n')) {
        const meldung = gelesen(zeile)
        if (meldung === undefined || erwartet?.test(meldung)) continue
        gefunden.push(meldung)
      }
    }
  }

  return gefunden
}
