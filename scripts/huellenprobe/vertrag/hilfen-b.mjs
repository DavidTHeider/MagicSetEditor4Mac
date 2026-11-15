// **Gemeinsame Hilfen der Gruppen `b-*`** — B01 (#2220).
//
// Keine Gruppe, sondern Werkzeug: hier steht, was mehrere Gruppen brauchen und
// was **keine Produktdatei kennt** — ein fremder Peer, den es nur in diesem
// Skript gibt, ein Einmalkennwort nach RFC 6238, ein winziges Bild. Alles
// davon sind **Testwerte dieser Anwendung**: nichts davon ist ein echtes
// Konto, ein echter Schlüssel oder ein echtes Geheimnis.
//
// **Warum ein fremder Peer von Hand gebaut wird.** Eine Einzelanwendung hat
// kein Gegenüber; für Kontakte, Einladungen, Handel und Leihe braucht sie aber
// eine **gültige** Peerkennung, die nicht ihre eigene ist. Die Gestalt ist
// nachgelesen in `internal/peer/kennung.go` (CIDv1, Multibase `b`, Ed25519 in
// libp2ps Hülle) und für die Einladung in `internal/domain/kontakteinladung.go`
// (Marke, Fassung, Vorsatz, Unterschrift über Vorsatz und Inhalt). Gegen das
// gebaute Programm geprüft mit `mse kontakt-aufnehmen` und `mse
// einladung-annehmen` in einer Wegwerfumgebung (2026-10-09).
import { createHmac, generateKeyPairSync, randomBytes, sign } from 'node:crypto'
import { writeFileSync } from 'node:fs'
import { deflateSync, crc32 } from 'node:zlib'

// b gibt den Platz dieser Gruppen im Kontext — **ein Unterobjekt**, damit
// keine Kennung einer fremden Gruppe überschrieben wird (`k.nutzer`,
// `k.profil` sind die Ausnahme: `vertrag/ersatz.mjs` liest sie).
export const b = (k) => (k.b ??= {})

const alphabet = 'abcdefghijklmnopqrstuvwxyz234567'

// base32 schreibt Bytes klein und ohne Auffüllung (RFC 4648, Multibase `b`).
function base32(puffer) {
  let bits = 0
  let wert = 0
  let aus = ''
  for (const byte of puffer) {
    wert = (wert << 8) | byte
    bits += 8
    while (bits >= 5) {
      aus += alphabet[(wert >>> (bits - 5)) & 31]
      bits -= 5
    }
  }
  if (bits > 0) aus += alphabet[(wert << (5 - bits)) & 31]

  return aus
}

// ausBase32 liest base32 (groß oder klein, ohne Auffüllung) zurück.
function ausBase32(text) {
  let bits = 0
  let wert = 0
  const bytes = []
  for (const zeichen of text.toLowerCase().replaceAll('=', '')) {
    const stelle = alphabet.indexOf(zeichen)
    if (stelle < 0) throw new Error(`kein base32-Zeichen: ${zeichen}`)
    wert = (wert << 5) | stelle
    bits += 5
    if (bits >= 8) {
      bytes.push((wert >>> (bits - 8)) & 255)
      bits -= 8
    }
  }

  return Buffer.from(bytes)
}

// fremder gibt eine **gültige**, nicht eigene Peeridentität unter einem Namen
// — im Kontext gemerkt, damit dieselbe in mehreren Einträgen dieselbe bleibt.
export function fremder(k, name) {
  const alle = (b(k).fremde ??= {})
  if (!alle[name]) {
    const { publicKey, privateKey } = generateKeyPairSync('ed25519')
    const oeffentlich = Buffer.from(publicKey.export({ format: 'jwk' }).x, 'base64url')
    // CIDv1 | libp2p-key | identity-Multihash | 36 | Protobuf (Typ 1, 32 Byte)
    const roh = Buffer.concat([Buffer.from([0x01, 0x72, 0x00, 36, 0x08, 0x01, 0x12, 0x20]), oeffentlich])
    alle[name] = { kennung: `b${base32(roh)}`, privateKey }
  }

  return alle[name]
}

// einladungVon schreibt eine Kontakteinladung, die der fremde Peer ausgestellt
// hätte. **Ohne Adresse**, also nicht einlösbar: angenommen wird sie, gemeldet
// wird nichts — es geht kein Byte ins Netz.
export function einladungVon(fremd, { name = 'Einladender', zweck = 'paarung' } = {}) {
  const ablauf = new Date(Date.now() + 24 * 3600_000).toISOString().replace(/\.\d{3}Z$/, 'Z')
  const inhalt = Buffer.from(JSON.stringify({ kennung: fremd.kennung, name, zweck, ablauf, nonce: randomBytes(16).toString('base64url') }))
  const daten = Buffer.concat([Buffer.from('mse4mac-kontakt/1\n'), inhalt])

  return `mse4mac-kontakt:1:${inhalt.toString('base64url')}.${sign(null, daten, fremd.privateKey).toString('base64url')}`
}

// totp rechnet den sechsstelligen Code zu einem Geheimnis (RFC 6238, SHA-1,
// dreißig Sekunden — `internal/totp/totp.go`). Das Geheimnis ist die
// base32-Form, die `FaktorEinrichten` einmal zurückgibt.
export function totp(geheimnis, jetzt = Date.now()) {
  const zaehler = Buffer.alloc(8)
  zaehler.writeBigUInt64BE(BigInt(Math.floor(jetzt / 1000 / 30)))
  const mac = createHmac('sha1', ausBase32(geheimnis)).update(zaehler).digest()
  const stelle = mac[mac.length - 1] & 15
  const zahl = ((mac[stelle] & 127) << 24) | (mac[stelle + 1] << 16) | (mac[stelle + 2] << 8) | mac[stelle + 3]

  return String(zahl % 1_000_000).padStart(6, '0')
}

// chunk packt einen PNG-Abschnitt.
function chunk(art, daten) {
  const laenge = Buffer.alloc(4)
  laenge.writeUInt32BE(daten.length)
  const rumpf = Buffer.concat([Buffer.from(art, 'latin1'), daten])
  const summe = Buffer.alloc(4)
  summe.writeUInt32BE(crc32(rumpf))

  return Buffer.concat([laenge, rumpf, summe])
}

// schreibePng legt ein kleines, gültiges Bild ab (8 × 8, einfarbig) — für die
// Wege, die eine Bilddatei verlangen (Profilbild, Abzeichen).
export function schreibePng(pfad, rot = 200, gruen = 60, blau = 60) {
  const kante = 8
  const zeilen = []
  for (let y = 0; y < kante; y += 1) {
    zeilen.push(Buffer.from([0, ...Array.from({ length: kante }, () => [rot, gruen, blau]).flat()]))
  }
  const kopf = Buffer.alloc(13)
  kopf.writeUInt32BE(kante, 0)
  kopf.writeUInt32BE(kante, 4)
  kopf.set([8, 2, 0, 0, 0], 8) // 8 Bit, Truecolor, ohne Besonderheiten
  writeFileSync(pfad, Buffer.concat([
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    chunk('IHDR', kopf),
    chunk('IDAT', deflateSync(Buffer.concat(zeilen))),
    chunk('IEND', Buffer.alloc(0)),
  ]))
}

// Die Testkennwörter dieser Anwendung — **Testwerte** (§17), nirgends sonst
// benutzt. Mindestlänge und Höchstlänge: `internal/kennwort`.
export const kennwort = { eins: 'vertrag-kennwort-1', zwei: 'vertrag-kennwort-2', drei: 'vertrag-kennwort-3' }
