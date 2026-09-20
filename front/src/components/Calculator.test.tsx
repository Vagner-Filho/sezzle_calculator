import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Calculator } from './Calculator';
import * as client from '../api/client';

vi.mock('../api/client');

const mockedCalculate = vi.mocked(client.calculate);

describe('Calculator', () => {
  beforeEach(() => {
    mockedCalculate.mockClear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders calculator with inputs and operations', () => {
    render(<Calculator />);
    expect(screen.getByTestId('input-a')).toBeInTheDocument();
    expect(screen.getByTestId('input-b')).toBeInTheDocument();
    expect(screen.getByTestId('calculate-btn')).toBeInTheDocument();
    expect(screen.getByTestId('clear-btn')).toBeInTheDocument();
  });

  it('shows input error for empty value', async () => {
    render(<Calculator />);
    fireEvent.click(screen.getByTestId('calculate-btn'));
    await waitFor(() => {
      expect(screen.getByTestId('input-error')).toHaveTextContent('Value is required');
    });
  });

  it('shows input error for invalid number', async () => {
    render(<Calculator />);
    fireEvent.change(screen.getByTestId('input-a'), { target: { value: 'abc' } });
    fireEvent.click(screen.getByTestId('calculate-btn'));
    await waitFor(() => {
      expect(screen.getByTestId('input-error')).toHaveTextContent('Value must be a valid number');
    });
  });

  it('shows input error for division by zero', async () => {
    render(<Calculator />);
    fireEvent.click(screen.getByTestId('op-divide'));
    fireEvent.change(screen.getByTestId('input-a'), { target: { value: '10' } });
    fireEvent.change(screen.getByTestId('input-b'), { target: { value: '0' } });
    fireEvent.click(screen.getByTestId('calculate-btn'));
    await waitFor(() => {
      expect(screen.getByTestId('input-error')).toHaveTextContent('Cannot divide by zero');
    });
  });

  it('calls API and displays result on successful calculation', async () => {
    mockedCalculate.mockResolvedValue(42);
    render(<Calculator />);
    fireEvent.change(screen.getByTestId('input-a'), { target: { value: '20' } });
    fireEvent.change(screen.getByTestId('input-b'), { target: { value: '22' } });
    fireEvent.click(screen.getByTestId('calculate-btn'));
    await waitFor(() => {
      expect(screen.getByTestId('result')).toHaveTextContent('42');
    });
    expect(mockedCalculate).toHaveBeenCalledWith('add', { a: 20, b: 22 });
  });

  it('shows toast on backend error', async () => {
    mockedCalculate.mockRejectedValue(new Error('Server is down'));
    render(<Calculator />);
    fireEvent.change(screen.getByTestId('input-a'), { target: { value: '5' } });
    fireEvent.change(screen.getByTestId('input-b'), { target: { value: '3' } });
    fireEvent.click(screen.getByTestId('calculate-btn'));
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Server is down');
    });
  });

  it('clears state when clear button is clicked', async () => {
    mockedCalculate.mockResolvedValue(8);
    render(<Calculator />);
    fireEvent.change(screen.getByTestId('input-a'), { target: { value: '5' } });
    fireEvent.change(screen.getByTestId('input-b'), { target: { value: '3' } });
    fireEvent.click(screen.getByTestId('calculate-btn'));
    await waitFor(() => {
      expect(screen.getByTestId('result')).toHaveTextContent('8');
    });
    fireEvent.click(screen.getByTestId('clear-btn'));
    expect(screen.getByTestId('input-a')).toHaveValue('');
    expect(screen.queryByTestId('result')).not.toBeInTheDocument();
  });
});
