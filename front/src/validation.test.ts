import { describe, it, expect } from 'vitest';
import { validateNumber, validateDivisor } from './validation';

describe('validateNumber', () => {
  it('returns null for valid numbers', () => {
    expect(validateNumber('42')).toBeNull();
    expect(validateNumber('-3.14')).toBeNull();
    expect(validateNumber('0')).toBeNull();
  });

  it('returns error for empty string', () => {
    expect(validateNumber('')).toBe('Value is required');
    expect(validateNumber('   ')).toBe('Value is required');
  });

  it('returns error for non-numeric strings', () => {
    expect(validateNumber('abc')).toBe('Value must be a valid number');
    expect(validateNumber('12.34.56')).toBe('Value must be a valid number');
  });

  it('returns error for infinite values', () => {
    expect(validateNumber('Infinity')).toBe('Value must be a finite number');
    expect(validateNumber('-Infinity')).toBe('Value must be a finite number');
  });
});

describe('validateDivisor', () => {
  it('returns null for non-zero divisors', () => {
    expect(validateDivisor('5')).toBeNull();
    expect(validateDivisor('-2')).toBeNull();
  });

  it('returns error for zero divisor', () => {
    expect(validateDivisor('0')).toBe('Cannot divide by zero');
  });

  it('returns number validation error for invalid input', () => {
    expect(validateDivisor('')).toBe('Value is required');
    expect(validateDivisor('abc')).toBe('Value must be a valid number');
  });
});
