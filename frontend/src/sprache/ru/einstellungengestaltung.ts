// **Appearance on the settings page** — UX-09 (#1681). Mirrors the German
// catalogue part of the same name.

export const gestaltungsnamen = {
  farbwelt: 'Цветовая схема',
  schriftbereich: 'Декоративный шрифт применяется к',
  hintergrund: 'Фон окна',
  gestaltung: 'Собственные цвета',
}

export const gestaltungswerte = {
  'farbwelt.pergament': 'Пергамент — тёплые тона, с волокнами',
  'farbwelt.klar': 'Чистая — нейтральные тона, без волокон',
  'farbwelt.schiefer': 'Сланец — холодный серо-голубой',
  'farbwelt.moos': 'Мох — приглушённый зелёный',
  'schriftbereich.ueberschriften': 'Только заголовки',
  'schriftbereich.ganz': 'Весь интерфейс',
  'hintergrund.aus': 'Выкл.',
  'hintergrund.fuellen': 'Заполнить',
  'hintergrund.einpassen': 'Вписать',
  'hintergrund.kacheln': 'Замостить',
  'schriftart.buch': 'Книжный — New York',
  'schriftart.rund': 'Округлый — SF Pro Rounded',
  'schriftart.serifenlos': 'Без засечек — системный шрифт',
}

export const gestaltungshinweise = {
  farbwelt:
    'Цвета интерфейса, светлые и тёмные, каждый пересчитан. При высокой контрастности действуют цвета «чернил», что бы здесь ни было выбрано.',
  schriftbereich:
    'Ко всему интерфейсу применяется только текстовый шрифт; декоративный шрифт остаётся у заголовков. Карты сохраняют шрифты своего шаблона.',
  hintergrund:
    'Помещает твоё собственное изображение под фон, под вуалью, которая сохраняет читаемость текста. Изображение выбирается ниже; **Выкл.** убирает его сразу.',
  gestaltung:
    'Собственные цвета для десяти свойств интерфейса, по одному в строке. Только цвета в шестнадцатеричной записи — ни адресов, ни шрифтов. То, что было бы нечитаемо, не применяется.',
}

export const gestaltungsbeiwerk = {
  fensterbild: 'Твой фон окна',
  vorschau: 'Предпросмотр фона окна',
  keinBild:
    'Собственного изображения пока нет. Выбери его — оно появится здесь и, как только фон окна не выключен, за интерфейсом.',
  aus: 'Изображение готово, но для фона окна выбрано значение **Выкл.**',
  waehlen: 'Выбрать изображение …',
  entfernen: 'Удалить изображение',
  merkmale: 'Собственные цвета',
  merkmaleHilfe:
    'По одной записи в строке, например «--mse-marke: #4e2f16;». Допустимы {merkmale}.',
  uebernehmen: 'Применить',
  zeilenfehler:
    'Строка {zeile}: ожидается запись вида «--mse-grund: #1e1b16;» — одно из допустимых свойств и цвет в шестнадцатеричной записи.',
  nichtLesbar: 'Не применено — так текст был бы нечитаем:',
  paar: '{vorn} на {hinten}: {verhaeltnis} : 1, требуется {schwelle} : 1',
  paarnamen: {
    'status:warnung': 'Предупреждения',
    'status:erfolg': 'Сообщения об успехе',
    'fensterbild:dunkel': '--mse-grund над тёмным фоном окна',
    'fensterbild:hell': '--mse-grund над светлым фоном окна',
  } as Record<string, string>,
  angewandt: 'Применено.',
  ruhtBeiKontrast:
    'Сохранено — при высокой контрастности собственные цвета не действуют; они вернутся, как только она будет выключена.',
  schmuckNurOben: 'Выбранный декоративный шрифт остаётся только у заголовков.',
  zuruecksetzenGescheitert: 'Не удалось сбросить оформление:',
  zurVorgabe: 'Сбросить собственные цвета до значений по умолчанию',
}
