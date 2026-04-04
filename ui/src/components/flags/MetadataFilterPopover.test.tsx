/**
 * @jest-environment jsdom
 */
import { render, screen, fireEvent } from '@testing-library/react';
import MetadataFilterPopover from './MetadataFilterPopover';
import { MetadataFilter } from '~/types/Flag';

// Radix Popover uses portals. We mock it so the content always renders inline.
jest.mock('~/components/Popover', () => ({
  Popover: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  PopoverTrigger: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  PopoverContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>
}));

describe('MetadataFilterPopover', () => {
  it('renders a "Filter" trigger button', () => {
    render(<MetadataFilterPopover availableKeys={[]} onAdd={jest.fn()} />);
    expect(screen.getByRole('button', { name: /^filter$/i })).toBeInTheDocument();
  });

  it('calls onAdd with the entered key and value when Add is clicked', () => {
    const onAdd = jest.fn();
    render(<MetadataFilterPopover availableKeys={['team']} onAdd={onAdd} />);

    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'team' }
    });
    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'backend' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    expect(onAdd).toHaveBeenCalledWith<[MetadataFilter]>({
      key: 'team',
      value: 'backend'
    });
  });

  it('does not call onAdd when key is empty', () => {
    const onAdd = jest.fn();
    render(<MetadataFilterPopover availableKeys={[]} onAdd={onAdd} />);

    fireEvent.change(screen.getByPlaceholderText(/value/i), {
      target: { value: 'backend' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    expect(onAdd).not.toHaveBeenCalled();
  });

  it('does not call onAdd when value is empty', () => {
    const onAdd = jest.fn();
    render(<MetadataFilterPopover availableKeys={[]} onAdd={onAdd} />);

    fireEvent.change(screen.getByPlaceholderText(/key/i), {
      target: { value: 'team' }
    });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    expect(onAdd).not.toHaveBeenCalled();
  });

  it('resets key and value inputs after a successful add', () => {
    const onAdd = jest.fn();
    render(<MetadataFilterPopover availableKeys={[]} onAdd={onAdd} />);

    const keyInput = screen.getByPlaceholderText(/key/i);
    const valueInput = screen.getByPlaceholderText(/value/i);

    fireEvent.change(keyInput, { target: { value: 'team' } });
    fireEvent.change(valueInput, { target: { value: 'backend' } });
    fireEvent.click(screen.getByRole('button', { name: /add filter/i }));

    expect((keyInput as HTMLInputElement).value).toBe('');
    expect((valueInput as HTMLInputElement).value).toBe('');
  });
});
