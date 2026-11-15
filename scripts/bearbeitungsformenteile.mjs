// **Die reinen Teile zu den Bearbeitungsformen** (#2393) — gelesen von
// `bearbeitungsformen.mjs`, das mit Netz erhebt, und von
// `bearbeitungsformen.test.mjs`, das ohne Netz prüft. **Getrennt, weil ein
// Import ausführt**: wer den Erzeuger einbände, um eine Funktion zu prüfen,
// lüde jedes Archiv.
import { readdirSync } from 'node:fs'
import { join } from 'node:path'

// repositorium macht aus dem Feld `repository` eines Pakets eine
// GitHub-Adresse und gibt Besitzer und Name — oder wirft, wenn es keine ist.
export function repositorium(paket) {
  const roh = typeof paket.repository === 'string' ? paket.repository : (paket.repository?.url ?? '')
  const glatt = roh
    .replace(/^git\+/, '')
    .replace(/^git@github\.com:/, 'https://github.com/')
    .replace(/^git:\/\//, 'https://')
    .replace(/^github:/, 'https://github.com/')
    .replace(/\.git$/, '')
    .replace(/\/$/, '')
  const voll = /^https?:\/\//.test(glatt) ? glatt : `https://github.com/${glatt}`
  const treffer = /^https:\/\/github\.com\/([^/]+)\/([^/]+)$/.exec(voll)
  if (!treffer) {
    throw new Error(`${paket.name}: kein GitHub-Repositorium (${roh})`)
  }

  return { besitzer: treffer[1], name: treffer[2], adresse: voll }
}

// ERSATZSTAENDE nennt Pakete, deren Bearbeitungsform nicht am Commit ihrer
// Veröffentlichung liegt — **benannt und begründet**, nie geraten. Ein Paket
// ohne erreichbaren Commit und ohne Zeile hier hält den Erzeuger an.
//
// - `tag`: ein anderer, öffentlicher Stand, dessen Code nachweislich derselbe
//   ist;
// - `gomodul`: die Quelle liegt schon über `ERZEUGTE_MODULE` bei — als
//   `modul:teil`, an der Fassung des Pakets.
export const ERSATZSTAENDE = {
  'react-remove-scroll-bar@2.3.8': {
    tag: 'v2.3.7',
    grund:
      '2.3.8 ist aus einem Stand veröffentlicht, der nie ins Repositorium kam (gitHead b3b1287 fehlt, das letzte Tag ist v2.3.7); gegen 2.3.7 ändert das Paket nur package.json — Fassung und den Bereich von react-style-singleton —, der Code ist bytegleich (verglichen am 2026-10-16)',
  },
  'react-style-singleton@2.2.3': {
    tag: 'v2.2.2',
    grund:
      '2.2.3 ist aus einem Stand veröffentlicht, der nie ins Repositorium kam (das letzte Tag ist v2.2.2, der Hauptzweig nennt 2.2.2); gegen 2.2.2 ändert das Paket nur package.json — Fassung und die Bereiche von react und @types/react —, der Code ist bytegleich (verglichen am 2026-10-16)',
  },
  '@wailsio/runtime@3.0.0-beta.30': {
    gomodul: 'github.com/wailsapp/wails/v3:internal/runtime/desktop/@wailsio/runtime',
    grund:
      'der TypeScript-Quelltext liegt in der Go-Fassung v3.0.0-beta.30 unter internal/runtime, die quellen/ schon trägt — das ganze Repositorium wöge 109 MB',
  },
}

// tagkandidaten sind die Tags, unter denen eine Fassung stehen kann — der
// Name mit Fassung zuerst, weil in einem Monorepo `v1.8.0` ein fremder Stand
// sein kann (floating-ui, gemessen am 2026-10-16).
export function tagkandidaten(name, fassung) {
  return [`${name}@${fassung}`, `v${fassung}`, fassung]
}

// nurTypen ist wahr, wenn ein Paketverzeichnis kein JavaScript trägt.
export function nurTypen(verzeichnis) {
  const js = (ort) =>
    readdirSync(ort, { withFileTypes: true }).some((eintrag) =>
      eintrag.isDirectory()
        ? eintrag.name !== 'node_modules' && js(join(ort, eintrag.name))
        : /\.(m|c)?js$/.test(eintrag.name),
    )

  return !js(verzeichnis)
}

// sqliteQuellen liest aus `build/download.sh` von go-sqlite3-wasm, was es
// lädt: die Quelle von sqlite.org und die Dateien am Tag des Repositoriums,
// je mit dem Ziel unter `sqlite/` — in den Unterordnern, in die das Skript
// wechselt (`ext/`, `test/`). **Liest es die erwartete Form nicht, wirft
// es**: eine geänderte Gestalt des Skripts soll auffallen, nicht eine kürzere
// Liste ergeben.
export function sqliteQuellen(skript) {
  const tag = /GITHUB_TAG="https:\/\/github\.com\/sqlite\/sqlite\/raw\/([^"]+)"/.exec(skript)?.[1]
  const autoconf = /curl -#OL "(https:\/\/sqlite\.org\/[^"]+)"/.exec(skript)?.[1]
  if (!tag || !autoconf) {
    throw new Error('download.sh nennt Tag oder Quelle nicht in der gelesenen Form')
  }
  const adressen = [[autoconf, `sqlite/${autoconf.split('/').pop()}`]]
  let ordner = ''
  for (const zeile of skript.split('\n')) {
    const wechsel = /^cd (\w+)\/$/.exec(zeile.trim())
    if (wechsel) ordner = `${wechsel[1]}/`
    if (zeile.trim() === 'cd ~-') ordner = ''
    const datei = /curl -#OL "\$GITHUB_TAG\/([^"]+)"/.exec(zeile)?.[1]
    if (datei) {
      adressen.push([`https://github.com/sqlite/sqlite/raw/${tag}/${datei}`, `sqlite/${ordner}${datei.split('/').pop()}`])
    }
  }

  return { tag, adressen }
}

// listeLesen liest `bearbeitungsformen.tsv` — Kommentare und Leerzeilen
// fallen weg.
export function listeLesen(text) {
  return text
    .split('\n')
    .filter((zeile) => zeile !== '' && !zeile.startsWith('#'))
    .map((zeile) => {
      const [art, name, fassung, herkunft, adresse, sha256, bytes, ziel, vermerk = ''] =
        zeile.split('\t')

      return { art, name, fassung, herkunft, adresse, sha256, bytes: Number(bytes), ziel, vermerk }
    })
}
