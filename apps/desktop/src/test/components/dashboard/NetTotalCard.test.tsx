import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { NetTotalCard } from '../../../components/dashboard/NetTotalCard';

/* The headline card: the real net over every account and its breakdown.
   Pure props, no state — the contract is the rendering of real sums. */

describe('NetTotalCard', () => {
  it('renders the signed net and every account breakdown row', () => {
    render(
      <NetTotalCard
        totalNet={125000}
        perAccount={[
          { id: 'acc-1', name: 'Nu', net: 150000 },
          { id: 'acc-2', name: 'Bancolombia', net: -25000 },
        ]}
      />,
    );

    expect(screen.getByText('+$1.250,00')).toBeTruthy();
    expect(screen.getByText('Nu')).toBeTruthy();
    expect(screen.getByText('+$1.500,00')).toBeTruthy();
    expect(screen.getByText('Bancolombia')).toBeTruthy();
    expect(screen.getByText('-$250,00')).toBeTruthy();
  });

  it('renders the headline alone when there is no breakdown', () => {
    render(<NetTotalCard totalNet={0} perAccount={[]} />);

    expect(screen.getByText('$0,00')).toBeTruthy();
    expect(screen.queryByText('Nu')).toBeNull();
  });
});
