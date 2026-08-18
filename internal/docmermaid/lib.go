package docmermaid

import _ "embed"

// Mermaid 11.4.1 (jsDelivr dist/mermaid.min.js), embedded for offline article rendering.
//
//go:embed mermaid.min.js
var LibJS []byte
