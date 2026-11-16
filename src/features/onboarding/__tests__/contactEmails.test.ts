import { describe, expect, it } from '@jest/globals';

import { contactEmails } from '../contactEmails';

describe('contactEmails', () => {
  it('normalises, dedupes and leaves out my own address', () => {
    const contacts = [
      { emails: [{ address: ' Aarav@Example.com ' }, { address: 'aarav@example.com' }] },
      { emails: [{ address: 'ME@example.com' }, { address: 'priya@example.in' }] },
      { emails: [] },
      {},
      { emails: [{}, { address: '' }, { address: 'not an email' }, { address: 'nodomain@' }] },
    ];
    expect(contactEmails(contacts, 'me@example.com')).toEqual(['aarav@example.com', 'priya@example.in']);
  });

  it('keeps everyone when I have no address yet', () => {
    expect(contactEmails([{ emails: [{ address: 'me@example.com' }] }])).toEqual(['me@example.com']);
  });
});
