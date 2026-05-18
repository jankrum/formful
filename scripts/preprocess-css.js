#!/usr/bin/env node
// Strips RTL language selectors from style.css before PurgeCSS runs.
// PurgeCSS v8 can't parse :is(:lang(...)) and silently drops any rule using it,
// including the LTR variants we need. This script removes the wrappers so
// PurgeCSS sees plain selectors.
//
// LTR rule:  selector:not(:is(:lang(ae),...)){...}  → selector{...}
// RTL rule:  selector:is(:lang(ae),...){...}        → removed

const fs = require('fs')

const [,, src, dest] = process.argv
if (!src || !dest) {
  console.error('Usage: preprocess-css.js <input> <output>')
  process.exit(1)
}

const RTL = ':is(:lang(ae),:lang(ar),:lang(arc),:lang(bcc),:lang(bqi),:lang(ckb),:lang(dv),:lang(fa),:lang(glk),:lang(he),:lang(ku),:lang(mzn),:lang(nqo),:lang(pnb),:lang(ps),:lang(sd),:lang(ug),:lang(ur),:lang(yi))'
const LTR_WRAPPER = ':not(' + RTL + ')'

let css = fs.readFileSync(src, 'utf8')

// 1. Strip LTR wrappers — rules become universal (LTR is the default anyway)
css = css.replaceAll(LTR_WRAPPER, '')

// 2. Remove RTL-only rules — selector contains RTL pattern + {declarations}
const escaped = RTL.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
css = css.replace(new RegExp('[^{}]*' + escaped + '[^{]*\\{[^}]*\\}', 'g'), '')

fs.writeFileSync(dest, css)
console.log(`Preprocessed ${src} → ${dest}`)
