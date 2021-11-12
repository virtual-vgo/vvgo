import { ApiResponse, ApiStatus } from "../../datasets";
import { ApiError } from "../../datasets/ApiError";

export class Client {
  target: string = "";
  readonly #token: string;

  constructor(token: string, target?: string) {
    if (target) this.target = target;
    this.#token = token;
  }

  fetch(url: string, init?: RequestInit): Promise<ApiResponse> {
    url = this.target + url;
    init ??= { method: "GET" };
    init.headers = {
      Authorization: "Bearer " + this.#token,
      ...init.headers,
    };
    console.log("Api Request:", init.method, url);

    return fetch(url, init)
      .then((resp) => resp.json())
      .then((respObj) => {
        const apiResponse = ApiResponse.fromApiJSON(respObj as object);
        console.log("Api Response:", apiResponse);
        if (apiResponse.status == ApiStatus.Error) {
          throw apiResponse.error ?? ApiError.UnknownError;
        }
        return apiResponse;
      });
  }
}

export default { Client };
