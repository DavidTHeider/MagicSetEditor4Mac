// Die Einstellungen der KI — die lokale Laufzeit, Apples Gerätemodell und
// die Adresse (A7, #1446; AI-05, #2042). **Aus `seiteneinstellungen.ts`
// herausgelöst**, als der Katalog über die Zeilengrenze aus §2 ging.

export const kinamen = {
  'ki.laufzeit': 'Entorno de ejecución de IA local',
  'ki.geraetemodell': 'Modelo en el dispositivo de Apple',
  'ki.vorrangmodell': 'Modelo de Ollama para la prioridad',
  'ki.adresse': 'Dirección del entorno de ejecución de IA',
} as const

export const kihinweise = {
  'ki.laufzeit':
    'Si se consulta siquiera un entorno de ejecución de IA. **Desactivado** significa que no sale ninguna llamada, ni siquiera a este ordenador. Desde aquí no se instala ni se inicia nada.',
  'ki.geraetemodell':
    'Si se consulta el modelo en el dispositivo de macOS para tareas de texto pequeñas: nombre de carta, texto de ambientación. **Automático** significa: si está activado en este Mac (Apple Intelligence activado, licencia de `fm` aceptada), se integra y se consulta primero; los datos no salen del equipo. **Desactivado** rige hasta que lo cambies; desactivarlo para una sesión está en **IA**.',
  'ki.vorrangmodell':
    'Qué modelo de Ollama consulta una sugerencia con **prioridad**, después del modelo del dispositivo. **Vacío** significa el primero que puede responder con una forma fija, y Ollama nombra primero el último descargado. Un modelo elegido que Ollama ya no nombra se notifica y no se sustituye en silencio por otro.',
  'ki.adresse':
    'Dónde escucha el entorno de ejecución. Vacío significa la dirección predeterminada de este ordenador. Una dirección que **no** sea de este ordenador clasifica todo modelo tras ella como no local; con solo local no se consulta nada en absoluto.',
} as const

export const kiwerte = {
  'ki.laufzeit.aus': 'Desactivado',
  'ki.laufzeit.an': 'Activado',
  'ki.geraetemodell.automatisch': 'Automático',
  'ki.geraetemodell.aus': 'Desactivado',
} as const
