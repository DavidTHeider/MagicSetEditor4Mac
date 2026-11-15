// **Der Tauschabgleich** — Werkstück T4 von Phase 5.6
// (#1469).
//
// # Warum der fremde Stand hereingereicht wird
//
// **Weil er eine Freigabe ist und keine Abfrage.** Was ein Gegenüber zeigt,
// entscheidet seine Freigabe (T3); wie es hierherkommt, entscheidet der Weg —
// heute von Hand, mit `T5` über eine Leitung. Die Gestalt ist dieselbe, die
// `Freigabesicht` hergibt: eine Zeile mit sechs Feldern.
import type { Sichtkarte } from './apisichtfreigabe'
import type { Umschlag } from './umschlag'

// Fremdstand ist, was von einem Gegenüber angekommen ist.
export interface Fremdstand {
  // angebot ist, was es hergeben kann.
  angebot: Sichtkarte[]

  // wuensche ist, was es sucht.
  wuensche: Sichtkarte[]
}

// derNeutraleRang und dieEngeArt sind die Vorgaben aus `domain`.
//
// **Sie stehen hier, damit die Anzeige sie wiedererkennt.** Was die Vorgabe
// ist, sagt an einer Zeile nichts — und eine Zeile, die „normal" und „genau"
// an jede Karte schreibt, versteckt die zwei Fälle, in denen es etwas sagt.
export const derNeutraleRang = 'normal'
export const dieEngeArt = 'genau'

// Tauschkandidat ist ein Treffer samt seiner Erklärung.
export interface Tauschkandidat {
  wunsch: string
  angebot: string

  // angebotDruck ist der Druck des Angebots — aus ihm entsteht eine
  // Handelszeile, wenn jemand den Vorschlag übernimmt (P11, #2379).
  angebotDruck: string

  // angebotZustand, angebotSprache und angebotFinish sind seine Variante —
  // die Zeile, die daraus entsteht, trägt sie (Befund 2 des Zweitblicks über
  // Phase 8).
  angebotZustand: string
  angebotSprache: string
  angebotFinish: string
  menge: number

  // fehlend ist, wie viele dem Wunsch danach noch fehlen.
  fehlend: number

  // rang ist die Dringlichkeit des Wunsches.
  rang: string

  // art sagt, **warum** es ein Treffer ist — `genau` oder `jederdruck`.
  art: string
}

// Tauschseite ist, was eine Seite gibt.
export interface Tauschseite {
  zeilen: Tauschkandidat[]
  stueck: number
  gedeckt: number
  offen: number

  // wert und waehrung gelten nur mit bepreist.
  wert: number
  waehrung: string
  bepreist: boolean
}

// Tauschvorschlag ist der Vorschlag für beide Seiten.
export interface Tauschvorschlag {
  gebe: Tauschseite
  bekomme: Tauschseite
}

// Tauschabgleichantwort ist das Ergebnis samt seinem Fehler.
export interface Tauschabgleichantwort extends Umschlag {
  kennung: string
  gebe: Tauschkandidat[]
  bekomme: Tauschkandidat[]
  vorschlag: Tauschvorschlag
  hatVorschlag: boolean

  // befunde sind die Stellen, an denen eine Kennung gefehlt hat.
  befunde: string[]
}

// Tauschapi ist der eine Weg zum Abgleich.
//
// **Einer und nicht zwei**: Kandidaten und Vorschlag entstehen aus derselben
// Rechnung über dieselben zwei Listen; zwei Rufe dafür ließen das Fenster
// zwei Stände nebeneinanderlegen, von denen einer der ältere ist.
export interface Tauschapi {
  Tauschabgleich: (
    kennung: string,
    stand: Fremdstand,
    behalten: number,
    hoechstzahl: number,
  ) => Promise<Tauschabgleichantwort>
}
