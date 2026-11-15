// **Der Kern des Mutationslaufs** — Katalog, Wahl, Urteil und das Mutieren
// selbst, herausgelöst aus `mutationslauf.mjs`, damit sie sich ohne `go test`
// prüfen lassen — T05 (#2108).
import { readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

// Jede Mutation: eine Invariante, ihre Fundstelle, die Verfälschung, der
// Lauf, der sie fangen soll — und **der Test, der sie fängt**. Bis zum
// 2026-10-06 galt jeder rote Lauf als gefangen; `erwartet` stammt aus dem
// Beleg vom 2026-10-05 (12 von 12, je mit seinem Test). `lauf` engt nur ein,
// wo ein Paket lange läuft — und steht dann im Beleg. **Eng heißt riskant**:
// der erste Lauf filterte `internal/service` auf `Partie|Empfang|…`, und M8
// „überlebte" — weil der Test, der sie fängt, anders heißt. Seitdem fährt
// `service` ganz; nur `app` bleibt eingeengt.
export const mutationen = [
  { kennung: 'M1', invariante: 'eine Rolle reicht nur so weit, wie sie reicht', datei: 'internal/domain/akteur.go', alt: 'if !rolle.Reicht(wirkung) {', neu: 'if false && !rolle.Reicht(wirkung) {', paket: './internal/domain', erwartet: 'TestDieAbsageNenntDieAchse' },
  { kennung: 'M2', invariante: 'ein nur lesendes Profil schreibt nicht', datei: 'internal/domain/akteur.go', alt: 'if a.NurLesend && wirkung != WirkungLesen {', neu: 'if false && a.NurLesend && wirkung != WirkungLesen {', paket: './internal/domain', erwartet: 'TestNurLesenSchlaegtJedeRolle' },
  { kennung: 'M3', invariante: 'eine Nonce wirkt einmal', datei: 'internal/peer/nonceliste.go', alt: 'if _, schon := l.gesehen[nonce]; schon {', neu: 'if _, schon := l.gesehen[nonce]; schon && false {', paket: './internal/peer', erwartet: 'TestDasGedaechtnisVergisstNachSeinerFrist' },
  { kennung: 'M4', invariante: 'eine behauptete Rahmenlänge wird geprüft, bevor sie zugeteilt wird', datei: 'internal/leitung/leitung.go', alt: 'if laenge == 0 || laenge > peer.Hoechstlaenge {', neu: 'if laenge == 0 {', paket: './internal/leitung', erwartet: 'TestEinBehaupteterRahmenStelltKeinenSpeicher' },
  { kennung: 'M5', invariante: 'ein Eintrag einer Sicherung führt nicht aus ihr hinaus — auch nicht über einen Umweg', datei: 'internal/sicherung/auspacken.go', alt: 'if sauber == ".." || strings.HasPrefix(sauber, "../") {', neu: 'if false && (sauber == ".." || strings.HasPrefix(sauber, "../")) {', paket: './internal/sicherung', erwartet: 'TestEinEintragDerHinausfuehrtWirdAbgewiesen' },
  { kennung: 'M6', invariante: 'ein falsches Kennwort wird abgewiesen', datei: 'internal/kennwort/kennwort.go', alt: 'if subtle.ConstantTimeCompare(gerechnet, hash) != 1 {', neu: 'if subtle.ConstantTimeCompare(gerechnet, hash) != 1 && false {', paket: './internal/kennwort', erwartet: 'TestEinFalschesKennwortIstKeinFormfehler' },
  { kennung: 'M7', invariante: 'eine angewendete Migration wird nicht nachträglich verändert', datei: 'internal/storage/migrationsbuchfuehrung.go', alt: 'if jetzt := pruefsumme(inhalt); jetzt != verbucht {', neu: 'if jetzt := pruefsumme(inhalt); jetzt != verbucht && false {', paket: './internal/storage', erwartet: 'TestEineNachtraeglichGeaenderteMigrationWirdAbgewiesen' },
  { kennung: 'M8', invariante: 'eine doppelt zugestellte Partienachricht wirkt einmal', datei: 'internal/service/partieempfang.go', alt: 'if urteil != domain.FolgeNeu {', neu: 'if urteil != domain.FolgeNeu && false {', paket: './internal/service', erwartet: 'TestEineAlteEpocheWirdGemeldetUndFaelltWeg' },
  { kennung: 'M9', invariante: 'ein Handel wird nicht halb verbucht', datei: 'internal/service/handelsverbuchung.go', alt: 'if offen > 0 {', neu: 'if offen > 0 && false {', paket: './internal/service', erwartet: 'TestEineZuKleineZusageBuchtNicht' },
  { kennung: 'M10', invariante: 'eine Sitzung staut höchstens so viel, wie ihre Tiefe erlaubt', datei: 'internal/service/sitzungsgriff.go', alt: 'if len(g.warten) < g.eltern.tiefe {', neu: 'if true {', paket: './internal/service', erwartet: 'TestEineVolleWarteschlangeStautZurueckUndWirftWeg' },
  { kennung: 'M11', invariante: 'eine Einlösung einer neueren Fassung wird als solche erkannt', datei: 'internal/domain/einloesung.go', alt: 'case fassung > Einloesefassung:', neu: 'case fassung > Einloesefassung+1000:', paket: './internal/domain', erwartet: 'TestEineNeuereEinloesefassungWirdBenannt' },
  { kennung: 'M12', invariante: 'eine beendete Partiesitzung wird freigegeben', datei: 'internal/app/partieaufnahme.go', alt: '\tschliessePartiesitzung(ctx, griff)\n}', neu: '\t_ = griff\n}', paket: './internal/app', lauf: 'Partie|Tisch|Beitritt|Aufstehen|Abriss', erwartet: 'TestNachDemAbrissGibtDerGastDieSitzungFrei' },
]

// waehle gibt die genannten Mutationen — alle, wenn keine genannt ist — und
// wirft bei einer unbekannten Kennung, statt sie still zu übergehen.
export function waehle(katalog, genannt) {
  if (genannt.length === 0) return katalog
  const fremd = genannt.filter((k) => !katalog.some((m) => m.kennung === k))
  if (fremd.length > 0) throw new Error(`unbekannte Mutation: ${fremd.join(', ')}`)

  return katalog.filter((m) => genannt.includes(m.kennung))
}

// rote gibt die Namen der roten Tests einer Ausgabe von `go test`.
const rote = (ausgabe) => [...ausgabe.matchAll(/--- FAIL: (\S+)/g)].map((t) => t[1])

// urteil entscheidet über eine Mutation aus dem Grundlauf ohne und dem Lauf
// mit ihr. Ein Lauf ist `{ status, signal, fehler, ausgabe }` wie bei
// `spawnSync`.
//
// **Gefangen ist nur, was der erwartete Test fängt** — und nur bei grünem
// Grundlauf. Eine Frist, ein Signal, ein roter Grundlauf entscheiden nichts;
// ein anderer roter Test heißt `anders rot` und zählt nicht.
export function urteil(m, grund, mutiert) {
  if (grund.status !== 0) {
    return { art: 'unentschieden', grund: `der Grundlauf ist rot: ${rote(grund.ausgabe).join(', ') || `Status ${String(grund.status ?? grund.signal)}`}` }
  }
  if (mutiert.fehler === 'ETIMEDOUT') return { art: 'unentschieden', grund: 'Frist' }
  if (mutiert.status === null) return { art: 'unentschieden', grund: `Signal ${String(mutiert.signal)}` }
  if (/\[build failed\]|cannot use|undefined:|declared and not used/.test(mutiert.ausgabe)) return { art: 'nicht übersetzt' }
  if (mutiert.status === 0) return { art: 'überlebt' }
  const gefallen = rote(mutiert.ausgabe)
  if (gefallen.includes(m.erwartet)) return { art: 'gefangen', von: m.erwartet }

  return { art: 'anders rot', von: gefallen.join(', ') }
}

// offen hält, was gerade verändert ist — damit auch ein Signal es zurücksetzt.
let offen = null

// zuruecksetzen stellt die gerade veränderte Datei wieder her.
export function zuruecksetzen() {
  if (offen) {
    writeFileSync(offen.pfad, offen.inhalt)
    offen = null
  }
}

// mutiere setzt die Mutation, fährt `ausfuehren` und setzt **immer** zurück —
// auch wenn der Lauf wirft. **Die Fundstelle muss genau einmal passen**,
// sonst fährt nichts: eine Mutation, die nicht stattfand, bewiese nichts.
export async function mutiere(m, wurzel, ausfuehren) {
  const pfad = join(wurzel, m.datei)
  const inhalt = readFileSync(pfad, 'utf8')
  const treffer = inhalt.split(m.alt).length - 1
  if (treffer !== 1) throw new Error(`${m.kennung}: die Fundstelle passt ${String(treffer)}-mal in ${m.datei} — erwartet genau einmal`)

  offen = { pfad, inhalt }
  try {
    writeFileSync(pfad, inhalt.replace(m.alt, m.neu))

    return await ausfuehren(m)
  } finally {
    zuruecksetzen()
  }
}
