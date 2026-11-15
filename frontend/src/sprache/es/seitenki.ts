// **Die Texte der KI** — Werkstück A6 von Phase 5.3
// (#1445),
// erweitert für A7
// (#1446).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als einer
// ankommt** (§9): `seiteneinstellungen.ts` stand mit ihnen an der
// Zeilengrenze aus §2.
export const ki = {
  kimodelleTitel: 'Modelos de IA',
  kimodelleKeine:
    'No hay ningún modelo de IA establecido. La lista de modelos de un proveedor todavía no es una capacidad demostrada: aquí solo aparecen los modelos que se han comprobado.',
  kiNurLokal:
    'Rige «solo local»: no se consulta ningún modelo fuera de este equipo.',
  kiAuchCloud:
    'Se permiten modelos de proveedores externos. Lo que se les envía sale de este equipo.',
  kiBeleg: '{faehigkeit} — comprobado el {datum}: {beleg}',
  kiKontextfenster: 'Ventana de contexto: {tokens} tokens — {beleg}',
  kiKontextfensterUnbekannt: 'Ventana de contexto: no indicada',
  kiHerkunftUnbekannt:
    'Procedencia desconocida: no se ha demostrado qué archivo se ejecuta realmente aquí.',
  kiEinrichtungAus:
    'No activado. No se consulta nada y nada sale de este equipo.',
  kiEinrichtungUnerreichbar:
    'Nadie responde en la dirección configurada. Esta aplicación no inicia ni instala nada allí.',
  kiEinrichtungOhneModell:
    'Responde, pero no tiene ningún modelo. Esta aplicación no descarga ninguno.',
  kiEinrichtungOhneFaehigkeit:
    'Hay modelos, pero ninguno tiene una capacidad demostrada que sirva de algo aquí.',
  kiEinrichtungBereit:
    'Al menos un modelo está listo con una capacidad demostrada.',
  kiLaufzeitOllama:
    'Si aquí no responde nada, Ollama no se está ejecutando. Esta aplicación no lo inicia ni lo instala: eso se hace en el propio Ollama.',
  kiLaufzeitTitel: 'IA local',
  kiLaufzeitFassung: 'Entorno de ejecución: {fassung}',
  kiAdresse: 'Se consulta {adresse}.',
  kiAdresseVorgabe: 'la dirección predeterminada de este equipo',
  kiCloudAus: 'La descarga de trabajo a la nube de Ollama está desactivada.',
  kiCloudAn: 'La descarga de trabajo a la nube de Ollama no está desactivada.',
  kiCloudUngewiss:
    'Desde aquí no se puede determinar si la descarga de trabajo a la nube de Ollama está desactivada.',
  kiCloudBeleg: 'Prueba: {beleg}',
  kiVorschlagTitel: 'Sugerencia de un modelo',
  kiVorschlagModell: 'Modelo',
  kiVorschlagAuftrag: 'Tarea',
  kiVorschlagAuftragHinweis: '¿Qué debe hacer el modelo?',
  kiVorschlagen: 'Sugerir',
  kiVorschlagLaeuft:
    'El modelo está trabajando; la primera vez puede tardar porque se está cargando.',
  kiVorschlagAbbrechen: 'Cancelar',
  kiVorschlagUebernehmen: 'Aplicar',
  kiVorschlagNichts:
    'El modelo no sugirió ningún campo que conozca la plantilla.',
  kiVorschlagAendertNichts:
    'Una sugerencia no cambia la carta. Solo «Aplicar» escribe.',
  kiVorschlagUebernommen: 'Se escribieron {felder} campos.',
  kiVorschlagAblauf: 'Aplicable hasta {ablauf}.',
  // **The second provider** — A8, #1447.
  kiAnbieterZeile: 'Proveedor {anbieter}: {stand}',
  kiVorschlagAnbieter: 'Proveedor y modelo',
  kiVorschlagModellWaehlen: 'Elige uno',
  kiVorschlagOhneModell:
    'No hay ningún modelo disponible. Lo que falta aparece en «Modelos de IA».',
  kiVorschlagAntwortVon: 'Ha respondido {anbieter}/{modell}.',
  // **Measuring what a model can do** — A9, #1448.
  kiMessbarTitel: 'Modelos sin descripción',
  kiAdresseSocket:
    'Se consulta mediante un socket Unix propio de este programa, sin puerto.',
  kiGeraetSitzung: 'Consultar en esta sesión',
  kiGeraetSitzungHinweis:
    'Rige hasta el próximo reinicio; después vuelve a regir el ajuste «Modelo en el dispositivo de Apple».',
  kiFeldVorschlagen: 'Sugerir',
  kiFeldVorschlagenFuer: '{feld}: sugerencia',
  kiFeldAuftragName:
    'Sugiere un nombre propio adecuado para esta carta — no el tipo de carta ni una descripción de sus reglas. Responde en español, solo con el nombre.',
  kiFeldAuftragFlavor:
    'Sugiere un texto de ambientación breve para esta carta, una o dos frases: atmósfera o historia que no repita el texto de reglas. Responde en español.',
  kiFeldKarte: 'La carta hasta ahora:',
  kiFeldUebernehmen: 'Aplicar',
  kiFeldVon: 'de {anbieter} · {weg}',
  kiFeldWegGewaehlt: 'elegido',
  kiFeldWegVorrang: 'preferente',
  kiFeldVorrangFragt: 'La prioridad consulta: {reihe}',
  kiFeldWegRueckfall: 'alternativa',
  kiFeldAbgelehnt: 'El modelo se negó: {grund}',
  kiFeldNichts: 'No llegó ninguna sugerencia para este campo.',
  kiMessbarSatz:
    '{anbieter} indica estos modelos, pero no dice nada sobre ellos. Una medición envía una tarea pequeña con un esquema fijo: cuesta lo que cuesta esa única tarea.',
  kiMessen: 'Medir',
  kiMessungVergessen: 'Descartar medición',
  kiMessungLaeuft: 'Midiendo …',
  kiMessungGeglueckt:
    '{alias} ha respondido: {faehigkeiten} quedan ahora demostradas.',
  // **Selection, comparison and provenance** — A11, #1450.
  kiVorschlagAuswahl: 'Aplicar: {feld}',
  kiVorschlagBisher: 'Actualmente: {wert}',
  kiVorschlagBisherLeer: 'Actualmente vacío.',
  kiVorschlagNichtsGewaehlt:
    'No hay ningún campo marcado. Se aplica lo que esté marcado.',
  kiHerkunftTitel: 'De dónde proceden los valores',
  kiHerkunftZeile: '{feld}: {anbieter}/{modell} el {zeitpunkt}',
  kiHerkunftAufgabe: 'La tarea era: {aufgabe}',
  kiHerkunftVeraendert: 'Editado a mano desde entonces.',
  kiHerkunftKeine: 'Ningún campo de esta carta procede de una sugerencia.',
  // **The image proposal** — A12, #1451.
  kiBildTitel: 'Imagen de un modelo',
  kiBildAendertNichts:
    'Generar no cambia la carta. Lo que aparece es una propuesta con caducidad: la miras, la recortas y solo entonces la insertas.',
  kiBildFeld: 'Campo de imagen',
  kiBildAufgabe: '¿Qué debe mostrar?',
  kiBildAufgabeHinweis: 'Un rayo sobre una llanura extensa, óleo',
  kiBildGroesse: 'Tamaño',
  kiBildGuete: 'Calidad',
  kiBildBearbeiten: 'Editar la imagen existente en lugar de generar una nueva',
  kiBildErzeugen: 'Generar imagen',
  kiBildLaeuft:
    'La imagen se está creando en {anbieter}/{modell}; tarda de segundos a minutos.',
  kiBildAbbrechen: 'Cancelar',
  kiBildAbgebrochen:
    'Cancelado. No se aplicó nada; lo que costó hasta entonces figura en el registro de costes.',
  kiBildVorschau: 'Propuesta para {feld}',
  kiBildMasse: '{breite} × {hoehe} píxeles, {kilobytes} kB',
  kiBildAblauf: 'Válido hasta {ablauf}.',
  kiBildUmschrift: 'Lo que se dibujó: {umschrift}',
  kiBildNachweisTitel: 'De dónde procede esta imagen',
  kiBildAngabeGroesse: 'Tamaño',
  kiBildAngabeGuete: 'Calidad',
  kiBildAngabeUmschrift: 'Prompt revisado',
  kiBildAngabeMedientyp: 'Formato de archivo',
  kiBildUebernehmen: 'Recortar e insertar',
  kiBildVerwerfen: 'Descartar',
  kiBildVerworfen: 'La propuesta ya no existe.',
  kiBildKosten:
    'Un encargo de imagen se aprueba por separado: la aprobación para texto no lo cubre.',
  // **Approving a paid request** —
  // #1539.
  kiFreigabeTitel: 'Esta petición cuesta dinero',
  kiFreigabeSatz:
    'El modelo elegido se ejecuta en {anbieter}. Aprueba un importe máximo: vale para esta única petición y caduca a los cinco minutos.',
  kiFreigabeBetrag:
    'Importe máximo en la unidad más pequeña (céntimos para EUR)',
  kiFreigabeVerfuegbar: 'Disponible hoy: {betrag}',
  kiFreigabeOhneBudget:
    'No hay nada aprobado para {anbieter}. Sin un presupuesto aprobado no sale ninguna petición de pago; el límite se fija en «Costes».',
  kiFreigabeErteilen: 'Aprobar y consultar',
  kiFreigabeVerzichten: 'No aprobar',
  kiFreigabeLaeuft: 'Emitiendo la aprobación …',

  // **The evidence-backed assessment** — A13.2a
  // (#1452).
  kiEinschaetzungTitel: 'Pedir una valoración',
  kiEinschaetzungErklaerung:
    'Un modelo valora esta carta a partir de pasajes del índice de búsqueda local. No cambia nada: ni la carta ni la colección.',
  kiEinschaetzungArt: 'De qué se trata',
  kiEinschaetzungArtRegel: 'Según las reglas',
  kiEinschaetzungArtBalance: 'En comparación',
  kiEinschaetzungFrage: 'Qué se debe valorar',
  kiEinschaetzungFragen: 'Obtener valoración',
  kiEinschaetzungLaeuft: 'El modelo está leyendo las pruebas …',
  kiEinschaetzungLeer: 'Todavía no se ha valorado nada.',
  kiEinschaetzungBelegt: 'Todo lo mencionado aparece en las pruebas.',
  kiEinschaetzungUnbelegt:
    'Esta respuesta menciona algo que no figura en ninguna prueba. Lo que es aparece debajo.',
  kiEinschaetzungBefunde: 'Lo que no está respaldado',
  kiEinschaetzungEinschraenkungen: 'Lo que no se tuvo en cuenta',
  kiEinschaetzungBelege: 'Los pasajes aportados',
  kiEinschaetzungStaende: 'Estado de las fuentes',

  // **The balance analysis** — A13.2b
  // (#1452).
  kiEinschaetzungFormat: 'Formato',
  kiEinschaetzungOhneFormat: 'ninguno indicado',
  kiEinschaetzungVergleiche: 'Con qué se comparó',
  kiEinschaetzungStrategie: 'Encontrado mediante',
  kiEinschaetzungOhneVergleich:
    'Esta respuesta no nombra ninguna carta de comparación. Una afirmación sobre cómo se sitúa una carta es una comparación.',
  kiEinschaetzungSynergien: 'Qué sinergias quedaron fuera',

  // **The batch and the atomic group** — A13.1b
  // (#1452).
  kiBatchTitel: 'Varias cartas a la vez',
  kiBatchSatz:
    'Un modelo responde a la misma tarea para cada carta seleccionada. Todavía no se escribe nada: eso es el segundo paso.',
  kiBatchZiele: '{gewaehlt} de {gesamt} cartas seleccionadas',
  kiBatchOhneKarten: 'Este set todavía no tiene cartas.',
  kiBatchOhneName: '(sin nombre)',
  kiBatchAuftrag: 'Tarea para cada carta',
  kiBatchFahren: 'Obtener sugerencias',
  kiBatchLaeuft: 'En curso …',
  kiBatchLagen: 'En qué quedó',
  kiBatchGanzOderGar:
    'Se aplica todo a la vez o nada. Si falla un destino, cada carta se queda como estaba, y las sugerencias siguen abiertas.',
  kiBatchGruppe: 'Aplicar como grupo ({karten})',
  kiBatchGruppeLaeuft: 'Aplicando …',
  kiBatchGeschrieben: 'Se escribieron {karten} cartas.',

  // **Card recognition from a photo** — A13.3
  // (#1452).
  kiErkennungTitel: 'Reconocer una carta a partir de una foto',
  kiErkennungSatz:
    'Un modelo lee lo que está impreso en la foto. Qué carta es la decide la misma búsqueda de impresiones que se usa con una lista escrita, y no se añade nada a la colección.',
  kiErkennungLesen: 'Elegir una foto y leerla',
  kiErkennungLaeuft: 'El modelo está leyendo la foto …',
  kiErkennungLeer: 'Todavía no se ha leído ninguna foto.',
  kiErkennungGelesen: 'Lo que se leyó',
  kiErkennungNichtGelesen: 'Sobre lo que no hubo nada que decir',
  kiErkennungUnlesbar: 'ilegible',
  kiErkennungOhneWert: 'no se leyó nada',
  kiErkennungArten: {
    name: 'Nombre de la carta',
    setcode: 'Set',
    nummer: 'Número de coleccionista',
    typzeile: 'Línea de tipo',
    layout: 'Diseño',
    kunstwerk: 'Ilustración',
  } as Record<string, string>,
  kiErkennungKandidaten: 'Qué impresiones encajan',
  kiErkennungDrucke: 'Encajan {drucke} impresiones.',
  kiErkennungEinDruck: 'Encaja exactamente una impresión.',
  kiErkennungWeitere: 'Y hay más detrás: solo se contó la primera página.',
  kiErkennungOhneKandidat:
    'Ninguna impresión encaja con lo que se leyó. Puede deberse a la foto o a que la carta no figura aquí.',
  kiErkennungGestuetzt: 'Se apoya en: {arten}',
  kiErkennungOhneStuetze:
    'Nada de esta impresión se puede reencontrar en la foto: figura solo porque la búsqueda la encontró para el término.',
  kiErkennungNichtsAngelegt:
    'No se añade nada. Para registrar un ejemplar, usa el formulario de arriba; así queda claro a qué impresión se refiere.',
  kiErkennungStand: 'Antigüedad de la búsqueda de impresiones: {alter}',
  kiErkennungModell: 'Leído por {modell} en {anbieter}',

  // **The citation question** — A13.4
  // (#1452).
  kiVorschlagBelegfrage:
    'Incluir pasajes de la base de conocimiento (opcional)',
  kiVorschlagBelegfrageHinweis: 'qué buscar; vacío significa: no buscar',
} as const
