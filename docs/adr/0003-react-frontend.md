# Use React for the application shell

**Status:** accepted

Moyan uses React + TypeScript + Vite for the application shell. Medict's Vue frontend is a source-project detail, not a product requirement; reusing its parser core does not justify carrying over its UI framework or feature coupling. React keeps the shell independently replaceable while matching the approved TypeScript requirement and the Wails v2 Vite integration.
