import { assert } from "chai";
import { describe, it } from "mocha";
import { Dataset } from "./Dataset";

describe("Dataset", () => {
  it("#fromApiArray", () => {
    const got = Dataset.fromApiArray([
      { cheese: "provolone", delicious: "yes" },
      { cheese: "swiss", delicious: "yes" },
    ]);
    const want = [
      new Map([
        ["cheese", "provolone"],
        ["delicious", "yes"],
      ]),
      new Map([
        ["cheese", "swiss"],
        ["delicious", "yes"],
      ]),
    ];
    assert.deepEqual(got, want);
  });
});
