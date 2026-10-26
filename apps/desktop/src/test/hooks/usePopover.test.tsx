import { fireEvent, render } from '@testing-library/react';
import { useState } from 'react';
import { describe, expect, it } from 'vitest';

import { usePopoverDismiss } from '../../hooks/usePopover';

/* The dismissal contract the dashboard relies on: NotificationBell and
   QuickActions compose it identically — the ref wraps the trigger AND
   the popover, so Escape and an outside pointer close it, and clicking
   the trigger toggles instead of closing. Both call sites depend on the
   hook honoring that, and nothing tested it. */

function Probe() {
  const [open, setOpen] = useState(false);
  const ref = usePopoverDismiss(() => setOpen(false));
  return (
    <div ref={ref}>
      <button onClick={() => setOpen((v) => !v)}>trigger</button>
      {open && <div data-testid="popover">content</div>}
    </div>
  );
}

function openPopover() {
  const utils = render(<Probe />);
  const trigger = utils.getByRole('button', { name: 'trigger' });
  fireEvent.pointerDown(trigger);
  fireEvent.click(trigger);
  return { ...utils, trigger };
}

describe('usePopoverDismiss', () => {
  it('toggles on the trigger instead of closing it', () => {
    const { trigger, queryByTestId } = openPopover();
    expect(queryByTestId('popover')).not.toBeNull();
    fireEvent.pointerDown(trigger);
    fireEvent.click(trigger);
    expect(queryByTestId('popover')).toBeNull();
  });

  it('closes on Escape', () => {
    const { queryByTestId } = openPopover();
    expect(queryByTestId('popover')).not.toBeNull();
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(queryByTestId('popover')).toBeNull();
  });

  it('closes on a pointer down outside', () => {
    const { queryByTestId } = openPopover();
    expect(queryByTestId('popover')).not.toBeNull();
    fireEvent.pointerDown(document.body);
    expect(queryByTestId('popover')).toBeNull();
  });
});
