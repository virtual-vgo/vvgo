import { assert } from "chai";
import { describe, it } from "mocha";
import { Highlight } from "./Highlight";

describe("Highlight", () => {
  it("#fromApiObject", () => {
    const got = Highlight.fromDatasetRow(
      new Map([
        ["Alt", "cannelloni"],
        ["Source", "fettuccine"],
      ])
    );
    assert.equal(got.alt, "cannelloni");
    assert.equal(got.source, "fettuccine");
  });
});
