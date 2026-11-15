// **Die Texte der Zugangsdaten** — herausgelöst mit
// #1534.
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als einer
// ankommt** (§9) — dieselbe Gestalt wie `seitenki.ts`. Mit der Beschriftung je
// Dienst ging `seiteneinstellungen.ts` über die Zeilengrenze aus §2, und der
// Schnitt folgt der Frage: dort steht, **was sich einstellen lässt**, hier,
// **womit ein Dienst sich ausweist**.
//
// **Zwei Wächter lesen diese Datei**: `internal/werkzeuge/zugangsangabenamen_test.go`
// hält die Namen gegen `domain.Dienste()`, und er endet rot, wenn ein Block
// hier fehlt — ein stilles Leerlesen gibt es nicht.
export const zugaenge = {
  // **Der Rückfall je Angabeart** — er gilt, wo `dienstangaben` nichts sagt.
  //
  // **Kein Anbietername im Rückfall** (#1534): `domain.ZugangMarke` bedient
  // **zwei** Dienste, und eine Zeichenkette kann nicht für beide stimmen —
  // „Patreon-Marke“ wäre für CardTrader falsch. Der Name eines Anbieters
  // gehört deshalb nach `dienstangaben`; hier steht nur, was für **jeden**
  // Dienst trägt.
  angaben: {
    marke: 'Zugangsmarke',
    schluessel: 'API-Schlüssel',
    team: 'Team-Kennung',
  },

  // **Und je Dienst, wo die Angabeart nicht genügt** (#1534).
  //
  // **Nachgelesen und nicht erschlossen** (§15): MTGJSONs Marke gibt es nur
  // für Patreon-Unterstützer ([0056](../../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)),
  // CardTraders kommt aus den Profileinstellungen seines eigenen Kontos
  // ([0095](../../../docs/entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md),
  // abgerufen 2026-09-11).
  //
  // **Ein Schlüssel, der keinen wirklichen Dienst nennt, fiele still auf den
  // Rückfall zurück** — also auf genau diesen Befund. `zugangsangabenamen_test.go`
  // hält beide Richtungen.
  dienstangaben: {
    'mtgjson.marke': 'Patreon-Marke',
    'cardtrader.marke': 'Marke aus den Profileinstellungen',
    'anthropic.schluessel': 'API-Schlüssel',
  },

  zugangLiegt: 'hinterlegt',
  zugangFehlt: 'fehlt',
  zugangSetzen: 'Eintragen',
  // **„Löschen“ und nicht „Entfernen“** — UX-10.AK-04: der Knopf ruft
  // `bundOderAnmeldebund().Loesche(…)`, und danach ist die Angabe im
  // Schlüsselbund weg. „Entfernen“ hieß in dieser Anwendung immer: die Sache
  // verlässt eine Ansicht und gibt es weiter.
  zugangLoeschen: 'Löschen',
  // **Gefragt wird vorher** — UX-10.AK-04 (#1682): zurück gibt den Wert nur,
  // wer ihn ausgegeben hat. Dieselben drei Sätze für ein Geheimnis eines
  // Anmeldeanbieters.
  zugangLoeschenFrage: '{angabe} für {name} löschen?',
  zugangLoeschenErklaerung:
    'Der Wert verschwindet aus dem Schlüsselbund dieses Rechners. Zurück bekommst du ihn nur bei dem, der ihn ausgegeben hat.',
  zugangLoeschenJa: 'Löschen',
}
