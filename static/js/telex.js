/**
 * telex.js – Real-time Telex → Vietnamese Unicode conversion.
 *
 * Telex rules (subset used in Vietnam):
 *   Tone marks (applied to the last vowel in the syllable):
 *     s → sắc (´)    e.g.  a+s = á
 *     f → huyền (`)  e.g.  a+f = à
 *     r → hỏi (?)    e.g.  a+r = ả
 *     x → ngã (~)    e.g.  a+x = ã
 *     j → nặng (.)   e.g.  a+j = ạ
 *
 *   Vowel modifiers:
 *     aa → â,  aw → ă
 *     ee → ê
 *     oo → ô,  ow → ơ
 *     uw → ư,  uw → ư
 *     dd → đ
 *
 * This implementation converts the full value of the textarea whenever a key
 * is released, so it stays in sync even after the user repositions the cursor.
 */

// ---------------------------------------------------------------------------
// Data tables
// ---------------------------------------------------------------------------

const VOWEL_MAP = {
  aa: 'â', aw: 'ă',
  ee: 'ê',
  oo: 'ô', ow: 'ơ',
  uw: 'ư',
  dd: 'đ',
};

// Base vowels that accept tone marks
const BASE_VOWELS = 'aeiouy';

// Precomposed tone tables for each base vowel (order: sắc, huyền, hỏi, ngã, nặng)
const TONE_MAP = {
  a: { s: 'á', f: 'à', r: 'ả', x: 'ã', j: 'ạ' },
  â: { s: 'ấ', f: 'ầ', r: 'ẩ', x: 'ẫ', j: 'ậ' },
  ă: { s: 'ắ', f: 'ằ', r: 'ẳ', x: 'ẵ', j: 'ặ' },
  e: { s: 'é', f: 'è', r: 'ẻ', x: 'ẽ', j: 'ẹ' },
  ê: { s: 'ế', f: 'ề', r: 'ể', x: 'ễ', j: 'ệ' },
  i: { s: 'í', f: 'ì', r: 'ỉ', x: 'ĩ', j: 'ị' },
  o: { s: 'ó', f: 'ò', r: 'ỏ', x: 'õ', j: 'ọ' },
  ô: { s: 'ố', f: 'ồ', r: 'ổ', x: 'ỗ', j: 'ộ' },
  ơ: { s: 'ớ', f: 'ờ', r: 'ở', x: 'ỡ', j: 'ợ' },
  u: { s: 'ú', f: 'ù', r: 'ủ', x: 'ũ', j: 'ụ' },
  ư: { s: 'ứ', f: 'ừ', r: 'ử', x: 'ữ', j: 'ự' },
  y: { s: 'ý', f: 'ỳ', r: 'ỷ', x: 'ỹ', j: 'ỵ' },
};

// Reverse: accented vowel → { base, tone }
const REVERSE_TONE = {};
for (const [base, tones] of Object.entries(TONE_MAP)) {
  for (const [tone, ch] of Object.entries(tones)) {
    REVERSE_TONE[ch] = { base, tone };
  }
}

const TONE_KEYS = new Set(['s', 'f', 'r', 'x', 'j']);

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

/** Apply vowel-modifier digraphs (aa→â, aw→ă, …, dd→đ). */
function applyVowelModifiers(text) {
  // Sort keys longest-first so "aw" isn't split before being checked.
  const keys = Object.keys(VOWEL_MAP).sort((a, b) => b.length - a.length);
  let result = text;
  for (const digraph of keys) {
    result = result.split(digraph).join(VOWEL_MAP[digraph]);
  }
  return result;
}

/**
 * Apply a tone key to the most recent eligible vowel in a word segment.
 * Returns the modified string, or null if no vowel was found.
 */
function applyTone(word, toneKey) {
  // Walk backwards and find the last vowel character that has a tone entry
  for (let i = word.length - 1; i >= 0; i--) {
    const ch = word[i];
    if (TONE_MAP[ch]) {
      const toned = TONE_MAP[ch][toneKey];
      if (toned) {
        return word.slice(0, i) + toned + word.slice(i + 1);
      }
    }
    // If this character is a consonant, keep searching backwards through the vowel cluster
    if (BASE_VOWELS.indexOf(ch) === -1 && !REVERSE_TONE[ch] && !VOWEL_MAP[ch]) {
      // hit a consonant after some vowels – stop searching
      if (i < word.length - 1) break;
    }
  }
  return null;
}

/**
 * Convert a single "word" token using Telex rules.
 * A word ends at a space, punctuation, or string boundary.
 */
function convertWord(raw) {
  let text = raw;

  // 1. Apply vowel modifiers
  text = applyVowelModifiers(text);

  // 2. Check if the last character is a tone key
  const last = text[text.length - 1];
  if (TONE_KEYS.has(last)) {
    const withoutTone = text.slice(0, -1);
    const toned = applyTone(withoutTone, last);
    if (toned !== null) {
      return toned;
    }
  }

  return text;
}

/**
 * Convert an entire buffer (potentially multi-word) through Telex rules.
 * We split on whitespace/punctuation, convert each token, then rejoin.
 */
function telexConvert(raw) {
  return raw.replace(/[^\s.,!?;:()\-"]+/g, convertWord);
}

// ---------------------------------------------------------------------------
// DOM integration
// ---------------------------------------------------------------------------

/**
 * Attach Telex processing to a textarea element.
 * @param {HTMLTextAreaElement} textarea
 */
function attachTelex(textarea) {
  textarea.addEventListener('input', () => {
    const { selectionStart, selectionEnd } = textarea;
    const converted = telexConvert(textarea.value);
    if (converted !== textarea.value) {
      textarea.value = converted;
      // Restore caret: conversion may shrink the string (digraphs)
      const delta = textarea.value.length - converted.length;
      textarea.setSelectionRange(selectionStart - delta, selectionEnd - delta);
    }
  });
}

export { attachTelex, telexConvert };
