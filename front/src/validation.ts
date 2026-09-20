/**
 * Validates calculator input values.
 * @param value - The raw input string to validate.
 * @returns An error message if invalid, or null if valid.
 */
export function validateNumber(value: string): string | null {
  if (value.trim() === '') {
    return 'Value is required';
  }

  const num = Number(value);
  if (Number.isNaN(num)) {
    return 'Value must be a valid number';
  }

  if (!Number.isFinite(num)) {
    return 'Value must be a finite number';
  }

  return null;
}

/**
 * Validates that a divisor is not zero.
 * @param value - The divisor value as a string.
 * @returns An error message if invalid, or null if valid.
 */
export function validateDivisor(value: string): string | null {
  const numError = validateNumber(value);
  if (numError) {
    return numError;
  }

  if (Number(value) === 0) {
    return 'Cannot divide by zero';
  }

  return null;
}
