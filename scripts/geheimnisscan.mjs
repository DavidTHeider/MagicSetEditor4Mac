// **Nach eingeschmuggelten Geheimnissen suchen** — I05
// (#2263).
//
// Ein Lauf legt **Marken** dorthin, wo Geheimnisse liegen — Bund, Adresse,
// Kopfzeile, Profil, Dateipfad — und sucht sie danach im Diagnosepaket und
// im Unified Log. Dieses Modul ist die Suche; der Lauf steht in
// `scripts/huellenprobe/i05geheimnisse.mjs`.
//
// # Warum verkleidet gesucht wird
//
// Ein Geheimnis landet selten so in einem Protokoll, wie es eingegeben wurde:
// in einer Adresse URL-kodiert, in einem Basic-Auth-Kopf als Base64 hinter
// `nutzer:` — und dort steht es je nach Länge des Vorsatzes an einer anderen
// Stelle im Dreierblock. Wer nur roh sucht, findet die gefährlichste Form
// nicht und hält das Paket für sauber.
//
// **Ein reines Modul**: es liest Dateien und Texte, ruft nichts auf und
// schreibt nichts.
import { createHash } from 'node:crypto'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'

// verkleidungen gibt die Formen, in denen eine Marke im Text stehen kann.
//
// **Base64 in drei Ausrichtungen**: steht die Marke `v` Bytes hinter einem
// Dreierblock, kodiert sich ihr Teil ab Byte `(3 − v) mod 3` in vollen
// Blöcken — gesucht wird dieser Kern, gekürzt auf ein Vielfaches von drei.
export function verkleidungen(marke) {
  const formen = [{ art: 'roh', text: marke }]
  const url = encodeURIComponent(marke)

  if (url !== marke) {
    formen.push({ art: 'url', text: url })
  }

  for (const versatz of [0, 1, 2]) {
    const kern = Buffer.from(marke).subarray((3 - versatz) % 3)
    const voll = kern.subarray(0, kern.length - (kern.length % 3))

    if (voll.length >= 3) {
      formen.push({ art: 'base64', text: voll.toString('base64') })
    }
  }

  return formen
}

// fundeIn gibt je Marke und Form den ersten Fund in einem Text.
export function fundeIn(text, marken) {
  const funde = []

  for (const marke of marken) {
    for (const form of verkleidungen(marke)) {
      const stelle = text.indexOf(form.text)

      if (stelle >= 0) {
        funde.push({ marke, art: form.art, stelle })

        // **Eine Form genügt je Marke** — die rohe zuerst, sie ist die
        // deutlichste Auskunft. Base64 einer rohen Fundstelle zählte sonst
        // doppelt.
        break
      }
    }
  }

  return funde
}

// dateienUnter gibt jede gewöhnliche Datei unter einem Verzeichnis, relativ
// zu ihm. **Sockets, Pipes und Geräte nicht**: sie zu lesen bricht ab oder
// wartet — der Ort einer Probe trägt den Socket der Abnahmeleitung.
function dateienUnter(wurzel, verzeichnis = wurzel) {
  return readdirSync(verzeichnis, { withFileTypes: true }).flatMap((eintrag) => {
    const pfad = join(verzeichnis, eintrag.name)

    if (eintrag.isDirectory()) {
      return dateienUnter(wurzel, pfad)
    }

    return eintrag.isFile() ? [relative(wurzel, pfad)] : []
  })
}

// durchsucheVerzeichnis sucht in einem ausgepackten Paket — **getrennt nach
// Ort**: im Dateinamen (Metadaten) und im Inhalt. Das Manifest ist eine Datei
// wie jede andere und steht mit seinem Namen da.
export function durchsucheVerzeichnis(verzeichnis, marken) {
  const funde = []

  for (const datei of dateienUnter(verzeichnis)) {
    for (const fund of fundeIn(datei, marken)) {
      funde.push({ ...fund, ort: 'dateiname', datei })
    }

    for (const fund of fundeIn(readFileSync(join(verzeichnis, datei), 'utf8'), marken)) {
      funde.push({ ...fund, ort: 'inhalt', datei })
    }
  }

  return funde
}

// pruefeManifest hält das Manifest gegen die Dateien: jede genannte liegt da,
// mit Größe und SHA-256 — und keine liegt daneben, die es nicht nennt.
export function pruefeManifest(verzeichnis, manifestname) {
  const manifest = JSON.parse(readFileSync(join(verzeichnis, manifestname), 'utf8'))
  const fehler = []
  const genannt = new Set([manifestname])

  for (const teil of manifest.teile ?? []) {
    genannt.add(teil.name)

    const pfad = join(verzeichnis, teil.name)
    let inhalt

    try {
      inhalt = readFileSync(pfad)
    } catch {
      fehler.push(`${teil.name}: fehlt`)

      continue
    }

    const summe = createHash('sha256').update(inhalt).digest('hex')

    if (summe !== teil.sha256) {
      fehler.push(`${teil.name}: Prüfsumme ${summe} statt ${teil.sha256}`)
    }

    if (statSync(pfad).size !== teil.groesse) {
      fehler.push(`${teil.name}: ${String(statSync(pfad).size)} Bytes statt ${String(teil.groesse)}`)
    }
  }

  for (const datei of dateienUnter(verzeichnis)) {
    if (!genannt.has(datei)) {
      fehler.push(`${datei}: nicht im Manifest`)
    }
  }

  return { gueltig: fehler.length === 0, fehler }
}

// ordneLog trennt die Einträge des Unified Log (`log show --style json`) in
// die der Anwendung und die des Systems — und gibt je Seite die Funde.
//
// **Getrennt, weil sie Verschiedenes sagen**: ein Fund bei der Anwendung ist
// ein Leck dieses Programms; einer bei einem Systemdienst kann eines sein,
// das dieses Programm ausgelöst hat (eine Adresse an `nsurlsessiond`), oder
// einer, der mit ihm nichts zu tun hat.
export function ordneLog(eintraege, prozess, marken) {
  const app = []
  const system = []

  for (const eintrag of eintraege) {
    const text = [eintrag.eventMessage, eintrag.composedMessage, eintrag.subsystem, eintrag.category]
      .filter((teil) => typeof teil === 'string')
      .join(' ')
    const eigen = (eintrag.processImagePath ?? '').endsWith(`/${prozess}`)

    for (const fund of fundeIn(text, marken)) {
      ;(eigen ? app : system).push({ ...fund, prozess: eintrag.processImagePath ?? '' })
    }
  }

  return { app, system }
}
