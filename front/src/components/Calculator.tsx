import { useState, useCallback } from 'react';
import type { Operation } from '../types';
import { calculate } from '../api/client';
import { validateNumber, validateDivisor } from '../validation';
import { Toast } from './Toast';

const OPERATIONS: { label: string; value: Operation; needsTwoInputs: boolean }[] = [
  { label: '+', value: 'add', needsTwoInputs: true },
  { label: '−', value: 'subtract', needsTwoInputs: true },
  { label: '×', value: 'multiply', needsTwoInputs: true },
  { label: '÷', value: 'divide', needsTwoInputs: true },
  { label: 'xʸ', value: 'pow', needsTwoInputs: true },
  { label: '√', value: 'sqrt', needsTwoInputs: false },
  { label: '%', value: 'percentage', needsTwoInputs: true },
];

/**
 * Main calculator component providing a responsive UI
 * for arithmetic operations backed by the REST API.
 */
export function Calculator() {
  const [a, setA] = useState('');
  const [b, setB] = useState('');
  const [selectedOp, setSelectedOp] = useState<Operation>('add');
  const [result, setResult] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [inputError, setInputError] = useState<string | null>(null);

  const currentOp = OPERATIONS.find((op) => op.value === selectedOp)!;
  const bDisabled = !currentOp.needsTwoInputs;

  const handleCalculate = useCallback(async () => {
    setError(null);
    setInputError(null);
    setResult(null);

    const aError = validateNumber(a);
    if (aError) {
      setInputError(aError);
      return;
    }

    if (!bDisabled) {
      const bError = selectedOp === 'divide' ? validateDivisor(b) : validateNumber(b);
      if (bError) {
        setInputError(bError);
        return;
      }
    }

    setLoading(true);
    try {
      const res = await calculate(selectedOp, {
        a: Number(a),
        b: bDisabled ? undefined : Number(b),
      });
      setResult(String(res));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Calculation failed');
    } finally {
      setLoading(false);
    }
  }, [a, b, selectedOp, bDisabled]);

  const clear = useCallback(() => {
    setA('');
    setB('');
    setResult(null);
    setError(null);
    setInputError(null);
    setSelectedOp('add');
  }, []);

  return (
    <section className="calculator" aria-label="Calculator">
      <output
        className="calculator-display"
        aria-live="polite"
        aria-label="Calculation result"
        data-testid="result-output"
      >
        {result !== null ? (
          <span className="result" data-testid="result">{result}</span>
        ) : (
          <span className="placeholder">Result</span>
        )}
      </output>

      <div className="operations" role="group" aria-label="Operations">
        {OPERATIONS.map((op) => (
          <button
            key={op.value}
            type="button"
            className={`op-btn ${selectedOp === op.value ? 'active' : ''}`}
            onClick={() => {
              setSelectedOp(op.value);
              setResult(null);
              setError(null);
              setInputError(null);
            }}
            aria-pressed={selectedOp === op.value}
            aria-label={op.label}
            data-testid={`op-${op.value}`}
          >
            {op.label}
          </button>
        ))}
      </div>

      <div className="inputs">
        <div>
          <label htmlFor="input-a">Value</label>
          <input
            id="input-a"
            type="text"
            inputMode="decimal"
            value={a}
            onChange={(e) => setA(e.target.value)}
            placeholder="Enter number"
            aria-invalid={!!inputError}
            aria-describedby={inputError ? 'input-error' : undefined}
            data-testid="input-a"
          />
        </div>

        <div>
          <label htmlFor="input-b" aria-disabled={bDisabled}>
            {selectedOp === 'divide' ? 'Divisor' : 'Value 2'}
          </label>
          <input
            id="input-b"
            type="text"
            inputMode="decimal"
            value={b}
            disabled={bDisabled}
            aria-disabled={bDisabled}
            onChange={(e) => setB(e.target.value)}
            placeholder={bDisabled ? 'Not used' : 'Enter number'}
            aria-invalid={!!inputError}
            aria-describedby={inputError ? 'input-error' : undefined}
            data-testid="input-b"
          />
        </div>
      </div>

      {inputError && (
        <div id="input-error" className="input-error" role="alert" data-testid="input-error">
          {inputError}
        </div>
      )}

      <div className="actions">
        <button type="button" onClick={handleCalculate} disabled={loading} className="calc-btn" data-testid="calculate-btn">
          {loading ? 'Calculating…' : 'Calculate'}
        </button>
        <button type="button" onClick={clear} className="clear-btn" data-testid="clear-btn">
          Clear
        </button>
      </div>

      {error && <Toast message={error} onClose={() => setError(null)} />}
    </section>
  );
}
