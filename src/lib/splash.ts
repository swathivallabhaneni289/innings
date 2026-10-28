// Whether the launch splash has been hidden, so the first screen can hold its entrance until it can be seen.
let hidden = false;

export function markSplashHidden() {
  hidden = true;
}

/** How long a screen mounting now should hold its entrance: long enough for the splash to fade, or 0 once it's gone. */
export function entranceLead() {
  return hidden ? 0 : 450;
}
