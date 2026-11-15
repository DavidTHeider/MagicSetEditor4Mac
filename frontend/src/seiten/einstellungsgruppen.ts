// **Die zehn Gruppen der Einstellungen** — UX-04.AK-01
// (#1676).
//
// # Was das Kriterium verlangt
//
// *„Ein geordneter Baum oder eine zweistufige Navigation bündelt Allgemein,
// Erscheinungsbild, Nutzer/Profil, Fachfunktionen, KI, Verbindungen,
// Dienste, Daten/Sicherung und Diagnose."*
//
// **Der Entwurf zeichnet sie mit Untertitel** (Artboard `Einstellungen`):
// „Sprache, Start, Kürzel" · „Thema, Schrift, Hintergrund" · „Anmeldung,
// Entsperren" · „Karten, Sammlung, Spiel" · „Anbieter, Modelle, Kosten" ·
// „Netz, Proxy, Gegenüber" · „Kartendienste, Preise, Markt" · „Bestand,
// Pfade, Sicherung" · „Protokolle, Technik, CLI". **Seit AP-10 (#2367) mit
// Abweichung**: CLI, MCP und REST stehen unter *Verbindungen*
// (`docs/PROTOTYP-0.6.0.md`).
//
// # Warum die Zuordnung je Schlüssel steht und nicht je Bereich
//
// **Weil der Entwurf anders schneidet als der Bestand.** `sprache` und
// `erscheinungsbild` liegen beide im Bereich `anwendung` — der Entwurf
// trennt sie. Eine Regel über Bereiche könnte das nicht; eine über Präfixe
// wäre eine Regel, die beim nächsten Schlüssel still danebengreift.
//
// **Eine Aufzählung erzwingt die Entscheidung.** Wer eine Einstellung
// aufnimmt, trägt sie hier ein — sonst wird
// `internal/werkzeuge/einstellungsgruppen_test.go` rot und nennt den
// Schlüssel. Dieselbe Gestalt wie [namensweg] dort.

// einstellungsgruppen sind die zehn — die neun des Entwurfs und die Schalter.
//
// **Von vorn nach hinten wie im Artboard**: was jeder braucht, steht oben;
// was Fachleute suchen, unten.
export const einstellungsgruppen = [
  'allgemein',
  // **Bereiche und Funktionen** — OPS-10 (#1908), eine zehnte Gruppe neben
  // dem Entwurf; die Abweichung steht in `PROTOTYP-0.6.0.md`.
  'bereiche',
  'erscheinungsbild',
  'nutzer',
  'fach',
  'ki',
  'verbindungen',
  'dienste',
  'daten',
  'diagnose',
] as const

export type Einstellungsgruppe = (typeof einstellungsgruppen)[number]

// nachGruppe ordnet jedem Schlüssel seine Gruppe zu.
//
// **Die Begründungen stehen bei den Gruppen und nicht bei jedem Eintrag** —
// sonst stünde dieselbe Aussage einundfünfzig Mal.
const nachGruppe: Readonly<Record<string, Einstellungsgruppe>> = {
  // **Allgemein: Sprache, Start, Kürzel.** Was das Programm als Ganzes
  // betrifft und mit keiner Fachsache zu tun hat.
  sprache: 'allgemein',
  'sprache.eigene': 'allgemein',
  darstellung: 'allgemein',
  zeitzone: 'allgemein',
  'kuerzel.ziffern': 'allgemein',
  aktualisierung: 'allgemein',
  ersteschritte: 'allgemein',

  // **Bereiche und Funktionen** — OPS-10 (#1908). Gezeichnet werden sie vom
  // `Bereichsschalter`, nicht als Wahl je Zeile.
  'schalter.sammeln': 'bereiche',
  'schalter.spielen': 'bereiche',
  'schalter.handeln': 'bereiche',
  'schalter.ressourcen': 'bereiche',
  'schalter.ki': 'bereiche',
  'schalter.netz': 'bereiche',
  'schalter.markt': 'bereiche',
  'schalter.turniere': 'bereiche',
  'schalter.dienst': 'bereiche',
  'schalter.werkzeugserver': 'bereiche',

  // **Erscheinungsbild: Thema, Schrift, Hintergrund.** Und das Fenster
  // dazu — wie groß es aufgeht und wie die Arbeitsplätze stehen, ist
  // Darstellung und keine Fachfrage.
  erscheinungsbild: 'erscheinungsbild',
  schriftgroesse: 'erscheinungsbild',
  schriftart: 'erscheinungsbild',
  farbwelt: 'erscheinungsbild',
  schriftbereich: 'erscheinungsbild',
  hintergrund: 'erscheinungsbild',
  gestaltung: 'erscheinungsbild',
  seitenleiste: 'erscheinungsbild',
  leistenlage: 'erscheinungsbild',
  fensterstart: 'erscheinungsbild',
  fensterbreite: 'erscheinungsbild',
  fensterhoehe: 'erscheinungsbild',
  'arbeitsplatz.navigator': 'erscheinungsbild',
  'arbeitsplatz.inspektor': 'erscheinungsbild',
  'arbeitsplatz.navigatorbreite': 'erscheinungsbild',
  'arbeitsplatz.inspektorbreite': 'erscheinungsbild',
  'arbeitsplatz.ansicht.sets': 'erscheinungsbild',
  'arbeitsplatz.ansicht.decks': 'erscheinungsbild',
  'arbeitsplatz.ansicht.sammlung': 'erscheinungsbild',
  'ausgabe.ziele': 'erscheinungsbild',

  // **Nutzer und Profil: Anmeldung, Entsperren.** Die Kennungen fremder
  // Anbieter sind der Weg hinein und gehören zu dem, wem etwas gehört.
  'anbieter.github.kennung': 'nutzer',
  'anbieter.google.kennung': 'nutzer',
  'anbieter.microsoft.kennung': 'nutzer',
  'anbieter.x.kennung': 'nutzer',

  // **Fachfunktionen: Karten, Sammlung, Spiel.** Womit ein neuer Datensatz
  // anfängt, und wie viel das Kartenformular zeigt.
  'vorgabe.zustand': 'fach',
  'vorgabe.waehrung': 'fach',
  'vorgabe.einzelspiel': 'fach',
  feldtiefe: 'fach',

  // **Gespeicherte Filter** — UX-12.AK-03 (#1684). Angelegt werden sie in
  // der Sicht, nach der sie fragen; hier stehen sie für Fachleute als Liste.
  'filter.gespeichert': 'fach',

  // **Wie der Tisch ordnet, ist eine Fachfunktion** — PLAY-04.AK-03
  // (#1712):
  // sie betrifft das Spiel, nicht das Aussehen der Anwendung.
  'tisch.autoanordnung': 'fach',

  // **Und wie er beschriftet ist, gleich daneben** — eine Frage der
  // Darstellung, aber gesucht wird sie beim Tisch und nicht beim
  // Erscheinungsbild des Fensters.
  'tisch.beschriftung': 'fach',

  // **KI: Anbieter, Modelle, Kosten.** AK-04 verlangt „einen eigenen
  // eindeutig benannten Bereich" — deshalb stehen die drei Sprachmodelle
  // **nicht** bei den Kartendiensten, obwohl ihre Schlüssel `dienst.`
  // heißen. Und die Kostengrenze steht hier, weil der Entwurf „Kosten" in
  // diesen Untertitel schreibt.
  'ki.laufzeit': 'ki',
  'ki.geraetemodell': 'ki',
  'ki.vorrangmodell': 'ki',
  'ki.adresse': 'ki',
  'dienst.anthropic': 'ki',
  'dienst.anthropic.adresse': 'ki',
  'dienst.openai': 'ki',
  'dienst.openai.adresse': 'ki',
  'dienst.xai': 'ki',
  'dienst.xai.adresse': 'ki',
  'richtlinie.kostengrenze': 'ki',

  // **Verbindungen: Netz, Proxy, Gegenüber.**
  'netz.proxy.http': 'verbindungen',
  'netz.proxy.https': 'verbindungen',
  'netz.proxy.aus': 'verbindungen',
  'netz.frist.verbindung': 'verbindungen',
  'netz.frist.kopf': 'verbindungen',
  'netz.frist.stillstand': 'verbindungen',
  'netz.schnittstelle': 'verbindungen',
  'netz.lauschen': 'verbindungen',
  'dienst.tls.zertifikat': 'verbindungen',
  'dienst.tls.schluessel': 'verbindungen',
  'dienst.tls.kette': 'verbindungen',
  'dienst.tls.anker': 'verbindungen',
  // **Und die Kontingente derselben Dienste** — QA-08.AK-03 (#1739).
  'grenzen.kontingente': 'verbindungen',

  // **Dienste: Kartendienste, Preise, Markt.** Die fünf Kartendienste, die
  // Frage, wer gefragt wird, der Blick aufs Marktkonto — und die Schranken
  // darüber: „kein fremder Dienst" und die vier des Handels urteilen über
  // genau diese Wege.
  'dienst.scryfall': 'dienste',
  'dienst.mtgapi': 'dienste',
  'dienst.mtgjson': 'dienste',
  'dienst.scrydex': 'dienste',
  'dienst.cardtrader': 'dienste',
  'frage.suche': 'dienste',
  'konto.schau': 'dienste',
  'richtlinie.nurlokal': 'dienste',
  'richtlinie.handel': 'dienste',
  'richtlinie.handel.markt': 'dienste',
  'richtlinie.handel.menge': 'dienste',
  'richtlinie.handel.wert': 'dienste',
  'richtlinie.handel.frist': 'dienste',

  // **Daten und Sicherung: Bestand, Pfade, Sicherung.** Wohin etwas
  // geschrieben wird — der Hochladepfad und der Weg bei einem
  // Bildnamenskonflikt sind beide Fragen an die Ablage, nicht an die Karte.
  'vorgabe.hochladepfad': 'daten',
  'vorgabe.ausfuhrpfad': 'daten',
  'vorgabe.sicherungspfad': 'daten',
  'vorgabe.diagnosepfad': 'daten',
  'bildausfuhr.ordner': 'daten',

  // **Der Ordner des Plans** (OPS01, #2347) — gezeichnet von der Fläche
  // „Letzte Sicherung" in derselben Gruppe.
  'sicherung.planordner': 'daten',
  bildkonflikt: 'daten',

  // **Das Protokoll unter Diagnose** — OPS-04 (#1721): der Entwurf schreibt
  // „Protokolle, Technik, CLI" in diesen Untertitel; das CLI steht seit AP-10
  // unter *Verbindungen*.
  'protokoll.stufe': 'diagnose',
  'protokoll.datei': 'diagnose',
  'protokoll.system': 'diagnose',
  'protokoll.ausgabe': 'diagnose',
  'protokoll.format': 'diagnose',
  'protokoll.grenze': 'diagnose',
  'protokoll.aufbewahrung': 'diagnose',
  'protokoll.komponenten': 'diagnose',
  'protokoll.ort': 'diagnose',

  // **Und die Messung** — OPS-05 (#1722): Spur und Leistungsprofil stehen
  // bei der Diagnose, wo auch die Zahlen stehen, die sie erklären.
  'messung.spur': 'diagnose',
  'messung.abtastung': 'diagnose',
  'messung.spurgrenze': 'diagnose',
  'messung.aufbewahrung': 'diagnose',
  'messung.profile': 'diagnose',

  // **Und die Grenze der Zwischenspeicher** — OPS-08 (#1725): dort, wo sie
  // zu sehen und zu leeren sind.
  'zwischenspeicher.grenze': 'diagnose',
}

// zugeordneteSchluessel sind alle, die eine Gruppe haben — für den Lauf, der
// eine Gruppe mit **jeder** ihrer Einstellungen zeichnet (UXR-19, #1927).
// Dass es alle sind, die es gibt, hält der Wächter in Go.
export const zugeordneteSchluessel: readonly string[] = Object.keys(nachGruppe)

// gruppeFuer sagt, in welche Gruppe ein Schlüssel gehört.
//
// **Ein unbekannter fällt nicht heraus** (§8): ein Bestand von morgen kann
// einen tragen, den dieses Fenster nicht kennt — und ihn stillschweigend zu
// verstecken wäre schlimmer, als ihn unter „Allgemein" zu zeigen. Dass keiner
// unbekannt **bleibt**, hält der Wächter.
export function gruppeFuer(schluessel: string): Einstellungsgruppe {
  return nachGruppe[schluessel] ?? 'allgemein'
}

// dienstlisten teilt, was `dienst.` heißt, in die Sprachmodelle und die
// Kartendienste — beide heißen so, und nur die Gruppe sagt, welche wohin
// gehört (UX-04.AK-04).
//
// **Der Schnitt ist erschöpfend und das ist Absicht**: was nicht zur KI
// gehört, ist ein Kartendienst — auch ein `dienst.neuer`, den Go kennt und
// dieses Fenster noch nicht. Wer stattdessen auf `=== 'dienste'` prüfte,
// ließe ihn aus **beiden** Listen fallen und zeigte ihn nirgends.
//
// **Außer dem Zeugnis** — UXR-19
// (#1927):
// `dienst.tls.*` gehört zu den Verbindungen, in den Kasten für Fachleute,
// und stand bis dahin zwischen Scryfall und CardTrader. Es fällt hier aus
// beiden Listen und wird dort als gewöhnliche Einstellung gezeichnet.
export function dienstlisten<T extends { schluessel: string }>(
  dienste: readonly T[],
): { kimodelle: T[]; kartendienste: T[] } {
  return {
    kimodelle: dienste.filter(
      (eintrag) => gruppeFuer(eintrag.schluessel) === 'ki',
    ),
    kartendienste: dienste.filter(
      (eintrag) =>
        !['ki', 'verbindungen'].includes(gruppeFuer(eintrag.schluessel)),
    ),
  }
}

// **Was nur Fachleute brauchen** — UX-04.AK-04
// (#1676):
// „Technische Expertenoptionen sind einklappbar." Die drei Fristen sind es:
// wer sie verstellt, weiß, was ein Kopf und ein Stillstand sind. **Der
// Proxy nicht** — ihn suchte man (Befund F14 der nativen Prüfung), und er
// entscheidet, ob diese Anwendung überhaupt hinauskommt.
export const fachschluessel: ReadonlySet<string> = new Set([
  'netz.frist.verbindung',
  'netz.frist.kopf',
  'netz.frist.stillstand',

  // **Die Schnittstelle der Suche** (NET-02.AK-03, #1706): wer sie wählt,
  // weiß, was `en0` ist; alle anderen fahren mit der Automatik.
  'netz.schnittstelle',

  // **Wo die eigenen Dienste lauschen** (G08, #2272): gewählt wird unter
  // *Werkzeuge ohne Fenster* aus der Liste dieses Rechners; als Text braucht sie,
  // wer Adressen von Hand nennt.
  'netz.lauschen',

  // **Das eigene Zeugnis der Dienste** (NET-03, #1707): wer eine eigene
  // Stelle betreibt, weiß, was ein Anker ist; ohne sie gilt die Identität.
  'dienst.tls.zertifikat',
  'dienst.tls.schluessel',
  'dienst.tls.kette',
  'dienst.tls.anker',

  // **Die Kontingente der Dienste** (QA-08.AK-03, #1739): abgeschaltet
  // werden sie für eine kontrollierte Umgebung, und wer eine hat, sucht sie
  // bei den Fachleuten.
  'grenzen.kontingente',

  // **Die gespeicherten Filter als Text** (UX-12.AK-03, #1684): angelegt
  // werden sie im Navigator ihrer Sicht; wer die Liste von Hand ändern
  // will, findet sie hier.
  'filter.gespeichert',

  // **Und die Ziele der Ausgabe je Set** (AP-10, #2367) — aus demselben
  // Grund: gesetzt werden sie unter „Wohin?“, hier steht die Liste als Text.
  'ausgabe.ziele',
])
