/**
 * @jest-environment jsdom
 * @jest-environment-options {"url": "https://test/"}
 */

import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MetadataForm } from './MetadataForm';
import { IFlagMetadata } from '../../../types/Flag';

describe('MetadataForm', () => {
  const mockOnChange = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders empty form when no metadata provided', () => {
    render(<MetadataForm onChange={mockOnChange} />);
    expect(screen.getByText('Add Metadata')).toBeInTheDocument();
  });

  it('updates metadata when changing values', async () => {
    render(<MetadataForm onChange={mockOnChange} />);

    await act(async () => {
      fireEvent.click(screen.getByText('Add Metadata'));
    });

    const keyInput = screen.getByTestId('metadata-key-0');
    const valueInput = screen.getByTestId('metadata-value-0');

    await act(async () => {
      fireEvent.change(keyInput, { target: { value: 'new-key' } });
      fireEvent.change(valueInput, { target: { value: 'test-value' } });
    });

    await waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith([
        { key: 'new-key', value: 'test-value', type: 'string' }
      ]);
    });
  });

  it('validates required fields', async () => {
    render(<MetadataForm onChange={mockOnChange} />);

    await act(async () => {
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

  it('handles different value types correctly', async () => {
    const metadata: IFlagMetadata[] = [
      { key: 'string-key', value: 'string-value', type: 'string' },
      { key: 'bool-key', value: true, type: 'boolean' },
      { key: 'number-key', value: 42, type: 'number' }
    ];
    render(<MetadataForm metadata={metadata} onChange={mockOnChange} />);

    expect(screen.getByDisplayValue('string-value')).toBeInTheDocument();
    expect(screen.getByDisplayValue('42')).toBeInTheDocument();

    const typeSelects = screen.getAllByRole('combobox');
    expect(typeSelects[1]).toHaveTextContent('Boolean');
  });

  it('handles type changes and value conversions', async () => {
    render(<MetadataForm onChange={mockOnChange} />);

    await act(async () => {
      fireEvent.click(screen.getByText('Add Metadata'));
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

    await waitFor(() => {
      expect(mockOnChange).toHaveBeenCalledWith([
        { key: '', value: false, type: 'boolean' }
      ]);
    });
  });

  it('handles multiple metadata entries', async () => {
    render(<MetadataForm onChange={mockOnChange} />);

    // Add first entry
    await act(async () => {
      fireEvent.click(screen.getByText('Add Metadata'));
    });

    const firstKeyInput = screen.getByTestId('metadata-key-0');
    const firstValueInput = screen.getByTestId('metadata-value-0');

    await act(async () => {
      fireEvent.change(firstKeyInput, { target: { value: 'key1' } });
      fireEvent.change(firstValueInput, { target: { value: 'value1' } });
    });

    // Add second entry
    await act(async () => {
      fireEvent.click(screen.getByText('Add Metadata'));
    });

    const keyInputs = screen.getAllByTestId(/metadata-key-\d+/);
    const valueInputs = screen.getAllByTestId(/metadata-value-\d+/);

    await act(async () => {
      fireEvent.change(keyInputs[1], { target: { value: 'key2' } });
      fireEvent.change(valueInputs[1], { target: { value: 'value2' } });
    });

    await waitFor(() => {
      expect(mockOnChange).toHaveBeenLastCalledWith([
        { key: 'key1', value: 'value1', type: 'string' },
        { key: 'key2', value: 'value2', type: 'string' }
      ]);
    });
  });

  it('converts keys to valid format', async () => {
    render(<MetadataForm onChange={mockOnChange} />);

    await act(async () => {
      fireEvent.click(screen.getByText('Add Metadata'));
    });

    const keyInput = screen.getByTestId('metadata-key-0');

    await act(async () => {
      fireEvent.change(keyInput, { target: { value: 'Test Key!' } });
    });

    await waitFor(() => {
      expect(keyInput).toHaveValue('test-key');
    });
  });

  it('validates key format on blur', async () => {
    render(<MetadataForm onChange={mockOnChange} />);

    await act(async () => {
      fireEvent.click(screen.getByText('Add Metadata'));
    });

    const keyInput = screen.getByTestId('metadata-key-0');

    await act(async () => {
      fireEvent.change(keyInput, { target: { value: 'Invalid Key!' } });
      fireEvent.blur(keyInput);
    });

    await waitFor(() => {
      expect(keyInput).toHaveValue('invalid-key');
    });
  });

  it('handles edge cases for value types', async () => {
    render(<MetadataForm onChange={mockOnChange} />);

    await act(async () => {
      fireEvent.click(screen.getByText('Add Metadata'));
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

    const valueInput = screen.getByRole('spinbutton');

    await act(async () => {
      fireEvent.change(valueInput, { target: { value: 'not-a-number' } });
      fireEvent.blur(valueInput);
    });

    await waitFor(() => {
      expect(valueInput).toHaveValue(null);
    });
  });

  it('handles disabled state for all inputs', () => {
    const metadata: IFlagMetadata[] = [
      { key: 'key1', value: 'value1', type: 'string' },
      { key: 'key2', value: true, type: 'boolean' },
      { key: 'key3', value: 42, type: 'number' }
    ];
    render(<MetadataForm metadata={metadata} onChange={mockOnChange} disabled />);

    screen.getAllByRole('textbox').forEach(input => {
      expect(input).toBeDisabled();
    });
    screen.getAllByRole('combobox').forEach(select => {
      expect(select).toBeDisabled();
    });
    screen.getAllByRole('button').forEach(button => {
      expect(button).toBeDisabled();
    });
  });
});
