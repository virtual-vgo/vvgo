import { get } from "lodash/fp";

export class ApiError {
  code: number;
  error: string;

  constructor(code?: number, error?: string) {
    this.code = code ?? 0;
    this.error = error ?? `unknown api error 😔`;
  }

  toString(): string {
    return `vvgo error [${this.code}]: ${this.error}`;
  }

  static UnknownError = new ApiError(0, `unknown api error 😔`);

  static fromApiJson(obj: object): ApiError {
    return new ApiError(get("Code", obj), get("Error", obj));
  }
}
