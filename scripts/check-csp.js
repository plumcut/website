#!/usr/bin/env node
/**
 * Keeps the careers pages' Content-Security-Policy in vercel.json in step with
 * the pages themselves.
 *
 * The CSP allows inline scripts only by sha256 hash, so editing any inline
 * <script> on careers/sales.html or careers/sales/thank-you.html (or regenerating
 * them from a changed template) silently blocks that script in the browser
 * until the hash list is updated. This check recomputes the hashes and fails
 * when vercel.json does not carry exactly the policy below.
 *
 *   node scripts/check-csp.js          # verify, exit 1 on drift
 *   node scripts/check-csp.js --write  # rewrite the policy value in vercel.json
 *
 * It also fails on inline event handlers (onload=, onclick=...), which a
 * hash-based CSP blocks, and on scripts loaded from hosts the policy omits.
 * Zero dependencies.
 */
'use strict';

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');

const ROOT = path.resolve(__dirname, '..');
const PAGES = ['careers/sales.html', 'careers/sales/thank-you.html'];
const SOURCES = ['/careers/sales', '/careers/sales/thank-you'];

// GA4 is the only third party these pages load (main.js injects gtag.js).
const GA_SCRIPT = ['https://www.googletagmanager.com'];
const GA_CONNECT = ['https://*.google-analytics.com', 'https://*.analytics.google.com', 'https://*.googletagmanager.com'];
const GA_IMG = ['https://*.google-analytics.com', 'https://*.googletagmanager.com'];

function inlineScripts(html) {
  const out = [];
  for (const m of html.matchAll(/<script(\s[^>]*)?>([\s\S]*?)<\/script>/gi)) {
    const attrs = m[1] || '';
    if (/\ssrc\s*=/i.test(attrs)) continue;
    const type = (attrs.match(/\stype\s*=\s*"([^"]*)"/i) || [])[1];
    if (type && !/^(text|application)\/(java|ecma)script$|^module$/i.test(type)) continue; // JSON-LD is data, not script
    out.push(m[2]);
  }
  return out;
}

function policy(hashes) {
  return [
    "default-src 'self'",
    ['script-src', "'self'", ...hashes.map((h) => `'sha256-${h}'`), ...GA_SCRIPT].join(' '),
    // The shared template uses style="" attributes and <style> blocks throughout.
    "style-src 'self' 'unsafe-inline'",
    ['img-src', "'self'", 'data:', ...GA_IMG].join(' '),
    ['connect-src', "'self'", ...GA_CONNECT].join(' '),
    "font-src 'self' data:", // main.css embeds an icon font as a data: URL
    "media-src 'self'",
    "object-src 'none'",
    "base-uri 'none'",
    "form-action 'self'",
    "frame-ancestors 'none'",
    'upgrade-insecure-requests',
  ].join('; ');
}

function expected() {
  const errors = [];
  const hashes = new Set();
  for (const rel of PAGES) {
    const file = path.join(ROOT, rel);
    if (!fs.existsSync(file)) {
      errors.push(`${rel}: missing`);
      continue;
    }
    const html = fs.readFileSync(file, 'utf8');
    for (const s of inlineScripts(html)) hashes.add(crypto.createHash('sha256').update(s, 'utf8').digest('base64'));
    const handler = html.match(/<[^>]+\son[a-z]+\s*=\s*["'][^>]*>/i);
    if (handler) errors.push(`${rel}: inline event handler is blocked by the CSP: ${handler[0].slice(0, 80)}`);
    for (const m of html.matchAll(/<script[^>]*\ssrc\s*=\s*"([^"]+)"/gi)) {
      const src = m[1];
      if (src.startsWith('/') && !src.startsWith('//')) continue;
      if (!GA_SCRIPT.some((h) => src.startsWith(h + '/'))) errors.push(`${rel}: script from a host the CSP does not allow: ${src}`);
    }
  }
  return { value: policy([...hashes].sort()), errors };
}

function cspEntries(config) {
  return (config.headers || [])
    .filter((h) => SOURCES.includes(h.source))
    .map((h) => ({ source: h.source, header: (h.headers || []).find((x) => x.key === 'Content-Security-Policy') }));
}

function main() {
  const write = process.argv.includes('--write');
  const vercelPath = path.join(ROOT, 'vercel.json');
  const raw = fs.readFileSync(vercelPath, 'utf8');
  const config = JSON.parse(raw);
  const { value, errors } = expected();
  const entries = cspEntries(config);

  for (const src of SOURCES) {
    const e = entries.find((x) => x.source === src);
    if (!e || !e.header) errors.push(`vercel.json: no Content-Security-Policy header for ${src}`);
  }

  if (write && !errors.length) {
    // Replace only the policy strings, so the rest of vercel.json keeps its
    // hand-written formatting.
    let next = raw;
    for (const old of new Set(entries.map((e) => e.header.value))) {
      next = next.split(JSON.stringify(old)).join(JSON.stringify(value));
    }
    if (next !== raw) fs.writeFileSync(vercelPath, next);
    console.log(next === raw ? 'check-csp: vercel.json already up to date' : 'check-csp: vercel.json updated');
    return;
  }

  for (const e of entries) {
    if (e.header && e.header.value !== value) {
      errors.push(`vercel.json: CSP for ${e.source} is out of date. Run: node scripts/check-csp.js --write`);
    }
  }
  if (errors.length) {
    for (const e of errors) console.error('  x ' + e);
    console.error(`[check-csp] ${errors.length} error(s).`);
    process.exit(1);
  }
  console.log('[check-csp] careers CSP matches the pages.');
}

if (require.main === module) main();

module.exports = { inlineScripts, policy, expected };
