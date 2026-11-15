# 0167 — CLF und UYW bleiben zweistellig: eine benannte Abweichung von ISO 4217

**Stand:** angenommen am 2026-10-04 · G8-Rest aus dem Zweitblick über Phase
6.3 (#1664),
Phase 6.10 zugeordnet
(#1671).
Marke `new`: MSE2 kennt keine Beträge.

## Der Befund

ISO 4217 führt für **CLF** (Unidad de Fomento, Chile) und **UYW** (Unidad
Previsional, Uruguay) **vier** Nachkommastellen; `domain.Waehrungsstellen`
gibt zwei, und `frontend/src/waehrungsstellen.ts` mit ihm. Gegengelesen am
2026-10-04 in CLDR über `Intl` (Node 24.15.0): `maximumFractionDigits` ist
für beide **4**. Der Kommentar „ISO 4217" über den Listen sagte damit mehr,
als die Listen hielten.

## Die Entscheidung

**Beide bleiben zweistellig**, und die Abweichung steht benannt an den
Listen. Geändert werden die Kommentare, nicht die Skala.

## Warum

**Die Stellenzahl ist die Skala gespeicherter Beträge**
([0127](0127-ein-betrag-steht-in-der-kleinsten-einheit.md)): ein Betrag
steht in der kleinsten Einheit. Vier statt zwei Stellen hieße, jeden
gespeicherten CLF- und UYW-Betrag mit hundert zu multiplizieren — und das
an Stellen, die sich nicht umschreiben lassen:

- **Unterschriebene Handelsrevisionen.** Eine `Handelsposition` trägt ihre
  Bewertung in der kleinsten Einheit, und ist sie geteilt
  (`BewertungGeteilt`), geht die Zahl in die Unterschriftsform und damit in
  den `Abdruck`. Umgeschrieben stimmte der Abdruck einer Revision nicht
  mehr, die ein Gegenüber bestätigt hat.
- **Der Draht.** Zwei Fassungen dieses Programms bedeuteten mit derselben
  Zahl das Hundertfache voneinander — eine brechende Änderung nach
  [0033](0033-was-ueber-fassungen-hinweg-zugesagt-wird.md), für zwei
  Rechnungseinheiten.
- **Zwölf Migrationen** führen Währungsspalten; eine Umrechnung müsste
  jede kennen und jede künftige auch.

## Was es kostet

**CLF und UYW lassen sich nur auf Hundertstel genau erfassen** — bei rund
39.000 CLP je CLF sind das Schritte von etwa 390 CLP. Beide sind
Rechnungseinheiten; **kein angebundener Dienst liefert Preise in ihnen**
(Scryfall und MTGJSON liefern USD, EUR und TIX — `internal/scryfall`,
`internal/mtgjsonpreise`). Wer sie von Hand einträgt, bekommt zwei
Stellen.

## Verworfen

- **Umskalieren mit einer Migration** — aus den drei Gründen oben.
- **Vier Stellen nur für neue Beträge** — dieselbe Zahl hätte je nach Alter
  zwei Bedeutungen; genau das schließt 0127 aus.

## Was es zurücknimmt

Wer CLF oder UYW auf vier Stellen bringen will, braucht eine Hauptfassung
des Drahtformats und eine Umrechnung der Handelskette, die ihre Abdrücke
neu bestätigt — und ändert zuerst diesen Vermerk.
