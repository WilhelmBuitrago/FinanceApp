import { describe, expect, it } from 'vitest';

import type { ReviewItem } from '../api/review';
import { payloadOf } from './reviewPayload';

const item = (payload?: string): ReviewItem => ({
  ID: 'rv_1',
  Kind: 'EXTRACTION',
  Status: 'PENDING',
  CreatedAt: '2026-10-24T00:00:00Z',
  PayloadJson: payload,
});

describe('payloadOf', () => {
  it('parses the parked payload JSON', () => {
    const p = payloadOf(item('{"reason":"no amount","movement_id":"mv_1"}'));
    expect(p.reason).toBe('no amount');
    expect(p.movement_id).toBe('mv_1');
  });

  it('returns an empty object when there is no payload', () => {
    expect(payloadOf(item(undefined))).toEqual({});
    expect(payloadOf(item(''))).toEqual({});
  });

  it('never throws on malformed JSON: one bad payload must not break the queue', () => {
    expect(payloadOf(item('{not json'))).toEqual({});
  });
});
