import { FormEvent, KeyboardEvent, useEffect, useRef, useState } from 'react'
import { chooseAndOpenDictionary, DictionaryEntry, lookupWord, searchCandidates } from './api'
import { headwordOnSubmit } from './search'

type Entry = DictionaryEntry & {
  pronunciation?: string
}

const suggestionLimit = 8
const shortcutLabel = typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform) ? '⌘K' : 'Ctrl K'

export function App() {
  const [query, setQuery] = useState('')
  const [submittedQuery, setSubmittedQuery] = useState('')
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [dictionaryName, setDictionaryName] = useState<string | null>(null)
  const [history, setHistory] = useState<string[]>([])
  const [recentOpen, setRecentOpen] = useState(true)
  const [result, setResult] = useState<Entry | null>(null)
  const [candidates, setCandidates] = useState<string[]>([])
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [activeSuggestion, setActiveSuggestion] = useState(-1)
  const [suggestOpen, setSuggestOpen] = useState(false)
  const [isSearching, setIsSearching] = useState(false)
  const [isImporting, setIsImporting] = useState(false)
  const [importError, setImportError] = useState('')
  const searchInputRef = useRef<HTMLInputElement>(null)
  const suggestRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!submittedQuery) {
      setResult(null)
      setCandidates([])
      setIsSearching(false)
      return
    }
    let cancelled = false
    setIsSearching(true)
    setCandidates([])
    async function search() {
      try {
        const entry = await lookupWord(submittedQuery)
        if (cancelled) return
        setResult(entry)
        if (!entry) {
          setCandidates(await searchCandidates(submittedQuery))
        }
      } catch {
        if (!cancelled) {
          setResult(null)
          setCandidates(await searchCandidates(submittedQuery))
        }
      } finally {
        if (!cancelled) setIsSearching(false)
      }
    }
    void search()
    return () => { cancelled = true }
  }, [submittedQuery, dictionaryName])

  useEffect(() => {
    const value = query.trim()
    if (!value || !dictionaryName) {
      setSuggestions([])
      setActiveSuggestion(-1)
      return
    }
    let cancelled = false
    const handle = window.setTimeout(async () => {
      const found = await searchCandidates(value)
      if (cancelled) return
      setSuggestions(found.slice(0, suggestionLimit))
      setActiveSuggestion(-1)
    }, 180)
    return () => {
      cancelled = true
      window.clearTimeout(handle)
    }
  }, [query, dictionaryName])

  useEffect(() => {
    if (dictionaryName) searchInputRef.current?.focus()
  }, [dictionaryName])

  useEffect(() => {
    function handleShortcut(event: globalThis.KeyboardEvent) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        searchInputRef.current?.focus()
        searchInputRef.current?.select()
        setSuggestOpen(true)
      }
    }
    window.addEventListener('keydown', handleShortcut)
    return () => window.removeEventListener('keydown', handleShortcut)
  }, [])

  useEffect(() => {
    function handlePointer(event: PointerEvent) {
      const target = event.target as Node
      if (suggestRef.current?.contains(target) || searchInputRef.current?.contains(target)) return
      setSuggestOpen(false)
    }
    window.addEventListener('pointerdown', handlePointer)
    return () => window.removeEventListener('pointerdown', handlePointer)
  }, [])

  async function importDictionary() {
    setIsImporting(true)
    setImportError('')
    try {
      const name = await chooseAndOpenDictionary()
      if (!name) return
      setDictionaryName(name)
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error)
      setImportError(`Couldn't import: ${message}`)
    } finally {
      setIsImporting(false)
    }
  }

  function lookUp(word: string, fromRecent = false) {
    const value = word.trim()
    if (!value || !dictionaryName) return
    setQuery(value)
    setSubmittedQuery(value)
    if (!fromRecent) {
      setHistory((current) => current.includes(value) ? current : [value, ...current].slice(0, 12))
    }
    setSuggestions([])
    setActiveSuggestion(-1)
    setSuggestOpen(false)
  }

  function submitSearch(event: FormEvent) {
    event.preventDefault()
    lookUp(headwordOnSubmit(query, suggestions, activeSuggestion))
  }

  function onSearchKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (!suggestOpen || suggestions.length === 0) {
      if (event.key === 'Escape') searchInputRef.current?.blur()
      return
    }
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setActiveSuggestion((current) => (current + 1) % suggestions.length)
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActiveSuggestion((current) => (current <= 0 ? suggestions.length - 1 : current - 1))
    } else if (event.key === 'Escape') {
      event.preventDefault()
      setSuggestOpen(false)
      setActiveSuggestion(-1)
    }
  }

  const showSuggestions = suggestOpen && suggestions.length > 0
  const pane = !dictionaryName
    ? 'welcome'
    : isSearching
      ? 'searching'
      : result
        ? 'entry'
        : submittedQuery
          ? 'miss'
          : 'ready'

  const panelTitle = pane === 'welcome'
    ? 'Reader'
    : pane === 'searching'
      ? submittedQuery
      : pane === 'miss'
        ? submittedQuery
        : result
          ? result.Headword
          : 'Reader'
  const panelMeta = pane === 'welcome' ? 'No dictionary' : dictionaryName

  return (
    <main className={`app-shell${sidebarOpen ? '' : ' is-collapsed'}`}>
      {sidebarOpen && (
        <aside className="sidebar" aria-label="Library and history">
          <header className="sidebar-brand">
            <button className="wordmark" type="button" onClick={() => setSidebarOpen(false)} aria-label="Hide library">
              Moyan
            </button>
          </header>

          <nav className="sidebar-nav">
            <section className="nav-section">
              <div className="nav-label">
                <span>Library</span>
                <button className="nav-action" type="button" onClick={importDictionary} disabled={isImporting}>
                  {isImporting ? '…' : dictionaryName ? 'Replace' : 'Import'}
                </button>
              </div>
              {dictionaryName ? (
                <div className="nav-row is-current">
                  <span className="nav-dot" aria-hidden="true" />
                  <span className="nav-title">{dictionaryName}</span>
                </div>
              ) : (
                <p className="nav-empty">None open</p>
              )}
              {importError && <p className="status-message" role="alert">{importError}</p>}
            </section>

            <section className="nav-section nav-grow">
              <div className="nav-label">
                <span>Recent</span>
                <button
                  className="nav-action"
                  type="button"
                  aria-expanded={recentOpen}
                  onClick={() => setRecentOpen((open) => !open)}
                >
                  {recentOpen ? 'Hide' : 'Show'}
                </button>
              </div>
              {recentOpen && (history.length > 0 ? (
                <ul className="nav-list">
                  {history.map((item) => (
                    <li key={item}>
                      <button
                        className="nav-row"
                        type="button"
                        onClick={() => lookUp(item, true)}
                      >
                        <span className="nav-dot" aria-hidden="true" />
                        <span className="nav-title">{item}</span>
                      </button>
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="nav-empty">No lookups yet</p>
              ))}
            </section>
          </nav>
        </aside>
      )}

      <div className="workspace">
        <header className="workspace-bar">
          {!sidebarOpen && (
            <button className="wordmark compact" type="button" onClick={() => setSidebarOpen(true)} aria-label="Show library">
              Moyan
            </button>
          )}
          <form className="search-form" role="search" onSubmit={submitSearch}>
            <div className="search-well" ref={suggestRef}>
              <input
                id="word-search"
                ref={searchInputRef}
                value={query}
                onChange={(event) => {
                  setQuery(event.target.value)
                  setSuggestOpen(true)
                }}
                onFocus={() => setSuggestOpen(true)}
                onKeyDown={onSearchKeyDown}
                placeholder={dictionaryName ? 'Look up a word' : 'Import a dictionary first'}
                autoComplete="off"
                spellCheck={false}
                role="combobox"
                aria-label="Look up a word"
                aria-autocomplete="list"
                aria-expanded={showSuggestions}
                disabled={!dictionaryName}
                aria-controls="word-suggestions"
                aria-activedescendant={activeSuggestion >= 0 ? `suggestion-${activeSuggestion}` : undefined}
              />
              {showSuggestions && (
                <ul className="suggestion-list" id="word-suggestions" role="listbox">
                  {suggestions.map((item, index) => (
                    <li key={item} role="presentation">
                      <button
                        id={`suggestion-${index}`}
                        className={`suggestion-item${index === activeSuggestion ? ' is-active' : ''}`}
                        type="button"
                        role="option"
                        aria-selected={index === activeSuggestion}
                        onMouseEnter={() => setActiveSuggestion(index)}
                        onClick={() => lookUp(item)}
                      >
                        {item}
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
            <button className="search-go" type="submit" disabled={!dictionaryName || isSearching || !query.trim()}>
              Search
            </button>
          </form>
        </header>

        <section className="panel" aria-live="polite" aria-busy={isSearching}>
          <header className="panel-bar">
            <span className="nav-dot is-on" aria-hidden="true" />
            <h1 className="panel-title">{panelTitle}</h1>
            {panelMeta && <span className="panel-meta">{panelMeta}</span>}
          </header>
          <div className="panel-body">
            {pane === 'welcome' && (
              <div className="empty-state">
                <p className="empty-copy">Choose a local MDX file. Lookup stays on this computer.</p>
                <button className="primary-button" type="button" onClick={importDictionary} disabled={isImporting}>
                  {isImporting ? 'Opening…' : 'Import dictionary'}
                </button>
              </div>
            )}

            {pane === 'ready' && (
              <div className="empty-state">
                <p className="empty-copy">Type a headword in the bar above. {shortcutLabel} focuses search.</p>
              </div>
            )}

            {pane === 'searching' && !result && (
              <div className="empty-state">
                <p className="empty-copy">Looking up “{submittedQuery}”…</p>
              </div>
            )}

            {pane === 'miss' && (
              <div className="empty-state">
                <p className="empty-copy">No entry for “{submittedQuery}”.</p>
                {candidates.length > 0 ? (
                  <ul className="candidate-list">
                    {candidates.map((candidate) => (
                      <li key={candidate}>
                        <button type="button" onClick={() => lookUp(candidate)}>{candidate}</button>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className="empty-copy">Try another spelling, or replace the dictionary from the library.</p>
                )}
              </div>
            )}

            {result && (
              <article className="entry">
                {result.pronunciation && <p className="pronunciation">{result.pronunciation}</p>}
                <DictionaryFrame key={`${result.Dictionary}-${result.Headword}`} html={result.HTML} css={result.CSS} />
              </article>
            )}
          </div>
        </section>
      </div>
    </main>
  )
}

function DictionaryFrame({ html, css }: { html: string; css: string }) {
  const frameRef = useRef<HTMLIFrameElement>(null)
  const [height, setHeight] = useState(120)
  const safeCss = css.replace(/<\/style/gi, '<\\/style')
  const srcDoc = `<!doctype html><html><head><meta charset="utf-8"><meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; img-src data:; media-src data:; font-src data:"><style>
html, body { margin: 0; padding: 0; background: transparent; }
body { overflow: visible; line-height: 1.7; overflow-wrap: anywhere; }
img, svg, video, canvas { max-width: 100%; height: auto; }
${safeCss}
</style></head><body>${html}</body></html>`

  function measure() {
    const doc = frameRef.current?.contentDocument
    if (!doc?.documentElement) return
    const next = Math.max(doc.documentElement.scrollHeight, doc.body?.scrollHeight ?? 0, 80)
    setHeight(next)
  }

  return (
    <iframe
      ref={frameRef}
      className="dictionary-frame"
      title="Dictionary entry"
      sandbox="allow-same-origin"
      srcDoc={srcDoc}
      height={height}
      onLoad={() => {
        measure()
        const doc = frameRef.current?.contentDocument
        if (!doc) return
        doc.querySelectorAll('img').forEach((img) => {
          if (!img.complete) img.addEventListener('load', measure)
        })
      }}
    />
  )
}
