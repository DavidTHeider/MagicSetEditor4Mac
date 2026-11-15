// **Appearance on the settings page** — UX-09 (#1681). Mirrors the German
// catalogue part of the same name.

export const gestaltungsnamen = {
  farbwelt: '配色',
  schriftbereich: '表示書体を適用する範囲',
  hintergrund: 'ウインドウの背景',
  gestaltung: 'カスタムカラー',
}

export const gestaltungswerte = {
  'farbwelt.pergament': '羊皮紙 — 温かみのある、繊維入り',
  'farbwelt.klar': 'クリア — ニュートラル、繊維なし',
  'farbwelt.schiefer': 'スレート — 涼しげなブルーグレー',
  'farbwelt.moos': 'モス — 落ち着いたグリーン',
  'schriftbereich.ueberschriften': '見出しのみ',
  'schriftbereich.ganz': 'インターフェイス全体',
  'hintergrund.aus': 'オフ',
  'hintergrund.fuellen': '塗りつぶし',
  'hintergrund.einpassen': '全体を表示',
  'hintergrund.kacheln': 'タイル状に並べる',
  'schriftart.buch': 'ブック — New York',
  'schriftart.rund': '丸み — SF Pro Rounded',
  'schriftart.serifenlos': 'サンセリフ — システム書体',
}

export const gestaltungshinweise = {
  farbwelt:
    'インターフェイスの色です。ライトとダークのどちらも、一つずつ計算し直しています。ハイコントラストのときは、ここでの選択にかかわらずインクの配色が適用されます。',
  schriftbereich:
    'インターフェイス全体に適用できるのは本文用の書体だけです。装飾書体は見出しにのみ使われます。カードにはテンプレートの書体がそのまま使われます。',
  hintergrund:
    '背景の下に、自分の画像を敷きます。文字が読みやすいように薄いベールを重ねます。画像は下で選択します。**オフ**にすると、すぐに外れます。',
  gestaltung:
    'インターフェイスの10個の項目に、1行に1つずつカスタムカラーを指定します。使えるのは16進数の色だけで、アドレスやフォントは指定できません。読みにくくなる色は適用されません。',
}

export const gestaltungsbeiwerk = {
  fensterbild: 'ウインドウの背景画像',
  vorschau: 'ウインドウの背景画像のプレビュー',
  keinBild:
    '自分の画像はまだありません。画像を選択すると、ここに表示され、ウインドウの背景がオフでなければインターフェイスの背後にも表示されます。',
  aus: '画像は用意できていますが、ウインドウの背景が**オフ**になっています。',
  waehlen: '画像を選択…',
  entfernen: '画像を削除',
  merkmale: 'カスタムカラー',
  merkmaleHilfe:
    '1行に1項目、たとえば「--mse-marke: #4e2f16;」のように入力します。使用できる項目は {merkmale} です。',
  uebernehmen: '適用',
  zeilenfehler:
    '{zeile} 行目：「--mse-grund: #1e1b16;」のような形式で、使用できる項目と16進数の色を指定してください。',
  nichtLesbar: '適用されませんでした。読みにくくなります：',
  paar: '{hinten}上の{vorn}：{verhaeltnis} : 1（必要な値：{schwelle} : 1）',
  paarnamen: {
    'status:warnung': '警告',
    'status:erfolg': '成功メッセージ',
    'fensterbild:dunkel': '暗いウインドウ背景画像の上の --mse-grund',
    'fensterbild:hell': '明るいウインドウ背景画像の上の --mse-grund',
  } as Record<string, string>,
  angewandt: '適用しました。',
  ruhtBeiKontrast:
    '保存しました。ハイコントラストがオンの間、カスタムカラーは適用されません。オフにすると再び適用されます。',
  schmuckNurOben: '選択した表示書体は装飾書体のため、見出しにのみ使われます。',
  zuruecksetzenGescheitert: '外観をリセットできませんでした：',
  zurVorgabe: 'カスタムカラーを初期設定に戻す',
}
