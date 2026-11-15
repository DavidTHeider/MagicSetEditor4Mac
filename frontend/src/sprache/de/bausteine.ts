// Was auf mehr als einer Seite vorkommt: Aktionen, Feldnamen, Zustandswörter.
import { dateiansicht } from './bausteinedatei'
import { ablage } from './bausteineablage'
import { geste } from './bausteinegeste'
import { laufstand } from './bausteinelauf'
import { objektmarken } from './bausteinemarken'
import { teilen } from './bausteineteilen'

export const bausteine = {
  dateiansicht,
  ablage,
  geste,
  laufstand,
  objektmarken,
  teilen,
  // **Eine Liste, zwei Gestalten** — UX-11 (#1683): Sets und Decks.
  // **Ein Entwurf und ein fremder Wert im selben Feld** — U01 bis U06
  // (#2119).
  entwurf: {
    fremdGeaendert:
      'Inzwischen steht hier »{wert}« — anderswo geändert. Dein Entwurf ist noch nicht gespeichert.',
    behalten: 'Meinen Entwurf speichern',
    nehmen: 'Den neuen Wert übernehmen',
  },
  listenansicht: {
    suche: 'Suchen',
    suchePlatzhalter: 'Name oder Spiel',
    sortierung: 'Ordnen',
    ansicht: 'Ansicht',
    liste: 'Liste',
    kacheln: 'Kacheln',
    zahl: '{gezeigt} von {gesamt}',
    nichtsPasst: 'Nichts passt zu »{suche}«.',
    sucheLeeren: 'Suche leeren',
    filterZuruecksetzen: 'Suche und Filter zurücksetzen',
    weitere: 'Weitere {anzahl} zeigen',
    wieImBestand: 'Wie im Bestand',
    zuletztBenutzt: 'Zuletzt benutzt zuerst',
    nachName: 'Nach Name',
    nachKarten: 'Nach Kartenzahl',
  },
  // **Ein Gegenüber aufnehmen, wo gehandelt und gespielt wird** —
  // UX-03.AK-03 (#1675).
  // **Ob das Eingetippte im Bestand steht** — UX-05.AK-03 (#1677).
  speicherstand: {
    schreibt: 'Wird gespeichert …',
    geschrieben: 'Jede Änderung steht im Bestand.',
    fehler: '„{feld}“ ist nicht gespeichert — der Grund steht am Feld.',
  },
  gegenueberTitel: 'Gegenüber aufnehmen oder bestätigen',
  // **Ein Gegenüber zur Wahl** — UX-08.AK-04 (#1680): Name und Fingerabdruck,
  // die Form, die zwei Menschen vergleichen.
  kontaktwahlEintrag: '{name} · {fingerabdruck}',
  kontaktwahlSuchen: 'Gegenüber suchen …',
  // **Das Hilfesymbol am Feld** — UX-15.AK-02
  // (#1771).
  //
  // **Es nennt das Feld**, damit es auch dann trägt, wenn dreißig davon auf
  // einer Seite stehen: „Hilfe“ allein sagt nicht, wozu (WCAG 2.4.6) —
  // derselbe Grund wie beim Satz am Pflichtfeld darunter.
  feldhilfe: 'Erklärung zu {feld}',

  // **Der Name des Hilfeknopfes an einem Erläuterungsabsatz** — UX-15.AK-01
  // (#1771).
  //
  // **Er nennt den Satz und nicht den Abschnitt.** Dreißig solche Absätze
  // stehen über den Baum verteilt, und ihre Abschnitte heißen nicht überall
  // gleich — der erste Satz steht dagegen unmittelbar daneben und ist
  // eindeutig. „Mehr dazu“ allein wäre bei dreißig Knöpfen dreißigmal
  // derselbe Name (WCAG 2.4.6).
  mehrZu: 'Mehr zu „{satz}“',

  // **Der Satz an einem leeren Pflichtfeld** — Werkstück F von Phase 4.10
  // (#1325).
  //
  // **Er nennt das Feld**, damit er auch dann trägt, wenn drei
  // nebeneinanderstehen: „ohne diesen Wert“ allein ließe offen, welcher
  // gemeint ist. Und er sagt, was zu tun ist (§6) — nicht bloß, dass etwas
  // fehlt.
  // **Woher etwas kam** — Werkstück 12 von Phase 4.12
  // (#1351).
  //
  // **Zwei Sätze und kein zusammengesetzter**: „aus einem Profil übernommen“
  // und „eingelesen“ sind verschiedene Aussagen, und ein gemeinsamer Rahmen
  // mit ausgetauschtem Wort stünde in einer der beiden Sprachen schief (§9).
  herkunftAusProfil: 'Übernommen aus dem Profil {quelle}.',
  herkunftAusEinfuhr: 'Eingelesen aus {quelle}.',

  pflichtfeldLeer: '{feld} braucht einen Wert.',

  // **Das Feld für einen Währungscode** — UX-06
  // (#1678).
  //
  // **Das Beispiel steht im Feld und nicht darunter** (AK-02) — das
  // Formular der Sammlung bleibt eine Zeile (UX-15). **Der Satz nennt, was
  // dastand, und was erwartet wird** (AK-03) — und er sagt „kein
  // Währungscode" und nicht „unbekannt“: drei Buchstaben sind die Gestalt
  // eines Codes, nicht der Beweis, dass es die Währung gibt (AK-04).
  waehrung: {
    beispiel: 'z. B. EUR',
    keineGestalt:
      '»{wert}« ist kein Währungscode — der hat drei Buchstaben, wie EUR.',
    grossSchreiben:
      '»{wert}« — ein Währungscode steht in Großbuchstaben: {gross}.',
  },

  // **Die Sprachen eines Drucks** — UX-06 (#1678), zu den Codes aus
  // `drucksprachen.ts`, nachgelesen bei Scryfall.
  drucksprachen: {
    beispiel: 'z. B. en',
    namen: {
      en: 'Englisch',
      es: 'Spanisch',
      fr: 'Französisch',
      de: 'Deutsch',
      it: 'Italienisch',
      pt: 'Portugiesisch',
      ja: 'Japanisch',
      ko: 'Koreanisch',
      ru: 'Russisch',
      zhs: 'Chinesisch (vereinfacht)',
      zht: 'Chinesisch (traditionell)',
      he: 'Hebräisch',
      la: 'Latein',
      grc: 'Altgriechisch',
      ar: 'Arabisch',
      sa: 'Sanskrit',
      ph: 'Phyrexianisch',
      qya: 'Quenya',
      dw: 'Zwergisch',
      tlh: 'Klingonisch',
    },
  },

  // **Das Kennwortfeld** — Werkstück 6 von Phase 4.12
  // (#1351).
  //
  // **Der Umschalter sagt, was er tun wird**, nicht wie es gerade steht — das
  // sagt `aria-pressed`.
  kennwort: {
    zeigen: 'Das Kennwort im Klartext zeigen',
    zeigenKurz: 'Zeigen',
    verbergen: 'Das Kennwort wieder verdecken',
    verbergenKurz: 'Verbergen',
  },

  // **Das Feld für den zweiten Faktor** — Werkstück 7 von Phase 4.12 (#1351).
  //
  // **Es fragt nach dem Code, nicht nach der Art**: sechs Ziffern aus der App
  // oder zwölf Zeichen vom Zettel sind für den, der sie eingibt, dasselbe.
  nachweis: {
    beschriftung: 'Code',
    abschicken: 'Bestätigen',
  },

  aktionen: {
    abbrechen: 'Abbrechen',
    oeffnen: 'Öffnen',
    schliessen: 'Schließen',
    speichern: 'Speichern',
  },

  // **Stile zur Wahl** — UX-05, S07.NAT-02 (#1677): gesucht wird über den
  // Namen und die Kennung.
  stilliste: {
    mitZusatz: '{kennung} · {zusatz}',
    keinerGewaehlt: '— Stil wählen —',
  },
  // **Drucke aus der eigenen Sammlung wählen** — UXR-F18 (#2033): die Wahl bei
  // „Ausgewählte Drucke“ und die Behaltemenge bei „Dubletten“ einer Sicht.
  druckwahl: {
    drucke: 'Drucke',
    hinweis:
      'Das Gegenüber sieht nur diese Drucke, mit der Menge, die du besitzt.',
    suchen: 'Drucke suchen',
    gewaehlt: '{anzahl} gewählt',
    nurSuche:
      'Deine Sammlung hat {anzahl}. Suche nach einem Namen, um welche zu wählen.',
    keiner: 'Kein Druck passt zu „{suche}“.',
    genauer: 'Es passen mehr Drucke, als hier stehen. Suche genauer.',
    leer: 'In deiner Sammlung steht noch kein Druck, den du freigeben könntest. Nimm zuerst Exemplare auf.',
    fehlt:
      'Wähle mindestens einen Druck — eine Auswahl ohne Drucke gibt nichts frei.',
    behalten: 'Davon behalte ich je Druck',
    behaltenHinweis:
      'Das Gegenüber sieht, was darüber liegt: bei 1 und drei Stück zwei.',
    detailBehalten: 'du behältst {behalten}',
  },

  // **Eine Auswahl mit Suche** — UXR-F1 (#2007), R7 aus
  // `docs/UIREVIEW.md`: ab fünfzehn Werten gehört zur Auswahl eine Suche.
  suchauswahl: {
    keiner: 'Nichts passt zu „{suche}“.',
    treffer: 'Treffer: {anzahl}',
    begrenzt: '{gezeigt} von {anzahl} Treffern gezeigt — Suche eingrenzen …',
  },
  // **Die Zeitzone als Auswahl** — leer heißt: die des Rechners.
  zeitzonenwahl: {
    system: 'Wie das System',
  },

  // **Der Aktionskatalog** — UX-10.AK-01 (#1682): je Aktion das Kurzlabel und
  // was sie überall bedeutet. Symbol, Kürzel und Bestätigung stehen in
  // `aktionen.ts`, denn sie sind keine Sprache.
  aktionskatalog: {
    aktionen: {
      anlegen: { kurz: 'Anlegen', heisst: 'Legt etwas Neues an.' },
      bearbeiten: {
        kurz: 'Bearbeiten',
        heisst: 'Ändert es an Ort und Stelle.',
      },
      sichern: { kurz: 'Sichern', heisst: 'Hält den Stand fest.' },
      zuruecknehmen: {
        kurz: 'Zurücknehmen',
        heisst: 'Nimmt den letzten Zug zurück.',
      },
      duplizieren: { kurz: 'Duplizieren', heisst: 'Legt ein zweites daneben.' },
      kopieren: { kurz: 'Kopieren', heisst: 'Legt es in die Zwischenablage.' },
      entfernen: {
        kurz: 'Entfernen',
        heisst: 'Nimmt es aus dieser Ansicht; anderswo bleibt es.',
      },
      loeschen: { kurz: 'Löschen', heisst: 'Löscht es endgültig, überall.' },
    },
  },
  // **Der Hinweis, dass hier nur gelesen wird** — Werkstück 16 von Phase
  // 4.12 (#1351).
  nurlesen: {
    hinweis:
      'Du siehst dieses Profil als Leser: du kannst alles ansehen und nichts ändern. Wechsle in dein eigenes Profil, um zu arbeiten — oder bitte die Person, die dieses verwaltet, um die Rolle »Mitarbeiter«.',
    knopf: 'Nur lesen — in diesem Profil darfst du nichts ändern.',
  },

  // **Die Kontextleiste** — Werkstück U4 von Phase 5.1 (#1415): wer wo als
  // was arbeitet, jederzeit sichtbar.
  kontext: {
    bezeichnung: 'Arbeitskontext',
    profilVon: 'von {eigner}',
    rollen: {
      leser: 'Leser',
      mitarbeiter: 'Mitarbeiter',
      verwalter: 'Verwalter',
    },
    nurLesen: 'nur lesen',
    zuDenNutzern: 'Profil wechseln …',
    zeile: '{name} · {profil}',
  },

  zustaende: {
    laedt: 'Wird geladen …',
    leer: 'Nichts vorhanden',
    fehler: 'Das ließ sich nicht laden.',
    fehlerMitGrund: 'Das ließ sich nicht laden. {grund}',
    // **Ein Fehlschlag bietet den nächsten Versuch an** — UX-08 S15.NAT-02.
    wiederholen: 'Erneut laden',
    // **Ein Konflikt bietet den heutigen Stand an** — UX-08.AK-02 (#1680).
    neuLaden: 'Neu laden',
  },
  // Suchen und Ersetzen **im Feld** (#1164).
  suche: {
    oeffnen: 'Im Text suchen',
    schliessen: 'Suche schließen',
    suchen: 'Suchen',
    ersetzen: 'Ersetzen durch',
    treffer: '{bei} von {ganz}',
    keinTreffer: 'Kein Treffer',
    weiter: 'Nächster Treffer',
    zurueck: 'Voriger Treffer',
    einsetzen: 'Ersetzen',
    alleEinsetzen: 'Alle ersetzen',
    grossKlein:
      'Groß- und Kleinschreibung zählt — in MSE-Script sind »card« und »Card« zweierlei.',
  },
  // Vervollständigung **im** Feld (#1164).
  // **Welche Karten hinausgehen** (#1167).
  ausgabeauswahl: {
    aufschrift: 'Welche Karten',
    ganzesSet: 'Das ganze Set ({zahl} Karten)',
    ganzesSetOhneZahl: 'Das ganze Set',
    teilmenge: 'Nur ausgewählte ({zahl})',
    ohneNamen: 'ohne Namen',
    eingrenzen: 'Karten eingrenzen',
  },

  vorschlaege: {
    bezeichnung: 'Vorschläge',
    uebernehmen: 'Mit ↹ übernehmen',
  },

  // **Die Regionen eines Arbeitsplatzes** — Werkstück U2 von Phase 5.1
  // (#1413),
  // nach Vermerk 0076. Der Knopf sagt, was er tun wird, nicht wie es steht —
  // das sagt `aria-expanded`.
  arbeitsplatz: {
    breiteVon: 'Breite von „{region}“',
    breiteWert: '{punkte} Punkte',
    navigatorAuf: 'Navigator einblenden',
    navigatorZu: 'Navigator ausblenden',
    inspectorAuf: 'Inspector einblenden',
    inspectorZu: 'Inspector ausblenden',
    navigatorEng:
      'Zu schmal für alle drei Spalten — der Navigator geht über dem Zentrum auf',
    inspectorEng:
      'Zu schmal für alle drei Spalten — der Inspector geht über dem Zentrum auf',
  },
  // **Der Druck ohne Alias** — Werkstück A2 von Phase 5.3
  // (#1441, 0086).
  druck: {
    vorlaeufig: 'vorläufig',
    vorlaeufigErklaerung:
      'Dieser Druck hat noch keine Scryfall-Kennung. Preise und Bilder bleiben aus, bis er aufgelöst ist.',
  },
  // **Die Kommandozentrale** — Werkstück U12 von Phase 5.1
  // (#1423).
  kommandozentrale: {
    titel: 'Kommandozentrale',
    erklaerung:
      'Ein Befehl ist ein Werkzeug des MCP-Servers, hier im Fenster gerufen — dieselbe Antwort wie über MCP und auf der Kommandozeile. Nichts geht durch eine Shell: was du eingibst, bleibt Daten.',
    lesend: 'liest',
    schreibend: 'ändert',
    gefaehrlich: 'nicht ohne Weiteres zurückzunehmen',
    faehigkeit: 'Fähigkeit',
    cli: 'Kommandozeile',
    rest: 'Dienst',
    parameter: 'Parameter',
    keineParameter: 'Dieser Befehl braucht keine Parameter.',
    pflicht: 'Pflicht',
    jsonHinweis:
      'Als JSON — eine Liste in eckigen, ein Objekt in geschweiften Klammern.',
    befundPflicht: 'Dieses Feld braucht einen Wert.',
    befundZahl: 'Das ist keine Zahl.',
    befundJson: 'Das ist kein JSON.',
    vorschau: 'Was geschickt wird',
    bestaetigen: 'Ich habe die Vorschau gelesen und will das so ausführen.',
    ausfuehren: 'Ausführen',
    abbrechen: 'Abbrechen',
    laeuft: 'Der Befehl läuft …',
    ergebnis: 'Ergebnis',
    ausgabe: 'Ausgabe des Befehls',
    alsText: 'Text',
    alsJson: 'JSON',
    dauer: '{ms} ms',
    verlauf: 'Verlauf dieser Sitzung',
    verlaufLeer:
      'Noch nichts ausgeführt. Geheimnisse — Token, Kennwörter — stehen hier nie im Klartext.',
    wiederholen: '{name} erneut laden',
    gelungen: 'gelungen',
    gescheitert: 'gescheitert',
    abgebrochen: 'abgebrochen',
    schliessen: 'Schließen',
  },
  // „Diagnose öffnen“ an einem Fehler mit Vorgangskennung (NET-02,
  // S13.NAT-04, #1706).
  diagnoseverweis: {
    oeffnen: 'Diagnose öffnen',
  },
  // „KI einrichten“ an einer KI-Meldung (AI-01.AK-04, #1742).
  kieinrichtung: {
    oeffnen: 'KI einrichten',
  },
  kikasten: {
    titel: 'KI zu dieser Karte',
  },
} as const
