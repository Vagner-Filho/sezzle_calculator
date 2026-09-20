import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Calculator } from './Calculator';
import { calculate } from '../api/client';

vi.mock('../api/client', () => ({
  calculate: vi.fn(),
}));

describe('Calculator UI accessibility and layout', () => {
  it('keeps input b rendered at all times', () => {
    render(<Calculator />);
    expect(screen.getByTestId('input-b')).toBeInTheDocument();
  });

  it('disables input b for unary operations and keeps layout stable', async () => {
    const user = userEvent.setup();
    render(<Calculator />);

    const sqrtButton = screen.getByTestId('op-sqrt');
    const inputB = screen.getByTestId('input-b') as HTMLInputElement;

    expect(inputB).not.toBeDisabled();

    await user.click(sqrtButton);

    expect(inputB).toBeDisabled();
    expect(inputB).toHaveAttribute('aria-disabled', 'true');
    expect(inputB).toBeInTheDocument();
  });

  it('enables input b when switching back to a binary operation', async () => {
    const user = userEvent.setup();
    render(<Calculator />);

    await user.click(screen.getByTestId('op-sqrt'));
    const inputB = screen.getByTestId('input-b') as HTMLInputElement;
    expect(inputB).toBeDisabled();

    await user.click(screen.getByTestId('op-add'));
    expect(inputB).not.toBeDisabled();
    expect(inputB).toHaveAttribute('aria-disabled', 'false');
  });

  it('uses semantic output element with live region for results', () => {
    render(<Calculator />);
    const output = screen.getByTestId('result-output');
    expect(output.tagName.toLowerCase()).toBe('output');
    expect(output).toHaveAttribute('aria-live', 'polite');
  });

  it('marks operation buttons with aria-pressed reflecting selection', async () => {
    const user = userEvent.setup();
    render(<Calculator />);

    const addButton = screen.getByTestId('op-add');
    expect(addButton).toHaveAttribute('aria-pressed', 'true');

    await user.click(screen.getByTestId('op-multiply'));
    expect(addButton).toHaveAttribute('aria-pressed', 'false');
    expect(screen.getByTestId('op-multiply')).toHaveAttribute('aria-pressed', 'true');
  });

  it('associates labels with their inputs via htmlFor', () => {
    render(<Calculator />);
    expect(screen.getByText('Value')).toHaveAttribute('for', 'input-a');
    expect(screen.getByText('Value 2')).toHaveAttribute('for', 'input-b');
  });
});

describe('Calculator handleCalculate', () => {
  beforeEach(() => {
    vi.mocked(calculate).mockReset();
  });

  it('calls the API and displays the result for a binary operation', async () => {
    const user = userEvent.setup();
    vi.mocked(calculate).mockResolvedValueOnce(8);

    render(<Calculator />);

    await user.type(screen.getByTestId('input-a'), '5');
    await user.type(screen.getByTestId('input-b'), '3');
    await user.click(screen.getByTestId('calculate-btn'));

    expect(calculate).toHaveBeenCalledWith('add', { a: 5, b: 3 });
    await waitFor(() => {
      expect(screen.getByTestId('result')).toHaveTextContent('8');
    });
  });

  it('calls the API with only operand a for a unary operation', async () => {
    const user = userEvent.setup();
    vi.mocked(calculate).mockResolvedValueOnce(4);

    render(<Calculator />);

    await user.click(screen.getByTestId('op-sqrt'));
    await user.type(screen.getByTestId('input-a'), '16');
    await user.click(screen.getByTestId('calculate-btn'));

    expect(calculate).toHaveBeenCalledWith('sqrt', { a: 16 });
    await waitFor(() => {
      expect(screen.getByTestId('result')).toHaveTextContent('4');
    });
  });

  it('does not send b when input b is disabled', async () => {
    const user = userEvent.setup();
    vi.mocked(calculate).mockResolvedValueOnce(3);

    render(<Calculator />);

    await user.click(screen.getByTestId('op-sqrt'));
    await user.type(screen.getByTestId('input-a'), '9');
    // Fill b while disabled; it should not be sent.
    await user.type(screen.getByTestId('input-b'), 'should-be-ignored');
    await user.click(screen.getByTestId('calculate-btn'));

    expect(calculate).toHaveBeenCalledWith('sqrt', { a: 9 });
  });

  it('shows a validation error when operand a is empty', async () => {
    const user = userEvent.setup();
    render(<Calculator />);

    await user.click(screen.getByTestId('calculate-btn'));

    expect(screen.getByTestId('input-error')).toHaveTextContent('Value is required');
    expect(calculate).not.toHaveBeenCalled();
  });

  it('shows a validation error when operand a is not a number', async () => {
    const user = userEvent.setup();
    render(<Calculator />);

    await user.type(screen.getByTestId('input-a'), 'not-a-number');
    await user.click(screen.getByTestId('calculate-btn'));

    expect(screen.getByTestId('input-error')).toHaveTextContent('Value must be a valid number');
    expect(calculate).not.toHaveBeenCalled();
  });

  it('shows a validation error when operand b is empty for a binary operation', async () => {
    const user = userEvent.setup();
    render(<Calculator />);

    await user.type(screen.getByTestId('input-a'), '5');
    await user.click(screen.getByTestId('calculate-btn'));

    expect(screen.getByTestId('input-error')).toHaveTextContent('Value is required');
    expect(calculate).not.toHaveBeenCalled();
  });

  it('shows a validation error when dividing by zero', async () => {
    const user = userEvent.setup();
    render(<Calculator />);

    await user.click(screen.getByTestId('op-divide'));
    await user.type(screen.getByTestId('input-a'), '10');
    await user.type(screen.getByTestId('input-b'), '0');
    await user.click(screen.getByTestId('calculate-btn'));

    expect(screen.getByTestId('input-error')).toHaveTextContent('Cannot divide by zero');
    expect(calculate).not.toHaveBeenCalled();
  });

  it('shows a toast when the API returns an error', async () => {
    const user = userEvent.setup();
    vi.mocked(calculate).mockRejectedValueOnce(new Error('division by zero'));

    render(<Calculator />);

    await user.type(screen.getByTestId('input-a'), '10');
    await user.type(screen.getByTestId('input-b'), '0');
    await user.click(screen.getByTestId('calculate-btn'));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('division by zero');
    });
  });

  it('disables the calculate button while the request is in flight', async () => {
    const user = userEvent.setup();
    let resolveCalculation: (value: number) => void;
    const calculationPromise = new Promise<number>((resolve) => {
      resolveCalculation = resolve;
    });
    vi.mocked(calculate).mockReturnValueOnce(calculationPromise);

    render(<Calculator />);

    await user.type(screen.getByTestId('input-a'), '2');
    await user.type(screen.getByTestId('input-b'), '3');
    await user.click(screen.getByTestId('calculate-btn'));

    expect(screen.getByTestId('calculate-btn')).toBeDisabled();
    expect(screen.getByTestId('calculate-btn')).toHaveTextContent('Calculating…');

    resolveCalculation!(6);
    await waitFor(() => {
      expect(screen.getByTestId('calculate-btn')).not.toBeDisabled();
    });
    expect(screen.getByTestId('calculate-btn')).toHaveTextContent('Calculate');
  });

  it('clears values, result, and errors when Clear is clicked', async () => {
    const user = userEvent.setup();
    vi.mocked(calculate).mockResolvedValueOnce(7);

    render(<Calculator />);

    await user.type(screen.getByTestId('input-a'), '4');
    await user.type(screen.getByTestId('input-b'), '3');
    await user.click(screen.getByTestId('calculate-btn'));

    await waitFor(() => {
      expect(screen.getByTestId('result')).toHaveTextContent('7');
    });

    await user.click(screen.getByTestId('clear-btn'));

    expect(screen.getByTestId('input-a')).toHaveValue('');
    expect(screen.getByTestId('input-b')).toHaveValue('');
    expect(screen.queryByTestId('result')).not.toBeInTheDocument();
    expect(screen.queryByTestId('input-error')).not.toBeInTheDocument();
    expect(screen.getByTestId('op-add')).toHaveAttribute('aria-pressed', 'true');
  });
});
