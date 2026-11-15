// **Die Texte der Recherche** — Werkstück A16 von Phase 5.3
// (#1515).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als einer
// ankommt** (§9).
export const recherche = {
  rechercheTitel: 'Im Wissensbestand nachschlagen',
  rechercheErklaerung:
    'Fragt Kartentexte, Rulings und das Regelwerk auf diesem Rechner — Volltext und Vektoren zusammengeführt. Für die Vektoren rechnet das Einbettungsmodell mit, lokal; ein Modell, das Texte schreibt, wird nicht gefragt. Nichts verlässt diesen Rechner, und nichts wird geändert.',

  rechercheFall: 'Wonach',
  rechercheFallKarten: 'Karten',
  rechercheFallRegeln: 'Regeln',
  rechercheFallRulings: 'Rulings',

  rechercheFrage: 'Frage',
  rechercheSuchen: 'Nachschlagen',
  rechercheSprache: 'Sprache',
  rechercheAlleSprachen: 'Jede',

  rechercheModus: 'Gesucht',
  rechercheModusHybrid: 'Volltext und Vektoren zusammengeführt',
  rechercheModusLexikalisch: 'nur Volltext — kein Einbettungsmodell',
  rechercheModusOhneIndex: 'gar nicht — es steht kein Suchindex da',

  // **Sechs Gründe, und fünf davon sagen, was zu tun ist.** „Nichts
  // gefunden" ist der einzige, der heißt, dass es die Antwort nicht gibt.
  //
  // **Hier steht der Rat und nicht die Lage.** Den Fall nennt der Dienst
  // („Es ist kein Suchindex aktiv.“), und er steht davor; beim ersten Anlauf
  // stand er in beiden Sätzen, und am Fenster stand zweimal dasselbe.
  rechercheLeereNichtsGefunden: 'Ein anderes Wort probieren?',
  rechercheLeereOhneIndex:
    'Unter „Einstellungen › Diagnose“ lässt sich einer bauen.',
  rechercheLeereOhneModell:
    'Ohne Einbettungsmodell findet die Suche Wörter, aber keine Ähnlichkeit.',
  rechercheLeereOhneQuelle:
    'Diese Frage braucht eine Quelle, die im Index nicht liegt.',
  rechercheLeereGesperrt: 'Der Zugriff auf diese Quelle ist nicht freigegeben.',
  rechercheLeereBudget: 'Die Grenze dieses Laufs war erreicht.',

  rechercheHerkunft: 'Stand {revision}, gültig {datum}, geholt {abgerufen}',
  rechercheWeitere: '{anzahl} weitere Abschnitte derselben Karte',
  rechercheOeffnen: 'Karte öffnen',
  rechercheEinschraenkung: 'Nicht getan',
  rechercheStaende: 'Stände',
  rechercheIndex: 'Suchindex',
  rechercheMehr: 'Mehr laden',

  // **Die Ähnlichkeit** — aus dem Karteneditor, mit dem Stand der Karte.
  aehnlichTitel: 'Ähnliche Karten',
  aehnlichErklaerung:
    'Sucht mit dem Text dieser Karte, nicht mit ihrem Namen. Die Karte selbst steht nicht in der Liste.',
  aehnlichSuchen: 'Ähnliche finden',
  aehnlichLeer: 'Noch nichts gesucht.',

  // **Die Freigabe** — was an ein Modell ginge, bevor es hinausgeht.
  freigabeTitel: 'Was an ein Modell ginge',
  freigabeErklaerung:
    'Stellt zusammen, was ein Modell zu sehen bekäme — und sendet es nicht. Wer danach generieren lässt, hat vorher gesehen, womit.',
  freigabeAufgabe: 'Aufgabe',
  freigabeAnbieter: 'Anbieter',
  freigabeModell: 'Modell',
  freigabeFenster: 'Kontextfenster',
  freigabeReserve: 'Ausgabereserve',
  freigabeZeigen: 'Zusammenstellen',
  freigabeOrtLokal: 'Gerechnet wird auf diesem Rechner.',
  freigabeOrtCloud: 'Gerechnet wird bei einem fremden Anbieter.',
  freigabeBudget: '{verbraucht} von {grenze} Tokens, {reserve} freigehalten',
  freigabeZaehlweise: 'geschätzt nach {zaehlweise}',
  freigabeZieladresse: 'Zieladresse',
  freigabeKategorien: 'Datenkategorien',
  freigabePrivat: 'Eigener Inhalt ist dabei.',
  freigabeOeffentlich: 'Nur öffentliche Quellen.',
  freigabeBelege: 'Belege',
  freigabeKeineBelege:
    'Keine Belege — so zusammengestellt bekäme ein Modell nichts zu sehen.',
  freigabeVeraltet: 'Nicht der aktive Stand',
  freigabeLaeuftAb: 'Gilt bis {zeit}',

  // **Die Freigabe des eigenen Bestands** — A13.4
  // (#1452).
  freigabeEigeneKarten: 'Meine eigenen Karten mitsuchen',
  freigabeBestandErklaerung:
    'Ohne dieses Kreuz bleibt der eigene Bestand außen — gesucht wird dann nur in öffentlichen Quellen, und die Antwort sagt es. Was mitgeht, verlässt bei einem Modell außerhalb dieses Rechners den Rechner.',
}
