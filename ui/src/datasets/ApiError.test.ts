import { assert } from "chai";
import { describe, it } from "mocha";
import { ApiError } from "./ApiError";

describe("ApiError", () => {
  it("#fromApiObject", () => {
    const got = ApiError.fromApiObject({
      Code: 42069,
      Error: "too much algorithm",
    });
    assert.isNotEmpty(got);
    assert.equal(got.code, 42069);
    assert.equal(got.error, "too much algorithm");
  });

  it("#constructor", () => {
    const got = new ApiError(42069, "too much algorithm");
    assert.isNotEmpty(got);
    assert.equal(got.code, 42069);
    assert.equal(got.error, "too much algorithm");
  });

  it("#toString()", () => {
    const got = new ApiError(42069, "too much algorithm").toString();
    assert.equal(got, "vvgo error [42069]: too much algorithm");
  });
});
