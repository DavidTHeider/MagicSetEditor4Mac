// Die Einstellungen der KI — die lokale Laufzeit, Apples Gerätemodell und
// die Adresse (A7, #1446; AI-05, #2042). **Aus `seiteneinstellungen.ts`
// herausgelöst**, als der Katalog über die Zeilengrenze aus §2 ging.

export const kinamen = {
  'ki.laufzeit': 'Local AI runtime',
  'ki.geraetemodell': 'Apple on-device model',
  'ki.vorrangmodell': 'Ollama model for priority',
  'ki.adresse': 'Address of the AI runtime',
} as const

export const kihinweise = {
  'ki.laufzeit':
    'Whether an AI runtime is asked at all. **Off** means no call goes out — not even to this machine. Nothing is installed or started from here.',
  'ki.geraetemodell':
    'Whether the macOS on-device model is asked for small text tasks — card names, flavor text. **Automatic** means: if it is enabled on this Mac (Apple Intelligence on, the `fm` license accepted), it is included and asked first; the data does not leave this machine. **Off** applies until you change it back; switching it off for one session is under **AI**.',
  'ki.vorrangmodell':
    'Which Ollama model a suggestion with **priority** asks — after the on-device model. **Empty** means the first one that can answer in a fixed form, and Ollama lists the most recently pulled first. A chosen model that Ollama no longer lists is reported, not silently replaced by another.',
  'ki.adresse':
    'Where the runtime listens. Empty means the default address on this machine. An address that is **not** on this machine classifies every model behind it as non-local — under local only nothing is asked at all.',
} as const

export const kiwerte = {
  'ki.laufzeit.aus': 'Off',
  'ki.laufzeit.an': 'On',
  'ki.geraetemodell.automatisch': 'Automatic',
  'ki.geraetemodell.aus': 'Off',
} as const
