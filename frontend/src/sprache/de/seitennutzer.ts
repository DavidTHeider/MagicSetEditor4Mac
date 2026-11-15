// **Die Seite „Nutzer und Profile“** — Phase 4.12
// (#1351).
//
// **Ein eigener Katalog und kein Anhang an `seiten.ts`** — die Zeilengrenze
// aus §2, und der Schnitt folgt der Frage: hier steht, was diese eine Seite
// sagt.
import { faktor, konto } from './seitenkonto'
import { sitzung } from './seitensitzung'

export const nutzer = {
  titel: 'Nutzer und Profile',

  // **Die drei Laschen** — auf Anweisung des Auftraggebers vom 2026-09-04.
  //
  // Der Schnitt ist der der Domäne: *Allgemeines* ist, was der Person gehört;
  // *Konto* ist, womit sie sich anmeldet; *Profil* ist die Arbeitsumgebung.
  // Die Oberfläche zeigt damit den Bau, statt ihn zu verstecken.
  laschen: {
    allgemein: 'Allgemeines',
    konto: 'Konto',
    profil: 'Profile',
    persoenlich: 'Persönliches',
  },

  allgemeinErklaerung:
    'Ein Nutzer ist die Person. Ein Profil ist eine Arbeitsumgebung dieser Person — eigene Sets, eigene Sammlung, eigene Einstellungen. Wer gerade wo arbeitet, steht in der Seitenleiste unter dem Anwendungsnamen; gewechselt wird hier.',

  kontoErklaerung:
    'Womit sich dieser Nutzer anmeldet. Solange es genau einen Nutzer mit einem Profil gibt, wird gar nicht gefragt.',

  profilErklaerung:
    'Wer für ein Turnier getrennt sammeln will, braucht kein zweites Konto, sondern ein zweites Profil.',

  // Die Liste
  niemand:
    'Noch kein weiterer Nutzer. Der erste ist beim ersten Start entstanden.',
  keineProfile: 'Dieser Nutzer hat noch kein Profil.',

  // Das Anlegen
  neuerNutzer: 'Neuen Nutzer anlegen',
  // **Der Zusatz seit dem 2026-09-05**: das Formular für die eigenen
  // Angaben (Werkstück 8) hat auch ein Feld „Anzeigename“, und zwei
  // gleich beschriftete Felder auf einer Seite sind für eine Vorlesehilfe
  // nicht zu unterscheiden.
  neuerNutzerName: 'Anzeigename des neuen Nutzers',
  neuerNutzerNameHinweis:
    'Eindeutig und ohne Leerzeichen am Rand — er unterscheidet zwei Nutzer',
  neuerNutzerKnopf: 'Nutzer anlegen',
  nutzerAngelegt: 'Angelegt: {name}',

  neuesProfil: 'Weiteres Profil anlegen',
  neuesProfilNutzer: 'Für welchen Nutzer',
  neuesProfilName: 'Name des Profils',
  neuesProfilNameHinweis:
    'Je Nutzer eindeutig — zwei Personen dürfen beide „Turnier“ führen',
  neuesProfilKnopf: 'Profil anlegen',
  // **Der Knopf im Kopf, der die Schublade öffnet** — UXR-20 (#1928), wie
  // im Artboard `Nutzer`. Die Auslassungspunkte sagen: davor kommt noch
  // eine Frage.
  profilAnlegenOeffnen: 'Profil anlegen …',
  profilAngelegt: 'Angelegt: {name}',

  // **Die Wahl** — Werkstück 3 (#1351).
  wahl: 'In welchem Profil gearbeitet wird',
  wahlErklaerung:
    'Sets, Decks, Sammlung, Turniere und Einstellungen gehören dem gewählten Profil. Ein Wechsel löscht nichts — was hier liegt, ist nach dem Zurückwechseln wieder da.',
  hier: 'Hier arbeitest du gerade',
  hierhin: 'Hierhin wechseln',
  hierhinFuer: 'Zu {nutzer} · {profil} wechseln',

  // **Wer arbeitet** — Werkstück 17 (#1400).
  //
  // **Die Erklärung nennt den Unterschied**, denn er ist nicht
  // selbsterklärend: in ein fremdes Profil zu sehen ist etwas anderes, als
  // jemand anderes zu sein.
  nutzerwahl: 'Wer gerade arbeitet',
  nutzerwahlErklaerung:
    'In ein fremdes Profil zu sehen ändert nicht, wer du bist — deine eigenen Freigaben entscheiden, wohin du darfst. Das ist etwas anderes als das Profil darüber. Wer jemand anderes wird, weist dessen zweiten Faktor nach; danach wird in dessen erstem Profil gearbeitet.',
  dasBinIch: 'Das bist du gerade',
  werden: 'Dieser Mensch werden',
  werdenFuer: '{nutzer} werden',

  // **Die Anmeldung** — Werkstück 4 (#1351).
  anmeldung: 'Anmeldung über einen Anbieter',
  anmeldungErklaerung:
    'Wer sich über einen Anbieter anmelden will, trägt die `client_id` unter **Einstellungen › Nutzer und Profil › Anmeldung über einen Anbieter** ein. MagicSetEditor4Mac liefert keine Zugangsdaten mit; die eigene OAuth-Anwendung legt man beim Anbieter an, und ein Geheimnis, falls er eines verlangt, kommt in den Schlüsselbund und nicht in den Bestand.',
  anmelden: 'Anmelden',
  anbieterstand: '{eingerichtet} · {geprueft}',
  eingerichtet: 'eingerichtet',
  nichtEingerichtet: 'nicht eingerichtet',
  geprueft: {
    simuliert: 'nur gegen eine Nachbildung geprüft',
    realGeprueft: 'gegen den echten Anbieter geprüft',
  },
  anmeldenBei: 'Bei {anbieter} anmelden',
  anmeldungAbbrechen: 'Abbrechen',
  anmeldungAbbrechenBei: 'Anmeldung bei {anbieter} abbrechen',
  verbunden: 'Verbunden',
  verbundenAls: 'Verbunden als {name}',
  loesen: 'Lösen',
  loesenVon: 'Die Verbindung zu {anbieter} lösen',

  // **Die Namen der Anbieter, wie sie sich selbst schreiben.**
  //
  // **Nicht übersetzt** (§9): ein Eigenname bleibt ein Eigenname, und wer „X“
  // sucht, findet „X“.
  anbieternamen: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  } as Record<string, string>,

  // **Ändern und Löschen** — Werkstück 5 (#1351).
  pflege: 'Nutzer und Profile bearbeiten',
  pflegeErklaerung:
    'Ein Name lässt sich ändern. Zum Löschen muss der Name stehen bleiben, wie er ist — er bestätigt, wer gemeint ist.',
  bearbeiten: '{name} bearbeiten',
  bearbeitenKurz: 'Bearbeiten',
  neuerNameFuer: 'Name',
  umbenennen: 'Umbenennen',
  loeschen: 'Löschen',
  loeschenBestaetigt: '{name} endgültig löschen',
  // **Und gefragt wird davor** — UX-10.AK-04 (#1682).
  loeschenFrage: '{name} endgültig löschen?',
  loeschenJa: 'Endgültig löschen',
  abbrechen: 'Abbrechen',
  nutzerLoeschenHinweis:
    'Ein gelöschter Nutzer nimmt seine Profile mit. Der letzte Nutzer lässt sich nicht löschen.',
  profilLoeschenHinweis:
    'Das letzte Profil eines Nutzers lässt sich nicht löschen. Was im Profil liegt, bleibt im Bestand, ist aber nicht mehr erreichbar.',

  // **Die Angaben zur Person** — Werkstück 8 (#1351), Punkte 1 bis 5, 14
  // und 15.
  angaben: 'Angaben zur Person',
  angabenErklaerung:
    'Alles außer dem Anzeigenamen ist freiwillig. Die **Anrede** ist ein freies Feld: jede Liste ließe jemanden aus, und eine Anwendung, die zwingt, sich für einen falschen Eintrag zu entscheiden, hat den Menschen falsch angeredet. Adresse, Rufnummer und Kennwort stehen unter **Konto**.',
  angabenSichern: 'Angaben sichern',
  felder: {
    name: 'Anzeigename',
    anrede: 'Anrede',
    namenszusatz: 'Namenszusatz',
    vorname: 'Vorname',
    weiterenamen: 'Weitere Vornamen',
    nachname: 'Nachname',
    slogan: 'Slogan',
    kurzebio: 'Kurze Biografie',
    langebio: 'Lange Biografie',
  },

  // **Das Profilbild** — Werkstück 8 (#1351).
  // **Die Profil-Einstellungen** — Werkstück 9 von Phase 4.12 (#1351).
  badge: 'Profil-Abzeichen',
  badgeVon: 'Abzeichen des Profils von {name}',
  badgeWaehlen: 'Abzeichen wählen …',
  badgeEntfernen: 'Abzeichen entfernen',
  profilangaben: 'Angaben zum Profil',
  profilangabenErklaerung:
    'Sie gehören **diesem Profil** und nicht dir — wer mehrere führt, unterscheidet sie hier. Der **Subslogan** steht unter deinem eigenen Slogan und ersetzt ihn nicht. Das **Abzeichen** liegt rechts unten über deinem Profilbild: es zeigt die Rolle, in der du gerade arbeitest.',
  profilangabenSichern: 'Profil sichern',
  profilfelder: {
    name: 'Profilname',
    subslogan: 'Subslogan',
    kurzebeschreibung: 'Kurze Beschreibung',
    langebeschreibung: 'Lange Beschreibung',
  },

  // **Technische Profile und ihre Zugriffstoken** — Werkstück 10 von
  // Phase 4.12 (#1351), Punkte 12 und 13.
  technisch: 'Technisches Profil — dahinter steht kein Mensch',
  technischErklaerung:
    'Ein technisches Profil trennt, was ein Dienst automatisch schreibt, von dem, was du selbst tust. Wer einen Dienst bedienen lässt, will nicht, dass er in **seinem** Profil schreibt: was dort entsteht, entsteht nicht von seiner Hand. Es gehört trotzdem dir, denn irgendwer muss es abschalten können.',
  ablauf: 'Läuft ab am (leer: gar nicht)',
  ablaufErklaerung:
    'Ein Zeitpunkt, ab dem dieses Profil abgelaufen ist — **kein Löschauftrag**. Nichts verschwindet von selbst.',

  token: 'Zugriffstoken',
  tokenErklaerung:
    'Damit weist sich ein Automat am REST-Zugang aus — in **diesem** Profil und mit der Erlaubnis, die hier steht. Am Fenster, an der Kommandozeile und über MCP wird kein Token verlangt: dort steht schon fest, wer arbeitet.',
  tokenLeer: 'Für dieses Profil gibt es noch kein Zugriffstoken.',
  tokenBezeichnung: 'Wofür es da ist',
  tokenScope: 'Erlaubnis',
  tokenLesen: 'Nur lesen',
  tokenSchreiben: 'Lesen und schreiben',
  tokenAblauf: 'Läuft ab am (leer: gar nicht)',
  tokenAnlegen: 'Token anlegen',
  tokenAendern: 'Übernehmen',
  tokenLoeschen: 'Zurückziehen',
  tokenLoeschenFrage: 'Den Token „{name}“ zurückziehen?',
  tokenLoeschenErklaerung:
    'Wer ihn benutzt, kommt danach nicht mehr herein. Ein neuer Token muss dort neu eingetragen werden.',
  tokenLoeschenJa: 'Zurückziehen',
  tokenAbgelaufen: 'abgelaufen',
  tokenUnbefristet: 'unbefristet',

  // **Der Satz steht VOR dem Anlegen und nicht danach.** Ein Feld, das man
  // später nachschlagen kann, kopiert niemand — und dann fehlt es.
  tokenEinmalig:
    'Der Klartext erscheint **genau einmal**, gleich nach dem Anlegen. Danach gibt es ihn nirgends mehr; wer ihn verliert, legt ein neues Token an.',
  tokenKlartext: 'Jetzt kopieren — dieser Wert kommt nicht wieder:',

  // **Rotation, Sammelwiderruf, Ablaufwahlen und die Einrichtung eines
  // MCP-Clients** — Werkstück O2 von Phase 5.2 (#1426).
  tokenPrincipal: 'Für wen es arbeitet',
  tokenAblaufwahl: 'Ablauf',
  tokenAblaufwahlen: {
    tag: 'Ein Tag',
    monat: 'Dreißig Tage',
    jahr: 'Ein Jahr',
    nie: 'Nie',
    eigen: 'Eigener Zeitpunkt',
  },
  tokenRotieren: 'Neues Geheimnis',
  tokenRotierenFrage: 'Ein neues Geheimnis erzeugen?',
  tokenRotierenErklaerung:
    'Das alte gilt noch eine Stunde, damit ein laufender Auftrag zu Ende kommt. Der neue Klartext erscheint genau einmal.',
  tokenAlleWiderrufen: 'Alle zurückziehen',
  tokenAlleWiderrufenFrage: 'Alle Zugriffstoken dieses Profils zurückziehen?',
  tokenAlleWiderrufenErklaerung:
    'Jeder Automat, der eines davon benutzt, kommt danach nicht mehr herein — auch mitten in einer Sitzung. Was damit angelegt wurde, bleibt.',
  tokenGezogen: '{zahl} Token zurückgezogen.',
  tokenLetzteNutzung: 'zuletzt benutzt {zeit}',
  tokenNieBenutzt: 'noch nie benutzt',
  tokenVorgaengerBis: 'altes Geheimnis gilt bis {zeit}',
  tokenOnboarding: 'So richtest du einen MCP-Client ein',
  tokenOnboardingSchritt1:
    'Lege ein Token an und kopiere den Klartext — er erscheint nur einmal.',
  tokenOnboardingSchritt2:
    'Schreibe ihn in eine Datei, die nur du lesen kannst (Rechte 0600). Eine Datei, die andere lesen dürfen, weist »mse-mcp« ab.',
  tokenOnboardingSchritt3:
    'Starte »mse-mcp« mit MSE_TOKEN_DATEI auf diese Datei — als Umgebungsvariable, nie als Argument: das stünde in der Prozessliste.',
  tokenOnboardingSchritt4:
    'Das Token wird bei jedem Werkzeugruf geprüft: ein Widerruf, ein Ablauf oder eine kleinere Erlaubnis wirkt beim nächsten Ruf. Ohne Token läuft »mse-mcp« wie bisher ohne Ausweis im gezeigten Profil.',
  tokenOnboardingBeispiel:
    '{ "mcpServers": { "mse": { "command": "/Applications/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp", "env": { "MSE_TOKEN_DATEI": "/Users/ich/.mse/mcp-token" } } } }',
  tokenOnboardingBeispielTitel: 'Beispiel für die Einrichtung',

  // **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
  // (#1351), Punkte 4 bis 7.
  vorgaben: 'Vorgaben dieses Profils',
  vorgabenErklaerung:
    'Womit ein neuer Datensatz in **diesem Profil** anfängt — ein Format, ein Spiel, eine Vorlage aus dem Bestand. Anders als die Einstellungen zeigen Vorgaben auf **Zeilen im Bestand**; was gewählt werden kann, hängt davon ab, was da ist.',
  vorgabenamen: {
    format: 'Vorgabeformat für neue Decks',
    spiel: 'Vorgabespiel für neue Sets',
    vorlage: 'Vorgabevorlage für neue Sets',
    deck: 'Präferenzdecks',
    lagerort: 'Zugeordnete Lagerorte',
  } as Record<string, string>,
  vorgabenOhne: 'ohne Vorgabe',
  vorgabenLeer: 'Dazu gibt es in diesem Bestand noch nichts.',
  vorgabenVerwaist: 'Zeigt ins Leere, weil es das nicht mehr gibt:',

  // **Rechte, Rollen und Sichtbarkeit** — Werkstück 11 von Phase 4.12
  // (#1351).
  freigaben: 'Wer an diesem Profil arbeiten darf',
  freigabenErklaerung:
    'Ein **Leser** sieht nur, dass es dieses Profil gibt — nicht, was darin steht. Ein **Mitarbeiter** wechselt hinein und arbeitet darin. Ein **Verwalter** vergibt selbst Freigaben. Dir gehört dieses Profil, wenn du es angelegt hast; dann verwaltest du es ohnehin.',
  freigabenNurLesen:
    'Freigaben vergeben darf nur, wer dieses Profil verwaltet.',
  freigabenLeer: 'Bisher darf niemand außer dem Eigner an dieses Profil.',
  freigabeNutzer: 'Wem',
  freigabeNiemand: 'niemandem',
  freigabeRolle: 'Als was',
  freigabeVergeben: 'Freigeben',
  freigabeZurueckziehen: 'Zurückziehen',
  rollennamen: {
    leser: 'Leser — sieht, dass es das Profil gibt',
    mitarbeiter: 'Mitarbeiter — wechselt hinein und arbeitet darin',
    verwalter: 'Verwalter — vergibt selbst Freigaben',
  } as Record<string, string>,

  bild: 'Profilbild',
  bildVon: 'Profilbild von {name}',
  bildWaehlen: 'Bild wählen …',
  bildEntfernen: 'Bild entfernen',
  ohneBild: '{name} hat kein Profilbild',

  // **Wohin der nächste Schritt führt, steht da** (§6: leere Zustände
  // erklären ihn, statt einen Knopf zu zeigen, der nichts tut).
  nochNicht:
    'Die Daten sind zwischen Profilen getrennt — solange niemand die Bestandsdatei mit einem fremden Werkzeug öffnet. Dagegen hilft die Verschlüsselung des Bestands unter „Einstellungen › Daten und Sicherung“.',

  // **Die vier Kontoarten** — Werkstück 6 (#1351), in einer eigenen Datei.
  konto,
  sitzung,

  // **Der zweite Faktor** — Werkstück 7 (#1351).
  faktor,

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351),
  // Punkte 9, 10 und 11.
  //
  // **Am 2026-09-06 aus `seiteneinstellungen` hierher gezogen**, nach dem
  // Zweitblick über die Phase: die Einstellungen einer Person lagen an zwei
  // Orten, und wer sie suchte, suchte an beiden.

  persoenlich: 'Mein Zugang',

  persoenlichErklaerung:
    'Diese Einstellungen gehören **dir** und nicht diesem Profil — du nimmst sie in jedes deiner Profile mit. Die beiden oberen übersteuern das Betriebssystem, und zwar **nur in eine Richtung**: sie schalten zu, nie ab.',

  auskunft: 'Meine Daten',

  auskunftErklaerung:
    'Was dieses Programm über dich gespeichert hat — Angaben, Anmeldewege, die **Arten** der zweiten Faktoren, Einstellungen und Profilnamen. **Ohne Kennwort und ohne Geheimnis**: beides ist kein Datum über dich, sondern ein Mittel, du zu sein. Sets, Decks und Sammlung stehen nicht darin — sie gehören einem Profil und gehen über ihre eigenen Wege hinaus.',

  auskunftSichern: 'Meine Daten sichern …',

  datenschutz: 'Was diese Anwendung über dich weiß',

  datenschutzHinweis:
    'Welche Angaben erfasst werden, wo sie liegen, was davon diesen Rechner verlässt — und was ausdrücklich nicht. Steht unter **Hilfe → Was diese Anwendung über dich weiß**.',
} as const
