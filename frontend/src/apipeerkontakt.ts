// **Wen dieses Profil kennt und was es ihm erlaubt** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// **Eine eigene Datei neben `apipeer.ts`**, weil sie eine andere Frage
// beantwortet: dort, wer **dieses** Programm ist — hier, wer die **anderen**
// sind.
import type { FehlerDTO } from './formen'
import type { Umschlag } from './umschlag'

// Peerkontakt ist ein bekanntes Gegenüber.
export interface Peerkontakt {
  // kennung ist der Nachweis — sie **ist** der öffentliche Schlüssel.
  kennung: string

  // fingerabdruck ist die Form, die zwei Menschen vergleichen.
  fingerabdruck: string

  // selbstgewaehlterName ist, wie das Gegenüber sich nennt.
  //
  // **Er belegt nichts.** Das Feld heißt so, damit keine Seite ihn für einen
  // Nachweis hält — das Issue sagt es wörtlich.
  selbstgewaehlterName: string

  // eigenerName ist, wie dieses Profil ihn nennt.
  eigenerName: string

  // stand ist `unbestaetigt`, `bestaetigt` oder `gesperrt`.
  stand: string

  // gesehen ist der Zeitpunkt der letzten Begegnung.
  gesehen: string

  // rechte sind die freigegebenen Zwecke.
  rechte: string[]

  // adresse ist, wo dieses Gegenüber erreichbar ist — leer heißt: nirgends
  // bekannt (N2, #1455).
  adresse: string

  // erstmals ist die **erste** Begegnung — leer heißt: nicht bekannt
  // (T3, #1468). Für einen Kontakt aus der Zeit davor weiß dieses Programm
  // sie nicht, und ein geratenes Datum sähe aus wie ein gemessenes.
  erstmals: string

  // notiz ist, was dieses Profil sich notiert hat — **lokal und nur lokal**.
  notiz: string

  // lage ist einer der Zustände aus NET-01.AK-03: `unbestaetigt`,
  // `schluesselGeaendert`, `bestaetigt` oder `widerrufen` (#1705).
  lage: string

  // vergleichscode steht nur an einem unbestätigten Kontakt — das, was beide
  // Seiten laut vergleichen. Leer heißt: nichts zu vergleichen, oder dieses
  // Profil hat noch keine Identität.
  vergleichscode: string
}

// Einloesemeldung sagt, ob sich dieses Profil nach einer Annahme beim
// Einladenden melden konnte (NET-01b). `grund` ist beim Scheitern das
// Ergebnis der Einlösung — der Schlüssel, unter dem der Katalog seinen Satz
// hat.
export interface Einloesemeldung {
  stand: 'gemeldet' | 'ohne' | 'gescheitert'
  grund?: string
  meldung?: string
}

// Peerkontakteantwort ist die Liste samt den möglichen Zwecken.
export interface Peerkontakteantwort extends Umschlag {
  kontakte: Peerkontakt[]

  // zwecke sind die, die es überhaupt gibt.
  //
  // **Sie kommen aus Go und werden hier nicht nachgebaut.** Eine zweite
  // Aufzählung liefe beim fünften Zweck auseinander.
  zwecke: string[]

  // gemeldet steht nur nach einer Annahme.
  gemeldet?: Einloesemeldung

  // nachladefehler heißt: **die Änderung ist geschrieben**, nur die Liste
  // danach war nicht zu lesen (§7, Regel 3; #1960). Er steht neben `fehler`
  // und nicht an seiner Stelle — `auspacken` lässt die Antwort deshalb
  // durch, und die Seite lädt die Liste selbst nach.
  nachladefehler?: FehlerDTO
}

// Kontaktapi sind die neun Wege zu den bekannten Gegenüber.
//
// **Neun und nicht einer**: aufnehmen ist ein Eintrag, bestätigen ein
// Vergleich, sperren ein Ausschluss, entsperren eine neue Prüfung, erlauben
// und entziehen betreffen einen **einzelnen** Zweck, die Adresse sagt,
// **wo** — nicht, **was** —, und die Notiz sagt gar nichts nach draußen
// (T3, #1468).
export interface Kontaktapi {
  Peerkontakte: () => Promise<Peerkontakteantwort>
  PeerkontaktAufnehmen: (
    kennung: string,
    name: string,
  ) => Promise<Peerkontakteantwort>
  PeerkontaktBestaetigen: (
    kennung: string,
    verglichen: string,
  ) => Promise<Peerkontakteantwort>
  PeerkontaktSperren: (kennung: string) => Promise<Peerkontakteantwort>
  PeerkontaktEntsperren: (kennung: string) => Promise<Peerkontakteantwort>
  PeerzweckErlauben: (
    kennung: string,
    zweck: string,
  ) => Promise<Peerkontakteantwort>
  PeerzweckEntziehen: (
    kennung: string,
    zweck: string,
  ) => Promise<Peerkontakteantwort>

  // PeerkontaktAdresse trägt ein, wo ein Gegenüber erreichbar ist —
  // **eine leere löscht die eingetragene**.
  PeerkontaktAdresse: (
    kennung: string,
    adresse: string,
  ) => Promise<Peerkontakteantwort>

  // PeerkontaktNotiz schreibt, was dieses Profil sich notiert —
  // **eine leere löscht die vorhandene** (T3, #1468).
  PeerkontaktNotiz: (
    kennung: string,
    notiz: string,
  ) => Promise<Peerkontakteantwort>
}

// Horchzustand sagt, ob und wo dieses Profil erreichbar ist.
export interface Horchzustand {
  laeuft: boolean

  // adresse ist die **wirkliche** — wer `:0` angibt, bekommt einen Port vom
  // Betriebssystem, und nur so erfährt ein Mensch, welchen er nennen muss.
  //
  // **Wo gebunden wurde, nicht wo jemand ankommt.** Für „jede Schnittstelle"
  // steht hier `[::]:41234`; wohin ein Gegenüber wählt, sagt `erreichbar`.
  adresse: string

  // erreichbarNamen nennt die geprüften Rechnernamen dieses Rechners.
  //
  // **Sie stehen vor den Adressen, weil sie länger gelten.** Eine IP-Adresse
  // kommt in den meisten Netzen über DHCP, und eine Lease endet; ein Name,
  // den das Netz auflöst, zeigt danach auf die neue Adresse.
  erreichbarNamen: string[]

  // erreichbarAdressen nennt die Adressen, die ein Gegenüber eintippen kann.
  //
  // **Beides leer heißt: von außen erreicht dich niemand** — und nicht
  // „unbekannt".
  erreichbarAdressen: string[]

  // sichtbar heißt: dieses Profil kündigt sich im selben Netz an (N2, #1455).
  //
  // **Ein zweiter Wahrheitswert und kein dritter Zustand**: erreichbar und
  // auffindbar sind zwei Dinge.
  sichtbar: boolean
}

// Nachbar ist ein Gegenüber, das sich im selben Netz gemeldet hat.
export interface Nachbar {
  // kennung ist, was es über sich sagt — **noch nicht geprüft**.
  kennung: string

  // fingerabdruck ist die Form, die zwei Menschen vergleichen.
  fingerabdruck: string

  // name ist der selbstgewählte — **kein Nachweis**.
  name: string

  // adresse ist, wohin gewählt würde.
  adresse: string

  // merkadresse ist, was in die Kontaktadresse gehört — der `.local`-Name,
  // wo es einen gibt, sonst dieselbe Adresse
  // (#1633).
  //
  // **Zwei Felder und nicht eines**: zum Wählen ist die Adresse schneller,
  // zum Merken ist der Name haltbarer — eine IP von heute ist nach dem
  // nächsten DHCP-Wechsel falsch.
  merkadresse: string

  // bekannt heißt: dieses Profil führt ihn schon in seiner Kontaktliste.
  bekannt: boolean

  // stand ist seiner — leer, solange er unbekannt ist.
  stand: string

  // gemerkteAdresse ist, was die Kontaktliste über ihn weiß; adresseGeaendert
  // sagt, ob er woanders steht (NET-02.NAT-02, #1706) — entschieden im
  // Dienst.
  gemerkteAdresse: string
  adresseGeaendert: boolean
}

// Nachbarnantwort ist die Liste samt dem, was nicht zu deuten war.
export interface Nachbarnantwort extends Umschlag {
  nachbarn: Nachbar[]

  // verworfen zählt die Antworten, die keine Peerkennung trugen.
  //
  // **Sie steht da, damit „niemand gefunden" nicht wie „nichts gehört"
  // aussieht.**
  verworfen: number
}

// Horchantwort ist der Stand samt seinem Fehler.
export interface Horchantwort extends Umschlag {
  stand: Horchzustand
}

// Horchapi sind die sechs Wege zum Erreichbar- und Auffindbarsein.
//
// **Kein Schalter**: Anfangen braucht eine Adresse, Aufhören nichts, und der
// Stand ist eine Frage. **Und die Ankündigung ist die zweite Entscheidung** —
// erreichbar sein gilt dem, der die Adresse kennt, auffindbar sein **jedem im
// Netz**.
export interface Horchapi {
  HorchStarten: (adresse: string) => Promise<Horchantwort>
  HorchBeenden: () => Promise<Horchantwort>
  HorchStand: () => Promise<Horchantwort>

  // SichtbarWerden kündigt dieses Profil im selben Netz an — **ein zweiter
  // Schritt**, der jedem im Netz sagt, dass es dieses Profil gibt.
  SichtbarWerden: (name: string) => Promise<Horchantwort>

  // SichtbarkeitBeenden nimmt die Ankündigung zurück — **das Horchen bleibt**.
  SichtbarkeitBeenden: () => Promise<Horchantwort>

  // Nachbarn sagt, wer sich im selben Netz meldet.
  Nachbarn: () => Promise<Nachbarnantwort>

  // FernHorchStarten gibt eine Einladung aus — **der dritte Schritt**, und
  // die weitreichendste Preisgabe: erreichbar für jeden, der sie bekommt.
  FernHorchStarten: () => Promise<Fernantwort>

  // FernHorchBeenden entwertet jede ausgegebene Einladung.
  FernHorchBeenden: () => Promise<Fernantwort>

  // FernHorchStand sagt, ob und unter welcher Einladung gehorcht wird.
  FernHorchStand: () => Promise<Fernantwort>
}

// Fernzustand sagt, ob dieses Profil über fremde Netze erreichbar ist.
export interface Fernzustand {
  laeuft: boolean

  // einladung ist, was ein Gegenüber zum Wählen braucht — leer heißt: nichts
  // auszugeben.
  //
  // **Sie stirbt mit dem Horchen**: der Serverschlüssel wird nicht
  // aufbewahrt, und das nächste Horchen gibt eine neue.
  einladung: string

  // wege sagt je Gegenüber, wie die Verbindung geht — NET-02.AK-03 (#1706).
  // **Leer heißt nicht „niemand"**: tailcat v0.6.0 nennt dort nie eine
  // Gegenstelle (SEC-21, #2162).
  wege: Fernweg[]

  // angebunden zählt die offenen Ströme — gezählt von diesem Programm, nicht
  // von tailcat.
  angebunden: number
}

// Fernweg ist ein Weg zu einem Gegenüber: `direkt` mit seiner Adresse,
// `relais` mit der DERP-Region, `aufbau`, solange es noch keinen gibt.
export interface Fernweg {
  art: 'direkt' | 'relais' | 'aufbau'
  adresse: string
}

// Fernantwort ist der Fernstand samt seinem Fehler.
export interface Fernantwort extends Umschlag {
  stand: Fernzustand
}
