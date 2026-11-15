// Die eingestellte Sprache — gelesen, geschrieben, gemerkt.
//
// # Warum im Browserspeicher und nicht in der Datenbank
//
// Die Datenbank hält den **Bestand**: Sets, Karten, Werte. Die Sprache der
// Oberfläche gehört zu diesem Fenster, nicht zu den Daten — und ein Bestand,
// der ohne Datenbank nicht anzeigbar ist, wäre eine Anwendung, die bei einem
// Migrationsfehler auch noch stumm ist. Kommt später eine Einstellungstabelle,
// zieht die Sprache mit; bis dahin ist der Browserspeicher die kleinere Zusage.
import { sprachen, type Sprache } from './kataloge'

const schluessel = 'mse.sprache'

// istSprache prüft, was aus dem Speicher oder vom System kommt.
//
// **Geprüft und nicht gecastet**: im Speicher steht, was jemand hineingeschrieben
// hat, und `sprachen` ist die einzige Wahrheit darüber, was wir führen.
export function istSprache(wert: unknown): wert is Sprache {
  return (
    typeof wert === 'string' && (sprachen as readonly string[]).includes(wert)
  )
}

// gespeicherteSprache gibt die eingestellte Sprache.
//
// Die Reihenfolge ist: was jemand gewählt hat, dann die Sprache des Systems,
// dann Deutsch als Ausgangssprache (§9).
export function gespeicherteSprache(
  speicher: Pick<Storage, 'getItem'>,
  systemsprache: string,
): Sprache {
  const gewaehlt = speicher.getItem(schluessel)
  if (istSprache(gewaehlt)) {
    return gewaehlt
  }

  // `de-AT` ist Deutsch. Der Teil vor dem Strich ist die Sprache, der Rest die
  // Gegend — und Gegenden führen wir nicht.
  const kurz = systemsprache.split('-')[0]

  return istSprache(kurz) ? kurz : 'de'
}

// spracheMerken schreibt die Wahl fort.
export function spracheMerken(
  speicher: Pick<Storage, 'setItem'>,
  sprache: Sprache,
): void {
  speicher.setItem(schluessel, sprache)
}

// spracheAus wählt die Sprache: die Wahl, sonst das System, sonst Deutsch.
//
// **`de-AT` ist Deutsch.** Der Teil vor dem Strich ist die Sprache, der Rest
// die Gegend — und Gegenden führen wir nicht.
export function spracheAus(gewaehlt: string, systemsprache: string): Sprache {
  if (istSprache(gewaehlt)) {
    return gewaehlt
  }

  const kurz = systemsprache.split('-')[0]

  return istSprache(kurz) ? kurz : 'de'
}
