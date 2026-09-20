export interface CalculationRequest {
  a: number;
  b?: number;
}

export interface CalculationResponse {
  result: number;
}

export interface ErrorDetail {
  code: string;
  message: string;
}

export interface ErrorResponse {
  error: ErrorDetail;
}

export type Operation = 'add' | 'subtract' | 'multiply' | 'divide' | 'pow' | 'sqrt' | 'percentage';
