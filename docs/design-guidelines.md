# Moyan Design Guidelines

## Visual direction

Moyan is a restrained dictionary reader with a warm editorial character: quiet, comfortable, and focused on the word being read.

- Canvas: warm parchment `#f5f4ed`.
- Accent: ink blue `#1B365D` as the only saturated accent.
- Supporting colors: warm grays with a yellow-brown undertone.
- No gradients, drop shadows, 3D treatments, decorative color systems, or stock imagery.
- Use generous whitespace and a tight editorial rhythm rather than dashboard density.

## Typography

Set the shell in Newsreader (self-hosted, variable weight). Chinese and other CJK copy falls back to a system serif so Library names stay in the same texture.

- Labels, controls, empty states, search, and the looked-up headword all use the Newsreader stack.
- Dictionary body copy stays under the imported package CSS.
- The UI ships in English. Dictionary entries keep the language of the imported package.

## Interaction

- The primary surface is a two-column layout: collapsible dictionary/search sidebar and reading pane.
- Motion is sparse and functional. Respect `prefers-reduced-motion` and provide no animation that competes with reading.
- Focus, hover, selected, loading, empty, and error states use tonal changes and thin rules rather than large shadows or bright fills.
- Do not invent a logo or substitute stock product imagery. Until a real mark exists, use the Moyan wordmark as text.

## Reading surface

- Preserve dictionary-authored CSS as far as the secure renderer permits.
- Support bounded, local CSS, images, fonts, and user-triggered audio from the MDX/MDD package and its same-directory static resources.
- Disable dictionary scripts, remote requests, video, plugins, and external navigation.
- The application shell may follow the system light/dark mode; dictionary content keeps its original palette rather than being force-inverted.
