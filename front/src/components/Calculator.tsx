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
  { label: '%', value: 'percentage', needsTwoInputs: false },
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

  const handleCalculate = useCallback(async () => {
    setError(null);
    setInputError(null);
    setResult(null);

    const aError = validateNumber(a);
    if (aError) {
      setInputError(aError);
      return;
    }

    if (currentOp.needsTwoInputs) {
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
        b: currentOp.needsTwoInputs ? Number(b) : undefined,
      });
      setResult(String(res));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Calculation failed');
    } finally {
      setLoading(false);
    }
  }, [a, b, selectedOp, currentOp]);

  const clear = useCallback(() => {
    setA('');
    setB('');
    setResult(null);
    setError(null);
    setInputError(null);
    setSelectedOp('add');
  }, []);

  return (
    <div className="calculator">
      <div className="calculator-display">
        {result !== null ? (
          <span className="result" data-testid="result">{result}</span>
        ) : (
          <span className="placeholder">Result</span>
        )}
      </div>

      <div className="operations">
        {OPERATIONS.map((op) => (
          <button
            key={op.value}
            className={`op-btn ${selectedOp === op.value ? 'active' : ''}`}
            onClick={() => {
              setSelectedOp(op.value);
              setResult(null);
              setError(null);
              setInputError(null);
            }}
            aria-pressed={selectedOp === op.value}
            data-testid={`op-${op.value}`}
          >
            {op.label}
          </button>
        ))}
      </div>

      <div className="inputs">
        <label>
          <span>Value</span>
          <input
            type="text"
            inputMode="decimal"
            value={a}
            onChange={(e) => setA(e.target.value)}
            placeholder="Enter number"
            aria-invalid={!!inputError}
            data-testid="input-a"
          />
        </label>

        {currentOp.needsTwoInputs && (
          <label>
            <span>{selectedOp === 'divide' ? 'Divisor' : 'Value 2'}</span>
            <input
              type="text"
              inputMode="decimal"
              value={b}
              onChange={(e) => setB(e.target.value)}
              placeholder="Enter number"
              aria-invalid={!!inputError}
              data-testid="input-b"
            />
          </label>
        )}
      </div>

      {inputError && (
        <div className="input-error" role="alert" data-testid="input-error">
          {inputError}
        </div>
      )}

      <div className="actions">
        <button onClick={handleCalculate} disabled={loading} className="calc-btn" data-testid="calculate-btn">
          {loading ? 'Calculating…' : 'Calculate'}
        </button>
        <button onClick={clear} className="clear-btn" data-testid="clear-btn">
          Clear
        </button>
      </div>

      {error && <Toast message={error} onClose={() => setError(null)} />}
    </div>
  );
}
