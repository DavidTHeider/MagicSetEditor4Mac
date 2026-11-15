# 0066 — Nutzer, Profile und die vier Entscheidungen vor der ersten Zeile

**Stand:** angenommen
**Datum:** 2026-09-04
**Anlass:** #1351
(Phase 4.12) verlangt vier Entscheidungen, **bevor** eine Zeile entsteht — und
nennt für jede ausdrücklich: nachzulesen, nicht zu erinnern (§16, Regel 1).
Dieser Vermerk hält fest, was gemessen ist und was daraus folgt.

## Lage

Der Bestand kennt keinen Nutzer. Sets, Sammlungen, Decks, Turniere und
Einstellungen liegen in einer Datei, und wer sie öffnet, sieht alles.

**Die Recherche ist am 2026-09-04 gelaufen**, teils selbst, teils über einen
zweiten Agenten. Jede Zeile unten trägt ihren Status.

## Entscheidung 1 — Verschlüsselung

**Nicht SQLCipher, und die naheliegende Begründung war falsch.**

| Frage | Status | Ergebnis |
| --- | --- | --- |
| Verschlüsselt `modernc.org/sqlite`? | **nicht belegt** | Die Dokumentation von v1.58.0 nennt weder SQLCipher noch `PRAGMA key` noch ein Verschlüsselungs-VFS. **Das ist ein Schluss aus dem Schweigen, kein Beleg** |
| SQLCipher über Go? | **belegt** | `mattn/go-sqlite3` (die dokumentierte Anbindung von SQLite3 Multiple Ciphers) verlangt cgo und `gcc` |
| Ein cgo-freier Weg? | **belegt** | [`ncruces/go-sqlite3`](https://github.com/ncruces/go-sqlite3): „is a `cgo`-free SQLite wrapper", „wraps a VFS to offer encryption at rest" — Adiantum oder AES-XTS, MIT, v0.35.4 vom 2026-09-01 |

**Hier stand als erster Schluss „also Dateiebene".** Der war voreilig: es gibt
einen cgo-freien Weg zu einer verschlüsselten SQLite-Datei. Die Wahl ist
deshalb nicht „cgo gegen Dateiebene", sondern **ein zweiter SQLite-Treiber
gegen Dateiebene** — und was ein Treiberwechsel an 33 Migrationen, jeder
Abfrage und jedem Test kostet, ist eine eigene Frage.

**Der Auftraggeber hat am 2026-09-04 entschieden: „Dann gibt es halt den
zweiten Treiber."** Damit wird die Verschlüsselung gebaut, über
`ncruces/go-sqlite3` und dessen Verschlüsselungs-VFS.

**Was das kostet, gehört daneben** — 33 Wanderungen, jede Abfrage, jeder Test
gegen einen zweiten Treiber. Deshalb ist es ein **eigenes Werkstück** und
nicht ein Zusatz an einem anderen: ein Treiberwechsel, der nebenbei geschieht,
ist der Fehler, den man erst drei Phasen später sieht.

**Der Reihenfolge nach zuletzt.** Trennung, Anmeldung und zweiter Faktor
stehen dann schon; die Verschlüsselung legt sich darunter, statt dass alles
andere auf sie wartet.

**Und die Grenze gehört ausgesprochen, gleich welcher Weg.** Der Schlüssel
läge im Schlüsselbund, und Apple sagt dazu: „keychain services only provides
the data after a successful user authentication event." Verschlüsselt ist die
**ruhende** Datei; ein Prozess, der schon berechtigt ist, liest sie. Die
Anmeldemaske darf nichts anderes behaupten.

## Entscheidung 2 — Trennung ohne RLS

**SQLite hat kein `ROW LEVEL SECURITY`**, und das Issue sagt es selbst. Der
ehrliche Ersatz hat drei Teile, und der dritte ist der, der trägt:

1. Eine `profil_id` in jeder Tabelle, die Nutzerdaten führt.
2. Jede Abfrage über eine Sicht oder eine Bedingung, die sie klemmt.
3. **Ein Wächter, der eine Abfrage ohne diese Klausel rot macht.**

Ohne den dritten ist es eine Konvention, und Konventionen halten nicht.

**Was der Ersatz nicht ist**, gehört danebengeschrieben: RLS wirkt in der
Datenbank, auch gegen ein fremdes Werkzeug. Eine Spalte plus Wächter wirkt in
**dieser Anwendung**. Die Abnahme des Issues verlangt aber ausdrücklich, dass
zwei Profile einander auch dann nicht sehen, „wenn man die Bestandsdatei mit
einem fremden Werkzeug öffnet" — **und das leistet nur Verschlüsselung je
Profil, nicht eine Spalte.**

**Erfüllbar wird sie erst mit der Verschlüsselung**, und die kommt (oben).
Bis dahin gilt: die Trennung wirkt in dieser Anwendung, und **das ist die
Zusage, die die Oberfläche machen darf** — keine andere.

Der Auftraggeber hat am 2026-09-04 zugestimmt: „Ist ok, solange Trennung
geht."

### Wie sie gebaut wurde — Werkstück 2, am 2026-09-04

**Zehn Wurzeln tragen die Spalte, die Abkömmlinge erben sie über ihren
Fremdschlüssel** (`0035_profiltrennung.sql`). **Fortgeschrieben durch
[0131](0131-ein-abkoemmling-fragt-nach-seiner-wurzel.md)**: seit dem
2026-09-23 fragt jeder Abkömmling in der Ablage nach seiner Wurzel — vorher
war er nur so sicher wie die Kennung, die ein Aufrufer nannte. Eine `profil_id` an
`card_values` wäre eine zweite Wahrheit neben `card_id`, und zwei Wahrheiten
laufen auseinander, sobald jemand eine Karte verschiebt.

**Sieben der elf Tabellen mussten neu gebaut werden**, weil ihr Schlüssel das
Profil einschließen muss: `settings` hat `schluessel` als Primärschlüssel —
ohne das Profil darin könnten zwei Profile nicht beide eine Sprache haben.
SQLite kann einen Primärschlüssel nicht erweitern; der Weg ist neu bauen,
kopieren, tauschen.

**Die vier übrigen bekamen sie per `ALTER TABLE`** — und damit `DEFAULT 0`,
denn anders rüstet SQLite eine `NOT NULL`-Spalte nicht nach. Profil 0 gibt es
nicht (`AUTOINCREMENT` beginnt bei 1), und ohne Gegenmaßnahme wäre das der
leiseste aller Fehler: eine vergessene Zuordnung glückt beim Schreiben, und
niemand sieht die Zeile je wieder. **Acht Trigger** weisen sie deshalb ab, für
`INSERT` und für `UPDATE`.

**Das Profil reist im Kontext**, nicht in der Signatur — sonst wären es 94
Signaturen, und kein Aufrufer, den es nichts angeht, bliebe unberührt.
[`storage.ProfilAus`](../../internal/storage/profil.go) gibt einen **Fehler**,
wenn keines gesetzt ist; eine stillschweigende Null gäbe null Zeilen, und die
Anwendung sähe aus, als wären alle Daten weg.

**Vier Flächen statten ihn aus, auf zwei Arten.** Das Fenster bekommt seinen
Kontext von Wails und kann ihn beim Entstehen nicht anfassen — dort hängt
`App.imProfil` ihn an **jede** Bindung, gespeist aus dem Zeiger, den der Start
gelesen hat. Kommandozeile, MCP und REST bauen ihren Wurzelkontext selbst und
statten ihn einmal aus (`storage.MitGezeigtemProfil`).

**Beim Dienst gehört eine Zeile dazu, die man vergisst**: `net/http` baut den
Kontext einer Anfrage aus `BaseContext` und nicht aus dem, mit dem jemand
`Serve` gerufen hat. Ohne sie trug der Wurzelkontext das Profil, und **kein
Behandler sah es**. Gefunden hat das nicht der Test des Dienstes — der war
grün —, sondern der Flächenlauf gegen den **gebauten** Dienst.

**Drei Wächter halten das Ganze**, und jeder hält etwas anderes:

| Wächter | Zusage | Was er **nicht** hält |
| --- | --- | --- |
| `internal/werkzeuge/profilklausel_test.go` | jede Abfrage auf eine Wurzeltabelle nennt `profil_id` | ob die Klausel richtig herum steht |
| `internal/werkzeuge/profilnaht_test.go` | jede Bindung beginnt mit `ctx = a.imProfil(ctx)` | ob die Bindung den Bestand überhaupt anfasst |
| `internal/storage/ohneprofil_test.go` | keine geklemmte Naht arbeitet ohne Profil | ob die Trennung danach wirkt |

**Der zweite hat sich beim Bauen selbst widerlegt**: seine erste Fassung las
nur einzeilige Signaturen, meldete grün — und **47 mehrzeilige** hatten die
Zeile nicht. Ein Wächter, der die halbe Zusage hält, macht die andere
unsichtbar (§10).

### Und der Wechsel — Werkstück 3, am 2026-09-05

**Es gibt einen Zeiger, und er liegt im Bestand.** Das Issue verlangt, dass
„der gewählte Nutzer und das gewählte Profil den Aufruf überleben" — sonst wäre
jeder Befehl an der Kommandozeile eine eigene Anmeldung. Ein Zeiger im Prozess
könnte das nicht.

**Was das kostet, gehört gesagt**: wer an der Kommandozeile wechselt, wechselt
auch für das Fenster. Bei **einem** Rechner mit **einer** arbeitenden Person ist
das richtig; wer für einen einzelnen Aufruf etwas anderes will, nimmt
`--profil=<kennung>`, und das verstellt den Zeiger nicht.

**Am Fenster räumt ein Wechsel den offenen Kontext** (§7, Regel 1) — und zwar
den ganzen: `App.tsx` trägt einen `key` an der **Anwendung**. Beim Set ist das
Bezugsobjekt eine Zeile im Bestand; beim Profil ist es alles. Ein offenes
Kartenformular zeigte danach eine Karte, die dieses Profil nicht hat, und sein
„Speichern" schriebe ins Leere.

**Zwei Befunde am eigenen Werk**, beide derselben Gestalt — *ein Stand, der
einmal gelesen und nie wieder gefragt wird*:

| Wo | Was |
| --- | --- |
| **Der Dienst** | `net/http` baut den Kontext einer Anfrage aus `BaseContext`, nicht aus dem von `Serve`. Der Wurzelkontext trug das Profil vom Start, und `PUT /profil` änderte den Zeiger, den keine Anfrage mehr las. Jetzt stattet eine Zwischenschicht **jede** Anfrage aus |
| **MCP** | Dieselbe Lage in einer langen Sitzung: `schreiben.profil.wechseln` änderte den Stand, aus dem der nächste Ruf liest — und der Ruf las den alten. Jetzt stattet der Server **jeden** Ruf aus |

**Den zweiten hätte kein vorhandener Lauf gefunden.** Die Vorrichtung im
Flächenlauf startete je Ruf einen eigenen Prozess und war damit großzügiger als
der Betrieb (§3); erst ein Lauf, der beide Rufe in **eine** Sitzung legt, macht
den Unterschied sichtbar. Die Gegenprobe ist gesehen: mit der alten Gestalt ist
er rot, der Lauf mit einem Prozess je Ruf bleibt grün.

**Was hier noch fehlt**: der zweite Faktor, den das Issue für den Wechsel
verlangt. Er kommt in Werkstück 5 und landet in
`service.Nutzerverwaltung.Wechseln` — ein Wechsel, den eine Prüfung abweist,
findet nicht statt. **Bis dahin verspricht ihn keine Fläche** (§10: auch eine
Ablehnung braucht Deckung).

## Entscheidung 3 — Der zweite Faktor

**TOTP wird gebaut, Biometrie nicht.**

| Frage | Status | Ergebnis |
| --- | --- | --- |
| TOTP | **belegt** | RFC 6238: `T = (Unix-Zeit − T0) / X`, X = 30 s, T0 = 0; HMAC-SHA-1, optional SHA-256/512. **Das Prüffenster ist eine Entscheidung des Prüfers**, keine Vorgabe der RFC |
| LocalAuthentication | **belegt** | ObjC-Framework. `LAPolicyDeviceOwnerAuthenticationWithBiometrics` (macOS 10.12.2+), `LAPolicyDeviceOwnerAuthentication` (10.11+), `…WithCompanion` und `…WithBiometricsOrCompanion` (beide 15.0+) — gelesen in `LocalAuthentication.framework/Headers/LAContext.h` des macOS-SDK |
| Ohne cgo erreichbar? | **belegt, Drittanbieter** | Eine von Apple dokumentierte Kommandozeilenschnittstelle gibt es **nicht**. [`tmc/apple`](https://github.com/tmc/apple) bindet die Frameworks „cgo-free, using purego" und führt ein `localauthentication`-Paket — MIT, v0.6.18 vom 2026-08-22, **keine stabile 1.0** |

**Beides wird gebaut** — Entscheidung des Auftraggebers vom 2026-09-04: „Beides
vorsehen, auch Biometrie. Vor 1.0 ist kein Hinderungsgrund."

| Weg | Womit | Was ihn trägt |
| --- | --- | --- |
| **TOTP** | `crypto/hmac`, `crypto/sha1` — reines Go | keine Plattformbindung, läuft überall |
| **Biometrie** | `tmc/apple/localauthentication` über purego | cgo-frei, MIT, aber **v0.6.18 und keine 1.0** |

**Die Bauform folgt daraus, und sie ist die eigentliche Entscheidung:** ein
Profil trägt **einen oder mehrere** zweite Faktoren, und die Anmeldung fragt
den, der auf diesem Rechner geht. **TOTP ist der, der immer geht** — er
funktioniert auch auf einem Mac ohne Touch ID, unter Windows und Linux (§12),
und er ist der Rückfall, wenn die Bindung sich als unreif erweist.

**Was die Unreife kostet, bleibt gesagt:** eine Bindung vor 1.0 kann ihre
Schnittstelle ändern. Sie liegt deshalb hinter einer eigenen Naht mit einem
Bautag, so wie `internal/schluesselbund` es für den Schlüsselbund tut — fällt
sie aus, fällt die Biometrie aus und nicht die Anmeldung.

## Entscheidung 4 — Was mit dem heutigen Bestand geschieht

**Eine Migration legt einen ersten Nutzer mit einem ersten Profil an und
ordnet ihm alles zu, was da ist.** Niemand ist danach ausgesperrt, und wer nie
einen zweiten anlegt, merkt von der Phase nichts.

**Ohne Anmeldung geht es weiter**, solange es genau einen Nutzer mit einem
Profil gibt. Das ist die Abnahme aus dem Issue („Ein bestehender Aufruf ohne
Anmeldung bricht nicht") und zugleich die Bedingung dafür, dass MCP,
Kommandozeile und Dienst nicht über Nacht eine Anmeldung brauchen.

## Wer entschieden hat

**Drei dieser vier Entscheidungen hat der Auftraggeber am 2026-09-04 anders
getroffen, als dieser Vermerk sie vorschlug** — und das ist der Grund, warum
sie hier stehen und nicht in einem Kopf: ein Vorschlag, der nirgends steht,
lässt sich nicht umkehren.

| | Vorschlag | Entschieden |
| --- | --- | --- |
| Verschlüsselung | nicht in dieser Phase | **gebaut**, mit zweitem Treiber |
| Trennung | Einschränkung benannt | **angenommen**, solange sie wirkt |
| Zweiter Faktor | nur TOTP | **beides**, auch Biometrie |
| Bestand | erste Wanderung | **angenommen** |

## Die Anbieter — vier, nicht fünf

**Gemessen am 2026-09-04**, jede Zeile mit Fundstelle im Issue-Nachtrag:

| Anbieter | PKCE | `state` | Identität aus | Status |
| --- | --- | --- | --- | --- |
| GitHub | ja, **nur `S256`** | ja | einem zweiten Ruf | **belegt** |
| Google | ja, `S256` oder `plain` | ja | `id_token` (OIDC) | **belegt** |
| Microsoft | ja, beide | ja | `id_token`, `nonce` **Pflicht** im hybrid flow | **belegt** |
| **X** | ja, `S256` oder `plain` | **Pflicht**, bis 500 Zeichen | einem zweiten Ruf | **belegt** |
| ~~Apple~~ | — | — | — | **scheidet aus** |
| OpenAI | — | — | — | **kein Weg belegt** |

**Apple scheidet aus, und zwar belegt:** Sign in with Apple verlangt über den
Services-ID-/REST-Weg einen Developer-Account, eine Services ID, einen privaten
Schlüssel **und eine bestehende App im App Store**. Die Mitgliedschaft kostet
99 USD im Jahr. Dieses Projekt hat keinen Account und soll keinen bekommen
([0003](0003-keine-notarisierung.md)). **X tritt an seine Stelle** — auf
ausdrückliche Anweisung des Auftraggebers vom 2026-09-04.

**OpenAI bleibt offen, nicht gestrichen.** „Sign in with ChatGPT" existiert als
Identitätsanbieter — Startpartner sind Airtable, GitLab, HubSpot, Notion,
Supabase, Vercel. Aber **jede** Entwicklerseite bei OpenAI beschreibt die
umgekehrte Richtung: ChatGPT als OAuth-**Klient** gegen einen fremden Server,
mit Rückadresse bei `chatgpt.com/aip/…`. Ein Weg, auf dem sich eine fremde
Desktop-Anwendung als Klient registriert, ist **nicht belegt**.

Das ist kein „unmöglich", sondern ein „nicht gefunden" — und der Unterschied
gehört stehen zu bleiben. Die Anbieternaht wird so gebaut, dass OpenAI eine
Zeile ist, sobald es einen Weg gibt.

## Zwei Bauformen, und das ist der Zuschnitt

Die vier Anbieter zerfallen in zwei Gestalten, und daran hängt der Code:

| Gestalt | Wer | Die Identität kommt |
| --- | --- | --- |
| **OIDC** | Google, Microsoft | aus dem `id_token`, geprüft gegen die Schlüssel des Anbieters |
| **OAuth pur** | GitHub, X | aus einem **zweiten Ruf** an eine Nutzerabfrage |

Wer das nicht trennt, baut vier Sonderfälle. Wer es trennt, baut zwei Wege und
vier Zeilen Konfiguration.

**Und bei X kostet der zweite Ruf etwas:** die freie Stufe der X-API führt
**100 Reads im Monat**. Ob eine Nutzerabfrage im Nutzerkontext dagegen zählt,
ist **nicht geprüft** — das gehört gemessen, bevor X als Anbieter angeboten
wird.

## Was der Bau gekostet hat — Werkstück 4, am 2026-09-05

**Drei Dinge waren vor der ersten Zeile nachzulesen** (§16, Regel 1), und alle
drei ändern den Bau:

| Anbieter | `client_secret` beim Tausch | Wortlaut |
| --- | --- | --- |
| **GitHub** | **Pflicht**, auch mit PKCE | „`client_secret` \| string \| **Required**" |
| **Google** | **optional** für Desktop-Klienten | „(**Optional**)" |
| **Microsoft** | **verboten** | „Public clients, which include native applications and single page apps, **must not use secrets or certificates** when redeeming an authorization code" |

**Ein Fluss, der das nicht je Anbieter führt, scheitert bei einem von dreien** —
und zwar erst am Tauschendpunkt, wo die Meldung niemandem etwas sagt. Deshalb
tragen `Anbieter.GeheimnisPflicht` und `…Verboten` die Entscheidung, und
`Beginne` weist beide Verstöße ab, bevor der Browser aufgeht.

### Und deshalb liefert dieses Projekt keine Kennungen mit

**Jeder Anbieter verlangt eine registrierte Anwendung**, GitHub dazu ein
Geheimnis. Ein mitgeliefertes wäre keines: es stünde in einem öffentlichen
Repositorium und in jedem ausgelieferten Bündel. **Wer sich anmelden will,
registriert seine eigene Anwendung** und trägt Kennung und Geheimnis ein — das
Geheimnis in den Schlüsselbund, aus demselben Grund wie jedes andere
([0056](0056-drei-weitere-kartendienste.md)).

**Was das für die Abnahme heißt, gehört gesagt**: ein Lauf gegen den **echten**
Anbieter hat hier nie stattgefunden und kann es nicht — es gibt keine
registrierte Anwendung. Geprüft ist der **Weg**, gegen einen selbst
geschriebenen Anbieter (§16): dass PKCE mitgeht, dass `state` sperrt, dass das
Geheimnis dort steht, wo es hingehört, und dass aus der Antwort eine Identität
wird.

### Der Rückruf geht auf die Schleifenadresse

**Nachgelesen bei Google**: „http://127.0.0.1:port or http://[::1]:port. Query
your platform for the relevant loopback IP address and start an HTTP listener
on a random available port." Und eigene Schemata sind **abgeschafft**: „Custom
URI schemes are no longer supported due to the risk of app impersonation."

Der Port steht deshalb erst, wenn der Horcher steht — und die Rückadresse
reicht der Horcher in den Lauf, statt dass der Lauf sie wüsste.

### Die Signatur des `id_token` wird nicht geprüft

**OIDC Core 3.1.3.7 erlaubt es**: „If the ID Token is received via direct
communication between the Client and the Token Endpoint (which it is in this
flow), the TLS server validation MAY be used to validate the issuer in place of
checking the token signature."

**Was das kostet**: die Identität hängt an der TLS-Kette des Tauschendpunkts.
Wer ein gültiges Zertifikat dafür vorweisen kann, kann sie fälschen — er könnte
dann allerdings ebenso den ganzen Tausch beantworten. Dafür entfallen JWKS,
Schlüsselwechsel und eine RSA-Prüfung.

**Der `nonce` wird trotzdem geprüft**, wenn einer dasteht: dieselbe Stelle
verlangt es, und ohne ihn ließe sich ein Token wiedereinspielen, das für eine
andere Anmeldung ausgestellt wurde.

## Was ausdrücklich nicht kommt

Newsletter und Support. Die Stammdaten werden so geführt, dass es später
möglich wäre — mehr nicht. Was nicht gebaut wird, wird auch nicht versprochen.

## Verworfen

- **SQLCipher über cgo.** Dieses Projekt ist cgo-frei, und
  `internal/schluesselbund` hat dieselbe Frage schon einmal so entschieden.
  Der zweite Treiber ist die Antwort, nicht cgo.
- **Eine Spalte als „RLS".** Das Wort wäre eine Behauptung über eine
  Fähigkeit, die die Ablage nicht hat.
- **Apple als Anbieter** — nicht aus Geschmack, sondern weil der Weg einen
  Account verlangt, den es nicht gibt.
