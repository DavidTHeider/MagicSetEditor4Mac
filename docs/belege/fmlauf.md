# Echtlauf gegen Apples Gerätemodell

**2026-10-04 01:50** — Commit d6da05431, Mac17,3, macOS 27.0.1 (26A434).

go test -tags fmlauf -run TestDasEchteGeraetemodell -v ./internal/applefm/

- **Kaltstart** bis /health: 0.06 s
- **TCP-Ports**, auf denen fm serve lauscht: keiner

| Fall | Ausgang | Antwort | Dauer |
| --- | --- | --- | ---: |
| Name, deutsch | fertig | Wachsame Kreatur | 0.53 s |
| Name, englisch | fertig | Alert Elk | 0.41 s |
| Flavortext, deutsch | fertig | Ein geheimnisvoller Nebelhirsch durchstreift die dunklen Wälder bei Nacht. Sein Ruf lässt die Stille erzittern. | 1.20 s |
| Flavortext, englisch | fertig | The Fog Elk moves silently through the mist, its antlers catching the dim light. | 0.70 s |
| Ablehnung, englisch | abgelehnt | The model's safety guardrails were triggered. | 0.12 s |

Die Aufträge stehen in internal/applefm/fmlauf_test.go.
