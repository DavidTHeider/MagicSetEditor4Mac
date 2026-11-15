# 0122 — Ein fremdes Konto wird gelesen und nicht bedient

**Stand:** entschieden · **Datum:** 2026-09-19 ·
**Anlass:** Werkstück `T11` von Phase 5.6
(#1476)

## Worum es geht

CardTrader ist seit `A4`
(#1443)
ein **Marktplatz**, den dieses Programm liest: was fremde Verkäufer für eine
Karte verlangen ([0095](0095-was-cardtrader-sagt-und-was-davon-gilt.md)). Die
Schnittstelle kann mehr. Sie kann das **eigene** Konto zeigen — und sie kann
darin schreiben.

`T11` fragt, was davon hier gebaut wird. Die Antwort hat zwei Hälften, und
die zweite ist die wichtigere.

## Was nachgelesen ist

**Abgerufen am 2026-09-19** von
<https://www.cardtrader.com/en/docs/api/full/reference>, im Browser gelesen
und nicht erinnert (§16, Regel 1).

| | |
| --- | --- |
| Grundadresse | `https://api.cardtrader.com/api/v2` |
| Anmeldung | `Authorization: Bearer <Marke>`; die Marke steht in den Profileinstellungen des Kontos |
| Grenze | **200 Anfragen je 10 Sekunden** über **alle** Endpunkte zusammen; darüber kommt `{"error": "Too many requests: max 200 requests per 10 seconds"}` |

**Die drei lesenden Wege zum Konto:**

| Weg | Was er gibt |
| --- | --- |
| `GET /info` | die App und ihre Nutzerkennung — der Weg, den der Anbieter selbst zum **Prüfen der Anmeldung** nennt |
| `GET /products/export` | die **eigenen** Angebote; wahlfrei nach `blueprint_id` oder `expansion_id`. Der Anbieter warnt, dass der ganze Abruf bei großen Beständen Sekunden dauert, und empfiehlt eine **Lesefrist von 120 bis 180 Sekunden** |
| `GET /orders` | die **eigenen** Bestellungen; `page` ab 1, `limit` 1–100 (ohne Angabe 20), dazu `from`/`to`, `from_id`/`to_id`, `state`, `order_as` und `sort` |

**Die acht Bestellzustände**: `paid` → `sent` → `arrived` → `done`; über
**CardTrader Zero** kommt `hub_pending` dazu; daneben `request_for_cancel`,
`canceled` und `lost`.

**Und die Tabelle, die alles entscheidet** — der Anbieter schreibt sie selbst
so hin:

| Zustand | über CardTrader Zero | Bestand abbuchen? |
| --- | --- | --- |
| `hub_pending` | ja | **ja** |
| `paid` | ja | **nein** |
| `paid` | nein | **ja** |

**Warum sie so aussieht**: CardTrader Zero legt zu jedem Verkauf **zwei**
Bestellungen an — eine im Zustand `hub_pending` und einmal je Woche eine
gebündelte im Zustand `paid`. Wer beide für Verkäufe hielte, nähme dasselbe
Stück zweimal aus dem Bestand. Der Anbieter sagt es wörtlich: „otherwise you
risk to do it twice."

**Bei Vorverkauf kommt eine vierte Bedingung dazu** — die Bestellkennung muss
der `hub_pending_order_id` der Position gleichen. Sie hängt an der
**Position** und nicht am Zustand; `domain.BewegtDenBestand` kennt sie
deshalb nicht, und das steht dort (§10: die Reichweite einer Messung ist die
Reichweite ihrer Aussage).

## Die Entscheidung

**Gelesen wird, geschrieben nicht.**

Gebaut sind die drei Wege oben — hinter **zwei** Schaltern, die beide
ausgeschaltet anfangen: dem Dienst `cardtrader` (0063) und der neuen
Einstellung `konto.schau`. Dazu der Riegel an der Netzgrenze
([0078](0078-der-wirksame-akteur.md)): `externSenden`, also mindestens
Mitarbeiter.

**Nicht gebaut sind alle Wege, die schreiben**, und sie sind namentlich
bekannt:

`POST /products` · `PUT /products/:id` · `DELETE /products/:id` ·
`POST /products/:id/increment` · `POST /products/bulk_create` ·
`POST /products/bulk_update` · `POST /products/bulk_destroy` ·
`POST /product_imports` · `POST /cart/add` · `POST /cart/remove` ·
**`POST /cart/purchase`** · `POST /wishlists` · `DELETE /wishlists/:id` ·
`PUT /orders/:id/tracking_code` · `PUT /orders/:id/ship` ·
`PUT /orders/:id/request-cancellation` ·
`PUT /orders/:id/confirm-cancellation` · `PATCH /app`

## Warum

**Weil eine Anbieterfreigabe fehlt, und weil sie das kleinere Hindernis
wäre.**

Das Issue sagt es selbst: „Falls reale notwendige API-/Testberechtigung
fehlt, Issue bleibt optional offen bzw dokumentiert konkret gesperrte
Teilfähigkeit; reine Adaptertests werden nicht als produktive
Anbieterabnahme ausgegeben." Genau das ist der Fall — es gibt hier kein
CardTrader-Konto und keine Marke, und eines zu beschaffen ist keine
Maßnahme, die im Rahmen dieses Werkstücks getroffen wird.

**Das Größere ist die Umkehrbarkeit.** `POST /cart/purchase` kauft. Was
dieses Programm bei einem Fehler anrichten könnte, endet nicht in seinem
Bestand, sondern auf einer Rechnung — und die nimmt niemand mit einer
Gegenbewegung zurück. Dafür bräuchte es, was das Issue selbst auflistet: ein
persistentes Operationsjournal, eine Abstimmung gegen den Anbieterstatus, und
die Antwort auf „ein Timeout nach einem `POST` heißt **Ausgang unbekannt**,
nicht gescheitert". Nichts davon lässt sich gegen eine selbst geschriebene
Gegenstelle **belegen**; es lässt sich nur behaupten.

**Und die Adaptertests belegen das Schreiben ohnehin nicht.** Sie belegen,
was dieser Code aus einer Antwort macht — nicht, was der Anbieter mit einer
Anfrage tut. Der Unterschied ist derselbe wie in
[0007](0007-kein-entfernter-lauf.md): gebaut ist nicht gelaufen.

## Was das kostet, und es gehört ausgesprochen

**Wer über dieses Programm verkaufen will, kann es nicht.** Angebot anlegen,
Preis ändern, Warenkorb, Kauf — alles bleibt im Browser des Anbieters.
`MarktlinkOeffnen` aus `T8`
(#1473)
führt dorthin, und das ist der ganze Weg, den dieses Programm anbietet.

**Und die Kontoschau nützt ohne Marke nichts.** Sie ist gebaut, geprüft und
ausgeschaltet; wer sie benutzen will, trägt eine Marke ein, die er sich beim
Anbieter holt.

## Was hier ausdrücklich nicht geschieht

- **Kein Konto wird angelegt** und keine Marke beschafft.
- **Kein Testkauf**, auch kein kleiner. Das Issue verbietet es, und es wäre
  auch ohne das Verbot der falsche Weg, einen Schreibpfad zu prüfen.
- **Das `shared_secret` aus `GET /info` wird nicht gelesen.** Der Anbieter
  schreibt „Do not share this information" daneben; die billigste Art, es
  nicht weiterzugeben, ist, es nicht zu haben — es steht in keinem Feld
  dieses Programms und damit in keinem Export, keinem Diagnosepaket und
  keinem Umschlag (§11).
- **Kein zweiter Schlüsselbundeintrag.** Es ist dieselbe Marke wie am
  Marktplatz; ein zweiter Eintrag wäre eine zweite Stelle, an der sie
  veralten kann.
- **Keine automatische Buchung.** Was `domain.BewegtDenBestand` sagt, ist
  eine **Auskunft**: „nach der Regel des Anbieters hätte diese Bestellung ein
  Stück genommen". Wer daraus eine Bewegung macht, tut es über die
  vorhandenen Schreibwege, und die verlangen seit `T10`
  (#1475)
  die Zustimmung eines Menschen.

## Verworfen

**Einen zweiten Dienst `cardtrader-konto` einzuführen.** Er hätte einen
eigenen Schalter gebracht — den es jetzt auch gibt — und dazu einen zweiten
Zugangseintrag für **dieselbe** Marke. Zwei Stellen für ein Geheimnis sind
eine zu viel.

**Die Kontoschau an den Dienstschalter zu hängen.** Sie wäre dann an, sobald
jemand Marktpreise sehen will — und das ist eine andere Entscheidung: der
Marktplatz beantwortet, was **fremde** Verkäufer verlangen, das Konto, was
**bei dir** liegt.

**Die Bestellzustände zu übersetzen.** `paid` wäre „bezahlt", und der
Vergleich gegen die Antwort des Anbieters ginge über eine Zuordnung, die beim
nächsten Zustand auseinanderläuft. Sie behalten ihre Namen; übersetzt wird
erst am Fenster (§9).
