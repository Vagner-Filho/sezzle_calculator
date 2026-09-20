import type { CalculationRequest, CalculationResponse, ErrorResponse, Operation } from '../types';

const API_BASE_URL = (import.meta.env.VITE_API_URL ?? 'http://localhost:8080').replace(/\/$/, '');

/**
 * Performs a calculation by calling the backend REST API.
 * @param operation - The arithmetic operation to perform.
 * @param request - The calculation request payload.
 * @returns A promise that resolves to the calculation result.
 * @throws Error if the request fails or the server returns an error.
 */
export async function calculate(
  operation: Operation,
  request: CalculationRequest
): Promise<number> {
  const response = await fetch(`${API_BASE_URL}/api/${operation}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(request),
  });

  if (!response.ok) {
    const errorData: ErrorResponse = await response.json().catch(() => ({ error: { code: 'unknown', message: 'Unknown error' } }));
    throw new Error(errorData.error?.message || `HTTP error ${response.status}`);
  }

  const data: CalculationResponse = await response.json();
  return data.result;
}
