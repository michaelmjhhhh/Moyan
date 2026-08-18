export type DictionaryEntry = {
  Headword: string
  HTML: string
}

const browserEntries: Record<string, DictionaryEntry> = {
  '你好': {
    Headword: '你好',
    HTML: '<p>hello</p><p>hi</p>',
  },
  hello: {
    Headword: 'hello',
    HTML: '<p>a greeting used when meeting someone</p>',
  },
}

type WailsApp = {
  LookupWord(word: string): Promise<DictionaryEntry>
}

declare global {
  interface Window {
    go?: { main?: { App?: WailsApp } }
  }
}

export async function lookupWord(word: string): Promise<DictionaryEntry | null> {
  const app = window.go?.main?.App
  if (app) {
    return app.LookupWord(word)
  }
  return browserEntries[word.trim()] ?? null
}
