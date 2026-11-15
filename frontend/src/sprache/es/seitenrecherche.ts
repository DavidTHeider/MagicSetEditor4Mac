// The research texts — work package A16 of phase 5.3
// (#1515).
export const recherche = {
  rechercheTitel: 'Consultar',
  rechercheErklaerung:
    'Consulta textos de cartas, rulings y las reglas en este equipo: texto completo y vectores combinados. El modelo de embeddings participa, en local; no se consulta ningún modelo que escriba texto. Nada sale de este equipo y nada cambia.',

  rechercheFall: 'Qué buscar',
  rechercheFallKarten: 'Cartas',
  rechercheFallRegeln: 'Reglas',
  rechercheFallRulings: 'Rulings',

  rechercheFrage: 'Pregunta',
  rechercheSuchen: 'Consultar',
  rechercheSprache: 'Idioma',
  rechercheAlleSprachen: 'Cualquiera',

  rechercheModus: 'Búsqueda',
  rechercheModusHybrid: 'texto completo y vectores combinados',
  rechercheModusLexikalisch: 'solo texto completo: sin modelo de embeddings',
  rechercheModusOhneIndex: 'ninguna: no hay índice de búsqueda',

  rechercheLeereNichtsGefunden: '¿Probar con otra palabra?',
  rechercheLeereOhneIndex: 'Se puede crear uno en «Ajustes › Diagnóstico».',
  rechercheLeereOhneModell:
    'Sin modelo de embeddings, la búsqueda encuentra palabras, no similitud.',
  rechercheLeereOhneQuelle:
    'Esta pregunta necesita una fuente que no está en el índice.',
  rechercheLeereGesperrt: 'No se ha concedido el acceso a esta fuente.',
  rechercheLeereBudget: 'Esta ejecución alcanzó su límite.',

  rechercheHerkunft:
    'Revisión {revision}, válida {datum}, obtenida {abgerufen}',
  rechercheWeitere: '{anzahl} secciones más de la misma carta',
  rechercheOeffnen: 'Abrir carta',
  rechercheEinschraenkung: 'No realizado',
  rechercheStaende: 'Revisiones',
  rechercheIndex: 'Índice de búsqueda',
  rechercheMehr: 'Cargar más',

  aehnlichTitel: 'Cartas similares',
  aehnlichErklaerung:
    'Busca con el texto de esta carta, no con su nombre. La propia carta no aparece en la lista.',
  aehnlichSuchen: 'Buscar similares',
  aehnlichLeer: 'Todavía no se ha buscado nada.',

  freigabeTitel: 'Lo que recibiría un modelo',
  freigabeErklaerung:
    'Reúne lo que vería un modelo, y no lo envía. Quien genere después sabrá de antemano con qué.',
  freigabeAufgabe: 'Tarea',
  freigabeAnbieter: 'Proveedor',
  freigabeModell: 'Modelo',
  freigabeFenster: 'Ventana de contexto',
  freigabeReserve: 'Reserva de salida',
  freigabeZeigen: 'Reunir',
  freigabeOrtLokal: 'Se calcula en este equipo.',
  freigabeOrtCloud: 'Se calcula en un proveedor externo.',
  freigabeBudget: '{verbraucht} de {grenze} tokens, {reserve} reservados',
  freigabeZaehlweise: 'estimado según {zaehlweise}',
  freigabeZieladresse: 'Dirección de destino',
  freigabeKategorien: 'Categorías de datos',
  freigabePrivat: 'Se incluye contenido propio.',
  freigabeOeffentlich: 'Solo fuentes públicas.',
  freigabeBelege: 'Citas',
  freigabeKeineBelege: 'Sin citas: reunido así, un modelo no vería nada.',
  freigabeVeraltet: 'No es la revisión activa',
  freigabeLaeuftAb: 'Válido hasta {zeit}',

  // **Releasing your own collection** — A13.4
  // (#1452).
  freigabeEigeneKarten: 'Buscar también en mis propias cartas',
  freigabeBestandErklaerung:
    'Sin marcar esta casilla, tus propias cartas quedan fuera: la búsqueda cubre entonces solo fuentes públicas, y la respuesta lo indica. Lo que se incluya sale de este equipo si el modelo se ejecuta en otro sitio.',
}
