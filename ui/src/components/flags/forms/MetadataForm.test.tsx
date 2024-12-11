import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MetadataForm } from '../MetadataForm';
import { IFlagMetadata } from '../../../../types/Flag';

describe('MetadataForm', () => {
  const mockOnChange = jest.fn();

  beforeEach(() => {
    mockOnChange.mockClear();
  });

  it('renders empty form with add button', () => {
    render(<MetadataForm onChange={mockOnChange} />);
    expect(screen.getByText('Add Metadata')).toBeInTheDocument();
  });

  it('adds new metadata entry when clicking add button', () => {
    render(<MetadataForm onChange={mockOnChange} />);
    fireEvent.click(screen.getByText('Add Metadata'));
    expect(mockOnChange).toHaveBeenCalledWith([
      { key: '', value: '', type: 'string' }
    ]);
  });

  it('renders existing metadata entries', () => {
    const metadata: IFlagMetadata[] = [
      { key: 'test-key', value: 'test-value', type: 'string' }
    ];
    render(<MetadataForm metadata={metadata} onChange={mockOnChange} />);
    expect(screen.getByDisplayValue('test-key')).toBeInTheDocument();
    expect(screen.getByDisplayValue('test-value')).toBeInTheDocument();
  });

  it('updates metadata when changing values', () => {
    const metadata: IFlagMetadata[] = [
      { key: 'test-key', value: 'test-value', type: 'string' }
    ];
    render(<MetadataForm metadata={metadata} onChange={mockOnChange} />);

    const keyInput = screen.getByDisplayValue('test-key');
    fireEvent.change(keyInput, { target: { value: 'new-key' } });

    expect(mockOnChange).toHaveBeenCalledWith([
      { key: 'new-key', value: 'test-value', type: 'string' }
    ]);
  });

  it('removes metadata entry when clicking remove button', () => {
    const metadata: IFlagMetadata[] = [
      { key: 'test-key', value: 'test-value', type: 'string' }
    ];
    render(<MetadataForm metadata={metadata} onChange={mockOnChange} />);

    fireEvent.click(screen.getByLabelText('Remove metadata entry'));
    expect(mockOnChange).toHaveBeenCalledWith([]);
  });

  it('handles different value types correctly', () => {
    const metadata: IFlagMetadata[] = [
      { key: 'string-key', value: 'string-value', type: 'string' },
      { key: 'bool-key', value: true, type: 'boolean' },
      { key: 'number-key', value: 42, type: 'number' }
    ];
    render(<MetadataForm metadata={metadata} onChange={mockOnChange} />);

    expect(screen.getByDisplayValue('string-value')).toBeInTheDocument();
    expect(screen.getByDisplayValue('true')).toBeInTheDocument();
    expect(screen.getByDisplayValue('42')).toBeInTheDocument();
  });

  it('disables inputs when disabled prop is true', () => {
    const metadata: IFlagMetadata[] = [
      { key: 'test-key', value: 'test-value', type: 'string' }
    ];
    render(<MetadataForm metadata={metadata} onChange={mockOnChange} disabled />);

    expect(screen.getByDisplayValue('test-key')).toBeDisabled();
    expect(screen.getByDisplayValue('test-value')).toBeDisabled();
    expect(screen.getByLabelText('Remove metadata entry')).toBeDisabled();
  });
});
