# Risikokatalog — was die Tests schützen müssen

**Die kritischen Pfade, ihre Invarianten und die Läufe, die sie halten** —
QA-05 (#1736).
Eine hohe Abdeckung ist hier kein Ziel; das Ziel ist, dass ein Fehler an
einer dieser Stellen **rot** wird. Ob er es wird, misst der Mutationslauf
([`belege/mutationen.md`](belege/mutationen.md)).

**Gemessen am 2026-10-01: 12 von 12 Mutationen gefangen** — jede
abgeschaltete Wache machte einen benannten Lauf rot. **Der erste Lauf meldete
11**: M8 „überlebte", weil der Lauf `internal/service` auf
`Partie|Empfang|…` eingeengt hatte und der Test, der sie fängt,
`TestEinDoppeltZugestelltesEreignisWirktNurEinmal` heißt. Eine Lücke im
Mutationslauf und nicht in den Tests — seitdem fährt `service` ganz (§10:
eine Gegenprobe, die nichts verschiebt, ist keine).

## Wie zu lesen

Je Bereich die **Invarianten** — was nie geschehen darf — und je Invariante
drei Arten von Läufen:

- **positiv**: der erlaubte Weg wirkt, und zwar fachlich (der Bestand steht
  danach so, wie er soll),
- **negativ**: der verbotene Weg wird abgewiesen, **und die Nebenwirkung
  bleibt aus** — nichts geschrieben, nichts gelöscht, nichts gesendet,
- **zustandsübergreifend**: ein Wechsel dazwischen — Profil, Sitzung,
  Leitung, ein zweiter Lauf — ändert daran nichts.

`M…` verweist auf eine Mutation im Mutationslauf: dort ist die Wache
abgeschaltet, und die genannten Läufe müssen rot werden.

## Eingabe

| Invariante | positiv | negativ | zustandsübergreifend | M |
| --- | --- | --- | --- | --- |
| Was ein Fremder schickt, ist begrenzt, **bevor** zugeteilt wird | `leitung/leitung_test.go`: `TestEineNachrichtGehtHinUndKommtAn` | `TestEinBehaupteterRahmenStelltKeinenSpeicher`, `TestEineZuGrosseNachrichtGehtNichtHinaus`; `FuzzRahmen` | `TestEinAbbruchUnterbrichtEinWartendesEmpfangen` | M4 |
| Eine Einlösung einer neueren Fassung wird als solche erkannt | `domain/einloesung_test.go` | `FuzzZerlegeEinloesung`; die Antwort „fassung" statt „gescheitert" | NET-01b: ältere Fassung antwortet „unbeantwortet" (`TestEineEinladungGenuegt`) | M11 |
| Eine freie Eingabe wird geprüft oder ist ausdrücklich ungeprüft | [`FELDPRUEFUNG.md`](FELDPRUEFUNG.md) je Datei | dort mit Grund | — | — |

## Autorisierung

| Invariante | positiv | negativ | zustandsübergreifend | M |
| --- | --- | --- | --- | --- |
| Eine Rolle reicht nur so weit, wie sie reicht | `domain/akteur_test.go`: `TestDieRolleBegrenztDieWirkung` | `TestEinTokenEngtEinUndErweitertNie`, `TestDieAbsageNenntDieAchse`; die Rollenmatrix über vier Flächen (`flaechen/rollenmatrix_test.go`) | `app/ohnerolle_test.go`: `TestOhneRolleWirdImFremdenProfilNichtsGeschrieben` | M1 |
| Ein nur lesendes Profil schreibt nicht | `TestNurLesenSchlaegtJedeRolle` | `query_only` am Bestand ([0073](entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)) | ein Profilwechsel räumt Sitzungen, Teilungen und Horchen (`app/teilprofil_intern_test.go`, `app/fernwechsel_intern_test.go`) | M2 |
| Ein falsches Kennwort wird abgewiesen, in gleicher Zeit | `kennwort/kennwort_test.go`: `TestEinKennwortLaesstSichHuetenUndWiedererkennen` | `TestEineKaputteFormWirdAbgewiesen` | `sicherung`: `TestEinFalschesKennwortWirdAmFeldAbgewiesen` | M6 |
| Was hinausgeht, braucht einen Nachweis | `TestWasHinausgehtBrauchtEinenNachweis` | ein Nachweis gilt einem Schritt und einmal ([0078](entscheidungen/0078-der-wirksame-akteur.md)) | ein geänderter Stand macht ihn ungültig | — |

## Datei

| Invariante | positiv | negativ | zustandsübergreifend | M |
| --- | --- | --- | --- | --- |
| Ein Eintrag einer Sicherung führt nicht aus ihr hinaus — auch nicht über einen Umweg | `sicherung/archiv_test.go`: `TestEineSicherungGibtHerausWasHineinging` | `fehlerwege_test.go`: `TestEinEintragDerHinausfuehrtWirdAbgewiesen`; `FuzzEntpacken` | `TestEinGescheitertesPackenLaesstNichtsLiegen` | M5 |
| Ein Pfad über REST und MCP bleibt in den erlaubten Orten | `domain/dienstpfad_test.go`: `TestPruefeDienstordner` | `TestPruefeDienstpfad` (verborgene Ordner, `~/Library`, Verknüpfungen) | — | — |
| Eine angewendete Migration wird nicht nachträglich verändert | `storage/migrator_test.go` | eine geänderte Prüfsumme sperrt aus | `storage/stolpermigrationen_test.go` | M7 |
| Ein Archivpaket wird vermessen, bevor es gelesen wird | `msepaket/archivumfang_test.go` | `ErrArchivZuGross` (#1886) | — | — |

## Handel

| Invariante | positiv | negativ | zustandsübergreifend | M |
| --- | --- | --- | --- | --- |
| Ein Handel wird nicht halb verbucht | `service/handelsverbuchung_test.go`: `TestEineVerbuchungBuchtBeideRichtungenUndUnterschreibt` | `TestOhneBestaetigteUebergabeWirdNichtVerbucht` | `TestZweimalVerbuchenBuchtEinmal` | M9 |
| Was geschrieben ist, bleibt | Auslöser gegen das Umschreiben von Beleg und Journal ([0152](entscheidungen/0152-was-geschrieben-ist-bleibt.md)) | eine überholte Nachricht schreibt nichts | ein Widerspruch führt in die Klärung | — |
| Zwei Umschläge, zwei Unterschriftsformen | `domain/handelsumschlag_test.go` | `FuzzHandelsumschlag` | — | — |

## Regel und Partie

| Invariante | positiv | negativ | zustandsübergreifend | M |
| --- | --- | --- | --- | --- |
| Eine doppelt zugestellte Partienachricht wirkt einmal | `service/partieempfang_test.go` | eine Lücke führt zur Wiederaufnahme (`service/partiestoerung_test.go`) | `FuzzEmpfaengerstand` | M8 |
| Die Priorität geht der Reihe nach, zustandsbasierte Aktionen greifen | `domain/prioritaet_test.go`: `TestDerAktiveSpielerBekommtSieZuerst` | `zustandsaktionen_test.go`: `TestBeiNeunVergiftungsmarkenGeschiehtNichts` | `TestAlleNacheinanderBeendenDenAbschnitt` | — |
| Eine beendete Partiesitzung wird freigegeben | `app/partiewiederkehr_intern_test.go`: `TestNachDemAufstehenGibtDerWirtDieSitzungFrei` | `TestNachDemAbrissGibtDerGastDieSitzungFrei` | `TestEinGastSetztSichAnEinenLaufendenTisch` | M12 |
| Eine Sitzung staut höchstens so viel, wie ihre Tiefe erlaubt | `service/sitzungtiefe_test.go` | der Absender erfährt „voll" | `service/partiewirtschub_test.go` | M10 |

## Netz und Gegenüber

| Invariante | positiv | negativ | zustandsübergreifend | M |
| --- | --- | --- | --- | --- |
| Eine Nonce wirkt einmal | `peer/handschlagschluss_test.go`: der Handschlag mit frischer Nonce | `peer/handschlaggrenzen_test.go`: `TestDieselbeNonceEinZweitesMalWirdAbgewiesen`, `TestEineZuKurzeZufallszahlIstKeine` | `peer/noncelistegrenze_intern_test.go`: `TestDieNoncelisteHaeltIhreGrenze`, `TestDieFristRaeumtVonVorn` | M3 |
| Ein Kennwort aus einer Proxyadresse erreicht kein Protokoll | `netz/proxykennwort_test.go` | `fernleitung/protokoll_intern_test.go` | — | — |

## Zurückspielen

| Invariante | positiv | negativ | zustandsübergreifend | M |
| --- | --- | --- | --- | --- |
| Eine abgewiesene Sicherung lässt den Bestand unberührt | `storage/sicherungspruefung_test.go`: `TestPruefeSicherungNimmtEineGueltigeDatei` | `TestPruefeSicherungWeistFremdeDateienAb`, `TestPruefeSicherungWeistEineNeuereFassungBenanntAb` | `service/rueckspielrest_intern_test.go`: `TestEinSpurlosVerschwundenerAuszugGiltNichtAlsGetauscht` | — |

## Abdeckung — Zähler und Nenner getrennt

- **Go misst Anweisungen, keine Zweige.** `go test -cover` zählt, welche
  Anweisungsblöcke liefen; ein `if` ohne `else`, dessen Bedingung nie
  falsch war, steht trotzdem als gedeckt da. **Deshalb gibt es diesen
  Katalog und den Mutationslauf**: sie fragen nach dem Zweig, den die Zahl
  nicht sieht.
- **Die Schwellen** stehen in `scripts/ci.sh`: 100 % in den Kernpaketen, 90 %
  im Gesamtprojekt — gemessen am 2026-10-01 **93,07 %, 117.872 von 126.644
  Anweisungen**. Die Nenner je Paket nennt jede Zeile der Teststufe
  (`internal/leitung 100 % (704 Anweisungen)`).
- **Das Frontend misst Zweige**: V8 über Vitest, 100 % in Anweisungen,
  Zweigen, Funktionen und Zeilen — am selben Tag **9.705 Zweige**.
- **85 bis 90 % sind für neue kritische Bereiche eine Untergrenze** und keine
  Erlaubnis, die 100 % der Kernpakete zu senken (QA-05.AK-02).
- **Ausnahmen**: 21 Einträge in `.covignore`, je mit Grund. Ausgenommen wird
  nur, was ohne echtes Fenster, echten Prozess oder echte Plattform nicht
  läuft — nie, weil ein Zweig unbequem ist (§3).

## Was ohne Instrumentierung läuft

Start, Fenster, Menüleiste, Dateidialoge und der Tisch im WebKit laufen nicht
unter `go test -cover`. **Ihr Beleg ist ein eigener Lauf an der gebauten
Anwendung**: die Hüllenprobe (`docs/belege/huellenprobe/`) und die WebKit-Probe
(`docs/belege/webkitprobe/`). Was nur ein Mensch belegen kann, steht in
[`HANDPRUEFUNGEN.md`](HANDPRUEFUNGEN.md).
