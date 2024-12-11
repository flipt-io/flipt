/**
 * @jest-environment jsdom
 * @jest-environment-options {"url": "https://test/"}
 */

import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MetadataForm } from '~/components/flags/forms/MetadataForm';
import type { IFlagMetadata } from '~/types/Flag';

describe('MetadataForm', () => {
  const mockOnChange = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should render empty form when no metadata is provided', async () => {
    await act(async () => {
      render(<MetadataForm onChange={mockOnChange} />);
      fireEvent.click(screen.getByText('Add Metadata'));
    });

    const keyInput = screen.getByTestId('metadata-key-0');
    const valueInput = screen.getByTestId('metadata-value-0');

    await act(async () => {
      fireEvent.change(keyInput, { target: { value: 'new-key' } });
      fireEvent.change(valueInput, { target: { value: 'new-value' } });
    });

    expect(mockOnChange).toHaveBeenCalledWith([
      { key: 'new-key', value: 'new-value', type: 'string' }
    ]);
  });

  it('should validate required fields', async () => {
    await act(async () => {
      render(<MetadataForm onChange={mockOnChange} />);
      fireEvent.click(screen.getByText('Add Metadata'));
    });

    const keyInput = screen.getByTestId('metadata-key-0');
    const valueInput = screen.getByTestId('metadata-value-0');

    await act(async () => {
      fireEvent.blur(keyInput);
      fireEvent.blur(valueInput);
    });

    expect(screen.getByText('Key is required')).toBeInTheDocument();
    expect(screen.getByText('Value is required')).toBeInTheDocument();
  });

  it('should handle boolean type selection', async () => {
    const metadata: IFlagMetadata[] = [
      {
        key: 'test-key',
        value: '',
        type: 'string'
      }
    ];

    await act(async () => {
      render(<MetadataForm metadata={metadata} onChange={mockOnChange} />);
    });

    const typeSelect = screen.getByTestId('metadata-type-0');

    // Open dropdown and wait for it to be visible
    await act(async () => {
      fireEvent.click(typeSelect);
      await waitFor(() => {
        expect(screen.getByText('Boolean')).toBeInTheDocument();
      });
      fireEvent.click(screen.getByText('Boolean'));
    });

    expect(mockOnChange).toHaveBeenCalledWith([
      { key: 'test-key', value: false, type: 'boolean' }
    ]);
  });

  it('should handle multiple metadata entries', async () => {
    await act(async () => {
      render(<MetadataForm onChange={mockOnChange} />);
      fireEvent.click(screen.getByText('Add Metadata'));
    });

    const firstKeyInput = screen.getByTestId('metadata-key-0');
    const firstValueInput = screen.getByTestId('metadata-value-0');

    await act(async () => {
      fireEvent.change(firstKeyInput, { target: { value: 'key1' } });
      fireEvent.change(firstValueInput, { target: { value: 'value1' } });
    });

    await act(async () => {
      fireEvent.click(screen.getByText('Add Metadata'));
    });

    const keyInputs = screen.getAllByTestId(/metadata-key-\d+/);
    const valueInputs = screen.getAllByTestId(/metadata-value-\d+/);

    await act(async () => {
      fireEvent.change(keyInputs[1], { target: { value: 'key2' } });
      fireEvent.change(valueInputs[1], { target: { value: 'value2' } });
    });

    expect(mockOnChange).toHaveBeenCalledWith([
      { key: 'key1', value: 'value1', type: 'string' },
      { key: 'key2', value: 'value2', type: 'string' }
    ]);
  });

  it('should format keys correctly', async () => {
    await act(async () => {
      render(<MetadataForm onChange={mockOnChange} />);
      fireEvent.click(screen.getByText('Add Metadata'));
    });

    const keyInput = screen.getByTestId('metadata-key-0');

    await act(async () => {
      fireEvent.change(keyInput, { target: { value: 'Test Key!' } });
    });

    await expect(keyInput).toHaveValue('test-key');
  });

  it('should handle special characters in keys', async () => {
    await act(async () => {
      render(<MetadataForm onChange={mockOnChange} />);
      fireEvent.click(screen.getByText('Add Metadata'));
    });

    const keyInput = screen.getByTestId('metadata-key-0');

    await act(async () => {
      fireEvent.change(keyInput, { target: { value: 'Invalid Key!' } });
    });

    await expect(keyInput).toHaveValue('invalid-key');
  });

  it('should handle number type selection', async () => {
    const metadata: IFlagMetadata[] = [
      {
        key: 'test-key',
        value: '',
        type: 'string'
      }
    ];

    await act(async () => {
      render(<MetadataForm metadata={metadata} onChange={mockOnChange} />);
    });

    const typeSelect = screen.getByTestId('metadata-type-0');

    // Open dropdown and wait for it to be visible
    await act(async () => {
      fireEvent.click(typeSelect);
      await waitFor(() => {
        expect(screen.getByText('Number')).toBeInTheDocument();
      });
      fireEvent.click(screen.getByText('Number'));
    });

    await expect(screen.getByTestId('metadata-value-0')).toHaveValue('');
  });

  it('should disable all inputs when disabled prop is true', async () => {
    const metadata: IFlagMetadata[] = [];
    await act(async () => {
      render(
        <MetadataForm metadata={metadata} onChange={mockOnChange} disabled />
      );
    });

    const input = screen.getByTestId('metadata-key-0');
    await expect(input).toBeDisabled();

    const select = screen.getByTestId('metadata-type-0');
    await expect(select).toBeDisabled();

    const button = screen.getByText('Add Metadata');
    await expect(button).toBeDisabled();
  });
});
